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
				dockerfileMapInstructionToState(map[string]dockerfileParseState{
					"add":         dockerfileParseStateShellArgs,
					"arg":         dockerfileParseStateShellArgs,
					"cmd":         dockerfileParseStateShellArgs,
					"copy":        dockerfileParseStateShellArgs,
					"entrypoint":  dockerfileParseStateShellArgs,
					"env":         dockerfileParseStateShellArgs,
					"expose":      dockerfileParseStateShellArgs,
					"from":        dockerfileParseStateFromArgs,
					"healthcheck": dockerfileParseStateHealthcheckArgs,
					"label":       dockerfileParseStateShellArgs,
					"maintainer":  dockerfileParseStateShellArgs,
					"onbuild":     dockerfileParseStateOnbuildArgs,
					"run":         dockerfileParseStateShellArgs,
					"shell":       dockerfileParseStateShellArgs,
					"stopsignal":  dockerfileParseStateShellArgs,
					"user":        dockerfileParseStateShellArgs,
					"volume":      dockerfileParseStateShellArgs,
					"workdir":     dockerfileParseStateShellArgs,
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

func dockerfileMapInstructionToState(instructionToNextState map[string]dockerfileParseState) parser.MapWithInputFn {
	lowercaseInstructionToNextState := make(map[string]dockerfileParseState, len(instructionToNextState))
	maxLength := 0
	for instruction, nextState := range instructionToNextState {
		maxLength = max(maxLength, len(instruction))
		lowercaseInstructionToNextState[strings.ToLower(instruction)] = nextState
	}

	return func(result parser.Result, iter parser.TrackingRuneIter, state parser.State) parser.Result {
		if result.NumConsumed > maxLength {
			return parser.Result{} // instruction is too long to match, fail to parse.
		}

		maybeInstruction := readInputString(iter, result.NumConsumed)
		nextState, ok := lowercaseInstructionToNextState[strings.ToLower(maybeInstruction)] // case insensitive
		if !ok {
			return parser.Result{} // no matching instruction, fail to parse.
		}

		// Matched an instruction, consume and transition to the next state
		token := parser.ComputedToken{
			Role: parser.TokenRoleKeyword,
			Length: result.NumConsumed,
		}
		return parser.Result{
			NumConsumed: result.NumConsumed,
			ComputedTokens: []parser.ComputedToken{token},
			NextState: nextState,
		}
	}
}

func dockerfileShellArgsParseFunc() parser.Func {
	// TODO
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
