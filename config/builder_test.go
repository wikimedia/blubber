package config_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"gitlab.wikimedia.org/repos/releng/blubber/build"
	"gitlab.wikimedia.org/repos/releng/blubber/config"
)

func TestBuilderConfigYAML(t *testing.T) {
	cfg, err := config.ReadYAMLConfig([]byte(`---
    version: v4
    base: foo
    builder:
      command: [make, -f, Makefile, test]
      requirements: [Makefile]
    variants:
      test: {}
      build:
        builder:
          command: [make]
          requirements: []
          mounts:
            - /src/main
            - from: test
              destination: /src/foo
              source: "/tmp"
          caches:
            - /var/cache/foo
            - destination: /var/cache/bar
              id: bar-cache
              access: locked`))

	if assert.NoError(t, err) {
		err := config.ExpandIncludesAndCopies(cfg, "test")
		assert.Nil(t, err)

		variant, err := config.GetVariant(cfg, "test")

		if assert.NoError(t, err) {
			assert.Equal(t, []string{"make", "-f", "Makefile", "test"}, variant.Builder.Command)
			assert.Equal(t, config.RequirementsConfig{
				{
					From:   config.LocalArtifactKeyword,
					Source: "Makefile",
				},
			}, variant.Builder.Requirements)
		}

		err = config.ExpandIncludesAndCopies(cfg, "build")
		assert.Nil(t, err)

		variant, err = config.GetVariant(cfg, "build")

		if assert.NoError(t, err) {
			assert.Equal(t, []string{"make"}, variant.Builder.Command)
			assert.Equal(t, config.RequirementsConfig{}, variant.Builder.Requirements)
			assert.Equal(
				t,
				config.MountsConfig{
					{
						From:        "local",
						Destination: "/src/main",
					},
					{
						From:        "test",
						Destination: "/src/foo",
						Source:      "/tmp",
					},
				},
				variant.Builder.Mounts,
			)
			assert.Equal(
				t,
				config.CachesConfig{
					{
						Destination: "/var/cache/foo",
					},
					{
						Destination: "/var/cache/bar",
						ID:          "bar-cache",
						Access:      "locked",
					},
				},
				variant.Builder.Caches,
			)
		}
	}
}

func TestBuilderConfigInstructions(t *testing.T) {
	t.Run("no requirements", func(t *testing.T) {
		cfg := config.BuilderConfig{Command: []string{"make", "-f", "Makefile"}}

		t.Run("PhasePreInstall", func(t *testing.T) {
			assert.Equal(t,
				[]build.Instruction{
					build.RunAllWithOptions{
						Runs: []build.Run{
							{
								"make",
								[]string{"-f", "Makefile"},
							},
						},
						Options: []build.RunOption{},
					},
				},
				cfg.InstructionsForPhase(build.PhasePreInstall),
			)
		})
	})

	t.Run("requirements", func(t *testing.T) {
		cfg := config.BuilderConfig{
			Command: []string{"make", "-f", "Makefile", "foo"},
			Requirements: config.RequirementsConfig{
				{
					From:        config.LocalArtifactKeyword,
					Source:      "Makefile",
					Destination: "",
				},
				{
					From:        config.LocalArtifactKeyword,
					Source:      "foo",
					Destination: "",
				},
				{
					From:        config.LocalArtifactKeyword,
					Source:      "bar/baz",
					Destination: "",
				},
			},
		}

		t.Run("PhasePreInstall", func(t *testing.T) {
			assert.Equal(t,
				[]build.Instruction{
					build.Copy{[]string{"Makefile", "foo"}, "./", []string{}},
					build.Copy{[]string{"bar/baz"}, "bar/", []string{}},
					build.RunAllWithOptions{
						Runs: []build.Run{
							{
								"make",
								[]string{"-f", "Makefile", "foo"},
							},
						},
						Options: []build.RunOption{},
					},
				},
				cfg.InstructionsForPhase(build.PhasePreInstall),
			)
		})
	})

	t.Run("mounts and caches", func(t *testing.T) {
		cfg := config.BuilderConfig{
			Command: []string{"make", "-f", "Makefile"},
			Mounts: config.MountsConfig{
				{
					From:        "foo",
					Destination: "/src/foo",
					Source:      "/foo/subdir",
				},
			},
			Caches: config.CachesConfig{
				{
					Destination: "/var/cache/foo",
					Access:      "locked",
				},
				{
					Destination: "/var/cache/bar",
					ID:          "bar-cache",
				},
			},
		}

		t.Run("PhasePreInstall", func(t *testing.T) {
			assert.Equal(t,
				[]build.Instruction{
					build.RunAllWithOptions{
						Runs: []build.Run{
							{
								"make",
								[]string{"-f", "Makefile"},
							},
						},
						Options: []build.RunOption{
							build.SourceMount{
								From:        "foo",
								Destination: "/src/foo",
								Source:      "/foo/subdir",
								Readonly:    true,
							},
							build.CacheMount{
								Destination: "/var/cache/foo",
								Access:      "locked",
								UID:         "$LIVES_UID",
								GID:         "$LIVES_GID",
							},
							build.CacheMount{
								Destination: "/var/cache/bar",
								ID:          "bar-cache",
								UID:         "$LIVES_UID",
								GID:         "$LIVES_GID",
							},
						},
					},
				},
				cfg.InstructionsForPhase(build.PhasePreInstall),
			)
		})
	})
}
