// Runner runs the container, the dump-loop and the IPC listener.
package runner

import (
	"net"
	"net/http"
	"net/rpc"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/Xarepo/msc-container-migration/internal/dump"
	"github.com/Xarepo/msc-container-migration/internal/env"
	"github.com/Xarepo/msc-container-migration/internal/ipc"
	"github.com/Xarepo/msc-container-migration/internal/remote_target"
	"github.com/Xarepo/msc-container-migration/internal/runc"
	"github.com/Xarepo/msc-container-migration/internal/runner/runner_context"
	. "github.com/Xarepo/msc-container-migration/internal/runner/runner_context"
	"github.com/Xarepo/msc-container-migration/internal/utils"
)

type Runner struct {
	RunnerContext
	RPCHandler
}

// Create a new runner.
//
// @param containerId: The id of the container to create.
// @param bundlePath: The path to the OCI-bundle used to create the container.
func New(containerId, bundlePath string) *Runner {
	runner := Runner{
		RunnerContext: runner_context.New(containerId, bundlePath),
	}
	runner.RPCHandler = RPCHandler{runner: &runner}
	return &runner
}

// Start the runner.
// This starts the runner's loop, IPC/RPC-listener and signals handler, and
// sets the runner's status to standby.
func (runner *Runner) Start() {
	// Handle signals
	go func() {
		for {
			c := make(chan os.Signal, 1)
			signal.Notify(c, syscall.SIGTERM, syscall.SIGINT)

			s := <-c
			log.Debug().Str("Signal", s.String()).Msg("Received signal")
			runner.SetStatus(runner_context.Terminated)
		}
	}()

	go runner.Loop()
	go runner.IPCListener.Listen(func(buf []byte) {
		ipc := ipc.ParseIPC(string(buf))
		if ipc != nil {
			ipc.Execute(&runner.RunnerContext)
		} else {
			log.Error().Msg("Failed to parse IPC")
		}
	})

	// RPC listener
	rpc.RegisterName("RPC", &runner.RPCHandler)
	rpc.HandleHTTP()
	l, e := net.Listen("tcp", ":1234")
	if e != nil {
		log.Fatal().Msgf("listen error:%s", e)
	}
	go http.Serve(l, nil)

	runner.SetStatus(runner_context.StandBy)
	log.Debug().Msg("Runner started, standing by")
	os.MkdirAll("/dumps", os.ModeDir)
}

// Start the container and set the status to running
func (runner *Runner) StartContainer() {
	go runner.runContainer()
	runner.SetStatus(runner_context.Running)
	log.Debug().Msg("Runner running")
}

// Restore the container and set the status to running
func (runner *Runner) RestoreContainer() {
	go runner.restoreContainer()
	runner.SetStatus(runner_context.Running)
	log.Debug().Msg("Runner restored")
}

// Wait for the runner to finish running.
func (runner *Runner) WaitForContainer() int {
	return <-runner.ContainerStatus
}

// Wait for the runner to finish running.
func (runner *Runner) WaitFor() {
	for runner.Status() != runner_context.Failed &&
		runner.Status() != runner_context.Stopped {
	}
}

func (runner *Runner) runContainer() {
	status, err := runc.Run(runner.ContainerId, runner.BundlePath)
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
		log.Info().Int("Status", status).Msg("Container exited")
	}
	runner.ContainerStatus <- status
}

func (runner *Runner) restoreContainer() {
	status, err := runc.Restore(
		runner.ContainerId,
		runner.LatestDump.Path(),
		runner.BundlePath,
	)
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
		log.Info().Int("Status", status).Msg("Container exited")
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
		case runner_context.Failed:
			log.Fatal().Msg("The runner has failed")
		case runner_context.Terminated:
			runner.WithLock(func() {
				log.Trace().
					Str("ContainerId", runner.ContainerId).
					Msg("Terminating runner")
				err := runc.Kill(runner.ContainerId)
				if err != nil {
					log.Error().
						Str("Error", err.Error()).
						Msg("Failed to kill container, assuming failed state")
					runner.SetStatusNoLock(runner_context.Failed)
					return
				}
				runner.SetStatusNoLock(runner_context.Stopped)
			})
			for runner.Status() == runner_context.Terminated {
			}
		}
	}
}

func (runner *Runner) loopRunning() {
	dumpFreq := env.Getenv().FULLDUMP_FREQ
	dumpTick := time.NewTicker(
		time.Duration(env.Getenv().DUMP_INTERVAL) * time.Second,
	)
	pingTick := time.NewTicker(
		time.Duration(env.Getenv().PING_INTERVAL) * time.Second,
	)
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
				nextDump := dump.FirstDump()
				parentPath := ""
				if runner.LatestDump != nil {
					nextDump = runner.LatestDump.NextDump(dumpFreq)
					// Parentpath is relative to the parent directory of the image path
					// so only the directory name (not the full path) should be used
					parentPath = runner.LatestDump.Base()
				}

				if nextDump.PreDump() {
					runc.PreDump(
						runner.ContainerId,
						nextDump.Path(),
						parentPath,
					)
				} else {
					runc.Dump(
						runner.ContainerId,
						nextDump.Path(),
						parentPath,
						true,
					)
				}

				// for _, target := range runner.Targets {
				// 	sftp.CopyToRemote(nextDump.Path(), &target)
				// }

				runner.LatestDump = nextDump
				runner.Chain.Push(*nextDump)
				for _, target := range runner.Targets {
					runner.Chain.Sync(&target)
				}
			})
		case <-pingTick.C:
			for _, target := range runner.Targets {
				log.Trace().Str("Target", target.RPCAddr()).Msg("Pinging remote")

				var client *rpc.Client
				var err error
				var reply bool
				var args struct{}
				// false/true in the channel indicates a failed/successful call, but
				// not necessarily a response.
				sync := make(chan bool)

				// Call RPC in a go routine in order to implement timeout behavior, as
				// net/rpc has no support for timeouts.
				go func() {
					client, err = rpc.DialHTTP("tcp", target.RPCAddr())
					if err != nil {
						log.Fatal().Msgf("dialing:%s", err)
					}
					err = client.Call("RPC.Ping", args, &reply)
					if err != nil {
						log.Warn().
							Str("Error", err.Error()).
							Str("Target", target.RPCAddr()).
							Msg("Failed to call PING RPC")
						runner.RemoveTarget(target)
						sync <- false
					} else {
						sync <- true
					}
				}()
				select {
				case success := <-sync:
					if success && reply == true {
						log.Trace().Str("Target", target.RPCAddr()).Msg("PING RECEIVED")
					}
				case <-time.After(3 * time.Second): // TODO: Don't hardcode ping timeout
					log.Warn().
						Str("Target", target.RPCAddr()).
						Msg("No ping received from target")
					runner.RemoveTarget(target)
					if client != nil {
						client.Close()
					}
				}
			}
		case <-done:
			return
		}
	}
}

func (runner *Runner) loopMigrating() {
	runner.WithLock(func() {
		log.Debug().
			Str("ContainerId", runner.ContainerId).
			Msg("Migrating container")

		// Pre-dump
		nextDump := runner.LatestDump.NextPreDump()
		runc.PreDump(
			runner.ContainerId,
			nextDump.Path(),
			runner.LatestDump.Base(),
		)
		// sftp.CopyToRemote(nextDump.Path(), &runner.Targets[0])
		runner.LatestDump = nextDump
		runner.Chain.Push(*nextDump)
		runner.Chain.Sync(&runner.Targets[0])

		// Dump
		nextDump = nextDump.NextFullDump()
		runc.Dump(
			runner.ContainerId,
			nextDump.Path(),
			runner.LatestDump.Base(),
			false)
		// sftp.CopyToRemote(nextDump.Path(), &runner.Targets[0])
		runner.LatestDump = nextDump
		runner.Chain.Push(*nextDump)
		runner.Chain.Sync(&runner.Targets[0])

		client, err := rpc.DialHTTP("tcp", runner.Targets[0].RPCAddr())
		if err != nil {
			log.Fatal().Msgf("dialing:%s", err)
		}

		var reply struct{}
		args := MigrateArgs{
			DumpPath:    runner.LatestDump.Base(),
			ContainerId: runner.ContainerId,
			BundlePath:  runner.BundlePath,
		}
		err = client.Call("RPC.Migrate", args, &reply)
		if err != nil {
			log.Error().Str("Error", err.Error()).Msg("Failed to call RPC")
			runner.SetStatusNoLock(runner_context.Failed)
			return
		}

		runner.SetStatusNoLock(runner_context.Stopped)
	})
}

func (runner *Runner) loopRestoring() {
	runner.WithLock(func() {
		log.Trace().Msg("Restoring container")
		go runner.restoreContainer()
		log.Info().
			Str("ContainerId", runner.ContainerId).
			Str("Dump", runner.LatestDump.Path()).
			Str("Bundle", runner.BundlePath).
			Msg("Container restored")
		runner.SetStatusNoLock(runner_context.Running)
	})
}

func (runner *Runner) loopJoining() {
	runner.WithLock(func() {
		log.Trace().Str("Remote", runner.Source).Msg("Joining cluster")

		client, err := rpc.DialHTTP("tcp", runner.Source)
		if err != nil {
			log.Error().Str("Error", err.Error()).Msg("Failed to dial RPC")
			runner.SetStatusNoLock(runner_context.Failed)
			return
		}

		var reply string
		args := runner.ToTarget()
		err = client.Call("RPC.Join", args, &reply)
		if err != nil {
			log.Error().Str("Error", err.Error()).Msg("Failed to call RPC")
			runner.SetStatusNoLock(runner_context.Failed)
			return
		}

		runner.ContainerId = reply

		log.Info().Str("ContainerId", reply).Msg("Successfully joined cluster")
		runner.SetStatusNoLock(runner_context.StandBy)
	})
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
			pingTimeout := time.Duration(env.Getenv().PING_TIMEOUT) * time.Second
			select {
			case <-time.After(pingTimeout):
				log.Warn().
					Msgf(
						"No ping received in %s. Assuming source is down. Starting recovery",
						pingTimeout.String(),
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

	runner.LatestDump = dump.Recover()
	runner.Source = ""

	log.Info().Str("Dump", runner.LatestDump.Path()).Msg("Recovering from dump")
	runner.RestoreContainer()
}

// Get the current runner represented as a remote target.
func (runner *Runner) ToTarget() remote_target.RemoteTarget {
	return remote_target.New(
		utils.GetLocalIP(),
		runner.RPCPort(),
		env.Getenv().DUMP_PATH,
		22, // TODO: Don't hard code port
	)
}
