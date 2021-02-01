package cli_commands

import (
	"github.com/rs/zerolog/log"

	"github.com/Xarepo/msc-container-migration/internal/ipc"
)

type Migrate struct {
	ContainerId *string
}

func (cmd Migrate) Execute() error {
	log.Trace().
		Str("ContainerId", *cmd.ContainerId).
		Msg("Executing migrate command")
	ipc := ipc.Migrate{ContainerId: *cmd.ContainerId}
	ipc.Send()

	return nil
}
