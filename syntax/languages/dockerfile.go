package languages

import (
	"github.com/aretext/aretext/syntax/parser"
)

type dockerfileParseState uint8

const (
	dockerfileParseStateToplevel = dockerfileParseState(iota)
	dockerfileParseStateInstructionArg
)

func (s dockerfileParseState) Equals(other parser.State) bool {
	otherState, ok := other.(dockerfileParseState)
	return ok && s == otherState
}

// DockerfileParseFunc returns a parser for a Dockerfile.
// See https://docs.docker.com/reference/dockerfile
func DockerfileParseFunc() parser.Func {
	parseComment := matchState(
		dockerfileParseStateToplevel,
		consumeString("#").
			ThenMaybe(consumeToNextLineFeed).
			Map(recognizeToken(parser.TokenRoleComment)))

	parseInstruction := matchState(
		dockerfileParseStateToplevel,
		dockerfileInstructionParseFunc().
			Map(setState(dockerfileParseStateInstructionArg)))

	// Transitions back to toplevel state at end of the instruction,
	// taking into account continuations (\ at end of line).
	parseInstructionArg := matchState(
		dockerfileParseStateInstructionArg,
		dockerfileInstructionArgParseFunc())

	skipInvalidInstructionOrComment := matchState(dockerfileParseStateToplevel, consumeToNextLineFeed)

	return initialState(
		dockerfileParseStateToplevel,
		parseComment.
			Or(parseInstruction).
			Or(parseInstructionArg).
			Or(skipInvalidInstructionOrComment),
	)
}

func dockerfileInstructionParseFunc() parser.Func {
	isAsciiLetter := func(r rune) bool { return r >= 'A' && r < 'z' }
	isPrecedingWhitespace := func(r rune) bool { return r == ' ' || r == '\t' }
	instructions := []string{
		"add", "arg", "cmd", "copy", "entrypoint", "env", "expose",
		"from", "healthcheck", "label", "maintainer", "onbuild",
		"run", "shell", "stopsignal", "user", "volume", "workdir",
	}
	return consumeRunesLike(isPrecedingWhitespace).
		MaybeBefore(consumeRunesLike(isAsciiLetter)).
		MapWithInput(recognizeKeywordOrConsume(instructions, false)) // case insensitive
}

func dockerfileInstructionArgParseFunc() parser.Func {
	consumeLineFeed := consumeSingleRuneLike(func(r rune) bool { return r == '\n' })
	consumeContinuationAndNewline := consumeSingleRuneLike(func(r rune) bool { return r == '\\' }).
		ThenMaybe(consumeSingleRuneLike(func(r rune) bool { return r == '\r' })).
		Then(consumeLineFeed)
	return BashParseFunc().
		Or(consumeContinuationAndNewline).
		Or(consumeLineFeed.Map(setState(dockerfileParseStateToplevel)))
}
