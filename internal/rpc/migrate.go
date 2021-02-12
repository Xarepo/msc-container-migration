package rpc

import (
	"errors"
	"fmt"

	"github.com/Xarepo/msc-container-migration/internal/image"
	"github.com/Xarepo/msc-container-migration/internal/runner/runner_context"
	. "github.com/Xarepo/msc-container-migration/internal/runner/runner_context"
)

type Migrate struct {
	containerId string
	imagePath   string
	bundlePath  string
}

func NewMigrate(containerId, imagePath, bundlePath string) *Migrate {
	return &Migrate{
		containerId: containerId,
		imagePath:   imagePath,
		bundlePath:  bundlePath,
	}
}

func (migrate Migrate) Execute(ctx *RunnerContext, remoteAddr string) {
	img := image.Restore(migrate.imagePath)
	ctx.ContainerId = migrate.containerId
	ctx.BundlePath = migrate.bundlePath
	ctx.LatestImage = img
	ctx.SetStatus(runner_context.Restoring)
}

func (migrate *Migrate) ParseFlags(fields []string) error {
	if len(fields) < 1 {
		return errors.New("Too few fields")
	}

	migrate.containerId = fields[0]
	migrate.imagePath = fields[1]
	migrate.bundlePath = fields[2]

	return nil
}

func (migrate Migrate) String() string {
	return fmt.Sprintf(
		"MIGRATE %s %s %s",
		migrate.containerId,
		migrate.imagePath,
		migrate.bundlePath,
	)
}
