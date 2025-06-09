package config

import (
	"encoding/json"

	"gitlab.wikimedia.org/repos/releng/blubber/build"
)

// BuilderConfig contains configuration for the definition of an arbitrary
// build command and the files required to successfully execute the command.
type BuilderConfig struct {
	Command      BuilderCommand     `json:"command"`
	Script       string             `json:"script"`
	Requirements RequirementsConfig `json:"requirements" validate:"omitempty,uniqueartifacts,dive"`
	Mounts       MountsConfig       `json:"mounts" validate:"omitempty,unique,dive"`
	Caches       CachesConfig       `json:"caches" validate:"omitempty,unique,dive"`
}

// Dependencies returns variant dependencies.
func (bc BuilderConfig) Dependencies() []string {
	return append(
		bc.Requirements.Dependencies(),
		bc.Mounts.Dependencies()...,
	)
}

// Merge takes another BuilderConfig and merges its fields into this one's,
// overwriting the builder command and requirements.
func (bc *BuilderConfig) Merge(bc2 BuilderConfig) {
	if bc2.Command != nil {
		bc.Command = bc2.Command
	}

	if bc2.Script != "" {
		bc.Script = bc2.Script
	}

	if bc2.Requirements != nil {
		bc.Requirements = bc2.Requirements
	}

	if bc2.Mounts != nil {
		bc.Mounts = bc2.Mounts
	}

	if bc2.Caches != nil {
		bc.Caches = bc2.Caches
	}
}

// InstructionsForPhase injects instructions into the build related to
// builder commands and required files.
//
// # PhasePreInstall
//
// Creates directories for requirements files, copies in requirements files,
// and runs the builder command.
func (bc BuilderConfig) InstructionsForPhase(phase build.Phase) []build.Instruction {
	instructions := bc.Requirements.InstructionsForPhase(phase)

	switch phase {
	case build.PhasePreInstall:
		opts := append(
			bc.Mounts.RunOptions(),
			bc.Caches.RunOptions()...,
		)

		if bc.Script != "" {
			instructions = append(instructions, build.RunScript{
				Script:  []byte(bc.Script),
				Options: opts,
			})
		} else if len(bc.Command) > 0 {
			run := build.Run{Command: bc.Command[0]}
			if len(bc.Command) > 1 {
				run.Arguments = bc.Command[1:]
			}

			instructions = append(instructions, build.RunAllWithOptions{
				Runs:    []build.Run{run},
				Options: opts,
			})
		}

	}

	return instructions
}

// BuilderCommand represents a single builder command to run.
type BuilderCommand []string

// UnmarshalJSON parses a shell command from either `["cmd", "arg"]` or `"cmd
// arg"` form.
func (bc *BuilderCommand) UnmarshalJSON(data []byte) error {
	var cmdString string

	err := json.Unmarshal(data, &cmdString)
	if err == nil && cmdString != "" {
		(*bc) = BuilderCommand([]string{cmdString})
		return nil
	}

	var cmd []string
	err = json.Unmarshal(data, &cmd)
	if err != nil {
		return err
	}

	(*bc) = BuilderCommand(cmd)

	return nil
}
