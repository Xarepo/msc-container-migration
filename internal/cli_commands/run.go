package cli_commands

import (
	"fmt"
	"time"

	"github.com/Xarepo/msc-container-migration/internal/runc"
	"github.com/rs/zerolog/log"
)

type Run struct {
	ContainerId *string
	BundlePath  *string
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
				if n%dumpFreq == 0 {
					runc.Dump(*cmd.ContainerId, fmt.Sprintf("dumps/d%d", n), parentPath)
				} else {
					runc.PreDump(*cmd.ContainerId, fmt.Sprintf("dumps/p%d", n),
						parentPath)
				}
				n++
			}
		}
	}()

	log.Trace().Msg("Waiting for container to exit")
	<-result // Wait for container to chanel
	return nil
}
