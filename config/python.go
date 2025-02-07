package config

import (
	"gitlab.wikimedia.org/repos/releng/blubber/build"
)

// PythonPoetryVenvs is the path where Poetry will create virtual environments.
const PythonPoetryVenvs = LocalLibPrefix + "/poetry"

// PythonVenv is the path of the virtualenv that will be created if use-virtualenv is true.
const PythonVenv = LocalLibPrefix + "/venv"

// PythnonUvVenvs is the path of the uv environment that will be created if use-virtualenv is true.
const PythnonUvVenvs = LocalLibPrefix + "/uv"

// PythonConfig holds configuration fields related to pre-installation of project
// dependencies via PIP.
type PythonConfig struct {
	// Python binary to use when installing dependencies
	Version string `json:"version"`

	// Install requirements from given files
	Requirements RequirementsConfig `json:"requirements" validate:"omitempty,unique,dive"`

	UseSystemSitePackages Flag `json:"use-system-site-packages"`

	UseNoDepsFlag Flag `json:"no-deps"`

	// Use Poetry for package management
	Poetry PoetryConfig `json:"poetry"`

	//Use UV for package management
	Uv UvConfig `json:"uv"`

	// Specify a specific version of tox to install (T346226)
	ToxVersion string `json:"tox-version"`
}

// UvConfig holds configuration fields for project dependencies installation
type UvConfig struct {
	Version string `json:"version" validate:"omitempty,pypkgver"`
	Devel   Flag   `json:"devel"`
	Variant string `json:"variant"`
}

// PoetryConfig holds configuration fields related to installation of project
// dependencies via Poetry.
type PoetryConfig struct {
	Version string `json:"version" validate:"omitempty,pypkgver"`
	Devel   Flag   `json:"devel"`
}

// Dependencies returns variant dependencies.
func (pc PythonConfig) Dependencies() []string {
	return pc.Requirements.Dependencies()
}

// Merge takes another PythonConfig and merges its fields into this one's,
// overwriting both the dependencies flag and requirements.
func (pc *PythonConfig) Merge(pc2 PythonConfig) {
	pc.UseSystemSitePackages.Merge(pc2.UseSystemSitePackages)
	pc.UseNoDepsFlag.Merge(pc2.UseNoDepsFlag)
	pc.Poetry.Merge(pc2.Poetry)
	if pc2.Version != "" {
		pc.Version = pc2.Version
	}

	if pc2.Requirements != nil {
		pc.Requirements = pc2.Requirements
	}

	if pc2.ToxVersion != "" {
		pc.ToxVersion = pc2.ToxVersion
	}
}

// Merge two PoetryConfig structs.
func (pc *PoetryConfig) Merge(pc2 PoetryConfig) {
	if pc2.Version != "" {
		pc.Version = pc2.Version
	}
	pc.Devel.Merge(pc2.Devel)
}

// InstructionsForPhase injects instructions into the build related to Python
// dependency installation.
//
// # PhasePrivileged
//
// Ensures that the newest versions of setuptools, wheel, tox, and pip are
// installed.
//
// # PhasePreInstall
//
// Sets up Python wheels under the shared library directory (/opt/lib/python)
// for dependencies found in the declared requirements files. Installing
// dependencies during the build.PhasePreInstall phase allows a compiler
// implementation (e.g. Docker) to produce cache-efficient output so only
// changes to the given requirements files will invalidate these steps of the
// image build.
//
// Injects build.Env instructions for PIP_WHEEL_DIR and PIP_FIND_LINKS that
// will cause future executions of `pip install` (and by extension, `tox`) to
// consider packages from the shared library directory first.
//
// # PhasePostInstall
//
// Injects a build.Env instruction for PIP_NO_INDEX that will cause future
// executions of `pip install` and `tox` to consider _only_ packages from the
// shared library directory, helping to speed up image builds by reducing
// network requests from said commands.
func (pc PythonConfig) InstructionsForPhase(phase build.Phase) []build.Instruction {
	ins := []build.Instruction{}

	if !pc.isEnabled() {
		return ins
	}

	// This only does something for build.PhasePreInstall
	ins = append(ins, pc.Requirements.InstructionsForPhase(phase)...)

	python := pc.version()

	switch phase {
	case build.PhasePreInstall:
		venvSetupCmd := []string{"-m", "venv", PythonVenv}
		if pc.UseSystemSitePackages.True {
			venvSetupCmd = append(venvSetupCmd, "--system-site-packages")
		}
		ins = append(ins, build.Run{python, venvSetupCmd})

		// "Activate" the virtualenv
		ins = append(ins, build.Env{map[string]string{
			"VIRTUAL_ENV": PythonVenv,
			"PATH":        PythonVenv + "/bin:$PATH",
		}})
		ins = append(ins, pc.setupPipAndPoetryAndUv()...)

		// Switch case for package managers
		switch {
		case pc.usePoetry():
			cmd := []string{"install", "--no-root"}
			if !pc.Poetry.Devel.True {
				cmd = append(cmd, "--no-dev")
			}
			ins = append(ins, build.CreateDirectory(PythonPoetryVenvs))
			ins = append(ins, build.Run{"poetry", cmd})

		case pc.useUv():
			cmd := []string{} // Initialize an empty list of commands

			// Check the uv variant
			if pc.Uv.Variant == "pip" {
				// Pip variant using uv pip install -r requirements.txt
				cmd = append(cmd, "pip", "install", "-r", "requirements.txt")
			} else {
				// Default uv sync
				cmd = append(cmd, "sync")
				if !pc.Uv.Devel.True {
					cmd = append(cmd, "--no-group", "dev")
				}
			}
			ins = append(ins, build.CreateDirectory(PythnonUvVenvs))
			ins = append(ins, build.Run{"uv", cmd})

		default:
			args := pc.RequirementsArgs()
			if args != nil {
				installCmd := []string{"-m", "pip", "install"}
				if pc.UseNoDepsFlag.True {
					installCmd = append(installCmd, "--no-deps")
				}
				ins = append(ins, build.Run{python, append(installCmd, args...)})
			}
		}

	case build.PhasePostInstall:
		if !pc.usePoetry() {
			if pc.UseNoDepsFlag.True {
				// Ensure requirements has all transitive dependencies
				ins = append(ins, build.Run{
					python, []string{
						"-m", "pip", "check",
					},
				})
			}
		}
	}

	return ins
}

func (pc PythonConfig) isEnabled() bool {
	return pc.Version != "" && pc.Requirements != nil
}

func (pc PythonConfig) setupPipAndPoetryAndUv() []build.Instruction {
	ins := []build.Instruction{}

	ins = append(ins, build.RunAll{[]build.Run{
		{pc.version(), []string{"-m", "pip", "install", "-U", "setuptools!=60.9.0"}},
		{pc.version(), []string{"-m", "pip", "install", "-U", "wheel", pc.toxPackage(), pc.pipPackage()}},
	}})

	if pc.usePoetry() {
		ins = append(ins, build.Env{map[string]string{
			"POETRY_VIRTUALENVS_PATH": PythonPoetryVenvs,
		}})
		ins = append(ins, build.Run{
			pc.version(), []string{
				"-m", "pip", "install", "-U", "poetry" + pc.Poetry.Version,
			},
		})
	} else {
		ins = append(ins, build.Env{map[string]string{
			"UV_VIRTUALENVS_PATH": PythnonUvVenvs,
		}})
		ins = append(ins, build.Run{
			pc.version(), []string{
				"-m", "pip", "install", "-U", "uv" + pc.Uv.Version,
			},
		})
	}

	return ins
}

// RequirementsArgs returns the configured requirements as pip `-r` arguments.
func (pc PythonConfig) RequirementsArgs() []string {
	if pc.Requirements == nil || len(pc.Requirements) == 0 {
		return nil
	}

	args := make([]string, len(pc.Requirements)*2)

	for i, req := range pc.Requirements {
		args[i*2] = "-r"
		args[(i*2)+1] = req.EffectiveDestination()
	}

	return args
}

func (pc PythonConfig) pipPackage() string {
	if pc.version()[0:7] == "python2" {
		return "pip<21"
	}

	return "pip"
}

func (pc PythonConfig) version() string {
	if pc.Version == "" {
		return "python"
	}

	return pc.Version
}

func (pc PythonConfig) usePoetry() bool {
	return pc.Poetry.Version != ""
}

func (pc PythonConfig) useUv() bool {
	return pc.Uv.Version != ""
}

func (pc PythonConfig) toxPackage() string {
	if pc.ToxVersion == "" {
		return "tox"
	}

	return "tox==" + pc.ToxVersion
}
