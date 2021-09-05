package languages

import (
	"github.com/aretext/aretext/syntax2/parser"
)

// GitCommitParseFunc parses a git commit.
func GitCommitParseFunc() parser.Func {
	parseCommentLine := consumeString("#").
		Then(consumeToEndOfLine).
		Map(recognizeToken(parser.TokenRoleComment))
	return parseCommentLine.Or(consumeToEndOfLine)
}
