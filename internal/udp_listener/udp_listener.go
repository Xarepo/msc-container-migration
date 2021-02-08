package udp_listener

import (
	"fmt"
	"net"
	"os"
	"strconv"

	"github.com/rs/zerolog/log"
)

type UDPListener struct {
}

func (udp_listener UDPListener) Listen(callback func(buf []byte)) {
	rpcIP := os.Getenv("RPC_IP")
	if rpcIP == "" {
		log.Warn().Msg("Environment variable RPC_IP not set, defaulting to localhost.")
		rpcIP = "localhost"
	}
	var rpcPort int
	var err error
	if os.Getenv("RPC_PORT") == "" {
		log.Warn().Msg("Environment variable RPC_PORT not set, defaulting to 1234.")
		rpcPort = 1234
	} else {
		rpcPort, err = strconv.Atoi(os.Getenv("RPC_PORT"))
		if err != nil {
			log.Panic().Str("Error", err.Error()).Msg("Failed to parse RPC port")
		}
	}

	p := make([]byte, 1024)
	addr := net.UDPAddr{
		Port: rpcPort,
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
		callback(buf)
	}
}
