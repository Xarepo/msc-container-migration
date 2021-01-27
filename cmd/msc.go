package main

import (
	"context"
	"fmt"
	"os"

	"github.com/Xarepo/msc-container-migration/internal/logger"
	runc "github.com/containerd/go-runc"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	cfg := &runc.Runc{}

	v, err := cfg.Version(context.Background())
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Println(v)

	if err := logger.InitLogger(zerolog.DebugLevel.String()); err != nil {
		log.Error().Msg("Failed to initialize logger, exiting...")
		os.Exit(1)
	}

	log.Debug().Str("Runc version", v.Runc).Send()
	log.Debug().Str("Commit", v.Commit).Send()
	log.Debug().Str("Runc-spec version", v.Spec).Send()
}
