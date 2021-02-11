package runner_context

import (
	"os"
	"strconv"
	"sync"

	"github.com/rs/zerolog/log"

	. "github.com/Xarepo/msc-container-migration/internal/image"
	"github.com/Xarepo/msc-container-migration/internal/ipc/utils"
	. "github.com/Xarepo/msc-container-migration/internal/ipc_listener"
	"github.com/Xarepo/msc-container-migration/internal/remote_target"
	. "github.com/Xarepo/msc-container-migration/internal/rpc_listener"
	"github.com/Xarepo/msc-container-migration/internal/udp_listener"
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
// The runner has been started, but is not running, it is waiting to either
// a) be run and transition into running status or b) be restored from a dumped
// image by migration request.
//
// Running:
// The runner is running. This means the runner has, or is the process of
// creating and starting the container.
//
// Migrating:
// The runner is in the process of migrating it's container to another host.
// NOTE: The runner should always have this status for EXACTLY one cycle of the
// runner's loop.
//
// Restoring:
// The runner is in the process of creating and running its container from a
// dumped image as part of a migration process.
// NOTE: The runner should always have this status for EXACTLY one cycle of the
// runner's loop.
//
// Joining:
// The runner is in the process of joining a cluster.
// NOTE: The runner should always have this status for EXACTLY one cycle of the
// runner's loop.
type RunnerStatus string

const (
	Stopped   RunnerStatus = "Stopped"
	StandBy                = "StandBy"
	Running                = "Running"
	Migrating              = "Migrating"
	Restoring              = "Restoring"
	Joining                = "Joining"
	Recovery               = "Recovery"
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
	// The latest checkpoint image that the runner has dumped, i.e. written to
	// disk.
	LatestImage *Image
	IPCListener
	RPCPort int
	RPCListener
	status        RunnerStatus
	lock          sync.Mutex
	Targets       []remote_target.RemoteTarget
	Source        string
	PingInterrupt chan bool
	// This channel can be read from when a RPC needs to wait for a corresponding
	// ACK message to be received.
	AckWait chan bool
}

func New(containerId, bundlePath, imagePath string) RunnerContext {
	var rpcPort int
	var err error
	if os.Getenv("RPC_PORT") == "" {
		log.Warn().Msg("Environment variable RPC_PORT not set, defaulting to 1234.")
		rpcPort = 1234
	} else {
		rpcPort, err = strconv.Atoi(os.Getenv("RPC_PORT"))
		if err != nil {
			log.Panic().Str("Error", err.Error()).Msg("Failed to parse RPC port")
		}
	}

	return RunnerContext{
		ContainerId:     containerId,
		ContainerStatus: make(chan int),
		BundlePath:      bundlePath,
		IPCListener:     USockListener{SockAddr: utils.GetSocketAddr(containerId)},
		RPCListener:     udp_listener.UDPListener{},
		RPCPort:         rpcPort,
		status:          Stopped,
		LatestImage:     nil,
		Targets:         []remote_target.RemoteTarget{},
		Source:          "",
		PingInterrupt:   make(chan bool),
		AckWait:         make(chan bool),
	}
}

func (ctx *RunnerContext) SetStatus(status RunnerStatus) {
	log.Debug().Str("Status", string(status)).Msg("Status set")
	ctx.status = status
}

func (ctx *RunnerContext) RunnerStatus() RunnerStatus {
	return ctx.status
}

func (ctx *RunnerContext) AddTarget(target remote_target.RemoteTarget) {
	ctx.Targets = append(ctx.Targets, target)
	log.Info().
		Str("RemoteTarget", target.Host()).
		Int("RPCPort", target.RPCPort()).
		Int("FileTransferPort", target.FileTransferPort()).
		Msg("Added target")
}

func (ctx *RunnerContext) WithLock(f func()) {
	ctx.lock.Lock()
	f()
	ctx.lock.Unlock()
}
