// Runner runs the container, the dump-loop and the IPC listener.
package runner

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/Xarepo/msc-container-migration/internal/image"
	"github.com/Xarepo/msc-container-migration/internal/ipc"
	"github.com/Xarepo/msc-container-migration/internal/rpc"
	"github.com/Xarepo/msc-container-migration/internal/runc"
	"github.com/Xarepo/msc-container-migration/internal/runner/runner_context"
	. "github.com/Xarepo/msc-container-migration/internal/runner/runner_context"
	"github.com/Xarepo/msc-container-migration/internal/sftp"
)

type Runner struct {
	RunnerContext
	lock sync.Mutex
}

// Create a new runner.
//
// @param containerId: The id of the container to create.
// @param bundlePath: The path to the OCI-bundle used to create the container.
func New(containerId, bundlePath, imagePath string) *Runner {
	return &Runner{
		RunnerContext: runner_context.New(containerId, bundlePath, imagePath),
	}
}

func (runner *Runner) Start() {
	go runner.Loop()
	go runner.IPCListener.Listen(func(buf []byte) {
		ipc := ipc.ParseIPC(string(buf))
		if ipc != nil {
			ipc.Execute(&runner.RunnerContext)
		} else {
			log.Error().Msg("Failed to parse IPC")
		}
	})
	go runner.RPCListener.Listen(runner.RPCPort, func(buf []byte, remoteAddr string) {
		log.Debug().Str("RPC", string(buf)).Msg("Received RPC")
		rpc := rpc.ParseRPC(string(buf))
		if rpc != nil {
			rpc.Execute(&runner.RunnerContext, remoteAddr)
		} else {
			log.Error().Msg("Failed to parse RPC")
		}
	})
	runner.SetStatus(runner_context.StandBy)
	log.Debug().Msg("Runner started, standing by")
	os.MkdirAll("/dumps", os.ModeDir)
}

func (runner *Runner) Run() {
	imagePath := ""
	if runner.LatestImage != nil {
		imagePath = runner.LatestImage.Path()
	}

	go runner.runContainer(imagePath)
	runner.SetStatus(runner_context.Running)
	log.Debug().Msg("Runner running")
}

func (runner *Runner) restoreContainer() {
	status, err := runc.Restore(runner.ContainerId, runner.LatestImage.Path(), runner.BundlePath)
	if err != nil {
		log.Error().Str("Error", err.Error()).Int("Status", status).Msg("Error running container")
	} else {
		log.Info().Int("Status", status).Msg("Container exited")
	}
	runner.ContainerStatus <- status
}

// Wait for the runner to finish running.
func (runner *Runner) WaitFor() int {
	status := <-runner.ContainerStatus
	for runner.RunnerStatus() == runner_context.Running ||
		runner.RunnerStatus() == runner_context.Migrating {
	}
	return status
}

// Run the container, either from scratch or by restoring it.
//
// @param imagePath: The path to the image from which to restore the container.
//	If this is empty, then the container is started from scratch.
func (runner *Runner) runContainer(imagePath string) {
	var status int
	var err error
	if imagePath == "" {
		status, err = runc.Run(runner.ContainerId, runner.BundlePath)
	} else {
		status, err = runc.Restore(
			runner.ContainerId,
			runner.LatestImage.Path(),
			runner.BundlePath,
		)
	}
	if err != nil {
		if status == 137 {
			log.Warn().Msg("Container exited with status 137 (SIGKILL), assuming it was checkpointed...")
		} else {
			log.Error().Str("Error", err.Error()).Int("Status", status).Msg("Error running container")
		}
	} else {
		log.Info().Str("Status", string(status)).Msg("Container exited")
	}
	runner.ContainerStatus <- status
}

func (runner *Runner) Loop() {
	dumpFreq := 3
	for {
		switch runner.RunnerStatus() {
		case runner_context.Running:
			select {
			case <-time.After(2 * time.Second):
				// Lock dumping stage as to avoid conflicted states
				runner.WithLock(func() {
					nextImg := image.FirstImage()
					parentPath := ""
					if runner.LatestImage != nil {
						nextImg = runner.LatestImage.NextImage(dumpFreq)
						// Parentpath is relative to the parent directory of the image path
						// so only the directory name (not the full path) should be used
						parentPath = runner.LatestImage.Base()
					}

					if nextImg.PreDump() {
						runc.PreDump(
							runner.ContainerId,
							nextImg.Path(),
							parentPath,
						)
					} else {
						runc.Dump(
							runner.ContainerId,
							nextImg.Path(),
							parentPath,
							true,
						)
					}

					for _, target := range runner.Targets {
						sftp.CopyToRemote(nextImg.Path(), &target)
					}

					runner.LatestImage = nextImg
				})
			case <-runner.TimerInterrupt:
				log.Trace().Msg("Runner timer interrupted, cancelling dumping")
			}
		case runner_context.Migrating:
			log.Debug().
				Str("ContainerId", runner.ContainerId).
				Msg("Migrating container")

			// Pre-dump
			nextImg := runner.LatestImage.NextPreDumpImage()
			runc.PreDump(
				runner.ContainerId,
				nextImg.Path(),
				runner.LatestImage.Base(),
			)
			sftp.CopyToRemote(nextImg.Path(), &runner.Targets[0])
			runner.LatestImage = nextImg

			// Dump
			nextImg = nextImg.NextDumpImage()
			runc.Dump(
				runner.ContainerId,
				nextImg.Path(),
				runner.LatestImage.Base(),
				false)
			sftp.CopyToRemote(nextImg.Path(), &runner.Targets[0])
			runner.LatestImage = nextImg

			// RPC
			remote := runner.Targets[0].RPCAddr()
			conn, err := net.Dial("udp", remote)
			defer conn.Close()
			if err != nil {
				log.Error().Str("Error", err.Error()).Msg("Failed to dial UDP")
				return
			}
			rpc := rpc.NewMigrate(runner.ContainerId, runner.LatestImage.Base(), runner.BundlePath)
			log.Trace().Str("RPC", rpc.String()).Str("Target", conn.RemoteAddr().String()).Msg("Sending RPC")
			fmt.Fprintf(conn, rpc.String())

			log.Info().
				Str("ContainerId", runner.ContainerId).
				Str("Remote", remote).
				Str("Image", nextImg.Path()).
				Msg("Container migrated")

			runner.SetStatus(runner_context.Stopped)
		case runner_context.Restoring:
			log.Trace().Msg("Restoring container")
			go runc.Restore(
				runner.ContainerId,
				runner.LatestImage.Path(),
				runner.BundlePath)
			log.Info().
				Str("ContainerId", runner.ContainerId).
				Str("Image", runner.LatestImage.Path()).
				Str("Bundle", runner.BundlePath).
				Msg("Container restored")
			runner.SetStatus(runner_context.Running)
		case runner_context.Joining:
			log.Trace().Str("Remote", runner.Source).Msg("Joining cluster")

			remote := runner.Source
			conn, err := net.Dial("udp4", remote)
			if err != nil {
				log.Error().Str("Error", err.Error()).Msg("Failed to dial UDP")
				return
			}
			defer conn.Close()

			fileTransferPort, err := strconv.Atoi(os.Getenv("FILE_TRANSFER_PORT"))
			if err != nil {
				log.Warn().
					Msg("Failed to parse file transfer port, defaulting to 22")
			}

			rpc := rpc.NewJoin(
				runner.RPCPort,
				fileTransferPort,
				os.Getenv("DUMP_PATH"),
			)
			log.Trace().
				Str("RPC", rpc.String()).
				Str("Target", conn.RemoteAddr().String()).
				Msg("Sending RPC")
			fmt.Fprintf(conn, rpc.String())

			runner.SetStatus(runner_context.StandBy)
		}
	}
}
