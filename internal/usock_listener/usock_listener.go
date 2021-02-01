package usock_listener

import (
	"net"
	"os"

	"github.com/rs/zerolog/log"
)

type USockListener struct {
	SockAddr string
}

func (usock USockListener) Listen(callback func(buf []byte)) {
	log.Trace().Str("SocketAddress", usock.SockAddr).Msg("Clearing socket")
	if err := os.RemoveAll(usock.SockAddr); err != nil {
		log.Error().Str("SocketAddress", usock.SockAddr).Msg("Failed to clear socket")
	} else {
		log.Trace().Str("SocketAddress", usock.SockAddr).Msg("Socket cleared")
	}

	l, err := net.Listen("unix", usock.SockAddr)
	if err != nil {
		log.Error().Msgf("Failed to listen: %s", err)
	}
	log.Info().Str("Address", usock.SockAddr).Msg("Listening on socket")
	defer l.Close()

	for {
		c, err := l.Accept()
		if err == nil {
			log.Trace().
				Str("Address", usock.SockAddr).
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
