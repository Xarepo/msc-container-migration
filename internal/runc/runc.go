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
