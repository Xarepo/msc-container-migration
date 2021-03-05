package dump

import (
	"fmt"
	"io/ioutil"
	"math"
	"path"
	"regexp"
	"strconv"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	. "github.com/Xarepo/msc-container-migration/internal/dump/type"
	dump_type "github.com/Xarepo/msc-container-migration/internal/dump/type"
	"github.com/Xarepo/msc-container-migration/internal/env"
)

type Dump struct {
	_type DumpType
	nr    int
}

// Construct a dump based on a dumpName.
func FromString(dumpName string) *Dump {
	re_nr := regexp.MustCompile("[0-9]+")
	re_type := regexp.MustCompile("[a-z]+")
	nr, err := strconv.Atoi(re_nr.FindString(dumpName))
	_type := re_type.FindString(dumpName)
	if err != nil {
		log.Error().Str("Error", err.Error()).Send()
	}
	_dumpType := dump_type.FromString(_type)

	return &Dump{_type: _dumpType, nr: nr}
}

// Construct a checkpoint dump from another dump.
func (dump *Dump) Checkpoint() *Dump {
	return &Dump{_type: dump_type.Checkpoint, nr: dump.nr + 1}
}

// Retrieve the latest dump possible to recover from, i.e. the last transferred
// full dump.
//
// The latest possible dump to recover from is determined by finding the
// directory fullfilling:
// 1) its name starts with the character 'd', i.e. the name is of the form
// "dX" (where X is an integer)
// 2) It has the highest number in its name of all full dumps (directories
// prefixed with 'd'), i.e. for all directories of the form "dX", its X is
// maximal.
func Recover() (*Dump, error) {
	entries, err := ioutil.ReadDir(env.Getenv().DUMP_PATH)
	if err != nil {
		log.Error().
			Str("Error", err.Error()).
			Msg("Failed to read recovery directory")
	}

	// Find the directory that fullfills the requirements described in the
	// function comment.
	latestDumpNum := math.Inf(-1)
	r := regexp.MustCompile("[0-9]+")
	for _, entry := range entries {
		num, _ := strconv.Atoi(r.FindString(entry.Name()))
		if entry.IsDir() && entry.Name()[0] == 'd' && float64(num) > latestDumpNum {
			latestDumpNum = float64(num)
		}
	}
	if latestDumpNum == math.Inf(-1) {
		return nil, errors.New("Failed to find a dump directory")
	}

	latestDump := fmt.Sprintf("d%d", int(latestDumpNum))
	log.Debug().
		Str("Dump", latestDump).
		Msg("Latest possible recovery dump determined")
	return &Dump{_type: dump_type.FullDump, nr: int(latestDumpNum)}, nil
}

func (dump Dump) Path() string {
	return path.Join(env.Getenv().DUMP_PATH, dump.Base())
}

func (dump Dump) Base() string {
	prefix := dump._type.ToChar()
	return fmt.Sprintf("%c%d", prefix, dump.nr)
}

// Return whether of not the dump is a predump
func (dump Dump) PreDump() bool {
	return dump._type == dump_type.PreDump
}

// Return the next dump to dump based on this dump and the current chain
// length.
func (dump Dump) NextDump(chainLength int) *Dump {
	t := dump_type.FullDump
	maxChainLength := env.Getenv().CHAIN_LENGTH
	if chainLength < maxChainLength-1 {
		t = dump_type.PreDump
	}
	return &Dump{_type: t, nr: dump.nr + 1}
}

// Return the next pre-dump based on this dump.
func (dump Dump) NextPreDump() *Dump {
	return &Dump{_type: dump_type.PreDump, nr: dump.nr + 1}
}

// Return the next full dump based on this dump.
func (dump Dump) NextFullDump() *Dump {
	return &Dump{_type: dump_type.FullDump, nr: dump.nr + 1}
}

// Return the first of all dumps, across all hosts.
func FirstDump() *Dump {
	return &Dump{_type: dump_type.PreDump, nr: 0}
}

// Return the first dump of the next chain
func (dump Dump) NextChainDump() *Dump {
	return &Dump{_type: dump_type.PreDump, nr: dump.nr + 1}
}

// Return the dump represented as a parent path to another dump.
func (dump Dump) ParentPath() string {
	return fmt.Sprintf("../%s", dump.Base())
}
