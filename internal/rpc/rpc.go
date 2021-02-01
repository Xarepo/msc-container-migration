package rpc

import (
	"strings"

	"github.com/rs/zerolog/log"

	. "github.com/Xarepo/msc-container-migration/internal/runner/runner_context"
)

type RPC interface {
	Execute(context *RunnerContext)
	ParseFlags([]string) error
	String() string
}

// Available RPCs
const RPC_MIGRATE = "MIGRATE"

func ParseRPC(message string) RPC {
	fields := strings.Split(message, " ")
	var rpc RPC = nil

	switch fields[0] {
	case RPC_MIGRATE:
		rpc = &Migrate{}
	default:
		log.Error().Str("RPC", fields[0]).Msg("Received unknown RPC")
		return nil
	}
	if err := rpc.ParseFlags(fields[1:]); err != nil {
		log.Error().Str("Error", err.Error()).Msg("Failed to parse flags")
		return nil
	}
	return rpc
}
