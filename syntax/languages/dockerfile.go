package languages

import (
	"github.com/aretext/aretext/syntax/parser"
)

type dockerfileParseState uint8

const (
	dockerfileParseStateToplevel = dockerfileParseState(iota)
	dockerfileParseStateFromArgs
	dockerfileParseStateHealthcheckArgs
	dockerfileParseStateOnbuildArgs
	dockerfileParseStateShellArgs
)

func (s dockerfileParseState) Equals(other parser.State) bool {
	otherState, ok := other.(dockerfileParseState)
	return ok && s == otherState
}

// DockerfileParseFunc returns a parser for a Dockerfile.
// See https://docs.docker.com/reference/dockerfile
func DockerfileParseFunc() parser.Func {
	parseToplevelComment := matchState(
		dockerfileParseStateToplevel,
		consumeString("#").
			ThenMaybe(consumeToNextLineFeed).
			Map(recognizeToken(parser.TokenRoleComment)))

	// FROM instruction
	parseFromInstruction := matchState(
		dockerfileParseStateToplevel,
		dockerfileInstructionParseFunc([]string{"from"}).Map(setState(dockerfileParseStateFromArgs)))
	parseFromInstructionArgs := matchState(
		dockerfileParseStateFromArgs,
		dockerfileFromInstructionArgsParseFunc().Map(setState(dockerfileParseStateToplevel)))

	// HEALTHCHECK instruction
	parseHealthcheckInstruction := matchState(
		dockerfileParseStateToplevel,
		dockerfileInstructionParseFunc([]string{"healthcheck"}).Map(setState(dockerfileParseStateHealthcheckArgs)))
	parseHealthcheckInstructionArgs := matchState(
		dockerfileParseStateHealthcheckArgs,
		dockefileHealthcheckInstructionArgsParseFunc().Map(setState(dockerfileParseStateToplevel)))

	// ONBUILD instruction
	parseOnbuildInstruction := matchState(
		dockerfileParseStateToplevel,
		dockerfileInstructionParseFunc([]string{"onbuild"}).Map(setState(dockerfileParseStateOnbuildArgs)))
	parseOnbuildInstructionArgs := matchState(
		dockerfileParseStateOnbuildArgs,
		dockefileOnbuildInstructionArgsParseFunc().Map(setState(dockerfileParseStateToplevel)))

	// All other valid instruction args are parsed as shell.
	// Some of these technically don't support all shell syntax, but there's enough overlap
	// that it ends up looking correct. For example, exec form (`["cmd", "arg1", "arg2"]`)
	// gets parsed into string tokens by the shell parser. Likewise, key value pairs
	// like "LABEL=value" and options like "--option=value" get parsed reasonably as shell.
	parseOtherInstruction := matchState(
		dockerfileParseStateToplevel,
		dockerfileInstructionParseFunc([]string{
			"run", "add", "cmd", "label", "maintainer", "expose", "env", "add",
			"copy", "entrypoint", "volume", "user", "workdir", "arg", "stopsignal", "shell",
		}).Map(setState(dockerfileParseStateShellArgs)))
	parseShellArgs := matchState(
		dockerfileParseStateShellArgs,
		dockerfileShellArgsParseFunc().Map(setState(dockerfileParseStateToplevel)))

	// For unrecognized arguments, consume to the end of the line.
	consumeInvalidLine := matchState(dockerfileParseStateToplevel, consumeToNextLineFeed)

	return initialState(
		dockerfileParseStateToplevel,
		parseToplevelComment.
			Or(parseFromInstruction).
			Or(parseFromInstructionArgs).
			Or(parseHealthcheckInstruction).
			Or(parseHealthcheckInstructionArgs).
			Or(parseOnbuildInstruction).
			Or(parseOnbuildInstructionArgs).
			Or(parseOtherInstruction).
			Or(parseShellArgs).
			Or(consumeInvalidLine))
}
