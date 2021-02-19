package cli_commands

import (
	"github.com/rs/zerolog/log"

	"github.com/Xarepo/msc-container-migration/internal/runner"
	"github.com/Xarepo/msc-container-migration/internal/runner/runner_context"
)

type Join struct {
	Remote string `kong:"arg,help='The RPC-address of the remote host to join'"`
}

func (cmd Join) Execute() error {
	// Prepare new runner by creating it with empty values
	r := runner.New("", ".")
	r.Source = cmd.Remote

	r.Start()
	r.SetStatus(runner_context.Joining)

	log.Trace().Msg("Waiting for container to exit")
	status := r.WaitForContainer()
	log.Info().Int("Status", status).Msg("Container exited")

	log.Info().Msg("Stopping runner...")
	r.WaitFor()
	r.SetStatus(runner_context.Stopped)

	return nil
}
