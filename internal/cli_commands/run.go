package cli_commands

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/bramvdbogaerde/go-scp"
	_ "github.com/joho/godotenv/autoload"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/ssh"

	"github.com/Xarepo/msc-container-migration/internal/runc"
)

type Run struct {
	ContainerId *string
	BundlePath  *string
}

func scpCopy(dumpName string) {
	user := os.Getenv("SCP_USER")
	password := os.Getenv("SCP_PASSWORD")
	remotePath := os.Getenv("SCP_REMOTE_PATH")

	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
	}

	client := scp.NewClient("localhost:22", config)
	err := client.Connect()
	if err != nil {
		fmt.Println("Couldn't establish a connection to the remote server ", err)
		return
	}
	defer client.Close()

	f, err := os.Open(dumpName)
	if err != nil {
		log.Error().Str("Error", err.Error()).Msg("Failed to open file")
	}
	defer f.Close()

	err = client.CopyFile(
		f,
		fmt.Sprintf("%s/%s", remotePath, filepath.Base(dumpName)),
		"0655")
	if err != nil {
		log.Error().Str("Error", err.Error()).Msg("Error while copying file ")
	}
}

// Execute the run command.
//
// The run command runs (i.e. create and starts) the container using runc. The
// container is run in a goroutine to not block execution.
// The function does not return until the container has exited.
func (cmd Run) Execute() error {
	log.Trace().
		Str("BundlePath", *cmd.BundlePath).
		Str("ContainerId", *cmd.ContainerId).
		Msg("Executing run command")
	result := make(chan int)
	go func() {
		status, err := runc.Run(*cmd.ContainerId, *cmd.BundlePath)
		if err != nil {
			log.Error().Str("Error", err.Error()).Msg("Error running container")
		} else {
			log.Info().Str("Status", string(status)).Msg("Container exited")
		}
		result <- status
	}()

	// Pre-dump loop
	ticker := time.NewTicker(2 * time.Second)
	dumpFreq := 3
	go func() {
		n := 0
		for {
			select {
			case <-ticker.C:
				parentPath := ""
				if n != 0 {
					parentPrefix := 'p'
					if (n-1)%dumpFreq == 0 {
						parentPrefix = 'd'
					}
					parentPath =
						fmt.Sprintf("dumps/%c%d", parentPrefix, n-1)
				}
				dumpName := fmt.Sprintf("dumps/p%d", n)
				if n%dumpFreq == 0 {
					dumpName = fmt.Sprintf("dumps/d%d", n)
					runc.Dump(*cmd.ContainerId, dumpName, parentPath)
				} else {
					runc.PreDump(*cmd.ContainerId, dumpName, parentPath)
				}
				scpCopy(dumpName)
				n++
			}
		}
	}()

	log.Trace().Msg("Waiting for container to exit")
	<-result // Wait for container to chanel
	return nil
}
