package config_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"gitlab.wikimedia.org/repos/releng/blubber/build"
	"gitlab.wikimedia.org/repos/releng/blubber/config"
)

func TestArgumentsConfigYAML(t *testing.T) {
	req := require.New(t)

	cfg, err := config.ReadYAMLConfig([]byte(`---
    version: v4
    base: foo
    arguments:
      FOO: foo
      BAR: bar
    variants:
      foo:
        arguments:
          BAR: baz`))

	req.Nil(err)

	err = config.ExpandIncludesAndCopies(cfg, "foo")
	variant, err := config.GetVariant(cfg, "foo")

	req.Equal(
		config.ArgumentsConfig{
			"FOO": "foo",
			"BAR": "baz",
		},
		variant.Arguments,
	)
}

func TestArgumentsConfigInstructions(t *testing.T) {
	req := require.New(t)
	cfg := config.ArgumentsConfig{
		"FOO": "foo",
		"BAR": "bar",
	}

	t.Run("PhasePrivileged", func(t *testing.T) {
		req.Equal(
			[]build.Instruction{
				build.StringArg{Name: "BAR", Default: "bar"},
				build.StringArg{Name: "FOO", Default: "foo"},
			},
			cfg.InstructionsForPhase(build.PhasePrivileged),
		)
	})

	t.Run("PhasePrivilegeDropped", func(t *testing.T) {
		req.Empty(cfg.InstructionsForPhase(build.PhasePrivilegeDropped))
	})

	t.Run("PhasePreInstall", func(t *testing.T) {
		req.Empty(cfg.InstructionsForPhase(build.PhasePreInstall))
	})

	t.Run("PhaseInstall", func(t *testing.T) {
		req.Empty(cfg.InstructionsForPhase(build.PhaseInstall))
	})

	t.Run("PhasePostInstall", func(t *testing.T) {
		req.Empty(cfg.InstructionsForPhase(build.PhasePostInstall))
	})
}
