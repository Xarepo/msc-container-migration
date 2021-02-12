package rpc

import (
	. "github.com/Xarepo/msc-container-migration/internal/runner/runner_context"
)

type Ping struct{}

func (ping Ping) Execute(ctx *RunnerContext, remoteAddr string) {
	ctx.PingInterrupt <- true
}

func (ping Ping) ParseFlags([]string) error {
	return nil
}

func (ping Ping) String() string {
	return RPC_PING
}
