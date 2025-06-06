package config

import (
	"gitlab.wikimedia.org/repos/releng/blubber/build"
)

// MountsConfig holds a number of [MountConfig] values.
type MountsConfig []MountConfig

// Dependencies returns the variant dependencies of the mounts.
func (mcs MountsConfig) Dependencies() []string {
	deps := []string{}

	for i := range mcs {
		deps = append(deps, mcs[i].Dependencies()...)
	}

	return deps
}

// RunOptions returns a number of [build.RunOption] for the mounts.
func (mcs MountsConfig) RunOptions() []build.RunOption {
	opts := []build.RunOption{}

	for _, mc := range mcs {
		opts = append(opts, mc.RunOptions()...)
	}

	return opts
}

// UnmarshalJSON implements json.Unmarshaler to handle both shorthand (just
// each destination path) and longhand configuration.
func (mcs *MountsConfig) UnmarshalJSON(unmarshal []byte) error {
	mcs2, err := unmarshalShorthand[MountsConfig](unmarshal, func(destination string) MountConfig {
		return MountConfig{
			From:        build.LocalContextKeyword,
			Destination: destination,
		}
	})

	if err != nil {
		return err
	}

	*mcs = mcs2
	return nil
}

// MountConfig holds configuration for a single source mount to be added to a
// [BuilderConfig] during execution.
type MountConfig struct {
	From        string `json:"from"`
	Destination string `json:"destination" validate:"required"`
	Source      string `json:"source" validate:"omitempty"`
}

// RunOptions returns a number of [build.RunOption] for the mount.
func (mc MountConfig) RunOptions() []build.RunOption {
	return []build.RunOption{
		build.SourceMount{
			From:        mc.From,
			Destination: mc.Destination,
			Source:      mc.Source,
			Readonly:    true,
		},
	}
}

// Dependencies returns the variant dependencies of the mount.
func (mc MountConfig) Dependencies() []string {
	// Since a mount can reference things other than another variant, only
	// return references that match the variant name validation regexp.
	if !variantNameRegexp.MatchString(mc.From) {
		return []string{}
	}

	return []string{mc.From}
}
