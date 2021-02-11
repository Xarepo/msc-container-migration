package udp_listener

import (
	"fmt"
	"net"
	"os"

	"github.com/rs/zerolog/log"
)

type UDPListener struct {
}

func (udp_listener UDPListener) Listen(port int, callback func(buf []byte, remoteAddr string)) {
	rpcIP := os.Getenv("RPC_IP")
	if rpcIP == "" {
		log.Warn().Msg("Environment variable RPC_IP not set, defaulting to localhost.")
		rpcIP = "localhost"
	}

	p := make([]byte, 1024)
	addr := net.UDPAddr{
		Port: port,
		IP:   net.ParseIP(rpcIP),
	}
	conn, err := net.ListenUDP("udp", &addr)
	if err != nil {
		log.Panic().Str("Error", err.Error()).Msg("Failed to listen to UDP")
	}
	for {
		nr, remoteaddr, err := conn.ReadFromUDP(p)
		buf := p[0:nr]
		log.Trace().
			Str("RemoteAddr", remoteaddr.String()).
			Msg("Received UDP message")
		if err != nil {
			fmt.Printf("Some error  %v", err)
			log.Error().Str("Error", err.Error()).Msg("Failed to read from UDP")
			continue
		}
		callback(buf, remoteaddr.String())
	}
}
