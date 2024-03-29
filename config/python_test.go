package config_test

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"

	"gitlab.wikimedia.org/repos/releng/blubber/build"
	"gitlab.wikimedia.org/repos/releng/blubber/config"
)

func TestPythonConfigYAMLMerge(t *testing.T) {
	cfg, err := config.ReadYAMLConfig([]byte(`---
    version: v4
    base: foo
    python:
      version: python2.7
      requirements: [requirements.txt]
      tox-version: 4.11.1
    variants:
      test:
        python:
          version: python3
          requirements: [other-requirements.txt, requirements-test.txt]
          tox-version: 4.11.2
          use-system-flag: true`))

	if assert.NoError(t, err) {
		assert.Equal(t, config.RequirementsConfig{
			{From: "local", Source: "requirements.txt"},
		}, cfg.Python.Requirements)
		assert.Equal(t, "python2.7", cfg.Python.Version)

		err = config.ExpandIncludesAndCopies(cfg, "test")
		assert.Nil(t, err)

		variant, err := config.GetVariant(cfg, "test")

		if assert.NoError(t, err) {
			assert.Equal(t, config.RequirementsConfig{
				{From: "local", Source: "other-requirements.txt"},
				{From: "local", Source: "requirements-test.txt"},
			}, variant.Python.Requirements)
			assert.Equal(t, "python3", variant.Python.Version)
			assert.Equal(t, true, variant.Python.UseSystemFlag.True)
			assert.Equal(t, "4.11.2", variant.Python.ToxVersion)
		}
	}
}

func TestPythonConfigYAMLMergeEmpty(t *testing.T) {
	cfg, err := config.ReadYAMLConfig([]byte(`---
    version: v4
    base: foo
    python:
      requirements: [requirements.txt]
    variants:
      test:
        python:
          requirements: []`))

	if assert.NoError(t, err) {
		assert.Equal(t, config.RequirementsConfig{
			{From: "local", Source: "requirements.txt"},
		}, cfg.Python.Requirements)

		err = config.ExpandIncludesAndCopies(cfg, "test")
		assert.Nil(t, err)

		variant, err := config.GetVariant(cfg, "test")

		if assert.NoError(t, err) {
			assert.Equal(t, config.RequirementsConfig{}, variant.Python.Requirements)
		}
	}
}

func TestPythonConfigInstructionsNoRequirementsWithVersion(t *testing.T) {
	cfg := config.PythonConfig{
		Version: "python2.7",
	}

	t.Run("PhasePrivileged", func(t *testing.T) {
		assert.Empty(t, cfg.InstructionsForPhase(build.PhasePrivileged))
	})

	t.Run("PhasePrivilegeDropped", func(t *testing.T) {
		assert.Empty(t, cfg.InstructionsForPhase(build.PhasePrivilegeDropped))
	})

	t.Run("PhasePreInstall", func(t *testing.T) {
		assert.Equal(t,
			[]build.Instruction{
				build.Env{map[string]string{
					"PYTHONPATH": "/opt/lib/python/site-packages:${PYTHONPATH}",
					"PATH":       "/opt/lib/python/site-packages/bin:${PATH}",
				}},
			},
			cfg.InstructionsForPhase(build.PhasePreInstall),
		)
	})

	t.Run("PhasePostInstall", func(t *testing.T) {
		assert.Empty(t, cfg.InstructionsForPhase(build.PhasePostInstall))
	})
}

func TestPythonConfigInstructionsNoRequirementsNoVersion(t *testing.T) {
	cfg := config.PythonConfig{}

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

func TestPythonConfigInstructionsWithRequirements(t *testing.T) {
	cfg := config.PythonConfig{
		Version: "python2.7",
		Requirements: config.RequirementsConfig{
			{From: "local", Source: "requirements.txt"},
			{From: "local", Source: "requirements-test.txt"},
			{From: "local", Source: "docs/requirements.txt"},
		},
	}

	t.Run("PhasePrivileged", func(t *testing.T) {
		assert.Equal(t,
			[]build.Instruction{
				build.Env{map[string]string{
					"PIP_BREAK_SYSTEM_PACKAGES": "1",
				}},
				build.RunAll{
					[]build.Run{
						{
							"python2.7",
							[]string{"-m", "pip", "install", "-U", "setuptools!=60.9.0"},
						},
						{
							"python2.7",
							[]string{"-m", "pip", "install", "-U", "wheel", "tox", "pip<21"},
						},
					},
				},
			},
			cfg.InstructionsForPhase(build.PhasePrivileged),
		)
	})

	t.Run("PhasePrivilegeDropped", func(t *testing.T) {
		assert.Empty(t, cfg.InstructionsForPhase(build.PhasePrivilegeDropped))
	})

	t.Run("PhasePreInstall", func(t *testing.T) {
		assert.Equal(t,
			[]build.Instruction{
				build.Copy{[]string{"requirements.txt", "requirements-test.txt"}, "./"},
				build.Copy{[]string{"docs/requirements.txt"}, "docs/"},
				build.Env{map[string]string{
					"PIP_WHEEL_DIR":  "/opt/lib/python",
					"PIP_FIND_LINKS": "file:///opt/lib/python",
				}},
				build.Run{"mkdir -p", []string{"/opt/lib/python"}},
				build.RunAll{[]build.Run{
					{"python2.7", []string{"-m", "pip", "wheel",
						"-r", "requirements.txt",
						"-r", "requirements-test.txt",
						"-r", "docs/requirements.txt",
					}},
					{"python2.7", []string{"-m", "pip", "install",
						"--target", "/opt/lib/python/site-packages",
						"-r", "requirements.txt",
						"-r", "requirements-test.txt",
						"-r", "docs/requirements.txt",
					}},
				}},
				build.Env{map[string]string{
					"PYTHONPATH": "/opt/lib/python/site-packages:${PYTHONPATH}",
					"PATH":       "/opt/lib/python/site-packages/bin:${PATH}",
				}},
			},
			cfg.InstructionsForPhase(build.PhasePreInstall),
		)
	})

	t.Run("PhasePostInstall", func(t *testing.T) {
		assert.Equal(t,
			[]build.Instruction{
				build.Env{map[string]string{
					"PIP_NO_INDEX": "1",
				}},
			},
			cfg.InstructionsForPhase(build.PhasePostInstall),
		)
	})
}

func TestPythonConfigUseSystemFlag(t *testing.T) {
	cfg := config.PythonConfig{
		Version: "python2.7",
		Requirements: config.RequirementsConfig{
			{From: "local", Source: "requirements.txt"},
			{From: "local", Source: "requirements-test.txt"},
			{From: "local", Source: "docs/requirements.txt"},
		},
		UseSystemFlag: config.Flag{True: true},
	}

	t.Run("PhasePreInstall", func(t *testing.T) {
		assert.Equal(t,
			[]build.Instruction{
				build.Copy{[]string{"requirements.txt", "requirements-test.txt"}, "./"},
				build.Copy{[]string{"docs/requirements.txt"}, "docs/"},
				build.Env{map[string]string{
					"PIP_WHEEL_DIR":  "/opt/lib/python",
					"PIP_FIND_LINKS": "file:///opt/lib/python",
				}},
				build.Run{"mkdir -p", []string{"/opt/lib/python"}},
				build.RunAll{[]build.Run{
					{"python2.7", []string{"-m", "pip", "wheel",
						"-r", "requirements.txt",
						"-r", "requirements-test.txt",
						"-r", "docs/requirements.txt",
					}},
					{"python2.7", []string{"-m", "pip", "install", "--system",
						"--target", "/opt/lib/python/site-packages",
						"-r", "requirements.txt",
						"-r", "requirements-test.txt",
						"-r", "docs/requirements.txt",
					}},
				}},
				build.Env{map[string]string{
					"PYTHONPATH": "/opt/lib/python/site-packages:${PYTHONPATH}",
					"PATH":       "/opt/lib/python/site-packages/bin:${PATH}",
				}},
			},
			cfg.InstructionsForPhase(build.PhasePreInstall),
		)
	})
}

func TestPythonConfigUseNoDepsFlag(t *testing.T) {
	cfg := config.PythonConfig{
		Version: "python3.9",
		Requirements: config.RequirementsConfig{
			{From: "local", Source: "requirements.txt"},
		},
		UseNoDepsFlag: config.Flag{True: true},
	}

	t.Run("PhasePreInstall", func(t *testing.T) {
		assert.Equal(t,
			[]build.Instruction{
				build.Copy{[]string{"requirements.txt"}, "./"},
				build.Env{map[string]string{
					"PIP_WHEEL_DIR":  "/opt/lib/python",
					"PIP_FIND_LINKS": "file:///opt/lib/python",
				}},
				build.Run{"mkdir -p", []string{"/opt/lib/python"}},
				build.RunAll{[]build.Run{
					{"python3.9", []string{"-m", "pip", "wheel", "--no-deps",
						"-r", "requirements.txt",
					}},
					{"python3.9", []string{"-m", "pip", "install", "--no-deps",
						"--target", "/opt/lib/python/site-packages",
						"-r", "requirements.txt",
					}},
				}},
				build.Env{map[string]string{
					"PYTHONPATH": "/opt/lib/python/site-packages:${PYTHONPATH}",
					"PATH":       "/opt/lib/python/site-packages/bin:${PATH}",
				}},
			},
			cfg.InstructionsForPhase(build.PhasePreInstall),
		)
	})

	t.Run("PhasePostInstall", func(t *testing.T) {
		assert.Equal(t,
			[]build.Instruction{
				build.Env{map[string]string{
					"PIP_NO_INDEX": "1",
				}},
				build.Run{"python3.9", []string{"-m", "pip", "check"}},
			},
			cfg.InstructionsForPhase(build.PhasePostInstall),
		)
	})

}

func TestPythonConfigRequirementsArgs(t *testing.T) {
	cfg := config.PythonConfig{
		Requirements: config.RequirementsConfig{
			{From: "local", Source: "foo"},
			{From: "local", Source: "bar"},
			{From: "local", Source: "baz/qux"},
		},
	}

	assert.Equal(t,
		[]string{
			"-r", "foo",
			"-r", "bar",
			"-r", "baz/qux",
		},
		cfg.RequirementsArgs(),
	)
}

func TestSliceInsert(t *testing.T) {
	t.Run("test inserting an element", func(t *testing.T) {
		got := config.InsertElement([]string{"Hello", "World"}, "Beautiful", 1)
		expected := []string{"Hello", "Beautiful", "World"}

		if !reflect.DeepEqual(got, expected) {
			t.Errorf("Expected '%v'; got '%v'", expected, got)
		}
	})

	t.Run("test inserting an element at the end", func(t *testing.T) {
		orig := []string{"Foo", "Bar", "Baz"}
		got := config.InsertElement(orig, "Beautiful", len(orig))
		expected := []string{"Foo", "Bar", "Baz", "Beautiful"}

		if !reflect.DeepEqual(got, expected) {
			t.Errorf("Expected '%v'; got '%v'", expected, got)
		}
	})

	t.Run("test inserting an element at the beginning", func(t *testing.T) {
		orig := []string{"Foo", "Bar", "Baz"}
		got := config.InsertElement(orig, "Beautiful", 0)
		expected := []string{"Beautiful", "Foo", "Bar", "Baz"}

		if !reflect.DeepEqual(got, expected) {
			t.Errorf("Expected '%v'; got '%v'", expected, got)
		}
	})
}

func TestPosFinding(t *testing.T) {
	t.Run("test finding string in slice", func(t *testing.T) {
		got := config.PosOf([]string{"foo", "bar"}, "foo")
		expected := 0
		if got != expected {
			t.Errorf("Expected '%v'; got '%v'", expected, got)
		}
	})

	t.Run("test finding string NOT in slice", func(t *testing.T) {
		got := config.PosOf([]string{"foo", "bar"}, "baz")
		expected := -1
		if got != expected {
			t.Errorf("Expected '%v'; got '%v'", expected, got)
		}
	})

}

func TestPythonConfigToxVersion(t *testing.T) {
	cfg := config.PythonConfig{
		Version: "python3",
		Requirements: config.RequirementsConfig{
			{From: "local", Source: "requirements.txt"},
		},
		ToxVersion: "1.23.4",
	}

	t.Run("tox version honored", func(t *testing.T) {
		assert.Equal(t,
			[]build.Instruction{
				build.Env{map[string]string{
					"PIP_BREAK_SYSTEM_PACKAGES": "1",
				}},
				build.RunAll{
					[]build.Run{
						{
							"python3",
							[]string{"-m", "pip", "install", "-U", "setuptools!=60.9.0"},
						},
						{
							"python3",
							[]string{"-m", "pip", "install", "-U", "wheel", "tox==1.23.4", "pip"},
						},
					},
				},
			},
			cfg.InstructionsForPhase(build.PhasePrivileged),
		)
	})
}

func TestPythonConfigInstructionsWithPoetry(t *testing.T) {
	cfg := config.PythonConfig{
		Version: "python3",
		Requirements: config.RequirementsConfig{
			{From: "local", Source: "pyproject.toml"},
			{From: "local", Source: "poetry.lock"},
		},
		Poetry: config.PoetryConfig{Version: "==10.0.1"},
	}

	t.Run("PhasePrivileged", func(t *testing.T) {
		assert.Equal(t,
			[]build.Instruction{
				build.Env{map[string]string{
					"PIP_BREAK_SYSTEM_PACKAGES": "1",
				}},
				build.RunAll{
					[]build.Run{
						{
							"python3",
							[]string{"-m", "pip", "install", "-U", "setuptools!=60.9.0"},
						},
						{
							"python3",
							[]string{"-m", "pip", "install", "-U", "wheel", "tox", "pip"},
						},
					},
				},
				build.Env{map[string]string{
					"POETRY_VIRTUALENVS_PATH": "/opt/lib/poetry",
				}},
				build.Run{
					"python3", []string{
						"-m", "pip", "install", "-U", "poetry==10.0.1",
					},
				},
			},
			cfg.InstructionsForPhase(build.PhasePrivileged),
		)
	})

	t.Run("PhasePrivilegeDropped", func(t *testing.T) {
		assert.Empty(t, cfg.InstructionsForPhase(build.PhasePrivilegeDropped))
	})

	t.Run("PhasePreInstall", func(t *testing.T) {
		assert.Equal(t,
			[]build.Instruction{
				build.Copy{[]string{"pyproject.toml", "poetry.lock"}, "./"},
				build.CreateDirectory("/opt/lib/poetry"),
				build.Run{
					"poetry", []string{"install", "--no-root", "--no-dev"},
				},
			},
			cfg.InstructionsForPhase(build.PhasePreInstall),
		)
	})

	t.Run("PhasePostInstall", func(t *testing.T) {
		assert.Equal(t,
			[]build.Instruction{},
			cfg.InstructionsForPhase(build.PhasePostInstall),
		)
	})
}
