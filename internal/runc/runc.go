// Package runc provides wrapper functions for the go-runc library.
package runc

import (
	"context"

	_runc "github.com/containerd/go-runc"
	"github.com/rs/zerolog/log"
)

// Return the version numbers for runc
func Version() (_runc.Version, error) {
	r := &_runc.Runc{}
	return r.Version(context.Background())
}

func Run(id, bundle string) (int, error) {
	r := &_runc.Runc{}
	io, err := _runc.NewSTDIO()
	if err != nil {
		log.Error().Str("Error", err.Error()).Msg("Failed to create new STDIO")
	}

	log.Debug().Str("Bundle", bundle).Str("Id", id).Msg("Running container")
	return r.Run(context.Background(), id, bundle, &_runc.CreateOpts{IO: io})
}

// PreDump the container, leaving it running.
func PreDump(id, imagePath, parentPath string) {
	log.Trace().
		Str("ContainerId", id).
		Str("ImagePath", imagePath).
		Str("ParentPath", parentPath).
		Msg("Pre-dumping container")

	r := &_runc.Runc{}
	opts := _runc.CheckpointOpts{ImagePath: imagePath}
	if parentPath != "" {
		opts.ParentPath = parentPath
	}
	err := r.Checkpoint(context.Background(), id, &opts, _runc.PreDump)
	if err != nil {
		log.Error().Str("Error", err.Error()).Msg("Failed to pre-dump container")
	}
}

// Dumps the entire container state.
func Dump(id, imagePath, parentPath string, leaveRunning bool) {
	log.Trace().
		Str("ContainerId", id).
		Str("ImagePath", imagePath).
		Str("ParentPath", parentPath).
		Msg("Dumping container")

	r := &_runc.Runc{}
	opts := _runc.CheckpointOpts{ImagePath: imagePath}
	if parentPath != "" {
		opts.ParentPath = parentPath
	}

	actions := []_runc.CheckpointAction{}
	if leaveRunning {
		actions = []_runc.CheckpointAction{_runc.LeaveRunning}
	}

	err := r.Checkpoint(context.Background(), id, &opts, actions...)
	if err != nil {
		log.Error().Str("Error", err.Error()).Msg("Failed to dump container")
	}
}

func Restore(id, imagePath, bundle string) (int, error) {
	log.Trace().
		Str("ContainerId", id).
		Str("ImagePath", imagePath).
		Msg("Restoring container")

	io, err := _runc.NewSTDIO()
	if err != nil {
		log.Error().Str("Error", err.Error()).Msg("Failed to create new STDIO")
	}

	r := &_runc.Runc{}
	opts := &_runc.RestoreOpts{
		IO:             io,
		CheckpointOpts: _runc.CheckpointOpts{ImagePath: imagePath},
	}
	// TODO: No use in returning here as this is run in a goroutine
	return r.Restore(context.Background(), id, bundle, opts)
}
