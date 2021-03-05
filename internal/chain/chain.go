package chain

import (
	"github.com/rs/zerolog/log"

	chain_node "github.com/Xarepo/msc-container-migration/internal/chain/node"
	"github.com/Xarepo/msc-container-migration/internal/dump"
	"github.com/Xarepo/msc-container-migration/internal/remote_target"
	"github.com/Xarepo/msc-container-migration/internal/sftp"
)

type DumpChain struct {
	latest *chain_node.ChainNode
	length int
}

func New() *DumpChain {
	return &DumpChain{
		latest: nil,
		length: 0,
	}
}

func (chain *DumpChain) Latest() *chain_node.ChainNode {
	return chain.latest
}

// Add a dump to the end of the chain
func (chain *DumpChain) Push(dump dump.Dump) {
	newNode := chain_node.New(&dump, chain.latest, false)
	chain.latest = newNode
	chain.length += 1
}

// Transfers all dumps in the chain regardless of whether or not they have
// been synced.
func (chain *DumpChain) FullTransfer(target *remote_target.RemoteTarget) {
	log.Debug().
		Str("Target", target.Host).
		Msg("Performing full transfer of chain to target")
	next := chain.latest
	for next != nil {
		sftp.CopyToRemote(next, target)
		next.SetSynced()
		next = next.GetPrev()
	}
}

// Transfers all dumps in the chain that has not previously been synced.
func (chain *DumpChain) Sync(target *remote_target.RemoteTarget) {
	log.Debug().
		Str("Target", target.Host).
		Int("FileTransferPort", target.FileTransferPort).
		Msg("Syncing chain to target")
	next := chain.latest
	for next != nil && !next.IsSynced() {
		sftp.CopyToRemote(next, target)
		next.SetSynced()
		next = next.GetPrev()
	}
}

func (chain DumpChain) Length() int {
	return chain.length
}
