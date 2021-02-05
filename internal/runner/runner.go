// Runner runs the container, the dump-loop and the IPC listener.
package runner

import (
	"fmt"
	"net"
	"os"
	"sync"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/Xarepo/msc-container-migration/internal/image"
	"github.com/Xarepo/msc-container-migration/internal/ipc"
	"github.com/Xarepo/msc-container-migration/internal/rpc"
	"github.com/Xarepo/msc-container-migration/internal/runc"
	"github.com/Xarepo/msc-container-migration/internal/runner/runner_context"
	. "github.com/Xarepo/msc-container-migration/internal/runner/runner_context"
	"github.com/Xarepo/msc-container-migration/internal/scp"
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
	go runner.RPCListener.Listen(func(buf []byte) {
		log.Debug().Str("RPC", string(buf)).Msg("Received RPC")
		rpc := rpc.ParseRPC(string(buf))
		if rpc != nil {
			rpc.Execute(&runner.RunnerContext)
		} else {
			log.Error().Msg("Failed to parse RPC")
		}
	})
	runner.SetStatus(runner_context.StandBy)
	log.Debug().Msg("Runner started, standing by")
}

func (runner *Runner) Run() {
	imagePath := ""
	if runner.LatestImage != nil {
		imagePath = runner.LatestImage.String()
	}

	go runner.runContainer(imagePath)
	runner.SetStatus(runner_context.Running)
	log.Debug().Msg("Runner running")
}

func (runner *Runner) restoreContainer() {
	status, err := runc.Restore(runner.ContainerId, runner.LatestImage.String(), runner.BundlePath)
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
			runner.LatestImage.String(),
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
						parentPath = runner.LatestImage.String()
					}
					if nextImg.PreDump() {
						runc.PreDump(
							runner.ContainerId,
							nextImg.String(),
							parentPath,
						)
					} else {
						runc.Dump(
							runner.ContainerId,
							nextImg.String(),
							parentPath,
							true,
						)
					}
					scp.CopyToRemote(nextImg.String())
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
				nextImg.String(),
				runner.LatestImage.String(),
			)
			scp.CopyToRemote(nextImg.String())
			runner.LatestImage = nextImg

			// Dump
			nextImg = nextImg.NextDumpImage()
			runc.Dump(
				runner.ContainerId,
				nextImg.String(),
				runner.LatestImage.String(),
				false)
			scp.CopyToRemote(nextImg.String())
			runner.LatestImage = nextImg

			// RPC
			remote := os.Getenv("MIGRATION_TARGET")
			conn, err := net.Dial("udp", remote)
			defer conn.Close()
			if err != nil {
				log.Error().Str("Error", err.Error()).Msg("Failed to dial UDP")
				return
			}
			rpc := rpc.NewMigrate(runner.ContainerId, runner.LatestImage.String(), runner.BundlePath)
			log.Trace().Str("RPC", rpc.String()).Msg("Sending RPC")
			fmt.Fprintf(conn, rpc.String())

			log.Info().
				Str("ContainerId", runner.ContainerId).
				Str("Remote", remote).
				Str("Image", nextImg.String()).
				Msg("Container migrated")

			runner.SetStatus(runner_context.Stopped)
		case runner_context.Restoring:
			log.Trace().Msg("Restoring container")
			go runc.Restore(
				runner.ContainerId,
				runner.LatestImage.String(),
				runner.BundlePath)
			log.Info().
				Str("ContainerId", runner.ContainerId).
				Str("Image", runner.LatestImage.String()).
				Str("Bundle", runner.BundlePath).
				Msg("Container restored")
			runner.SetStatus(runner_context.Running)
		}
	}
}