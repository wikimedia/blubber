package build_test

import (
	"testing"

	"github.com/moby/buildkit/solver/pb"
	"gitlab.wikimedia.org/repos/releng/blubber/build"
	"gitlab.wikimedia.org/repos/releng/blubber/util/testtarget"
)

func TestCacheMount(t *testing.T) {
	t.Run("default ID and access", func(t *testing.T) {
		_, req := testtarget.Compile(t,
			testtarget.NewTargets("foo"),
			build.RunAllWithOptions{
				[]build.Run{
					{"apt-get install", []string{"build-essentials"}},
				},
				[]build.RunOption{
					build.CacheMount{
						Destination: "/var/cache/apt",
					},
				},
			},
		)

		_, eops := req.ContainsNExecOps(1)

		req.Equal(
			[]string{"/bin/sh", "-c", `apt-get install "build-essentials"`},
			eops[0].Exec.Meta.Args,
		)
		req.Len(eops[0].Exec.Mounts, 2)
		mnt := eops[0].Exec.Mounts[1]

		req.Equal("/var/cache/apt", mnt.Dest)
		req.Equal(pb.MountType_CACHE, mnt.MountType)
		req.NotNil(mnt.CacheOpt)
		req.Equal(pb.CacheSharingOpt_SHARED, mnt.CacheOpt.Sharing)
		req.Equal("/var/cache/apt", mnt.CacheOpt.ID)
	})

	t.Run("with ID and access", func(t *testing.T) {
		_, req := testtarget.Compile(t,
			testtarget.NewTargets("foo"),
			build.RunAllWithOptions{
				[]build.Run{
					{"apt-get install", []string{"build-essentials"}},
				},
				[]build.RunOption{
					build.CacheMount{
						Destination: "/var/cache/apt",
						ID:          "apt-cache",
						Access:      "locked",
					},
				},
			},
		)

		_, eops := req.ContainsNExecOps(1)

		req.Equal(
			[]string{"/bin/sh", "-c", `apt-get install "build-essentials"`},
			eops[0].Exec.Meta.Args,
		)
		req.Len(eops[0].Exec.Mounts, 2)
		mnt := eops[0].Exec.Mounts[1]

		req.Equal("/var/cache/apt", mnt.Dest)
		req.Equal(pb.MountType_CACHE, mnt.MountType)
		req.NotNil(mnt.CacheOpt)
		req.Equal(pb.CacheSharingOpt_LOCKED, mnt.CacheOpt.Sharing)
		req.Equal("apt-cache", mnt.CacheOpt.ID)
	})

	t.Run("with UID/GID", func(t *testing.T) {
		_, req := testtarget.Compile(t,
			testtarget.NewTargets("foo"),
			build.RunAllWithOptions{
				[]build.Run{
					{"apt-get install", []string{"build-essentials"}},
				},
				[]build.RunOption{
					build.CacheMount{
						Destination: "/var/cache/apt",
						UID:         "123",
						GID:         "321",
					},
				},
			},
		)

		_, eops := req.ContainsNExecOps(1)

		req.Equal(
			[]string{"/bin/sh", "-c", `apt-get install "build-essentials"`},
			eops[0].Exec.Meta.Args,
		)
		req.Len(eops[0].Exec.Mounts, 2)
		mnt := eops[0].Exec.Mounts[1]

		req.Equal("/var/cache/apt", mnt.Dest)
		req.Equal("/cache", mnt.Selector)
		req.Equal(pb.MountType_CACHE, mnt.MountType)
		req.NotNil(mnt.CacheOpt)
		req.Equal(pb.CacheSharingOpt_SHARED, mnt.CacheOpt.Sharing)
	})
}
