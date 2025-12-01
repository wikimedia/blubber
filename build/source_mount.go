package build

import "github.com/moby/buildkit/client/llb"

// SourceMount mounts a dependency build context ("local" or another variant)
// at a directory during a [Run] instruction's execution.
type SourceMount struct {
	From        string
	Destination string
	Source      string
	Readonly    bool
}

// RunOption returns an [llb.RunOption] for this source mount.
func (sm SourceMount) RunOption(target *Target) (llb.RunOption, error) {
	mopts := []llb.MountOption{}

	if sm.Source != "" {
		mopts = append(mopts, llb.SourcePath(target.ExpandEnv(sm.Source)))
	}

	if sm.Readonly {
		mopts = append(mopts, llb.Readonly)
	}

	destination := "."
	if sm.Destination != "" {
		destination = sm.Destination
	}

	state, _, err := target.NamedContext(sm.From, ContextOpt{})
	if err != nil {
		return nil, err
	}

	return llb.AddMount(
		target.ExpandEnv(destination),
		state,
		mopts...,
	), nil
}
