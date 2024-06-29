package build

import (
	"github.com/moby/buildkit/client/llb"
	"github.com/moby/buildkit/solver/result"
)

// Scanner is a function for producing an attestion results from the internal
// core llb.State and dependency llb.State of a build.Target.
type Scanner func(core llb.State, dependencies map[string]llb.State) (result.Attestation[*llb.State], error)
