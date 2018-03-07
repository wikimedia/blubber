package config

import (
	"phabricator.wikimedia.org/source/blubber/build"
)

// CommonConfig holds the configuration fields common to both the root config
// and each configured variant.
//
type CommonConfig struct {
	Base         string       `yaml:"base" validate:"omitempty,baseimage"` // name/path to base image
	Apt          AptConfig    `yaml:"apt"`                                 // APT related
	Node         NodeConfig   `yaml:"node"`                                // Node related
	Python       PythonConfig `yaml:"python"`                              // Python related
	Lives        LivesConfig  `yaml:"lives"`                               // application owner/dir
	Runs         RunsConfig   `yaml:"runs"`                                // runtime environment
	SharedVolume Flag         `yaml:"sharedvolume"`                        // use volume for app
	EntryPoint   []string     `yaml:"entrypoint"`                          // entry-point executable
}

// Merge takes another CommonConfig and merges its fields this one's.
//
func (cc *CommonConfig) Merge(cc2 CommonConfig) {
	if cc2.Base != "" {
		cc.Base = cc2.Base
	}

	cc.Apt.Merge(cc2.Apt)
	cc.Node.Merge(cc2.Node)
	cc.Python.Merge(cc2.Python)
	cc.Lives.Merge(cc2.Lives)
	cc.Runs.Merge(cc2.Runs)
	cc.SharedVolume.Merge(cc2.SharedVolume)

	if len(cc.EntryPoint) < 1 {
		cc.EntryPoint = cc2.EntryPoint
	}
}

// PhaseCompileableConfig returns all fields that implement
// build.PhaseCompileable in the order that their instructions should be
// injected.
//
func (cc *CommonConfig) PhaseCompileableConfig() []build.PhaseCompileable {
	return []build.PhaseCompileable{cc.Apt, cc.Node, cc.Python, cc.Lives, cc.Runs}
}

// InstructionsForPhase injects instructions into the given build phase for
// each member field that supports it.
//
func (cc *CommonConfig) InstructionsForPhase(phase build.Phase) []build.Instruction {
	instructions := []build.Instruction{}

	for _, phaseCompileable := range cc.PhaseCompileableConfig() {
		instructions = append(instructions, phaseCompileable.InstructionsForPhase(phase)...)
	}

	return instructions
}
