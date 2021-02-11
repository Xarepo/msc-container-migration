package rpc

import (
	"fmt"
	"net"
	"strconv"

	"github.com/rs/zerolog/log"

	"github.com/Xarepo/msc-container-migration/internal/remote_target"
	"github.com/Xarepo/msc-container-migration/internal/runner/runner_context"
)

type Join struct {
	rpcPort          int
	dumpPath         string
	fileTransferPort int
}

func NewJoin(rpcPort, fileTransferPort int, dumpPath string) *Join {
	return &Join{
		rpcPort:          rpcPort,
		fileTransferPort: fileTransferPort,
		dumpPath:         dumpPath,
	}
}

func (j *Join) Execute(ctx *runner_context.RunnerContext, remoteAddr string) {
	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		log.Error().Msg("Failed to parse address")
	}
	log.Trace().
		Str("Host", host).
		Int("RPCPort", j.rpcPort).
		Int("FileTransferPort", j.fileTransferPort).
		Msg("Executing JOIN RPC")

	target := remote_target.New(host, j.rpcPort, j.dumpPath, j.fileTransferPort)

	conn, err := net.Dial("udp4", target.RPCAddr())
	if err != nil {
		log.Error().Str("Error", err.Error()).Msg("Failed to dial UDP")
		return
	}
	defer conn.Close()
	rpc := NewJoinAck(ctx.ContainerId)
	fmt.Fprintf(conn, rpc.String())

	ctx.AddTarget(target)
}

func (j *Join) ParseFlags(fields []string) error {
	rpcPort, err := strconv.Atoi(fields[0])
	if err != nil {
		log.Error().Str("RPC", j.String()).Msg("Failed to parse RPC PORT")
		return err
	}

	fileTransferPort, err := strconv.Atoi(fields[2])
	if err != nil {
		log.Error().Str("RPC", j.String()).Msg("Failed to parse file transfer port")
		return err
	}

	j.rpcPort = rpcPort
	j.dumpPath = fields[1]
	j.fileTransferPort = fileTransferPort

	return nil
}

func (j *Join) String() string {
	return fmt.Sprintf(
		"%s %d %s %d",
		RPC_JOIN,
		j.rpcPort,
		j.dumpPath,
		j.fileTransferPort,
	)
}
