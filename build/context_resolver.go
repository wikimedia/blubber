package build

import (
	"context"

	"github.com/moby/buildkit/client/llb"
	oci "github.com/opencontainers/image-spec/specs-go/v1"
)

// ContextOpt defines a number of options for use in resolving a named
// context.
type ContextOpt struct {
	Platform *oci.Platform
}

// ContextResolver returns an initialzed llb.State for a build context.
type ContextResolver func(context.Context) (*llb.State, error)

// NamedContextResolver returns an initialzed llb.State for a build context by
// name.
type NamedContextResolver func(context.Context, string, ContextOpt) (*llb.State, *oci.Image, error)
