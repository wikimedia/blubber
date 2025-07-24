package config_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.wikimedia.org/repos/releng/blubber/build"
	"gitlab.wikimedia.org/repos/releng/blubber/config"
)

func TestBuilderConfigYAML(t *testing.T) {
	t.Run("command", func(t *testing.T) {
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
            command: "make install"
            requirements: []
            mounts:
              - local
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
				assert.Equal(t, config.BuilderCommand([]string{"make", "-f", "Makefile", "test"}), variant.Builder.Command)
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
				assert.Equal(t, config.BuilderCommand([]string{"make install"}), variant.Builder.Command)
				assert.Equal(t, config.RequirementsConfig{}, variant.Builder.Requirements)
				assert.Equal(
					t,
					config.MountsConfig{
						{
							From: "local",
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
	})

	t.Run("script command", func(t *testing.T) {
		req := require.New(t)

		cfg, err := config.ReadYAMLConfig([]byte(`---
      version: v4
      base: an.example/base/image
      variants:
        build:
          builder:
            command: "foo"
`))
		req.NoError(err)

		err = config.ExpandIncludesAndCopies(cfg, "build")
		req.NoError(err)

		variant, err := config.GetVariant(cfg, "build")
		req.NoError(err)

		req.Equal(config.BuilderCommand([]string{"foo"}), variant.Builder.Command)
	})

	t.Run("script", func(t *testing.T) {
		req := require.New(t)

		cfg, err := config.ReadYAMLConfig([]byte(`---
      version: v4
      base: an.example/base/image
      variants:
        build:
          builder:
            script: |
              #!/bin/bash
              something_with foo
            requirements: [foo]
`))
		req.NoError(err)

		err = config.ExpandIncludesAndCopies(cfg, "build")
		req.NoError(err)

		variant, err := config.GetVariant(cfg, "build")
		req.NoError(err)

		req.Equal("#!/bin/bash\nsomething_with foo\n", variant.Builder.Script)
	})

	t.Run("script and command are mutually exclusive", func(t *testing.T) {
		req := require.New(t)

		_, err := config.ReadYAMLConfig([]byte(`---
      version: v4
      base: an.example/base/image
      variants:
        build:
          builder:
            command: "foo"
            script: |
              #!/bin/bash
              something_with foo
            requirements: [foo]
`))
		req.Error(err)
		msg := config.HumanizeValidationError(err)
		req.Contains(msg, `command: is not allowed if any of field(s) "script" is declared/included`, msg)
		req.Contains(msg, `script: is not allowed if any of field(s) "command" is declared/included`, msg)
	})

	t.Run("validation applies to builders.[].custom", func(t *testing.T) {
		req := require.New(t)

		_, err := config.ReadYAMLConfig([]byte(`---
      version: v4
      base: an.example/base/image
      variants:
        build:
          builders:
            - custom:
                command: "foo"
                script: |
                  #!/bin/bash
                  something_with foo
                requirements: [foo]
`))
		req.Error(err)
		msg := config.HumanizeValidationError(err)
		req.Contains(msg, `command: is not allowed if any of field(s) "script" is declared/included`, msg)
		req.Contains(msg, `script: is not allowed if any of field(s) "command" is declared/included`, msg)
	})
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

	t.Run("script", func(t *testing.T) {
		cfg := config.BuilderConfig{
			Script: "foo\nbar",
		}

		t.Run("PhasePreInstall", func(t *testing.T) {
			assert.Equal(t,
				[]build.Instruction{
					build.RunScript{
						Script:  []byte("foo\nbar"),
						Options: []build.RunOption{},
					},
				},
				cfg.InstructionsForPhase(build.PhasePreInstall),
			)
		})
	})

	t.Run("no command and no script", func(t *testing.T) {
		cfg := config.BuilderConfig{}

		t.Run("PhasePreInstall", func(t *testing.T) {
			assert.Equal(t, []build.Instruction{}, cfg.InstructionsForPhase(build.PhasePreInstall))
		})
	})
}
