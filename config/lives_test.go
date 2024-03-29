package config_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"gitlab.wikimedia.org/repos/releng/blubber/build"
	"gitlab.wikimedia.org/repos/releng/blubber/config"
)

func TestLivesConfigYAML(t *testing.T) {
	cfg, err := config.ReadYAMLConfig([]byte(`---
    version: v4
    base: foo
    lives:
      in: /some/directory
      as: foouser
      uid: 123
      gid: 223
    variants:
      development: {}`))

	if assert.NoError(t, err) {
		assert.Equal(t, "/some/directory", cfg.Lives.In)
		assert.Equal(t, "foouser", cfg.Lives.As)
		assert.Equal(t, uint(123), cfg.Lives.UID)
		assert.Equal(t, uint(223), cfg.Lives.GID)

		err = config.ExpandIncludesAndCopies(cfg, "development")
		assert.Nil(t, err)

		variant, err := config.GetVariant(cfg, "development")

		if assert.NoError(t, err) {
			assert.Equal(t, "/some/directory", variant.Lives.In)
			assert.Equal(t, "foouser", variant.Lives.As)
			assert.Equal(t, uint(123), variant.Lives.UID)
			assert.Equal(t, uint(223), variant.Lives.GID)
		}
	}
}

func TestLivesConfigDefaults(t *testing.T) {
	cfg, err := config.ReadYAMLConfig([]byte(`---
    version: v4
    base: foo`))

	if assert.NoError(t, err) {
		assert.Equal(t, "somebody", cfg.Lives.As)
		assert.Equal(t, uint(65533), cfg.Lives.UID)
		assert.Equal(t, uint(65533), cfg.Lives.GID)
	}
}

func TestLivesConfigInstructions(t *testing.T) {
	cfg := config.LivesConfig{
		In: "/some/directory",
		UserConfig: config.UserConfig{
			As:  "foouser",
			UID: 123,
			GID: 223,
		},
	}

	t.Run("PhasePrivileged", func(t *testing.T) {
		assert.Equal(t,
			[]build.Instruction{
				build.NewStringArg("LIVES_AS", "foouser"),
				build.NewUintArg("LIVES_UID", 123),
				build.NewUintArg("LIVES_GID", 223),
				build.RunAll{[]build.Run{
					{
						"(getent group %s || groupadd -o -g %s -r %s)",
						[]string{"$LIVES_GID", "$LIVES_GID", "$LIVES_AS"},
					},
					{
						"(getent passwd %s || useradd -l -o -m -d %s -r -g %s -u %s %s)",
						[]string{"$LIVES_UID", "/home/$LIVES_AS", "$LIVES_GID", "$LIVES_UID", "$LIVES_AS"},
					},
					{"mkdir -p", []string{"/some/directory"}},
					{"chown %s:%s", []string{"$LIVES_UID", "$LIVES_GID", "/some/directory"}},
					{"mkdir -p", []string{"/opt/lib"}},
					{"chown %s:%s", []string{"$LIVES_UID", "$LIVES_GID", "/opt/lib"}},
				}},
			},
			cfg.InstructionsForPhase(build.PhasePrivileged),
		)
	})

	t.Run("PhasePrivilegeDropped", func(t *testing.T) {
		assert.Equal(t,
			[]build.Instruction{
				build.WorkingDirectory{"/some/directory"},
			},
			cfg.InstructionsForPhase(build.PhasePrivilegeDropped),
		)
	})

	t.Run("PhasePreInstall", func(t *testing.T) {
		assert.Empty(t, cfg.InstructionsForPhase(build.PhasePreInstall))
	})

	t.Run("PhasePostInstall", func(t *testing.T) {
		assert.Empty(t, cfg.InstructionsForPhase(build.PhasePostInstall))
	})
}

func TestLivesConfigValidation(t *testing.T) {
	t.Run("in", func(t *testing.T) {
		t.Run("ok", func(t *testing.T) {
			err := config.Validate(config.LivesConfig{
				In: "/foo",
			})

			assert.False(t, config.IsValidationError(err))
		})

		t.Run("optional", func(t *testing.T) {
			err := config.Validate(config.LivesConfig{})

			assert.False(t, config.IsValidationError(err))
		})

		t.Run("non-root", func(t *testing.T) {
			err := config.Validate(config.LivesConfig{
				In: "/",
			})

			if assert.True(t, config.IsValidationError(err)) {
				msg := config.HumanizeValidationError(err)

				assert.Equal(t, `in: "/" is not a valid absolute non-root path`, msg)
			}
		})

		t.Run("non-root tricky", func(t *testing.T) {
			err := config.Validate(config.LivesConfig{
				In: "/foo/..",
			})

			if assert.True(t, config.IsValidationError(err)) {
				msg := config.HumanizeValidationError(err)

				assert.Equal(t, `in: "/foo/.." is not a valid absolute non-root path`, msg)
			}
		})

		t.Run("absolute", func(t *testing.T) {
			err := config.Validate(config.LivesConfig{
				In: "foo/bar",
			})

			if assert.True(t, config.IsValidationError(err)) {
				msg := config.HumanizeValidationError(err)

				assert.Equal(t, `in: "foo/bar" is not a valid absolute non-root path`, msg)
			}
		})
	})
}
