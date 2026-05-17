package languages

import (
	"github.com/aretext/aretext/syntax/parser"
)

type dockerfileParseState uint8

const (
	dockerfileParseStateToplevel = dockerfileParseState(iota)
	dockerfileParseStateInstructionArg
)

func DockerfileParseFunc() parser.Func {
	parseComment := matchState(
		dockerfileParseStateToplevel,
		consumeString("#").
			ThenMaybe(consumeTonextLineFeed).
			Map(recognizeToken(parser.TokenRoleComment)))

	parseInstruction := matchState(
		dockerfileParseStateToplevel,
		dockerfileInstructionParseFunc().
			Map(setState(dockerfileParseStateInstructionArg)))

	parseInstructionArg := matchState(
		dockerfileParseStateInstructionArg,
		dockerfileInstructionArgParseFunc())

	return initialState(
		dockerfileParseStateToplevel,
		parseComment.Or(parseInstruction).Or(parseInstructionArg),
	)
}
