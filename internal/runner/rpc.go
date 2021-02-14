package runner

import (
	"github.com/rs/zerolog/log"

	"github.com/Xarepo/msc-container-migration/internal/image"
	"github.com/Xarepo/msc-container-migration/internal/remote_target"
	"github.com/Xarepo/msc-container-migration/internal/runner/runner_context"
)

// RPCHandler is a struct encapsulating all RPCs. This makes sure only RPC
// methods are tried when registering, and not other runner methods (which
// would fail to be registered), like Start() or Run().
type RPCHandler struct {
	runner *Runner
}

func (handler *RPCHandler) Join(
	target *remote_target.RemoteTarget,
	reply *string,
) error {
	log.Trace().
		Str("Host", target.Host).
		Int("RPCPort", target.RPCPort).
		Int("FileTransferPort", target.FileTransferPort).
		Msg("Executing JOIN RPC")

	handler.runner.AddTarget(*target)
	*reply = handler.runner.ContainerId
	return nil
}

func (handler *RPCHandler) Ping(args *struct{}, reply *struct{}) error {
	log.Debug().Msg("PING received")
	handler.runner.PingInterrupt <- true
	return nil
}

type MigrateArgs struct {
	ImagePath, ContainerId, BundlePath string
}

func (handler *RPCHandler) Migrate(args *MigrateArgs, reply *struct{}) error {
	log.Debug().Msg("Migration request received")
	img := image.Restore(args.ImagePath)
	handler.runner.ContainerId = args.ContainerId
	handler.runner.BundlePath = args.BundlePath
	handler.runner.LatestImage = img
	handler.runner.SetStatus(runner_context.Restoring)
	return nil
}
