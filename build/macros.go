package build

import (
	"fmt"
)

// ApplyUser wraps any build.Copy instructions as build.CopyAs using the given
// UID/GID.
//
func ApplyUser(uid uint, gid uint, instructions []Instruction) []Instruction {
	applied := make([]Instruction, len(instructions))

	for i, instruction := range instructions {
		switch instruction.(type) {
		case Copy, CopyFrom:
			applied[i] = CopyAs{uid, gid, instruction}
		default:
			applied[i] = instruction
		}
	}

	return applied
}

// Chown returns a build.Run instruction for setting ownership on the given
// path.
//
func Chown(uid uint, gid uint, path string) Run {
	return Run{"chown %s:%s", []string{fmt.Sprint(uid), fmt.Sprint(gid), path}}
}

// CreateDirectories returns a build.Run instruction for creating all the
// given directories.
//
func CreateDirectories(paths []string) Run {
	return Run{"mkdir -p", paths}
}

// CreateDirectory returns a build.Run instruction for creating the given
// directory.
//
func CreateDirectory(path string) Run {
	return CreateDirectories([]string{path})
}

// CreateUser returns build.Run instructions for creating the given user
// account and group.
//
func CreateUser(name string, uid uint, gid uint) []Run {
	return []Run{
		{
			"(getent group %s || groupadd -o -g %s -r %s)",
			[]string{fmt.Sprint(gid), fmt.Sprint(gid), name},
		},
		{
			"(getent passwd %s || useradd -l -o -m -d %s -r -g %s -u %s %s)",
			[]string{fmt.Sprint(uid), homeDir(name), name, fmt.Sprint(uid), name},
		},
	}
}

// Home returns a build.Env instruction for setting the user's home directory.
//
func Home(name string) Env {
	return Env{map[string]string{"HOME": homeDir(name)}}
}

func homeDir(name string) string {
	if name == "root" {
		return "/root"
	}

	return "/home/" + name
}
