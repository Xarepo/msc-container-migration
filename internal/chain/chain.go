package chain

import (
	"github.com/rs/zerolog/log"

	"github.com/Xarepo/msc-container-migration/internal/dump"
	"github.com/Xarepo/msc-container-migration/internal/remote_target"
	"github.com/Xarepo/msc-container-migration/internal/sftp"
)

type chainNode struct {
	el     *dump.Dump
	prev   *chainNode
	synced bool
}

type DumpChain struct {
	latest *chainNode
	length int
}

func New() *DumpChain {
	return &DumpChain{
		latest: nil,
		length: 0,
	}
}

func (chain *DumpChain) Latest() chainNode {
	return *chain.latest
}

func (chain *DumpChain) Push(dump dump.Dump) {
	newNode := &chainNode{
		el:     &dump,
		prev:   chain.latest,
		synced: false,
	}
	chain.latest = newNode
	chain.length += 1
}

func (chain *DumpChain) FullTransfer(target *remote_target.RemoteTarget) {
	log.Debug().
		Str("Target", target.Host).
		Msg("Performing full transfer of chain to target")
	next := chain.latest
	for next != nil {
		sftp.CopyToRemote(next.el.Path(), target)
		next.synced = true
		next = next.prev
	}
}

func (chain *DumpChain) Sync(target *remote_target.RemoteTarget) {
	log.Debug().
		Str("Target", target.Host).
		Int("FileTransferPort", target.FileTransferPort).
		Msg("Syncing chain to target")
	next := chain.latest
	for next != nil && !next.synced {
		sftp.CopyToRemote(next.el.Path(), target)
		next.synced = true
		next = next.prev
	}
}
