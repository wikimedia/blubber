// Package config provides the internal representation of Blubber
// configuration parsed from YAML. Each configuration type may implement
// its own hooks for injecting build instructions into the compiler.
package config

import (
	"github.com/pkg/errors"

	"gitlab.wikimedia.org/repos/releng/blubber/build"
)

// Config holds the root fields of a Blubber configuration.
type Config struct {
	CommonConfig  `json:",inline"`
	Variants      map[string]VariantConfig `json:"variants" validate:"variants,dive"`
	VersionConfig `json:",inline"`

	IncludesDepGraph *DepGraph
	CopiesDepGraph   *DepGraph
}

// VariantCompileables returns a map of [build.TargetCompileable]s for the
// given variant and its dependencies.
func (cfg *Config) VariantCompileables(variant string) (map[string]build.TargetCompileable, error) {
	variants, err := cfg.CopiesDepGraph.GetDeps(variant)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get variant dependencies")
	}

	variants = append(variants, variant)

	compileables := make(map[string]build.TargetCompileable, len(variants))

	for _, variant := range variants {
		vcfg, err := GetVariant(cfg, variant)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to get variant %s", variant)
		}

		compileables[variant] = vcfg
	}

	return compileables, nil
}
