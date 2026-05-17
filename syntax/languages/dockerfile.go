package languages

import (
	"github.com/aretext/aretext/syntax/parser"
)

type dockerfileParseState uint8

const (
	dockerfileParseStateToplevel = dockerfileParseState(iota)
	dockerfileParseStateInstructionArg
)

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

	return initialState(
		dockerfileParseStateToplevel,
		parseComment.Or(parseInstruction).Or(parseInstructionArg),
	)
}

func dockerfileInstructionParseFunc() parser.Func {
	isAciiLetter := func(r rune) bool { return r >= 'A' && r < 'z' }
	instructions := []string{
		"add", "arg", "cmd", "copy", "entrypoint", "env", "expose",
		"from", "healthcheck", "label", "maintainer", "onbuild",
		"run", "shell", "stopsignal", "user", "volume", "workdir",
	}
	return consumeRunesLike(isAsciiLetter).
		MapWithInput(recognizeKeywordOrConsume(instructions, false)) // case insensitive
}

func dockerfileInstructionArgParseFunc() parser.Func {
	consumeContinuationAndNewline := consumeSingleRuneLike(func(r rune) bool { r == '\\' }).
		ThenMaybe(consumeSingleRuneLike(func (r rune) bool { r == '\r' })).
		Then(consumeSingleRuneLike(func(r rune) bool { r == '\n' }))
	return bashParseFunc().Or(consumeContinuationAndNewline)
}
