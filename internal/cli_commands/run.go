package cli_commands

import (
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
	log.Trace().Str("BundlePath", *cmd.BundlePath).Str("ContainerId", *cmd.ContainerId).Msg("Executing run command")
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
	ticker := time.NewTicker(500 * time.Millisecond)
	go func() {
		for {
			select {
			case <-ticker.C:
				log.Trace().Str("ContainerId", *cmd.ContainerId).Msg("Pre-dumping")
			}
		}
	}()

	log.Trace().Msg("Waiting for container to exit")
	<-result // Wait for container to chanel
	return nil
}
