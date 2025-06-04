package build_test

import (
	"testing"

	"github.com/moby/buildkit/solver/pb"
	"gitlab.wikimedia.org/repos/releng/blubber/build"
	"gitlab.wikimedia.org/repos/releng/blubber/util/testtarget"
)

func TestSourceMount(t *testing.T) {
	t.Run("without source path", func(t *testing.T) {
		_, req := testtarget.Setup(t,
			testtarget.NewTargets("bar", "foo"),
			func(bar *build.Target) {
				bar.RunShell("echo yo >> /output/msg")
			},
			func(foo *build.Target) {
				foo.WorkingDirectory("/srv/foo")
				i := build.RunAllWithOptions{
					[]build.Run{{"cat /bar/output/msg", []string{}}},
					[]build.RunOption{
						build.SourceMount{
							From:        "bar",
							Destination: "/bar",
						},
					},
				}
				i.Compile(foo)
			},
		)

		ops, eops := req.ContainsNExecOps(2)
		req.Len(eops[1].Exec.Mounts, 2)
		mnt := eops[1].Exec.Mounts[1]

		req.Equal("/bar", mnt.Dest)
		req.Equal(pb.MountType_BIND, mnt.MountType)
		req.False(mnt.Readonly)

		iops := req.HasValidInputs(ops[1])
		_, sops := req.ContainsNSourceOps(2)
		req.Equal(sops[0], iops[0].Op)
	})

	t.Run("with source path and readonly", func(t *testing.T) {
		_, req := testtarget.Setup(t,
			testtarget.NewTargets("bar", "foo"),
			func(bar *build.Target) {
				bar.RunShell("echo yo >> /output/msg")
			},
			func(foo *build.Target) {
				foo.WorkingDirectory("/srv/foo")
				i := build.RunAllWithOptions{
					[]build.Run{{"cat /bar/msg", []string{}}},
					[]build.RunOption{
						build.SourceMount{
							From:        "bar",
							Destination: "/bar",
							Source:      "/output",
							Readonly:    true,
						},
					},
				}
				i.Compile(foo)
			},
		)

		ops, eops := req.ContainsNExecOps(2)
		req.Len(eops[1].Exec.Mounts, 2)
		mnt := eops[1].Exec.Mounts[1]

		req.Equal("/bar", mnt.Dest)
		req.Equal(pb.MountType_BIND, mnt.MountType)
		req.Equal("/output", mnt.Selector)
		req.True(mnt.Readonly)

		iops := req.HasValidInputs(ops[1])
		_, sops := req.ContainsNSourceOps(2)
		req.Equal(sops[0], iops[0].Op)
	})
}
