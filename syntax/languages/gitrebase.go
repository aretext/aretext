package languages

import (
	"unicode"

	"github.com/aretext/aretext/syntax/parser"
)

// GitRebaseParseFunc parses a git rebase.
func GitRebaseParseFunc() parser.Func {
	keywords := []string{
		"pick", "reword", "edit", "squash", "fixup",
		"exec", "break", "drop", "label", "reset", "merge",
		"p", "r", "e", "s", "f", "e", "b", "d", "l", "t", "m",
	}

	isAlphaNumPunct := func(r rune) bool {
		return unicode.IsLetter(r) || unicode.IsDigit(r) || unicode.IsPunct(r)
	}

	return consumeRunesLike(isAlphaNumPunct).
		MapWithInput(recognizeKeywordOrConsume(keywords, true)).
		Map(func(result parser.Result) parser.Result {
			if len(result.ComputedTokens) == 0 {
				// Fail if we didn't recognize a token so the parser
				// falls back to the git commit parser.
				return parser.FailedResult
			}
			return result
		}).
		Or(GitCommitParseFunc())
}
