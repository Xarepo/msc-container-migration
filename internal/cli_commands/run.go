package cli_commands

import (
	_ "github.com/joho/godotenv/autoload"
	"github.com/rs/zerolog/log"

	"github.com/Xarepo/msc-container-migration/internal/runner"
)

type Run struct {
	ContainerId *string
	BundlePath  *string
}

// Execute the run command.
//
// The run command runs (i.e. create and starts) the container using runc. The
// container is run in a goroutine to not block execution.
// The function does not return until the container has exited.
func (cmd Run) Execute() error {
	log.Trace().
		Str("BundlePath", *cmd.BundlePath).
		Str("ContainerId", *cmd.ContainerId).
		Msg("Executing run command")

	runner := runner.New(*cmd.ContainerId, *cmd.BundlePath, "")

	runner.Start()
	runner.Run()

	log.Trace().Msg("Waiting for container to exit")
	status := runner.WaitFor()

	log.Info().Int("Status", status).Msg("Container exited")
	return nil
}
