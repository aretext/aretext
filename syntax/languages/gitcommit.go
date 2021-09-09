package languages

import (
	"github.com/aretext/aretext/syntax/parser"
)

// GitCommitParseFunc parses a git commit.
func GitCommitParseFunc() parser.Func {
	parseCommentLine := consumeString("#").
		ThenMaybe(consumeToNextLineFeed).
		Map(recognizeToken(parser.TokenRoleComment))
	return parseCommentLine.Or(consumeToNextLineFeed)
}
