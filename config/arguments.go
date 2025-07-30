package config

import (
	"maps"
	"slices"

	"gitlab.wikimedia.org/repos/releng/blubber/build"
)

// ArgumentsConfig represents a number of build arguments and their default
// values.
type ArgumentsConfig map[string]string

// Merge adds the given arguments to these ones.
func (args *ArgumentsConfig) Merge(other ArgumentsConfig) {
	if *args == nil {
		(*args) = make(ArgumentsConfig)
	}

	for k, v := range other {
		(*args)[k] = v
	}
}

// InstructionsForPhase returns build instructions for all defined arguments.
func (args ArgumentsConfig) InstructionsForPhase(phase build.Phase) []build.Instruction {
	if phase != build.PhasePrivileged {
		return []build.Instruction{}
	}

	ins := make([]build.Instruction, len(args))

	for i, k := range slices.Sorted(maps.Keys(args)) {
		ins[i] = build.StringArg{
			Name:    k,
			Default: args[k],
		}
	}

	return ins
}
