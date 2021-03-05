package runner

import (
	"regexp"
	"sort"
	"strconv"

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

	// Transfer chains
	handler.runner.WithLock(func() {
		if handler.runner.PrevChain != nil {
			handler.runner.PrevChain.FullTransfer(target)
		}
		handler.runner.Chain.FullTransfer(target)
	})

	return nil
}

func (handler *RPCHandler) Ping(args struct{}, reply *bool) error {
	log.Debug().Msg("PING received")
	handler.runner.PingInterrupt <- true
	*reply = true
	return nil
}

type MigrateArgs struct {
	DumpNames               []string
	ContainerId, BundlePath string
}

func (handler *RPCHandler) Migrate(args *MigrateArgs, reply *struct{}) error {
	log.Debug().Strs("DumpNames", args.DumpNames).Msg("Migration request received")

	// Sort the names according to their numbers, in ascending order.
	sort.SliceStable(args.DumpNames, func(i, j int) bool {
		re_nr := regexp.MustCompile("[0-9]+")
		n1, _ := strconv.Atoi(re_nr.FindString(args.DumpNames[i]))
		n2, _ := strconv.Atoi(re_nr.FindString(args.DumpNames[j]))
		return n1 < n2
	})

	for _, name := range args.DumpNames {
		dump := dump.Restore(name)
		handler.runner.Chain.Push(*dump)
	}
	handler.runner.ContainerId = args.ContainerId
	handler.runner.BundlePath = args.BundlePath
	handler.runner.SetStatus(runner_context.Restoring)
	return nil
}
