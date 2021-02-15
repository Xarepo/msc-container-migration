package usock_listener

import (
	"net"
	"os"

	"github.com/rs/zerolog/log"
)

type USockListener struct{}

const SOCK_ADDR = "/tmp/msc.sock"

func (usock USockListener) Listen(callback func(buf []byte)) {
	log.Trace().Str("SocketAddress", SOCK_ADDR).Msg("Clearing socket")
	if err := os.RemoveAll(SOCK_ADDR); err != nil {
		log.Error().Str("SocketAddress", SOCK_ADDR).Msg("Failed to clear socket")
	} else {
		log.Trace().Str("SocketAddress", SOCK_ADDR).Msg("Socket cleared")
	}

	l, err := net.Listen("unix", SOCK_ADDR)
	if err != nil {
		log.Error().Msgf("Failed to listen: %s", err)
	}
	log.Debug().
		Str("Address", SOCK_ADDR).
		Msg("Listening for IPC messages on socket")
	defer l.Close()

	for {
		c, err := l.Accept()
		if err == nil {
			log.Trace().
				Str("Address", SOCK_ADDR).
				Msg("Received message from socket")
			buf := make([]byte, 512)
			nr, err := c.Read(buf)
			if err != nil {
				return
			}

			data := buf[0:nr]
			log.Info().Str("Command", string(data)).Msg("Received command")

			callback(data)
		} else {
			log.Error().Msgf("Listener failed to accept: %s", err)
		}
	}
}
