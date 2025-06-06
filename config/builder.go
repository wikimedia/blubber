package config

import (
	"gitlab.wikimedia.org/repos/releng/blubber/build"
)

// BuilderConfig contains configuration for the definition of an arbitrary
// build command and the files required to successfully execute the command.
type BuilderConfig struct {
	Command      []string           `json:"command"`
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

		run := build.Run{}

		if bc.Script != "" {
			script := bc.Script

			if !shebangRegexp.MatchString(script) {
				script = "#!/bin/sh"
			}

			run = build.Run
		} else {
			run.Command = bc.Command[0]
			if len(bc.Command) > 1 {
				run.Arguments = bc.Command[1:]
			}
		}

		instructions = append(instructions, build.RunAllWithOptions{
			Runs:    []build.Run{run},
			Options: opts,
		})
	}

	return instructions
}
