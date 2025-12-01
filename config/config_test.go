package config_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.wikimedia.org/repos/releng/blubber/config"
)

func TestConfigYAML(t *testing.T) {
	cfg, err := config.ReadYAMLConfig([]byte(`---
    version: v4
    variants:
      foo: {}`))

	if assert.NoError(t, err) {
		assert.Equal(t, "v4", cfg.Version)
		assert.Contains(t, cfg.Variants, "foo")
		assert.IsType(t, config.VariantConfig{}, cfg.Variants["foo"])
	}
}

func TestConfigValidation(t *testing.T) {
	t.Run("variants", func(t *testing.T) {
		t.Run("ok", func(t *testing.T) {
			err := config.Validate(config.Config{
				VersionConfig: config.VersionConfig{Version: "v4"},
				Variants: map[string]config.VariantConfig{
					"build": config.VariantConfig{},
					"foo":   config.VariantConfig{},
				},
			})

			assert.False(t, config.IsValidationError(err))
		})

		t.Run("bad", func(t *testing.T) {
			err := config.Validate(config.Config{
				VersionConfig: config.VersionConfig{Version: "v4"},
				Variants: map[string]config.VariantConfig{
					"build foo": config.VariantConfig{},
					"foo bar":   config.VariantConfig{},
				},
			})

			if assert.True(t, config.IsValidationError(err)) {
				msg := config.HumanizeValidationError(err)

				assert.Equal(t, `variants: contains a bad variant name`, msg)
			}
		})
	})
}

func TestVariantCompileables(t *testing.T) {
	t.Run("base references", func(t *testing.T) {
		req := require.New(t)

		cfg := &config.Config{
			VersionConfig: config.VersionConfig{
				Version: "v4",
			},
			Variants: map[string]config.VariantConfig{
				"repo": {
					CommonConfig: config.CommonConfig{Base: "local"},
				},
				"foo": {
					CommonConfig: config.CommonConfig{Base: "repo"},
				},
			},
		}

		req.NoError(config.ExpandIncludesAndCopies(cfg, "foo"))

		compileables, err := cfg.VariantCompileables("foo")

		req.NoError(err)
		req.Contains(compileables, "foo")
		req.Contains(compileables, "repo")
		req.Len(compileables, 2)
	})
}
