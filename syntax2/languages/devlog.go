package languages

import (
	"github.com/aretext/aretext/syntax2/parser"
)

// devlogParseState represents the state of the devlog parser.
type devlogParseState struct {
	AtStartOfLine bool
}

func (s devlogParseState) Equals(other parser.State) bool {
	otherState, ok := other.(devlogParseState)
	return ok && s == otherState
}

// DevlogParseFunc returns a parse func for the devlog file format.
// See https://devlog-cli.org/ for details.
func DevlogParseFunc() parser.Func {
	// Custom token roles
	const toDoRole = parser.TokenRoleCustom1
	const inProgressRole = parser.TokenRoleCustom2
	const completedRole = parser.TokenRoleCustom3
	const blockedRole = parser.TokenRoleCustom4
	const tildeSeparatorRole = parser.TokenRoleCustom6
	const codeBlockRole = parser.TokenRoleCustom5

	// Parser states
	startOfLineState := devlogParseState{AtStartOfLine: true}
	withinLineState := devlogParseState{AtStartOfLine: false}

	// Parse funcs
	parseTaskToDo := matchState(
		startOfLineState,
		consumeString("*").
			Map(recognizeToken(toDoRole)).
			Map(setState(withinLineState)),
	)

	parseTaskInProgress := matchState(
		startOfLineState,
		consumeString("^").Then(consumeToEndOfLine).
			Map(recognizeToken(inProgressRole)).
			Map(setState(startOfLineState)),
	)

	parseTaskCompleted := matchState(
		startOfLineState,
		consumeString("+").Then(consumeToEndOfLine).
			Map(recognizeToken(completedRole)).
			Map(setState(startOfLineState)),
	)

	parseTaskBlocked := matchState(
		startOfLineState,
		consumeString("-").Then(consumeToEndOfLine).
			Map(recognizeToken(blockedRole)).
			Map(setState(startOfLineState)),
	)

	parseTildeSeparator := matchState(
		startOfLineState,
		consumeString("~~").Then(consumeRunesLike(func(r rune) bool { return r == '~' })).
			Map(recognizeToken(tildeSeparatorRole)).
			Map(setState(withinLineState)),
	)

	parseCodeBlock := consumeString("```").Then(consumeToString("```")).
		Map(recognizeToken(codeBlockRole)).
		Map(setState(withinLineState))

	parseEndOfLine := consumeString("\n").
		Map(setState(startOfLineState))

	consumeUntilNextParseable := consumeUntilEofOrRuneLike(func(r rune) bool {
		return r == '`' || r == '\n'
	}).Map(setState(withinLineState))

	// Construct the full parse func.
	return initialState(
		startOfLineState,
		parseTaskToDo.
			Or(parseTaskInProgress).
			Or(parseTaskCompleted).
			Or(parseTaskBlocked).
			Or(parseCodeBlock).
			Or(parseTildeSeparator).
			Or(parseEndOfLine).
			Or(consumeUntilNextParseable),
	)
}
