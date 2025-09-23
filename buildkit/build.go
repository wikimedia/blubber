package buildkit

import (
	"context"
	"sync"

	"github.com/containerd/containerd/platforms"
	"github.com/moby/buildkit/client/llb"
	"github.com/moby/buildkit/client/llb/sourceresolver"
	"github.com/moby/buildkit/frontend"
	"github.com/moby/buildkit/frontend/attestations/sbom"
	"github.com/moby/buildkit/frontend/dockerui"
	"github.com/moby/buildkit/frontend/gateway/client"
	"github.com/moby/buildkit/solver/pb"
	"github.com/moby/buildkit/solver/result"
	dockerspec "github.com/moby/docker-image-spec/specs-go/v1"
	oci "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/pkg/errors"

	"gitlab.wikimedia.org/repos/releng/blubber/build"
	"gitlab.wikimedia.org/repos/releng/blubber/config"
)

const (
	dockerignoreFilename = ".dockerignore"
	configLang           = "YAML"
)

// Build handles BuildKit client requests for the Blubber gateway.
//
// When performing a multi-platform build, the final exported manifest will be
// an OCI image index (aka "fat" manifest) and multiple sub manifests will be
// created for each platform that contain the actual image layers.
//
// See https://github.com/opencontainers/image-spec/blob/main/image-index.md
//
// For a single platform build, the export will be a normal single manifest
// with image layers.
//
// See https://github.com/opencontainers/image-spec/blob/main/manifest.md
func Build(ctx context.Context, c client.Client) (*client.Result, error) {
	bc, err := dockerui.NewClient(c)

	if err != nil {
		return nil, errors.Wrap(err, "failed to create dockerui client")
	}

	buildOptions, err := ParseBuildOptions(bc.BuildOpts())

	if err != nil {
		return nil, errors.Wrap(err, "failed to parse build options")
	}

	// Ensure image metadata resolution occurs via the gateway client and its
	// buildkitd bridge connection in case registry mirrors are configured.
	buildOptions.MetaResolver = c

	// Inherit the dockerui client configuration to ensure docker toolchain
	// compatibility.
	buildOptions.BuildArgs = bc.Config.BuildArgs
	buildOptions.Labels = bc.Config.Labels
	buildOptions.TargetPlatforms = bc.Config.TargetPlatforms

	// Ensure --no-cache client options work
	buildOptions.NoCache = bc.IsNoCache

	if len(bc.Config.BuildPlatforms) > 0 {
		buildOptions.BuildPlatform = bc.Config.BuildPlatforms[0]
	}

	if bc.Config.Target != "" {
		buildOptions.Variant = bc.Config.Target
	}

	buildOptions.BuildContext = func(ctx context.Context) (*llb.State, error) {
		return bc.MainContext(ctx)
	}

	cfg, err := readBlubberConfig(ctx, bc)

	if err != nil {
		if config.IsValidationError(err) {
			err = errors.New(config.HumanizeValidationError(err))
		}
		return nil, errors.Wrap(err, "failed to read blubber config")
	}

	err = config.ExpandIncludesAndCopies(cfg, buildOptions.Variant)

	if err != nil {
		if config.IsValidationError(err) {
			err = errors.New(config.HumanizeValidationError(err))
		}
		return nil, errors.Wrap(err, "failed to expand includes and copies")
	}

	var scanner sbom.Scanner

	if bc.SBOM != nil {
		scanner, err = sbom.CreateSBOMScanner(
			ctx, c, bc.SBOM.Generator,
			sourceresolver.Opt{
				ImageOpt: &sourceresolver.ResolveImageOpt{
					ResolveMode: resolveModeName(bc.ImageResolveMode),
				},
			},
			bc.SBOM.Parameters,
		)

		if err != nil {
			return nil, err
		}
	}

	scanTargets := sync.Map{}

	rb, err := bc.Build(
		ctx,
		func(ctx context.Context, platform *oci.Platform, idx int) (
			client.Reference,
			*dockerspec.DockerOCIImage,
			*dockerspec.DockerOCIImage,
			error,
		) {
			target, err := Compile(ctx, buildOptions, cfg, platform)

			if err != nil {
				return nil, nil, nil, errors.Wrap(err, "failed to compile target")
			}

			def, img, err := target.Marshal(ctx)

			if err != nil {
				return nil, nil, nil, errors.Wrap(err, "failed to marshal target")
			}

			res, err := c.Solve(ctx, client.SolveRequest{
				Definition:   def.ToPB(),
				CacheImports: bc.CacheImports,
			})

			if err != nil {
				return nil, nil, nil, errors.Wrap(err, "failed to solve")
			}

			ref, err := res.SingleRef()
			if err != nil {
				return nil, nil, nil, err
			}

			dimg := dockerspec.DockerOCIImage{
				Image: *img,
				Config: dockerspec.DockerOCIImageConfig{
					ImageConfig: img.Config,
				},
			}

			p := platforms.DefaultSpec()
			if platform != nil {
				p = *platform
			}
			scanTargets.Store(platforms.Format(platforms.Normalize(p)), target)

			return ref, &dimg, nil, nil
		},
	)

	if err != nil {
		return nil, err
	}

	if scanner != nil {
		err = rb.EachPlatform(ctx, func(ctx context.Context, id string, _ oci.Platform) error {
			v, ok := scanTargets.Load(id)
			if !ok {
				return errors.Errorf("no scan targets for %s", id)
			}

			target, ok := v.(*build.Target)
			if !ok {
				return errors.Errorf("invalid scan targets for %T", v)
			}

			att, err := target.Scan(func(core llb.State, dependencies map[string]llb.State) (result.Attestation[*llb.State], error) {
				return scanner(ctx, id, core, dependencies)
			})

			if err != nil {
				return err
			}

			attSolve, err := result.ConvertAttestation(&att, func(st *llb.State) (client.Reference, error) {
				def, err := st.Marshal(ctx)
				if err != nil {
					return nil, err
				}
				r, err := c.Solve(ctx, frontend.SolveRequest{
					Definition: def.ToPB(),
				})
				if err != nil {
					return nil, err
				}
				return r.Ref, nil
			})
			if err != nil {
				return err
			}
			rb.AddAttestation(id, *attSolve)
			return nil
		})

		if err != nil {
			return nil, err
		}
	}

	return rb.Finalize()
}

func readBlubberConfig(ctx context.Context, bc *dockerui.Client) (*config.Config, error) {
	cfgSrc, err := bc.ReadEntrypoint(ctx, configLang)
	if err != nil {
		return nil, err
	}

	cfg, err := config.ReadYAMLConfig(cfgSrc.Data)
	if err != nil {
		if config.IsValidationError(err) {
			return nil, errors.Wrapf(err, "config is invalid:\n%v", config.HumanizeValidationError(err))
		}

		return nil, errors.Wrap(err, "error reading config")
	}

	return cfg, nil
}

func resolveModeName(mode llb.ResolveMode) string {
	switch mode {
	case llb.ResolveModeForcePull:
		return pb.AttrImageResolveModeForcePull
	case llb.ResolveModePreferLocal:
		return pb.AttrImageResolveModePreferLocal
	}
	return pb.AttrImageResolveModeDefault
}
