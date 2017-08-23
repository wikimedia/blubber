package docker

import (
	"errors"
	"fmt"
	"strings"

	"phabricator.wikimedia.org/source/blubber.git/build"
)

func NewDockerInstruction(instruction build.Instruction) (DockerInstruction, error) {
	switch instruction.(type) {
	case build.Run, build.RunAll:
		var dockerInstruction DockerRun
		dockerInstruction.arguments = instruction.Compile()
		return dockerInstruction, nil
	case build.Copy:
		var dockerInstruction DockerCopy
		dockerInstruction.arguments = instruction.Compile()
		return dockerInstruction, nil
	case build.Env:
		var dockerInstruction DockerEnv
		dockerInstruction.arguments = instruction.Compile()
		return dockerInstruction, nil
	}
	return nil, errors.New("Unable to create DockerInstruction")
}

type DockerInstruction interface {
	Compile() string
	Arguments() []string
}

type abstractDockerInstruction struct {
	arguments []string
}

func (di abstractDockerInstruction) Arguments() []string {
	return di.arguments
}

type DockerRun struct{ abstractDockerInstruction }

func (dr DockerRun) Compile() string {
	return fmt.Sprintf(
		"RUN %s\n",
		removeNewlines(strings.Join(dr.arguments, "")))
}

type DockerCopy struct{ abstractDockerInstruction }

func (dc DockerCopy) Compile() string {
	return fmt.Sprintf(
		"COPY [%s]\n",
		removeNewlines(strings.Join(dc.arguments, ", ")))
}

type DockerEnv struct{ abstractDockerInstruction }

func (de DockerEnv) Compile() string {
	return fmt.Sprintf(
		"ENV %s\n",
		removeNewlines(strings.Join(de.arguments, " ")))
}

func removeNewlines(instructions string) string {
	out := strings.Replace(instructions, "\n", "\\n", -1)
	return out
}