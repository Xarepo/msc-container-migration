package cli_commands

import (
	"github.com/rs/zerolog/log"

	"github.com/Xarepo/msc-container-migration/internal/rpc_listener"
	"github.com/Xarepo/msc-container-migration/internal/runner"
)

type Listen struct {
	RPCListener rpc_listener.RPCListener
}

func (cmd Listen) Execute() error {
	log.Info().
		Msg("Listening for migrations")

	// Prepare new runner by creating it with empty values
	r := runner.New("", ".", "")

	go r.Start()
	r.WaitFor()

	return nil
}
