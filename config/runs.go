package config

import (
	"gitlab.wikimedia.org/repos/releng/blubber/build"
)

// RunsConfig holds configuration fields related to the application's
// runtime environment.
type RunsConfig struct {
	UserConfig  `json:",inline"`
	Environment map[string]string `json:"environment" validate:"envvars"`  // environment variables
	In          string            `json:"in" validate:"omitempty,abspath"` // runtime directory
	Insecurely  Flag              `json:"insecurely"`                      // runs user owns application files
}

// Merge takes another RunsConfig and overwrites this struct's fields. All
// fields except Environment are overwritten if set. The latter is an additive
// merge.
func (run *RunsConfig) Merge(run2 RunsConfig) {
	run.UserConfig.Merge(run2.UserConfig)
	run.Insecurely.Merge(run2.Insecurely)

	if run2.In != "" {
		run.In = run2.In
	}

	if run.Environment == nil {
		run.Environment = make(map[string]string)
	}

	for name, value := range run2.Environment {
		run.Environment[name] = value
	}
}

// InstructionsForPhase injects build instructions related to the runtime
// configuration.
//
// # PhasePrivileged
//
// Creates LocalLibPrefix directory and unprivileged user home directory,
// creates the unprivileged user and its group, and sets up directory
// permissions.
//
// # PhasePrivilegeDropped
//
// Injects build.Env instructions for all names/values defined by
// RunsConfig.Environment.
func (run RunsConfig) InstructionsForPhase(phase build.Phase) []build.Instruction {
	switch phase {
	case build.PhasePrivileged:
		return []build.Instruction{
			build.NewStringArg("RUNS_AS", run.As),
			build.NewUintArg("RUNS_UID", run.UID),
			build.NewUintArg("RUNS_GID", run.GID),
			build.RunAll{
				build.CreateUser("$RUNS_AS", "$RUNS_UID", "$RUNS_GID"),
			},
		}
	case build.PhasePrivilegeDropped:
		if len(run.Environment) > 0 {
			return []build.Instruction{
				build.Env{run.Environment},
			}
		}
	case build.PhasePostInstall:
		if run.In != "" {
			return []build.Instruction{
				build.WorkingDirectory{run.In},
			}
		}
	}

	return []build.Instruction{}
}
