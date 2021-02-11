package rpc

import (
	"strings"

	"github.com/rs/zerolog/log"

	. "github.com/Xarepo/msc-container-migration/internal/runner/runner_context"
)

type RPC interface {
	Execute(context *RunnerContext, remoteAddr string)
	ParseFlags([]string) error
	String() string
}

// Available RPCs
const (
	RPC_MIGRATE  = "MIGRATE"
	RPC_JOIN     = "JOIN"
	RPC_PING     = "PING"
	RPC_JOIN_ACK = "JOIN_ACK"
)

func ParseRPC(message string) RPC {
	fields := strings.Split(message, " ")
	var rpc RPC = nil

	switch fields[0] {
	case RPC_MIGRATE:
		rpc = &Migrate{}
	case RPC_JOIN:
		rpc = &Join{}
	case RPC_PING:
		rpc = &Ping{}
	case RPC_JOIN_ACK:
		rpc = &JoinAck{}
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
