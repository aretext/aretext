package languages

import (
	"unicode"

	"github.com/aretext/aretext/syntax/parser"
)

type todoTxtParseState struct {
	AtStartOfLine bool
}

func (s todoTxtParseState) Equals(other parser.State) bool {
	otherState, ok := other.(todoTxtParseState)
	return ok && s == otherState
}

const (
	todoTxtCompletedTaskRole = parser.TokenRoleCustom1
	todoTxtPriorityRole      = parser.TokenRoleCustom2
	todoTxtDateRole          = parser.TokenRoleCustom3
	todoTxtProjectTagRole    = parser.TokenRoleCustom4
	todoTxtContextTagRole    = parser.TokenRoleCustom5
	todoTxtKeyTagRole        = parser.TokenRoleCustom6
	todoTxtValTagRole        = parser.TokenRoleCustom7
)

// TodoTxtParseFunc returns a parse func for the todo.txt file format.
// See https://github.com/todotxt/todo.txt for details.
func TodoTxtParseFunc() parser.Func {
	// States to keep track of whether we're at the start of a line.
	startOfLineState := todoTxtParseState{AtStartOfLine: true}
	withinLineState := todoTxtParseState{AtStartOfLine: false}

	// Helper parse funcs.
	consumeToEndOfWord := consumeRunesLike(func(r rune) bool {
		return !unicode.IsSpace(r)
	})

	consumeNumbers := func(n int) parser.Func {
		return func(iter parser.TrackingRuneIter, state parser.State) parser.Result {
			for i := 0; i < n; i++ {
				r, err := iter.NextRune()
				if !(err == nil && (r >= '0' && r <= '9')) {
					return parser.FailedResult
				}
			}
			return parser.Result{
				NumConsumed: uint64(n),
				NextState:   state,
			}
		}
	}

	// Parse a completed task. This is an "x " at the start of a line, then all chars to the end of the line.
	// Transitions: startOfLineState -> startOfLineState
	parseCompletedTask := matchState(
		startOfLineState,
		consumeString("x ").
			ThenMaybe(consumeToNextLineFeed).
			Map(recognizeToken(todoTxtCompletedTaskRole)),
	)

	// Parse a priority, like "(A)" at the start of a line, followed by a space.
	// Transitions: startOfLineState -> withinLineState
	parsePriority := matchState(
		startOfLineState,
		consumeString("(").
			Then(consumeSingleRuneLike(func(r rune) bool { return r >= 'A' && r <= 'Z' })).
			Then(consumeString(")")).
			Map(recognizeToken(todoTxtPriorityRole)).
			Then(consumeString(" ")).
			Map(setState(withinLineState)))

	// Parse date formatted as YYYY-MM-DD.
	// Transitions: any -> withinLineState
	parseDate := consumeNumbers(4).
		Then(consumeString("-")).
		Then(consumeNumbers(2)).
		Then(consumeString("-")).
		Then(consumeNumbers(2)).
		Map(recognizeToken(todoTxtDateRole)).
		Map(setState(withinLineState))

	// Parse project tag like "+project".
	// Transitions: any -> withinLineState
	parseProjectTag := consumeString("+").
		Then(consumeToEndOfWord).
		Map(recognizeToken(todoTxtProjectTagRole)).
		Map(setState(withinLineState))

	// Parse context tag like "@context"
	// Transitions: any -> withinLineState
	parseContextTag := consumeString("@").
		Then(consumeToEndOfWord).
		Map(recognizeToken(todoTxtContextTagRole)).
		Map(setState(withinLineState))

	// Parse a key-value tag like "key:val"
	// Transitions: any -> withinLineState
	parseKey := consumeRunesLike(func(r rune) bool {
		return !unicode.IsSpace(r) && r != ':'
	}).Then(consumeString(":")).
		Map(recognizeToken(todoTxtKeyTagRole))

	parseVal := consumeToEndOfWord.
		Map(recognizeToken(todoTxtValTagRole))

	parseKeyValTag := parseKey.Then(parseVal).
		Map(setState(withinLineState))

	// Fallback to transition from startOfLineState -> withinLineState
	// if none of the other parse funcs succeed.
	parseOtherStartOfLine := matchState(
		startOfLineState,
		consumeSingleRuneLike(func(r rune) bool { return r != '\n' }).
			Map(setState(withinLineState)),
	)

	// Transition back to startOfLineState once we hit a newline.
	parseNewline := consumeString("\n").Map(setState(startOfLineState))

	// Construct parse func for incomplete tasks.
	parseIncompleteTask := parsePriority.
		Or(parseDate).
		Or(parseProjectTag).
		Or(parseContextTag).
		Or(parseKeyValTag).
		Or(consumeToEndOfWord).
		Or(parseOtherStartOfLine).
		Or(parseNewline)

	// Construct the final parse func (either complete or incomplete tasks).
	return initialState(startOfLineState, parseCompletedTask.Or(parseIncompleteTask))
}
