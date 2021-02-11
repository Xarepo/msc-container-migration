package ipc

import (
	"errors"
	"fmt"

	"github.com/rs/zerolog/log"

	"github.com/Xarepo/msc-container-migration/internal/runner/runner_context"
)

type Migrate struct {
	ContainerId string
}

func (migrate Migrate) Send() {
	msg := []byte(fmt.Sprintf("%s %s", IPC_MIGRATE, migrate.ContainerId))
	sendMessage(&msg, migrate.ContainerId)
}

func (migrate Migrate) Execute(ctx *runner_context.RunnerContext) {
	log.Trace().
		Str("ContainerId", migrate.ContainerId).
		Msg("Executing migrate IPC")
	ctx.SetStatus(runner_context.Migrating)
}

func (migrate *Migrate) ParseFlags(flags []string) error {
	if len(flags) < 1 {
		return errors.New("Too few arguments")
	}

	migrate.ContainerId = flags[0]

	return nil
}
