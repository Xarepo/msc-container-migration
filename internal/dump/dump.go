package dump

import (
	"fmt"
	"io/ioutil"
	"math"
	"regexp"
	"strconv"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/Xarepo/msc-container-migration/internal/env"
)

// dumpType represents the type of the dump
//
// preDump:
// This dump only contains the memory changed since last dump. It is not
// possiblle to restore from dumps of this type.
//
// fullDump:
// This dump contains the full state of the application. It is possible to
// restore from dumps of this type.
//
// checkpoint:
// This dump is the same as fullDump, but it has been issued manually via an
// IPC rather than automatically by the system. Note that checkpoint dumps may
// have the same number as another fulldump or preDump dump.
type dumpType int

const (
	preDump dumpType = iota
	fullDump
	checkpoint
)

type Dump struct {
	_type dumpType
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
	var _dumpType dumpType = -1
	switch _type {
	case "p":
		_dumpType = preDump
	case "d":
		_dumpType = fullDump
	default:
		log.Panic().Str("DumpName", dumpName).Msg("Failed to reconstruct dump")
	}

	return &Dump{_type: _dumpType, nr: nr}
}

// Construct a checkpoint dump from another dump.
func (dump *Dump) Checkpoint() *Dump {
	return &Dump{_type: checkpoint, nr: dump.nr + 1}
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
	return &Dump{_type: fullDump, nr: int(latestDumpNum)}, nil
}

func (dump Dump) Path() string {
	return fmt.Sprintf("%s/%s", env.Getenv().DUMP_PATH, dump.Base())
}

func (dump Dump) Base() string {
	prefix := '_'
	switch dump._type {
	case preDump:
		prefix = 'p'
	case fullDump:
		prefix = 'd'
	case checkpoint:
		prefix = 'c'
	default:
		log.Panic().Int("Type", int(dump._type)).Msg("Dump has invalid type")
	}
	return fmt.Sprintf("%c%d", prefix, dump.nr)
}

// Return whether of not the dump is a predump
func (dump Dump) PreDump() bool {
	return dump._type == preDump
}

// Return the next dump to dump based on this dump and the current chain
// length.
func (dump Dump) NextDump(chainLength int) *Dump {
	t := fullDump
	maxChainLength := env.Getenv().CHAIN_LENGTH
	if chainLength < maxChainLength-1 {
		t = preDump
	}
	return &Dump{_type: t, nr: dump.nr + 1}
}

// Return the next pre-dump based on this dump.
func (dump Dump) NextPreDump() *Dump {
	return &Dump{_type: preDump, nr: dump.nr + 1}
}

// Return the next full dump based on this dump.
func (dump Dump) NextFullDump() *Dump {
	return &Dump{_type: fullDump, nr: dump.nr + 1}
}

// Return the first of all dumps, across all hosts.
func FirstDump() *Dump {
	return &Dump{_type: preDump, nr: 0}
}

// Return the first dump of the next chain
func (dump Dump) NextChainDump() *Dump {
	return &Dump{_type: preDump, nr: dump.nr + 1}
}

// Return the dump represented as a parent path to another dump.
func (dump Dump) ParentPath() string {
	return fmt.Sprintf("../%s", dump.Base())
}
