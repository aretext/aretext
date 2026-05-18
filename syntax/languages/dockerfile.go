package languages

import (
	"strings"
	"unicode"

	"github.com/aretext/aretext/syntax/parser"
)

type dockerfileParseState uint8

const (
	dockerfileParseStateToplevel = dockerfileParseState(iota)
	dockerfileParseStateFromArgs
	dockerfileParseStateHealthcheckArgs
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
	// instruction's arguments. Otherwise, the parse fails.
	parseInstruction := matchState(
		dockerfileParseStateToplevel,
		// preceding whitespace
		consumeRunesLike(func(r rune) bool { return r == ' ' || r == '\t' }).MaybeBefore(
			// maybe keyword
			consumeRunesLike(func(r rune) bool { return r >= 'A' && r <= 'z' }).MapWithInput(
				dockerfileMapInstructionToState(map[string]dockerfileParseState{
					// FROM needs its own parsing to recognize "AS" keyword.
					"from": dockerfileParseStateFromArgs,
					// HEALTHCHECK needs its own parsing to recognize "NONE" and "CMD" keywords.
					"healthcheck": dockerfileParseStateHealthcheckArgs,
					// ONBUILD is a prefix to some other command, so go back to the toplevel state.
					"onbuild": dockerfileParseStateToplevel,
					// We're cheating a little bit here by using the shell parser for all other
					// commands, which happens to work reasonably even for non-shell forms
					// like exec (`["cmd", "arg"]`) and key-value pairs (`LABEL=label`).
					"add":        dockerfileParseStateShellArgs,
					"arg":        dockerfileParseStateShellArgs,
					"cmd":        dockerfileParseStateShellArgs,
					"copy":       dockerfileParseStateShellArgs,
					"entrypoint": dockerfileParseStateShellArgs,
					"env":        dockerfileParseStateShellArgs,
					"expose":     dockerfileParseStateShellArgs,
					"label":      dockerfileParseStateShellArgs,
					"maintainer": dockerfileParseStateShellArgs,
					"run":        dockerfileParseStateShellArgs,
					"shell":      dockerfileParseStateShellArgs,
					"stopsignal": dockerfileParseStateShellArgs,
					"user":       dockerfileParseStateShellArgs,
					"volume":     dockerfileParseStateShellArgs,
					"workdir":    dockerfileParseStateShellArgs,
				}))))

	// These transition back to toplevel state if they see a newline, ignoring continuation `\`
	parseShellArgs := matchState(
		dockerfileParseStateShellArgs,
		dockerfileShellArgsParseFunc())
	parseFromInstructionArgs := matchState(
		dockerfileParseStateFromArgs,
		dockerfileFromInstructionArgsParseFunc())
	parseHealthcheckInstructionArgs := matchState(
		dockerfileParseStateHealthcheckArgs,
		dockerfileHealthcheckInstructionArgsParseFunc())

	// For unrecognized arguments, consume to the end of the line.
	consumeInvalidLine := matchState(dockerfileParseStateToplevel, consumeToNextLineFeed)

	return initialState(
		dockerfileParseStateToplevel,
		parseToplevelComment.
			Or(parseInstruction).
			Or(parseShellArgs).
			Or(parseFromInstructionArgs).
			Or(parseHealthcheckInstructionArgs).
			Or(consumeInvalidLine))
}

// dockerfileMapInstructionToState checks if the consumed content matches an instruction (case insensitive).
// If it does, transition to the next state for that instruction.
// Otherwise, fail to parse.
func dockerfileMapInstructionToState(instructionToNextState map[string]dockerfileParseState) parser.MapWithInputFn {
	lowercaseInstructionToNextState := make(map[string]dockerfileParseState, len(instructionToNextState))
	var maxLength uint64
	for instruction, nextState := range instructionToNextState {
		maxLength = max(maxLength, uint64(len(instruction)))
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
			Role:   parser.TokenRoleKeyword,
			Length: result.NumConsumed,
		}
		return parser.Result{
			NumConsumed:    result.NumConsumed,
			ComputedTokens: []parser.ComputedToken{token},
			NextState:      nextState,
		}
	}
}

func dockerfileShellArgsParseFunc() parser.Func {
	// TODO: explain the assumption here that bash doesn't consume the continuation or newline
	parseLineContinuation := consumeString("\\").Then(consumeString("\n").Or(consumeString("\r\n")))
	parseShell := BashParseFunc()
	return parseLineContinuation.
		Or(consumeSingleRuneLike(func(r rune) bool { return r == '\n' }).Map(setState(dockerfileParseStateToplevel))).
		Or(parseShell)
}

func dockerfileFromInstructionArgsParseFunc() parser.Func {
	parseAsKeyword := consumeRunesLike(func(r rune) bool { return r >= 'A' && r <= 'z' }).
		MapWithInput(recognizeKeywordOrConsume([]string{"as"}, false)) // case insensitive

	parseEqualOperator := consumeString("=").Map(recognizeToken(parser.TokenRoleOperator))

	isVariableNameRune := func(r rune) bool { return unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' }
	parseVariable := consumeString("$").
		Then(consumeRunesLike(isVariableNameRune)).
		Map(recognizeToken(bashTokenRoleVariable)) // TODO: dockerfile

	parseVariableBrace := consumeString("$").
		Then(bashExpansionParseFunc('{')).
		Map(recognizeToken(bashTokenRoleVariable)) // TODO: dockerfile const

	return parseAsKeyword.
		Or(parseEqualOperator).
		Or(parseVariable).
		Or(parseVariableBrace)
}

func dockerfileHealthcheckInstructionArgsParseFunc() parser.Func {
	// TODO: recognize "NONE" (case insensitive) as a keyword if it occurs first (optionally with preceding whitespace)
	// TODO: skip anything else until "CMD", recognize "CMD" as a keyword and transition to dockerfileParseStateShellArgs state.
	// TODO: handle line continuations:
	/*
		HEALTHCHECK --interval=5m --timeout=3s \
			  CMD curl -f http://localhost/ || exit 1
	*/
}
