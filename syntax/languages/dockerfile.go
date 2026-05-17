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
		dockerfileParseStateTopLevel,
		dockerfileInstructionParseFunc([]string{"from"}).Map(setState(dockerfileParseStateFromArgs)))
	parseFromInstructionArgs := matchState(
		dockerfileParseStateFromArgs,
		dockerfileFromInstructionArgsParseFunc().Map(setState(dockerfileParseStateToplevel)))

	// HEALTHCHECK instruction
	parseHealthcheckInstruction := matchState(
		dockerfileParseStateTopLevel,
		dockerfileInstructionParseFunc([]string{"healthcheck"}).Map(setState(dockerfileParseStateHealthcheckArgs)))
	parseHealthcheckInstructionArgs := matchState(
		dockerfileParseStateHealthcheckArgs,
		dockefileHealthcheckInstructionArgsParseFunc().Map(setState(dockerfileParseStateToplevel)))

	// ONBUILD instruction
	parseOnbuildInstruction := matchState(
		dockerfileParseStateTopLevel,
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
		dockerfileParseStateTopLevel,
		dockerfileInstructionParseFunc([]string{
			"run", "add", "cmd", "label", "maintainer", "expose", "env", "add",
			"copy", "entrypoint", "volume", "user", "workdir", "arg", "stopsignal", "shell",
		}).Map(setState(dockerfileParseStateShellArgs)))
	parseShellArgs := matchState(
		dockerfileParseStateShellArgs,
		dockerfileShellArgsParseFunc().Map(setState(dockerfileParseStateToplevel)))

	// For unrecognized arguments, consume to the end of the line.
	consumeInvalidLine := matchState(dockerfileParseStateTopLevel, consumeToNextLineFeed)

	return initialState(
		dockerfileParseStateToplevel,
		matchState(
			dockerfileParseStateTopLevel,
			parseToplevelComment.
				Or(parseFromInstruction).
				Or(parseHealthcheckInstruction).
				Or(parseOnbuildInstruction).
				Or(parseOtherInstruction).
				Or(consumeInvalidLine)).
			Or(
				matchState(dockerfileParseStateFromArgs, parseFromInstructionArgs).
					Or(matchState(dockerfileParseStateHealthcheckArgs, parseHealthcheckInstructionArgs)).
					Or(matchState(dockerfileParseStateOnbuildInstruction, parseOnbuildInstructionArgs)).
					Or(matchState(dockerfileParseStateShellArgs, parseShellArgs)).
					Map(setState(dockerfileParseStateTopLevel))))
}
