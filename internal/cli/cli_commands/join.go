package cli_commands

import (
	"github.com/Xarepo/msc-container-migration/internal/runner"
	"github.com/Xarepo/msc-container-migration/internal/runner/runner_context"
)

type Join struct {
	Remote string `kong:"arg,help='The RPC-address of the remote host to join'"`
}

func (cmd Join) Execute() error {
	// Prepare new runner by creating it with empty values
	r := runner.New("", ".", "")
	r.Source = cmd.Remote

	r.Start()
	r.SetStatus(runner_context.Joining)
	r.WaitFor()

	return nil
}
