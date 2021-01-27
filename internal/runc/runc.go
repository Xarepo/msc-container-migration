// Package runc provides wrapper functions for the go-runc library.
package runc

import (
	"context"

	_runc "github.com/containerd/go-runc"
)

// Return the version numbers for runc
func Version() (_runc.Version, error) {
	r := &_runc.Runc{}
	return r.Version(context.Background())
}
