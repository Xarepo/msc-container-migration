package cli_commands

import (
	"fmt"

	"github.com/Xarepo/msc-container-migration/internal/runc"
	"github.com/rs/zerolog/log"
)

type Run struct {
	ContainerId *string
	BundlePath  *string
}

func (cmd Run) Execute() error {
	log.Trace().Str("BundlePath", *cmd.BundlePath).Str("ContainerId", *cmd.ContainerId).Msg("Executing run command")
	status, err := runc.Run(*cmd.ContainerId, *cmd.BundlePath)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	log.Debug().Str("Status", string(status)).Msg("Run exited")
	return nil
}
