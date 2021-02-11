package rpc

import (
	"errors"
	"fmt"

	. "github.com/Xarepo/msc-container-migration/internal/runner/runner_context"
)

type JoinAck struct {
	containerId string
}

func NewJoinAck(containerId string) *JoinAck {
	return &JoinAck{
		containerId: containerId,
	}
}

func (joinAck *JoinAck) Execute(ctx *RunnerContext, remoteAddr string) {
	ctx.ContainerId = joinAck.containerId
	ctx.AckWait <- true
}

func (joinAck *JoinAck) ParseFlags(flags []string) error {
	if len(flags) < 1 {
		return errors.New("Too few arguments")
	}
	joinAck.containerId = flags[0]
	return nil
}

func (joinAck *JoinAck) String() string {
	return fmt.Sprintf("%s %s", RPC_JOIN_ACK, joinAck.containerId)
}
