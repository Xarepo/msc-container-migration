package rpc

import (
	"errors"
	"fmt"

	"github.com/Xarepo/msc-container-migration/internal/image"
	"github.com/Xarepo/msc-container-migration/internal/runner/runner_context"
	. "github.com/Xarepo/msc-container-migration/internal/runner/runner_context"
)

type Migrate struct {
	ContainerId string
	ImagePath   string
	BundlePath  string
}

func NewMigrate(containerId, imagePath, bundlePath string) *Migrate {
	return &Migrate{
		ContainerId: containerId,
		ImagePath:   imagePath,
		BundlePath:  bundlePath,
	}
}

func (migrate Migrate) Execute(ctx *RunnerContext) {
	img := image.Restore(migrate.ImagePath)
	ctx.ContainerId = migrate.ContainerId
	ctx.BundlePath = migrate.BundlePath
	ctx.LatestImage = img
	ctx.SetStatus(runner_context.Restoring)
}

func (migrate *Migrate) ParseFlags(fields []string) error {
	if len(fields) < 1 {
		return errors.New("Too few fields")
	}

	migrate.ContainerId = fields[0]
	migrate.ImagePath = fields[1]
	migrate.BundlePath = fields[2]

	return nil
}

// String returns the serialized verisoned of the struct
func (migrate Migrate) String() string {
	return fmt.Sprintf("MIGRATE %s %s %s", migrate.ContainerId, migrate.ImagePath, migrate.BundlePath)
}
