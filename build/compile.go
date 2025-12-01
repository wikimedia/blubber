package build

import (
	"context"
	"unicode/utf8"

	"github.com/moby/buildkit/client/llb"
	oci "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/pkg/errors"
	"golang.org/x/sync/singleflight"
)

// TargetCompileable defines a [PhaseCompileable] with a name and base ref.
type TargetCompileable interface {
	PhaseCompileable
	BaseRef() string
	Name() string
}

// Compile takes a map of [BaseCompileable]s and build options and returns a
// compiled [Target] and its dependency [Target]s.
func Compile(
	ctx context.Context,
	compileables map[string]TargetCompileable,
	bo Options,
	platform oci.Platform,
) (*Result, error) {
	maxNameLength := 0
	for name := range compileables {
		nameLength := utf8.RuneCountInString(name)
		if nameLength > maxNameLength {
			maxNameLength = nameLength
		}
	}
	bo.NameLogWidth = maxNameLength

	// Ensure each target is only built once
	once := new(singleflight.Group)
	dependencies := []*Target{}

	compile := func(name string, opt ContextOpt) (*Target, error) {
		res, err, _ := once.Do(name, func() (any, error) {
			compileable, ok := compileables[name]
			if !ok {
				return nil, nil
			}

			target := NewTarget(compileable.Name(), compileable.BaseRef(), opt.Platform, bo)
			err := target.Initialize(ctx)
			if err != nil {
				return nil, err
			}

			for _, phase := range Phases() {
				for _, instruction := range compileable.InstructionsForPhase(phase) {
					err := instruction.Compile(target)
					if err != nil {
						return nil, errors.Wrap(err, "failed to compile instruction")
					}
				}
			}

			return target, nil
		})

		if err != nil {
			return nil, err
		}

		if target, ok := res.(*Target); ok {
			if name != bo.Variant {
				dependencies = append(dependencies, target)
			}
			return target, nil
		}

		return nil, nil
	}

	// Wrap the client provided [Options.NamedContext] function and if no client
	// context is matched for the given name, try one of the
	// [TargetCompileable]s provided.
	clientContext := bo.NamedContext
	bo.NamedContext = func(ctx context.Context, name string, opt ContextOpt) (*llb.State, *oci.Image, error) {
		// Try client provided contexts first
		state, img, err := clientContext(ctx, name, opt)
		if err != nil {
			return nil, nil, err
		}

		if state != nil {
			return state, img, nil
		}

		// No client context. Compile a variant
		target, err := compile(name, opt)
		if err != nil {
			return nil, nil, err
		}

		if target == nil {
			return nil, nil, nil
		}

		targetState := target.State()

		return &targetState, target.Image.OCI(), nil
	}

	finalTarget, err := compile(bo.Variant, ContextOpt{Platform: &platform})
	if err != nil {
		return nil, err
	}

	return &Result{Target: finalTarget, Dependencies: dependencies}, nil
}
