package runner

import (
	"fmt"

	"github.com/rs/zerolog/log"

	"github.com/Xarepo/msc-container-migration/internal/dump"
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
	handler.runner.Chain.FullTransfer(target)
	return nil
}

func (handler *RPCHandler) Ping(args struct{}, reply *bool) error {
	log.Debug().Msg("PING received")
	handler.runner.PingInterrupt <- true
	*reply = true
	return nil
}

type MigrateArgs struct {
	DumpPath, ContainerId, BundlePath string
}

func (handler *RPCHandler) Migrate(args *MigrateArgs, reply *struct{}) error {
	log.Debug().Msg("Migration request received")
	dump := dump.Restore(args.DumpPath)
	handler.runner.ContainerId = args.ContainerId
	handler.runner.BundlePath = args.BundlePath
	handler.runner.Chain.Push(*dump)
	fmt.Printf("LENGTH: %d", handler.runner.Chain.Length())
	handler.runner.SetStatus(runner_context.Restoring)
	return nil
}
