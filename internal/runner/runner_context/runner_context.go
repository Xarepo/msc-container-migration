package runner_context

import (
	"sync"

	"github.com/rs/zerolog/log"

	. "github.com/Xarepo/msc-container-migration/internal/dump"
	"github.com/Xarepo/msc-container-migration/internal/env"
	. "github.com/Xarepo/msc-container-migration/internal/ipc_listener"
	"github.com/Xarepo/msc-container-migration/internal/remote_target"
	. "github.com/Xarepo/msc-container-migration/internal/usock_listener"
)

// RunnerStatus is an enum describing the status of the runner. The status of
// the runner determines its next action in the runner's loop.
//
// Stopped:
// The container has either yet not been started or has been stopped, possibly
// by a migration request.
//
// Standby:
// The runner has been started, but is not running, it is waiting to either a)
// be run and transition into running status or b) be restored from a dump by a
// migration request. At this point the runner's loop, IPC-listener and
// RPC-listener has been started, but not yet the container.
//
// Running:
// The runner is running. This means the runner has, or is in the process of
// creating and starting the container.
//
// Migrating:
// The runner is in the process of migrating it's container to another host.
// NOTE: The runner should always have this status for EXACTLY one cycle of the
// runner's loop.
//
// Restoring:
// The runner is in the process of creating and running its container from a
// dump as part of a migration process.
// NOTE: The runner should always have this status for EXACTLY one cycle of the
// runner's loop.
//
// Joining:
// The runner is in the process of joining a cluster.
// NOTE: The runner should always have this status for EXACTLY one cycle of the
// runner's loop.
//
// Recovery:
// The runner has lost connection to the source, i.e. pings have timed out, and
// is in the process of recovery from the latest possible dump.
// NOTE: The runner should always have this status for EXACTLY one cycle of the
// runner's loop.
//
// Failed:
// The runner or its container has failed, and the runner is in a
// non-recoverable state.
//
// Terminated:
// The runner has been deliberatly terminated via user input, e.g. signals
// SIGTERM or SIGINT.
type RunnerStatus string

const (
	Stopped    RunnerStatus = "Stopped"
	StandBy                 = "StandBy"
	Running                 = "Running"
	Migrating               = "Migrating"
	Restoring               = "Restoring"
	Joining                 = "Joining"
	Recovery                = "Recovery"
	Failed                  = "Failed"
	Terminated              = "Terminated"
)

// RunnerContext represents the state of the runner.
type RunnerContext struct {
	// The id of the container the runner is running or is about to run.
	// May be empty if the runner is in standby and is waiting for a migration
	// request.
	ContainerId string
	// The goroutine that runs the container writes the container's exit status
	// to this after the container has exited.
	ContainerStatus chan int
	// The path to the OCI-bundle that the runner's container is created from.
	BundlePath string
	// The latest dump that the runner has made, i.e. written to disk.
	LatestDump *Dump
	IPCListener
	rpcPort int
	status  RunnerStatus
	lock    sync.Mutex
	// A list of targets of which to replicate when the runner is running.
	Targets []remote_target.RemoteTarget
	// The address of the source node to listen to for migrations. This will be
	// empty if the runner is running.
	Source        string
	PingInterrupt chan bool
}

func New(containerId, bundlePath string) RunnerContext {
	return RunnerContext{
		ContainerId:     containerId,
		ContainerStatus: make(chan int),
		BundlePath:      bundlePath,
		IPCListener:     &USockListener{},
		rpcPort:         env.Getenv().RPC_PORT,
		status:          Stopped,
		LatestDump:      nil,
		Targets:         []remote_target.RemoteTarget{},
		Source:          "",
		PingInterrupt:   make(chan bool),
	}
}

// Sets the status of the runner after locking
func (ctx *RunnerContext) SetStatus(status RunnerStatus) {
	ctx.WithLock(func() {
		log.Debug().Str("Status", string(status)).Msg("Status set")
		ctx.status = status
	})
}

// Sets the status of the runner without locking.
// Useful for when needing to set the status from within the callback passed to
// WithLock().
func (ctx *RunnerContext) SetStatusNoLock(status RunnerStatus) {
	log.Debug().Str("Status", string(status)).Msg("Status set")
	ctx.status = status
}

// Return the status of the runner.
func (ctx *RunnerContext) Status() RunnerStatus {
	return ctx.status
}

// Add a target to the targets list.
func (ctx *RunnerContext) AddTarget(target remote_target.RemoteTarget) {
	ctx.Targets = append(ctx.Targets, target)
	log.Info().
		Str("RemoteTarget", target.Host).
		Int("RPCPort", target.RPCPort).
		Int("FileTransferPort", target.FileTransferPort).
		Msg("Added target")
}

// Remove a target from the targets list
func (ctx *RunnerContext) RemoveTarget(target remote_target.RemoteTarget) {
	index := -1
	for i, t := range ctx.Targets {
		if t.RPCAddr() == target.RPCAddr() {
			index = i
		}
	}
	if index != -1 {
		ctx.Targets = append(ctx.Targets[:index], ctx.Targets[index+1:]...)
		log.Warn().
			Str("Target", target.RPCAddr()).
			Msg("Removed target")
	}
}

// Return the RPC port of the runner.
func (ctx *RunnerContext) RPCPort() int {
	return ctx.rpcPort
}

// Locks the context's lock and calls a function.
// Handles both locking and unlocking of the lock.
func (ctx *RunnerContext) WithLock(f func()) {
	ctx.lock.Lock()
	f()
	ctx.lock.Unlock()
}
