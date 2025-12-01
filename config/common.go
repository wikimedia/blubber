package config

import (
	"gitlab.wikimedia.org/repos/releng/blubber/build"
)

// CommonConfig holds the configuration fields common to both the root config
// and each configured variant.
type CommonConfig struct {
	Base       string          `json:"base" validate:"omitempty,context"`
	Arguments  ArgumentsConfig `json:"arguments" validate:"envvars"`
	Apt        AptConfig       `json:"apt"`
	Builders   BuildersConfig  `json:"builders" validate:"uniquetypesexcept=config.BuilderConfig,notallowedwith=node php python builder,dive"`
	Node       NodeConfig      `json:"node"`
	Php        PhpConfig       `json:"php"`
	Python     PythonConfig    `json:"python"`
	Builder    BuilderConfig   `json:"builder"`
	Lives      LivesConfig     `json:"lives"`
	Runs       RunsConfig      `json:"runs"`
	EntryPoint []string        `json:"entrypoint"`
}

// Dependencies returns variant dependencies.
func (cc *CommonConfig) Dependencies() []string {
	dependents := []VariantDependent{
		cc.Node,
		cc.Php,
		cc.Python,
		cc.Builder,
		cc.Builders,
	}

	deps := []string{}
	for _, dependent := range dependents {
		deps = append(deps, dependent.Dependencies()...)
	}

	return deps
}

// Merge takes another CommonConfig and merges its fields this one's.
func (cc *CommonConfig) Merge(cc2 CommonConfig) {
	if cc2.Base != "" {
		cc.Base = cc2.Base
	}

	cc.Arguments.Merge(cc2.Arguments)
	cc.Apt.Merge(cc2.Apt)
	cc.Builders.Merge(cc2.Builders)
	cc.Node.Merge(cc2.Node)
	cc.Php.Merge(cc2.Php)
	cc.Python.Merge(cc2.Python)
	cc.Builder.Merge(cc2.Builder)
	cc.Lives.Merge(cc2.Lives)
	cc.Runs.Merge(cc2.Runs)

	if cc2.EntryPoint != nil {
		cc.EntryPoint = cc2.EntryPoint
	}
}

// PhaseCompileableConfig returns all fields that implement
// build.PhaseCompileable in the order that their instructions should be
// injected.
func (cc *CommonConfig) PhaseCompileableConfig() []build.PhaseCompileable {
	return []build.PhaseCompileable{
		cc.Arguments,
		cc.Apt,
		cc.Builders,
		cc.Node,
		cc.Php,
		cc.Python,
		cc.Builder,
		cc.Lives,
		cc.Runs,
	}
}

// InstructionsForPhase injects instructions into the given build phase for
// each member field that supports it.
func (cc *CommonConfig) InstructionsForPhase(phase build.Phase) []build.Instruction {
	instructions := []build.Instruction{}

	if !cc.IsScratch() {
		for _, phaseCompileable := range cc.PhaseCompileableConfig() {
			instructions = append(instructions, phaseCompileable.InstructionsForPhase(phase)...)
		}
	}

	return instructions
}

// IsScratch returns whether this is configuration for a scratch image (no
// base image).
func (cc *CommonConfig) IsScratch() bool {
	return cc.Base == ""
}
