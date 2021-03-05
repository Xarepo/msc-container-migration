package dump_type

import "github.com/rs/zerolog/log"

// DumpType represents the type of the dump
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
type DumpType int

const (
	PreDump DumpType = iota
	FullDump
	Checkpoint
)

func FromString(s string) DumpType {
	switch s {
	case "p":
		return PreDump
	case "d":
		return FullDump
	case "c":
		return Checkpoint
	default:
		log.Panic().Str("DumpName", s).Msg("Failed to reconstruct dump")
		return -1 // Avoid linter complaining about missing return
	}
}

func (dt DumpType) ToChar() rune {
	switch dt {
	case PreDump:
		return 'p'
	case FullDump:
		return 'd'
	case Checkpoint:
		return 'c'
	default:
		log.Panic().Int("Type", int(dt)).Msg("Dump has invalid type")
		return '_' // Avoid linter complaining about missing return
	}
}
