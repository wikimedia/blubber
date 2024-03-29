package config_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"gitlab.wikimedia.org/repos/releng/blubber/build"
	"gitlab.wikimedia.org/repos/releng/blubber/config"
)

func TestVariantConfigYAML(t *testing.T) {
	cfg, err := config.ReadYAMLConfig([]byte(`---
    version: v4
    base: foo
    variants:
      build:
        copies: [local]
      production:
        copies:
          - from: build
            source: /foo/src
            destination: /foo/dst
          - from: build
            source: /bar/src
            destination: /bar/dst`))

	if assert.NoError(t, err) {
		err := config.ExpandIncludesAndCopies(cfg, "build")
		assert.Nil(t, err)

		variant, err := config.GetVariant(cfg, "build")

		if assert.NoError(t, err) {
			assert.Len(t, variant.Copies, 1)
		}

		err = config.ExpandIncludesAndCopies(cfg, "production")
		assert.Nil(t, err)

		variant, err = config.GetVariant(cfg, "production")

		if assert.NoError(t, err) {
			assert.Len(t, variant.Copies, 2)
		}
	}
}

func TestVariantDependencies(t *testing.T) {
	vcfg := config.NewVariantConfig("foo")
	vcfg.Copies = config.CopiesConfig{
		{From: "dep0"},
	}
	vcfg.Node = config.NodeConfig{
		Requirements: config.RequirementsConfig{
			{From: "dep1"},
		},
	}
	vcfg.Php = config.PhpConfig{
		Requirements: config.RequirementsConfig{
			{From: "dep2"},
		},
	}
	vcfg.Python = config.PythonConfig{
		Requirements: config.RequirementsConfig{
			{From: "dep3"},
		},
	}
	vcfg.Builder = config.BuilderConfig{
		Requirements: config.RequirementsConfig{
			{From: "dep4"},
		},
	}
	vcfg.Builders = config.BuildersConfig{
		config.BuilderConfig{
			Requirements: config.RequirementsConfig{
				{From: "dep5"},
			},
		},
	}

	assert.Equal(
		t,
		[]string{"dep0", "dep1", "dep2", "dep3", "dep4", "dep5"},
		vcfg.Dependencies(),
	)
}

func TestVariantLoops(t *testing.T) {
	cfg := config.Config{
		VersionConfig: config.VersionConfig{Version: "v4"},
		Variants: map[string]config.VariantConfig{
			"foo": config.VariantConfig{Includes: []string{"bar"}},
			"bar": config.VariantConfig{Includes: []string{"foo"}}}}

	cfgTwo := config.Config{
		VersionConfig: config.VersionConfig{Version: "v4"},
		Variants: map[string]config.VariantConfig{
			"foo": config.VariantConfig{},
			"bar": config.VariantConfig{Includes: []string{"foo"}}}}

	// Configuration that contains a loop in "Includes" should error
	err := config.ExpandIncludesAndCopies(&cfg, "bar")
	assert.Error(t, err)

	errTwo := config.ExpandIncludesAndCopies(&cfgTwo, "bar")
	assert.NoError(t, errTwo)
}

func TestVariantConfigInstructions(t *testing.T) {
	t.Run("PhasePrivileged", func(t *testing.T) {
		t.Run("with no given base", func(t *testing.T) {
			cfg := config.NewVariantConfig("foovariant")

			ins := cfg.InstructionsForPhase(build.PhasePrivileged)

			if assert.NotEmpty(t, ins) {
				assert.Equal(t, build.ScratchBase{Stage: "foovariant"}, ins[0])
			}
		})

		t.Run("with a given base", func(t *testing.T) {
			cfg := config.NewVariantConfig("foovariant")
			cfg.Base = "foobase"

			ins := cfg.InstructionsForPhase(build.PhasePrivileged)

			if assert.NotEmpty(t, ins) {
				assert.Equal(t, build.Base{Image: "foobase", Stage: "foovariant"}, ins[0])
			}
		})
	})

	t.Run("PhaseInstall", func(t *testing.T) {
		t.Run("without copies", func(t *testing.T) {
			cfg := config.VariantConfig{}

			assert.Empty(t, cfg.InstructionsForPhase(build.PhaseInstall))
		})

		t.Run("with copies", func(t *testing.T) {
			cfg := config.VariantConfig{
				CommonConfig: config.CommonConfig{
					Base:  "foobase",
					Lives: config.LivesConfig{UserConfig: config.UserConfig{UID: 123, GID: 223}},
				},
				Copies: config.CopiesConfig{
					{From: "local"},
					{From: "build", Source: "/foo/src", Destination: "/foo/dst"},
				},
			}

			assert.Equal(t,
				[]build.Instruction{
					build.CopyAs{"$LIVES_UID", "$LIVES_GID", build.Copy{[]string{"."}, "."}},
					build.CopyAs{"$LIVES_UID", "$LIVES_GID", build.CopyFrom{"build", build.Copy{[]string{"/foo/src"}, "/foo/dst"}}},
				},
				cfg.InstructionsForPhase(build.PhaseInstall),
			)
		})
	})

	t.Run("PhasePostInstall", func(t *testing.T) {
		t.Run("with entrypoint", func(t *testing.T) {
			cfg := config.VariantConfig{
				CommonConfig: config.CommonConfig{
					EntryPoint: []string{"/foo", "bar"},
				},
			}

			assert.Equal(t,
				[]build.Instruction{
					build.EntryPoint{[]string{"/foo", "bar"}},
				},
				cfg.InstructionsForPhase(build.PhasePostInstall),
			)
		})

		t.Run("without Runs.Insecurely", func(t *testing.T) {
			cfg := config.VariantConfig{
				CommonConfig: config.CommonConfig{
					Base: "foobase",
					Lives: config.LivesConfig{
						UserConfig: config.UserConfig{
							As: "foouser",
						},
					},
					Runs: config.RunsConfig{
						Insecurely: config.Flag{True: false},
						UserConfig: config.UserConfig{
							As:  "baruser",
							UID: 1000,
						},
					},
					EntryPoint: []string{"/foo", "bar"},
				},
			}

			assert.Equal(t,
				[]build.Instruction{
					build.User{UID: "$RUNS_UID"},
					build.Env{map[string]string{"HOME": "/home/$RUNS_AS"}},
					build.EntryPoint{[]string{"/foo", "bar"}},
				},
				cfg.InstructionsForPhase(build.PhasePostInstall),
			)
		})

		t.Run("with Runs.Insecurely", func(t *testing.T) {
			cfg := config.VariantConfig{
				CommonConfig: config.CommonConfig{
					Lives: config.LivesConfig{
						UserConfig: config.UserConfig{
							As: "foouser",
						},
					},
					Runs: config.RunsConfig{
						Insecurely: config.Flag{True: true},
						UserConfig: config.UserConfig{
							As: "baruser",
						},
					},
					EntryPoint: []string{"/foo", "bar"},
				},
			}

			assert.Equal(t,
				[]build.Instruction{
					build.EntryPoint{[]string{"/foo", "bar"}},
				},
				cfg.InstructionsForPhase(build.PhasePostInstall),
			)
		})
	})
}

func TestVariantConfigValidation(t *testing.T) {
	t.Run("includes", func(t *testing.T) {
		t.Run("ok", func(t *testing.T) {
			_, err := config.ReadYAMLConfig([]byte(`---
        version: v4
        variants:
          build: {}
          foo: { includes: [build] }`))

			assert.False(t, config.IsValidationError(err))
		})

		t.Run("optional", func(t *testing.T) {
			_, err := config.ReadYAMLConfig([]byte(`---
        version: v4
        variants:
          build: {}
          foo: {}`))

			assert.False(t, config.IsValidationError(err))
		})

		t.Run("bad", func(t *testing.T) {
			_, err := config.ReadYAMLConfig([]byte(`---
        version: v4
        variants:
          build: {}
          foo: { includes: [build, foobuild, foo_build] }`))

			if assert.True(t, config.IsValidationError(err)) {
				msg := config.HumanizeValidationError(err)

				assert.Equal(t, strings.Join([]string{
					`includes[1]: references an unknown variant "foobuild"`,
					`includes[2]: references an unknown variant "foo_build"`,
				}, "\n"), msg)
			}
		})
	})

	t.Run("copies", func(t *testing.T) {

		t.Run("should not contain duplicates", func(t *testing.T) {
			_, err := config.ReadYAMLConfig([]byte(`---
        version: v4
        variants:
          build: {}
          foo: { copies: [foo, bar, foo] }`))

			if assert.True(t, config.IsValidationError(err)) {
				msg := config.HumanizeValidationError(err)

				assert.Equal(t, `copies: cannot contain duplicates`, msg)
			}
		})
	})
}
