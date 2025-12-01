package build

import (
	"io/fs"
	"strconv"

	"github.com/moby/buildkit/client/llb"
)

const (
	cacheMountSourceDir = "/cache"
	cacheMountMode      = fs.FileMode(0o0755)
)

// CacheMount mounts a persistent cache at the given directory during a [Run]
// instruction's execution.
type CacheMount struct {
	Destination string
	ID          string
	Access      string
	UID         string
	GID         string
}

// RunOption returns an [llb.RunOption] for this cache mount.
func (cm CacheMount) RunOption(target *Target) (llb.RunOption, error) {
	id := cm.ID
	if id == "" {
		id = target.ExpandEnv(cm.Destination)
	}

	mode := llb.CacheMountShared

	switch cm.Access {
	case "private":
		mode = llb.CacheMountPrivate
	case "locked":
		mode = llb.CacheMountLocked
	}

	state := llb.Scratch()
	opts := []llb.MountOption{
		llb.AsPersistentCacheDir(id, mode),
	}

	// If a UID or GID was provided, we must create an initial directory with
	// the right ownership and set that as the mount's source. Note that UID and
	// GID are string representations of int values and may contain build args
	// (i.e. $LIVES_UID and $LIVES_GID) that require expansion.
	uid := 0
	gid := 0

	if cm.UID != "" {
		id, err := strconv.Atoi(target.ExpandEnv(cm.UID))
		if err == nil {
			uid = id
		}
	}

	if cm.GID != "" {
		id, err := strconv.Atoi(target.ExpandEnv(cm.GID))
		if err == nil {
			gid = id
		}
	}

	if uid != 0 || gid != 0 {
		state = state.File(
			llb.Mkdir(
				cacheMountSourceDir,
				cacheMountMode,
				llb.WithUIDGID(uid, gid),
			),
			target.Describef(
				"%s preparing cache mount permissions (%s %d:%d)",
				emojiLocal, cacheMountMode, uid, gid,
			),
		)
		opts = append(opts, llb.SourcePath(cacheMountSourceDir))
	}

	return llb.AddMount(
		target.ExpandEnv(cm.Destination),
		state,
		opts...,
	), nil
}
