package chain

import (
	"os"
	"path"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	chain_node "github.com/Xarepo/msc-container-migration/internal/chain/node"
	"github.com/Xarepo/msc-container-migration/internal/dump"
	"github.com/Xarepo/msc-container-migration/internal/env"
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
		sftp.TransferDump(next, target)
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
		sftp.TransferDump(next, target)
		next.SetSynced()
		next = next.GetPrev()
	}
}

func (chain *DumpChain) GetNames() []string {
	next := chain.latest
	var names []string
	for next != nil {
		names = append(names, next.Dump().Base())
		next = next.GetPrev()
	}
	return names
}

func (chain DumpChain) Length() int {
	return chain.length
}

// Reconstructs a chain based on the path to the latest dump of the chain.
//
// Uses the parent symlinks of the dump directories in order to recursively
// reconstruct the chain and its nodes.
// Stops when there is no parent symlink in the specified dump path.
func ReconstructChain(dumpPath string) ([]string, error) {
	linkPath := path.Join(dumpPath, "parent")
	fi, err := os.Lstat(linkPath)
	if err != nil {
		log.Trace().
			Str("File", dumpPath).
			Msg("Could not find parent symlink, assuming end of chain")
		return []string{path.Base(dumpPath)}, nil
	}

	if (fi.Mode() & os.ModeSymlink) != 0 {
		parentDest, err := os.Readlink(linkPath)
		if err != nil {
			return []string{dumpPath}, errors.Wrap(err, "Failed to read symlink")
		}
		parentDestAbs := path.Join(env.Getenv().DUMP_PATH, path.Base(parentDest))
		res, err := ReconstructChain(parentDestAbs)
		if err != nil {
			return []string{}, err
		}
		return append(res, path.Base(dumpPath)), nil
	} else {
		return []string{}, errors.New("File is not symlink")
	}
}
