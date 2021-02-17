package usock_listener

import (
	"net"
	"os"

	"github.com/rs/zerolog/log"
)

type USockListener struct{}

const SOCK_ADDR = "/tmp/msc.sock"

func (usock *USockListener) Listen(callback func(buf []byte)) {
	var err error
	log.Trace().Str("SocketAddress", SOCK_ADDR).Msg("Clearing socket")
	if err := os.RemoveAll(SOCK_ADDR); err != nil {
		log.Error().Str("SocketAddress", SOCK_ADDR).Msg("Failed to clear socket")
	} else {
		log.Trace().Str("SocketAddress", SOCK_ADDR).Msg("Socket cleared")
	}

	conn, err := net.ListenUnixgram(
		"unixgram",
		&net.UnixAddr{
			Name: SOCK_ADDR,
			Net:  "unixgram",
		},
	)
	if err != nil {
		log.Error().Msgf("Failed to listen: %s", err)
	}
	log.Debug().
		Str("Address", SOCK_ADDR).
		Msg("Listening for IPC messages on socket")

	for {
		var buf [1024]byte
		nr, err := conn.Read(buf[:])
		if err != nil {
			panic(err)
		}
		data := buf[0:nr]
		log.Info().Str("Command", string(data)).Msg("Received command")

		callback(data)
	}
}
