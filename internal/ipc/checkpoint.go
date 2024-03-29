package ipc

import (
	"fmt"

	"github.com/rs/zerolog/log"

	"github.com/Xarepo/msc-container-migration/internal/runc"
	"github.com/Xarepo/msc-container-migration/internal/runner/runner_context"
)

type Checkpoint struct {
}

func (cp Checkpoint) Send() {
	msg := []byte(fmt.Sprintf("%s", IPC_CHECKPOINT))
	sendMessage(&msg)
}

func (cp Checkpoint) Execute(ctx *runner_context.RunnerContext) {
	log.Trace().
		Msg("Executing checkpoint IPC")
	// Take lock so that no other routine can dump at the same time
	ctx.WithLock(func() {
		checkpointImg := ctx.Chain.Latest().Dump().Checkpoint()
		runc.Dump(ctx.ContainerId, checkpointImg.Path(), "", true)
	})
}

func (cp *Checkpoint) ParseFlags(flags []string) error {
	return nil
}
