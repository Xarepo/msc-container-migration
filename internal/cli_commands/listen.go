package cli_commands

import (
	"github.com/Xarepo/msc-container-migration/internal/rpc_listener"
	"github.com/Xarepo/msc-container-migration/internal/runner"
	"github.com/Xarepo/msc-container-migration/internal/runner/runner_context"
)

type Listen struct {
	RPCListener rpc_listener.RPCListener
	Remote      string
}

func (cmd Listen) Execute() error {
	// Prepare new runner by creating it with empty values
	r := runner.New("", ".", "")
	r.Source = cmd.Remote

	r.Start()
	r.SetStatus(runner_context.Joining)
	r.WaitFor()

	return nil
}
