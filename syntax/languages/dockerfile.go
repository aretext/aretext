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

	// This parser consumes the first word (ascii) of a line. If it matches
	// a valid docker instruction, it transitions to a state to parse the
	// instruction's arguments.
	//
	// We're cheating a little bit here by using the shell parser for most
	// commands, which happens to work reasonably even for non-shell forms
	// like exec (`["cmd", "arg"]`) and key-value pairs (`LABEL=label`).
	parseInstruction := matchState(
		dockerfileParseStateToplevel,
		consumeRunesLike(func(r rune) bool { return r >= 'A' && r <= 'z' }).
			MapWithInput(
				dockerfileMapInstructionToState(map[string][dockerfileParseState]{
					"add": dockerfileParseStateShellArgs,
					"arg": dockerfileParseStateShellArgs,
					"cmd": dockerfileParseStateShellArgs,
					"copy":dockerfileParseStateShellArgs,
					"entrypoint":dockerfileParseStateShellArgs,
					"env":dockerfileParseStateShellArgs,
					"expose": dockerfileParseStateShellArgs,
					"from": dockerfileParseStateFromArgs,
					"healthcheck": dockerfileParseStateHealthcheckArgs,
					"label": dockerfileParseStateShellArgs,
					"maintainer":dockerfileParseStateShellArgs,
					"onbuild":dockerfileParseStateOnbuildArgs,
					"run":dockerfileParseStateShellArgs,
					"shell":dockerfileParseStateShellArgs,
					"stopsignal":dockerfileParseStateShellArgs,
					"user":dockerfileParseStateShellArgs,
					"volume":dockerfileParseStateShellArgs,
					"workdir":dockerfileParseStateShellArgs,
				})))

	parseShellArgs := matchState(
		dockerfileParseStateShellArgs,
		dockerfileShellArgsParseFunc().Map(setState(dockerfileParseStateToplevel)))

	parseFromInstructionArgs := matchState(
		dockerfileParseStateFromArgs,
		dockerfileFromInstructionArgsParseFunc().Map(setState(dockerfileParseStateToplevel)))

	parseHealthcheckInstructionArgs := matchState(
		dockerfileParseStateHealthcheckArgs,
		dockerfileHealthcheckInstructionArgsParseFunc().Map(setState(dockerfileParseStateToplevel)))

	parseOnbuildInstructionArgs := matchState(
		dockerfileParseStateOnbuildArgs,
		dockerfileOnbuildInstructionArgsParseFunc().Map(setState(dockerfileParseStateToplevel)))

	// For unrecognized arguments, consume to the end of the line.
	consumeInvalidLine := matchState(dockerfileParseStateToplevel, consumeToNextLineFeed)

	return initialState(
		dockerfileParseStateToplevel,
		parseToplevelComment.
			Or(parseInstruction).
			Or(parseShellArgs).
			Or(parseFromInstructionArgs).
			Or(parseHealthcheckInstructionArgs).
			Or(parseOnbuildInstructionArgs).
			Or(consumeInvalidLine))
}

func dockerfileFromInstructionArgsParseFunc() parser.Func {
	// TODO
}

func dockerfileHealthcheckInstructionArgsParseFunc() parser.Func {
	// TODO
}

func dockerfileOnbuildInstructionArgsParseFunc() parser.Func {
	// TODO
}

func dockerfileShellArgsParseFunc() parser.Func {
	// TODO
}
