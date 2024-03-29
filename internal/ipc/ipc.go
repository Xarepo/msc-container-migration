package ipc

import (
	"net"
	"strings"

	"github.com/rs/zerolog/log"

	"github.com/Xarepo/msc-container-migration/internal/runner/runner_context"
	"github.com/Xarepo/msc-container-migration/internal/usock_listener"
)

func sendMessage(msg *[]byte) {
	sockAddr := usock_listener.SOCK_ADDR
	c, err := net.DialUnix(
		"unixgram",
		nil,
		&net.UnixAddr{Name: sockAddr, Net: "unixgram"},
	)
	if err != nil {
		log.Error().Str("Error", err.Error()).Msgf("Failed to connect IPC socket")
		panic(err)
	}
	defer c.Close()

	_, err = c.Write(*msg)
	if err != nil {
		log.Error().Msgf("Failed to write to socket: %s", err.Error())
	}
}

type IPC interface {
	Send()
	Execute(RunnerContext *runner_context.RunnerContext)
	ParseFlags([]string) error
}

// Available IPCs
const (
	IPC_MIGRATE    = "MIGRATE"
	IPC_CHECKPOINT = "CHECKPOINT"
)

func ParseIPC(message string) IPC {
	fields := strings.Split(message, " ")
	var ipc IPC = nil

	switch fields[0] {
	case IPC_MIGRATE:
		ipc = &Migrate{}
	case IPC_CHECKPOINT:
		ipc = &Checkpoint{}
	default:
		log.Error().Str("IPC", fields[0]).Msg("Received unknown IPC")
		return nil
	}
	if err := ipc.ParseFlags(fields[1:]); err != nil {
		log.Error().Str("Error", err.Error()).Msg("Failed to parse flags")
		return nil
	}
	return ipc
}
