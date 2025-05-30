package config_test

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.wikimedia.org/repos/releng/blubber/build"
	"gitlab.wikimedia.org/repos/releng/blubber/config"
)

func TestAptConfigYAML(t *testing.T) {
	cfg, err := config.ReadYAMLConfig([]byte(`---
    version: v4
    apt:
      packages:
        - libfoo
        - libbar
      proxies:
        - url: http://proxy.example:8080
          source: http://security.debian.org
        - https://proxy.example:8081
      sources:
        - url: http://apt.wikimedia.org
          distribution: buster-wikimedia
          components: [component/pygments]
        - url: http://packages.microsoft.com
          distribution: buster
          components: [main]
          signed-by: |
            -----BEGIN PGP PUBLIC KEY BLOCK-----
            Version: GnuPG v1.4.7 (GNU/Linux)

            foo
            -----END PGP PUBLIC KEY BLOCK-----

    variants:
      build:
        apt:
          packages:
            default:
              - libfoo-dev
            baz-backports:
              - libbaz-dev`))

	if assert.NoError(t, err) {
		assert.Equal(t,
			config.AptPackages{"default": {"libfoo", "libbar"}},
			cfg.Apt.Packages,
		)

		assert.Equal(t,
			[]config.AptProxy{
				{URL: "http://proxy.example:8080", Source: "http://security.debian.org"},
				{URL: "https://proxy.example:8081"},
			},
			cfg.Apt.Proxies,
		)

		assert.Equal(t,
			[]config.AptSource{
				{
					URL:          "http://apt.wikimedia.org",
					Distribution: "buster-wikimedia",
					Components:   []string{"component/pygments"},
					SignedBy:     "",
				},
				{
					URL:          "http://packages.microsoft.com",
					Distribution: "buster",
					Components:   []string{"main"},
					SignedBy: `-----BEGIN PGP PUBLIC KEY BLOCK-----
Version: GnuPG v1.4.7 (GNU/Linux)

foo
-----END PGP PUBLIC KEY BLOCK-----
`,
				},
			},
			cfg.Apt.Sources,
		)

		err = config.ExpandIncludesAndCopies(cfg, "build")

		if assert.NoError(t, err) {
			variant, err := config.GetVariant(cfg, "build")

			if assert.NoError(t, err) {
				assert.Equal(t,
					[]config.AptSource{
						{
							URL:          "http://apt.wikimedia.org",
							Distribution: "buster-wikimedia",
							Components:   []string{"component/pygments"},
						},
						{
							URL:          "http://packages.microsoft.com",
							Distribution: "buster",
							Components:   []string{"main"},
							SignedBy: `-----BEGIN PGP PUBLIC KEY BLOCK-----
Version: GnuPG v1.4.7 (GNU/Linux)

foo
-----END PGP PUBLIC KEY BLOCK-----
`,
						},
					},
					variant.Apt.Sources,
				)

				assert.Equal(t,
					[]config.AptProxy{
						{URL: "http://proxy.example:8080", Source: "http://security.debian.org"},
						{URL: "https://proxy.example:8081"},
					},
					variant.Apt.Proxies,
				)

				assert.Equal(t,
					config.AptPackages{
						"default":       {"libfoo", "libbar", "libfoo-dev"},
						"baz-backports": {"libbaz-dev"},
					},
					variant.Apt.Packages,
				)
			}
		}
	}
}

func TestAptConfigMerge(t *testing.T) {
	cfg := config.AptConfig{
		Packages: config.AptPackages{
			"default":       {"libfoo", "libbar"},
			"baz-backports": {"libbaz"},
		},
		Proxies: []config.AptProxy{
			{URL: "http://proxy.example:8080"},
		},
	}

	cfg.Merge(
		config.AptConfig{
			Packages: config.AptPackages{
				"baz-backports": {"libqux"},
			},
			Proxies: []config.AptProxy{
				{
					URL:    "https://proxy.example:8081",
					Source: "http://security.debian.org",
				},
			},
		},
	)

	assert.Equal(t, config.AptConfig{
		Packages: config.AptPackages{
			"default":       {"libfoo", "libbar"},
			"baz-backports": {"libbaz", "libqux"},
		},
		Proxies: []config.AptProxy{
			{
				URL: "http://proxy.example:8080",
			},
			{
				URL:    "https://proxy.example:8081",
				Source: "http://security.debian.org",
			},
		},
	}, cfg)
}

func TestAptConfigInstructions(t *testing.T) {
	cfg := config.AptConfig{
		Packages: config.AptPackages{
			"default":       {"libfoo", "libbar"},
			"baz-backports": {"libbaz"},
		},
		Sources: []config.AptSource{
			{
				URL:          "http://apt.wikimedia.org",
				Distribution: "buster-wikimedia",
				Components:   []string{"components/pygments"},
			},
			{
				URL:          "https://packages.microsoft.com",
				Distribution: "buster",
				Components:   []string{"main"},
				SignedBy:     "foo",
			},
		},
		Proxies: []config.AptProxy{{
			URL:    "http://proxy.example:8080",
			Source: "http://security.debian.org",
		}},
	}

	t.Run("PhasePrivileged", func(t *testing.T) {
		assert.Equal(t,
			[]build.Instruction{
				build.Env{map[string]string{
					"DEBIAN_FRONTEND": "noninteractive",
				}},
				build.File{
					Path:    "/etc/apt/apt.conf.d/99blubber-proxies",
					Content: []byte(`Acquire::http::Proxy::security.debian.org "http://proxy.example:8080";` + "\n"),
					Mode:    os.FileMode(config.AptFileMode),
				},
				build.File{
					Path:    "/etc/apt/keyrings/2c26b46b68ffc68ff99b453c1d30413413422d706483bfa0f98a5e886266e7ae.asc",
					Content: []byte("foo"),
					Mode:    os.FileMode(config.AptFileMode),
				},
				build.RunAll{[]build.Run{
					build.Run{"apt-get update", []string{}},
					build.Run{"apt-get install -y", []string{"ca-certificates"}},
				}},
				build.File{
					Path: "/etc/apt/sources.list.d/99blubber.list",
					Content: []byte(strings.Join([]string{
						"deb http://apt.wikimedia.org buster-wikimedia components/pygments\n",
						"deb [signed-by=/etc/apt/keyrings/2c26b46b68ffc68ff99b453c1d30413413422d706483bfa0f98a5e886266e7ae.asc] https://packages.microsoft.com buster main\n",
					}, "")),
					Mode: os.FileMode(config.AptFileMode),
				},
				build.RunAll{[]build.Run{
					build.Run{"apt-get update", []string{}},
					build.Run{"apt-get install -y -t", []string{"baz-backports", "libbaz"}},
					build.Run{"apt-get install -y", []string{"libfoo", "libbar"}},
					build.Run{"rm -rf /var/lib/apt/lists/*", []string{}},
					build.Run{"rm -f", []string{"/etc/apt/apt.conf.d/99blubber-proxies"}},
				}}},
			cfg.InstructionsForPhase(build.PhasePrivileged),
		)
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

func TestAptConfigValidation(t *testing.T) {
	t.Run("packages", func(t *testing.T) {
		t.Run("ok", func(t *testing.T) {
			err := config.Validate(config.AptConfig{
				Packages: map[string][]string{
					"default": {
						"f1",
						"foo-fighter",
						"bar+b.az",
						"bar+b.az=0:0.1~foo1-1",
						"bar+b.az/stable",
						"bar+b.az/jessie-wikimedia",
					}},
			})

			assert.False(t, config.IsValidationError(err))
		})

		t.Run("bad", func(t *testing.T) {
			err := config.Validate(config.AptConfig{
				Packages: map[string][]string{
					"default": {
						"f1",
						"foo fighter",
						"bar_baz",
						"bar=0.1*bad version",
						"bar/0bad_release",
					}},
			})

			if assert.True(t, config.IsValidationError(err)) {
				msg := config.HumanizeValidationError(err)

				assert.Equal(t, strings.Join([]string{
					`packages[default][1]: "foo fighter" is not a valid Debian package name`,
					`packages[default][2]: "bar_baz" is not a valid Debian package name`,
					`packages[default][3]: "bar=0.1*bad version" is not a valid Debian package name`,
					`packages[default][4]: "bar/0bad_release" is not a valid Debian package name`,
				}, "\n"), msg)
			}
		})
	})

	t.Run("proxies", func(t *testing.T) {
		t.Run("url", func(t *testing.T) {
			t.Run("ok - http", func(t *testing.T) {
				err := config.Validate(config.AptProxy{
					URL: "http://proxy.example:8080",
				})

				assert.False(t, config.IsValidationError(err))
			})

			t.Run("ok - https", func(t *testing.T) {
				err := config.Validate(config.AptProxy{
					URL: "https://proxy.example:8080",
				})

				assert.False(t, config.IsValidationError(err))
			})

			t.Run("bad - missing", func(t *testing.T) {
				err := config.Validate(config.AptProxy{})

				if assert.True(t, config.IsValidationError(err)) {
					assert.Equal(t,
						`url: is required`,
						config.HumanizeValidationError(err),
					)
				}
			})

			t.Run("bad - invalid scheme", func(t *testing.T) {
				err := config.Validate(config.AptProxy{
					URL: "bad://proxy.example",
				})

				if assert.True(t, config.IsValidationError(err)) {
					assert.Equal(t,
						`url: "bad://proxy.example" is not a valid HTTP/HTTPS URL`,
						config.HumanizeValidationError(err),
					)
				}
			})
		})

		t.Run("source", func(t *testing.T) {
			t.Run("ok - http", func(t *testing.T) {
				err := config.Validate(config.AptProxy{
					URL:    "http://proxy.example:8080",
					Source: "http://security.debian.org/",
				})

				assert.False(t, config.IsValidationError(err))
			})

			t.Run("ok - https", func(t *testing.T) {
				err := config.Validate(config.AptProxy{
					URL:    "http://proxy.example:8080",
					Source: "https://security.debian.org/",
				})

				assert.False(t, config.IsValidationError(err))
			})

			t.Run("ok - missing", func(t *testing.T) {
				err := config.Validate(config.AptProxy{
					URL: "http://proxy.example:8080",
				})

				assert.False(t, config.IsValidationError(err))
			})

			t.Run("bad - invalid scheme", func(t *testing.T) {
				err := config.Validate(config.AptProxy{
					URL:    "http://proxy.example:8080",
					Source: "bad://security.debian.org/",
				})

				if assert.True(t, config.IsValidationError(err)) {
					assert.Equal(t,
						`source: "bad://security.debian.org/" is not a valid HTTP/HTTPS URL`,
						config.HumanizeValidationError(err),
					)
				}
			})
		})
	})
}

func TestAptProxyConfiguration(t *testing.T) {
	t.Run("url only - http", func(t *testing.T) {
		cfg := config.AptProxy{
			URL: "http://proxy.example:8080",
		}

		assert.Equal(t,
			`Acquire::http::Proxy "http://proxy.example:8080";`,
			cfg.Configuration(),
		)
	})

	t.Run("url only - https", func(t *testing.T) {
		cfg := config.AptProxy{
			URL: "https://proxy.example:8080",
		}

		assert.Equal(t,
			`Acquire::https::Proxy "https://proxy.example:8080";`,
			cfg.Configuration(),
		)
	})

	t.Run("specific source - http", func(t *testing.T) {
		cfg := config.AptProxy{
			URL:    "https://proxy.example:8080",
			Source: "http://security.debian.org",
		}

		assert.Equal(t,
			`Acquire::http::Proxy::security.debian.org "https://proxy.example:8080";`,
			cfg.Configuration(),
		)
	})

	t.Run("specific source - https", func(t *testing.T) {
		cfg := config.AptProxy{
			URL:    "http://proxy.example:8080",
			Source: "https://security.debian.org",
		}

		assert.Equal(t,
			`Acquire::https::Proxy::security.debian.org "http://proxy.example:8080";`,
			cfg.Configuration(),
		)
	})
}

func TestAptSourceKeyringPath(t *testing.T) {
	req := require.New(t)

	source := config.AptSource{
		SignedBy: "foo",
	}

	req.Equal("/etc/apt/keyrings/2c26b46b68ffc68ff99b453c1d30413413422d706483bfa0f98a5e886266e7ae.asc", source.KeyringPath())
}
