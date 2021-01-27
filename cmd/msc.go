package main

import (
	"os"

	"github.com/Xarepo/msc-container-migration/internal/logger"
	"github.com/Xarepo/msc-container-migration/internal/runc"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	if err := logger.InitLogger(zerolog.DebugLevel.String()); err != nil {
		log.Error().Msg("Failed to initialize logger, exiting...")
		os.Exit(1)
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
}
