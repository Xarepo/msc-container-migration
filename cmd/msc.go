package main

import (
	"os"

	"github.com/rs/zerolog/log"

	"github.com/Xarepo/msc-container-migration/internal/cli"
	"github.com/Xarepo/msc-container-migration/internal/env"
	"github.com/Xarepo/msc-container-migration/internal/logger"
	"github.com/Xarepo/msc-container-migration/internal/runc"
)

func main() {
	// Use log level directly here, before env.Init(), to get nice logs.
	logLevel := os.Getenv("LOG_LEVEL")
	if err := logger.InitLogger(logLevel); err != nil {
		log.Fatal().Msg("Failed to initialize logger, exiting...")
	}

	err := env.Init()
	if err != nil {
		log.Fatal().Str("Error", err.Error()).Msg("Failed to initialize environment")
	}

	// Print version numbers
	v, err := runc.Version()
	if err != nil {
		log.Error().Str("Error", err.Error()).Msg("Failed to retrieve runc version")
		os.Exit(1)
	}
	log.Debug().Str("Runc version", v.Runc).Send()
	log.Debug().Str("Commit", v.Commit).Send()
	log.Debug().Str("Runc-spec version", v.Spec).Send()

	cmd := cli.Parse()
	cmd.Execute()
}
