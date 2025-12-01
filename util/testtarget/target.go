package testtarget

import (
	"context"
	"testing"

	"github.com/moby/buildkit/client/llb"
	oci "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/stretchr/testify/require"
	"golang.org/x/sync/singleflight"

	"gitlab.wikimedia.org/repos/releng/blubber/build"
	"gitlab.wikimedia.org/repos/releng/blubber/util/testmetaresolver"
	"gitlab.wikimedia.org/repos/releng/llbtest/llbtest"
)

// TargetFn represents a [build.Target] callback. See [Setup]
type TargetFn func(*build.Target)

// NewTarget returns a boilerlplate [build.Target] for use in tests.
func NewTarget(name string) *build.Target {
	targets := NewTargets(name)
	return targets[0]
}

// NewTargets returns a boilerlplate [build.TargetGroup] for use in tests.
func NewTargets(names ...string) []*build.Target {
	return NewTargetsWithBaseImage(
		names,
		oci.Image{
			Config: oci.ImageConfig{
				User:       "root",
				Env:        []string{},
				WorkingDir: "/srv/app",
			},
		},
	)
}

// NewTargetsWithBaseImage returns a boilerlplate set of [build.Target] for
// use in tests and uses the given oci.Image as the resolved base image.
func NewTargetsWithBaseImage(names []string, baseImage oci.Image) []*build.Target {
	targets := make([]*build.Target, len(names))

	for i, name := range names {
		baseImageRef := "testtarget.test/base/" + name

		options := build.NewOptions()
		options.MetaResolver = testmetaresolver.New(
			baseImageRef,
			baseImage,
		)

		targets[i] = build.NewTarget(
			name,
			baseImageRef,
			nil,
			*options,
		)
	}

	return targets
}

// Setup is a test helper that takes a number of given [build.Target]s and
// sets up a [build.NamedContextResolver] that compiles each target using the
// given function that matches the target's index. Returns an unmarshaled
// [oci.Image] and [Assertions] for the last target for additional assertions.
func Setup(
	t *testing.T,
	targets []*build.Target,
	fn ...TargetFn) (*oci.Image, *Assertions) {

	t.Helper()

	ctx := context.TODO()
	req := require.New(t)

	req.Positive(len(targets))

	once := new(singleflight.Group)
	targetCompiler := func(name string) *build.Target {
		target, _, _ := once.Do(name, func() (any, error) {
			for i, target := range targets {
				if target.Name == name {
					req.NoError(target.Initialize(ctx))

					if i < len(fn) {
						fn[i](target)
					}

					return target, nil
				}
			}

			return nil, nil
		})

		if bt, ok := target.(*build.Target); ok {
			return bt
		}

		return nil
	}

	// Set up the target compiler function as a named context resolver so
	// references to other targets resolve to the compiled version
	for _, target := range targets {
		nc := target.Options.NamedContext
		target.Options.NamedContext = func(ctx context.Context, name string, opt build.ContextOpt) (*llb.State, *oci.Image, error) {
			state, img, err := nc(ctx, name, opt)
			if err != nil {
				return nil, nil, err
			}
			if state != nil {
				return state, img, nil
			}

			target := targetCompiler(name)
			if target == nil {
				return nil, nil, nil
			}

			targetState := target.State()
			return &targetState, target.Image.OCI(), nil
		}
	}

	// Finally, call all target compilers directly
	for _, target := range targets {
		targetCompiler(target.Name)
	}

	def, image, err := targets[len(targets)-1].Marshal(ctx)
	req.NoError(err)

	return image, &Assertions{
		LLBAssertions: llbtest.New(t, def),
		Assertions:    require.New(t),
		t:             t,
		target:        targets[len(targets)-1],
	}
}

// Compile is a simplified version of [Setup] without the callback that simply
// compiles the given [build.Instruction] to the first of the given targets.
func Compile(
	t *testing.T,
	targets []*build.Target,
	ins build.Instruction) (*oci.Image, *Assertions) {

	t.Helper()

	return Setup(
		t,
		targets,
		func(target *build.Target) {
			require.NoError(t, ins.Compile(target))
		},
	)
}
