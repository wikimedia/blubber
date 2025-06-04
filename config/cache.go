package config

import (
	"gitlab.wikimedia.org/repos/releng/blubber/build"
)

// CachesConfig holds a number of [CacheConfig] values.
type CachesConfig []CacheConfig

// RunOptions returns a number of [build.RunOption] for the mounts.
func (ccs CachesConfig) RunOptions() []build.RunOption {
	opts := []build.RunOption{}

	for _, cc := range ccs {
		opts = append(opts, cc.RunOptions()...)
	}

	return opts
}

// UnmarshalJSON implements json.Unmarshaler to handle both shorthand (just
// each destination path) and longhand configuration.
func (ccs *CachesConfig) UnmarshalJSON(unmarshal []byte) error {
	ccs2, err := unmarshalShorthand[CachesConfig](unmarshal, func(destination string) CacheConfig {
		return CacheConfig{
			Destination: destination,
		}
	})

	if err != nil {
		return err
	}

	*ccs = ccs2
	return nil
}

// CacheConfig holds configuration for a single cache mount to be added to a
// [BuilderConfig] during execution.
type CacheConfig struct {
	Destination string `json:"destination" validate:"required"`
	ID          string `json:"id"`
	Access      string `json:"access"`
}

// RunOptions returns a number of [build.RunOption] for the mount.
func (cc CacheConfig) RunOptions() []build.RunOption {
	return []build.RunOption{
		build.CacheMount{
			Destination: cc.Destination,
			ID:          cc.ID,
			Access:      cc.Access,
			UID:         "$LIVES_UID",
			GID:         "$LIVES_GID",
		},
	}
}
