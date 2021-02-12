// Runner runs the container, the dump-loop and the IPC listener.
package runner

import (
	"os"
	"strconv"
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
	go runner.RPCListener.Listen(runner.RPCPort(), func(buf []byte, remoteAddr string) {
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
	status, err := runc.Restore(
		runner.ContainerId,
		runner.LatestImage.Path(),
		runner.BundlePath,
	)
	if err != nil {
		log.Error().
			Str("Error", err.Error()).
			Int("Status", status).
			Msg("Error running container")
	} else {
		log.Info().Int("Status", status).Msg("Container exited")
	}
	runner.ContainerStatus <- status
}

// Wait for the runner to finish running.
func (runner *Runner) WaitFor() int {
	status := <-runner.ContainerStatus
	for runner.Status() == runner_context.Running ||
		runner.Status() == runner_context.Migrating {
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
			log.Error().
				Str("Error", err.Error()).
				Int("Status", status).
				Msg("Error running container")
		}
	} else {
		log.Info().Str("Status", string(status)).Msg("Container exited")
	}
	runner.ContainerStatus <- status
}

func (runner *Runner) Loop() {
	for {
		switch runner.Status() {
		case runner_context.Running:
			runner.loopRunning()
		case runner_context.Migrating:
			runner.loopMigrating()
		case runner_context.Restoring:
			runner.loopRestoring()
		case runner_context.Joining:
			runner.loopJoining()
		case runner_context.StandBy:
			runner.loopStandby()
		case runner_context.Recovery:
			runner.loopRecovery()
		}
	}
}

func (runner *Runner) loopRunning() {
	dumpFreq := 3
	dumpTick := time.NewTicker(2 * time.Second)
	pingTick := time.NewTicker(1 * time.Second)
	done := make(chan bool)
	go func() {
		for runner.Status() == runner_context.Running {
		}
		done <- true
	}()
	for {
		select {
		case <-dumpTick.C:
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
		case <-pingTick.C:
			for _, target := range runner.Targets {
				log.Trace().Msg("Pinging remote")

				ping := rpc.Ping{}
				rpc.Send(ping, target.RPCAddr())
			}
		case <-done:
			log.Trace().Msg("DONE RUNNING")
			return
		}
	}
}

func (runner *Runner) loopMigrating() {
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

	migrate := rpc.NewMigrate(
		runner.ContainerId,
		runner.LatestImage.Base(),
		runner.BundlePath,
	)
	err := rpc.Send(migrate, runner.Targets[0].RPCAddr())
	if err != nil {
		log.Error().
			Str("RPC", migrate.String()).
			Str("Error", err.Error()).
			Msg("Failed to send RPC")
	} else {
		log.Info().
			Str("ContainerId", runner.ContainerId).
			Str("TargetHost", runner.Targets[0].Host()).
			Str("Image", nextImg.Path()).
			Msg("Container migrated")
	}

	runner.SetStatus(runner_context.Stopped)
}

func (runner *Runner) loopRestoring() {
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
}

func (runner *Runner) loopJoining() {
	log.Trace().Str("Remote", runner.Source).Msg("Joining cluster")

	fileTransferPort, err := strconv.Atoi(os.Getenv("FILE_TRANSFER_PORT"))
	if err != nil {
		log.Warn().
			Msg("Failed to parse file transfer port, defaulting to 22")
		fileTransferPort = 22
	}

	join := rpc.NewJoin(
		runner.RPCPort(),
		fileTransferPort,
		os.Getenv("DUMP_PATH"),
	)

	err = rpc.Send(join, runner.Source)
	if err != nil {
		log.Error().
			Str("RPC", join.String()).
			Str("Error", err.Error()).
			Msg("Failed to send RPC")
		runner.SetStatus(runner_context.Stopped)
		return
	}

	<-runner.AckWait
	log.Info().Msg("Successfully joined cluster")
	runner.SetStatus(runner_context.StandBy)
}

func (runner *Runner) loopStandby() {
	if runner.Source != "" {
		done := make(chan bool)
		go func() {
			for runner.Status() == runner_context.StandBy {
			}
			done <- true
		}()
		for {
			duration := 5 * time.Second
			select {
			case <-time.After(duration):
				log.Warn().
					Msgf(
						"No ping received in %s. Assuming source is down. Starting recovery",
						duration.String(),
					)
				runner.SetStatus(runner_context.Recovery)
			case <-runner.PingInterrupt:
			case <-done:
				return
			}
		}
	}
}

func (runner *Runner) loopRecovery() {
	log.Trace().Msg("Recovering")

	runner.LatestImage = image.Recover()
	runner.Source = ""

	log.Info().Str("Dump", runner.LatestImage.Path()).Msg("Recovering from dump")
	runner.Run()
}
