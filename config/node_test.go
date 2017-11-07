package config_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"phabricator.wikimedia.org/source/blubber/build"
	"phabricator.wikimedia.org/source/blubber/config"
)

func TestNodeConfig(t *testing.T) {
	cfg, err := config.ReadConfig([]byte(`---
    base: foo
    node:
      dependencies: true
      env: foo
    variants:
      build:
        node:
          dependencies: false
          env: bar`))

	assert.Nil(t, err)

	assert.Equal(t, true, cfg.Node.Dependencies.True)
	assert.Equal(t, "foo", cfg.Node.Env)

	variant, err := config.ExpandVariant(cfg, "build")

	assert.Equal(t, false, variant.Node.Dependencies.True)
	assert.Equal(t, "bar", variant.Node.Env)
}

func TestNodeConfigInstructionsNoDependencies(t *testing.T) {
	cfg := config.NodeConfig{Dependencies: config.Flag{True: false}}

	t.Run("PhasePrivileged", func(t *testing.T) {
		assert.Empty(t, cfg.InstructionsForPhase(build.PhasePrivileged))
	})

	t.Run("PhasePrivilegeDropped", func(t *testing.T) {
		assert.Empty(t, cfg.InstructionsForPhase(build.PhasePrivilegeDropped))
	})

	t.Run("PhasePreInstall", func(t *testing.T) {
		assert.Empty(t, cfg.InstructionsForPhase(build.PhasePreInstall))
	})

	t.Run("PhasePostInstall", func(t *testing.T) {
		assert.Empty(t, cfg.InstructionsForPhase(build.PhasePostInstall))
	})
}

func TestNodeConfigInstructionsNonProduction(t *testing.T) {
	cfg := config.NodeConfig{Dependencies: config.Flag{True: true}, Env: "foo"}

	t.Run("PhasePrivileged", func(t *testing.T) {
		assert.Empty(t, cfg.InstructionsForPhase(build.PhasePrivileged))
	})

	t.Run("PhasePrivilegeDropped", func(t *testing.T) {
		assert.Empty(t, cfg.InstructionsForPhase(build.PhasePrivilegeDropped))
	})

	t.Run("PhasePreInstall", func(t *testing.T) {
		assert.Equal(t,
			[]build.Instruction{
				build.Copy{[]string{"package.json"}, "/opt/lib"},
				build.RunAll{[]build.Run{
					{"cd", []string{"/opt/lib"}},
					{"npm install", []string{}},
				}},
			},
			cfg.InstructionsForPhase(build.PhasePreInstall),
		)
	})

	t.Run("PhasePostInstall", func(t *testing.T) {
		assert.Equal(t,
			[]build.Instruction{
				build.Env{map[string]string{
					"NODE_ENV":  "foo",
					"NODE_PATH": "/opt/lib/node_modules",
					"PATH":      "/opt/lib/node_modules/.bin:${PATH}",
				}},
			},
			cfg.InstructionsForPhase(build.PhasePostInstall),
		)
	})
}

func TestNodeConfigInstructionsProduction(t *testing.T) {
	cfg := config.NodeConfig{Dependencies: config.Flag{True: true}, Env: "production"}

	t.Run("PhasePrivileged", func(t *testing.T) {
		assert.Empty(t, cfg.InstructionsForPhase(build.PhasePrivileged))
	})

	t.Run("PhasePrivilegeDropped", func(t *testing.T) {
		assert.Empty(t, cfg.InstructionsForPhase(build.PhasePrivilegeDropped))
	})

	t.Run("PhasePreInstall", func(t *testing.T) {
		assert.Equal(t,
			[]build.Instruction{
				build.Copy{[]string{"package.json"}, "/opt/lib"},
				build.RunAll{[]build.Run{
					{"cd", []string{"/opt/lib"}},
					{"npm install", []string{"--production"}},
					{"npm dedupe", []string{}},
				}},
			},
			cfg.InstructionsForPhase(build.PhasePreInstall),
		)
	})

	t.Run("PhasePostInstall", func(t *testing.T) {
		assert.Equal(t,
			[]build.Instruction{
				build.Env{map[string]string{
					"NODE_ENV":  "production",
					"NODE_PATH": "/opt/lib/node_modules",
					"PATH":      "/opt/lib/node_modules/.bin:${PATH}",
				}},
			},
			cfg.InstructionsForPhase(build.PhasePostInstall),
		)
	})
}

func TestNodeConfigInstructionsEnvironmentOnly(t *testing.T) {
	cfg := config.NodeConfig{Env: "production"}

	t.Run("PhasePrivileged", func(t *testing.T) {
		assert.Empty(t, cfg.InstructionsForPhase(build.PhasePrivileged))
	})

	t.Run("PhasePrivilegeDropped", func(t *testing.T) {
		assert.Empty(t, cfg.InstructionsForPhase(build.PhasePrivilegeDropped))
	})

	t.Run("PhasePreInstall", func(t *testing.T) {
		assert.Empty(t, cfg.InstructionsForPhase(build.PhasePreInstall))
	})

	t.Run("PhasePostInstall", func(t *testing.T) {
		assert.Equal(t,
			[]build.Instruction{
				build.Env{map[string]string{
					"NODE_ENV":  "production",
					"NODE_PATH": "/opt/lib/node_modules",
					"PATH":      "/opt/lib/node_modules/.bin:${PATH}",
				}},
			},
			cfg.InstructionsForPhase(build.PhasePostInstall),
		)
	})
}

func TestNodeConfigValidation(t *testing.T) {
	t.Run("env", func(t *testing.T) {
		t.Run("ok", func(t *testing.T) {
			_, err := config.ReadConfig([]byte(`---
        node:
          env: production`))

			assert.False(t, config.IsValidationError(err))
		})

		t.Run("optional", func(t *testing.T) {
			_, err := config.ReadConfig([]byte(`---
        node: {}`))

			assert.False(t, config.IsValidationError(err))
		})

		t.Run("bad", func(t *testing.T) {
			_, err := config.ReadConfig([]byte(`---
        node:
          env: foo bar`))

			if assert.True(t, config.IsValidationError(err)) {
				msg := config.HumanizeValidationError(err)

				assert.Equal(t, `env: "foo bar" is not a valid Node environment name`, msg)
			}
		})
	})
}