package config_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"gitlab.wikimedia.org/repos/releng/blubber/config"
)

func TestCopiesConfigExpand(t *testing.T) {
	cfg := config.CopiesConfig{
		{From: "foo"},
		{From: "local"},
		{From: "foo", Source: "./foo/dir", Destination: "./bar/dir"},
		{From: "baz", Source: "./baz/dir"},
	}

	expanded := cfg.Expand("/app/dir")

	assert.Equal(t, config.CopiesConfig{
		{From: "foo", Source: "/app/dir", Destination: "/app/dir"},
		{From: "foo", Source: "/opt/lib", Destination: "/opt/lib"},
		{From: "local", Source: ".", Destination: "."},
		{From: "foo", Source: "./foo/dir", Destination: "./bar/dir"},
		{From: "baz", Source: "./baz/dir", Destination: "./baz/dir"},
	}, expanded)
}

func TestCopiesConfigMerge(t *testing.T) {
	cfg := config.CopiesConfig{
		{From: "local"},
		{From: "foo", Source: "/app/dir", Destination: "/app"},
		{From: "bar"},
	}

	cfg.Merge(config.CopiesConfig{
		{From: "foo", Source: "/app/dir", Destination: "/app"},
		{From: "bar", Source: "/some/dir", Destination: "/dir"},
	})

	assert.Equal(t, config.CopiesConfig{
		{From: "local"},
		{From: "bar"},
		{From: "foo", Source: "/app/dir", Destination: "/app"},
		{From: "bar", Source: "/some/dir", Destination: "/dir"},
	}, cfg)
}

func TestCopiesConfigUnmarshalJSON(t *testing.T) {
	t.Run("strings", func(t *testing.T) {
		cfg := config.CopiesConfig{}
		err := cfg.UnmarshalJSON([]byte(`["foo", "bar"]`))

		if assert.NoError(t, err) {
			assert.Len(t, cfg, 2)
			assert.Equal(t, "foo", cfg[0].From)
			assert.Equal(t, "bar", cfg[1].From)
		}
	})

	t.Run("objects", func(t *testing.T) {
		cfg := config.CopiesConfig{}
		err := cfg.UnmarshalJSON([]byte(`[{ "from": "foo" }, { "from": "bar", "source": "/foo", "destination": "/bar" }]`))

		if assert.NoError(t, err) {
			assert.Len(t, cfg, 2)
			assert.Equal(t, "foo", cfg[0].From)
			assert.Equal(t, "bar", cfg[1].From)
			assert.Equal(t, "/foo", cfg[1].Source)
			assert.Equal(t, "/bar", cfg[1].Destination)
		}
	})
}
