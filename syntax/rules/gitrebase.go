package rules

import (
	"fmt"
	"strings"

	"github.com/aretext/aretext/syntax/parser"
)

func GitRebaseRules() []parser.TokenizerRule {
	gitCommitRules := GitCommitRules()
	gitRebaseKeywords := []string{
		"p", "pick",
		"r", "reword",
		"e", "edit",
		"s", "squash",
		"f", "fixup",
		"x", "exec",
		"b", "break",
		"d", "drop",
		"l", "label",
		"t", "reset",
		"m", "merge",
	}
	gitRebaseRules := []parser.TokenizerRule{
		{
			Regexp:    fmt.Sprintf("(^|\n)(%s)", strings.Join(gitRebaseKeywords, `|`)),
			TokenRole: parser.TokenRoleNone,
			SubRules: []parser.TokenizerRule{
				{
					Regexp:    `[^\n]+`,
					TokenRole: parser.TokenRoleKeyword,
				},
			},
		},
		// This prevents the keyword rule from matching substrings of a symbol.
		{
			Regexp:    `-?([a-zA-Z0-9._\-])+`,
			TokenRole: parser.TokenRoleNone,
		},
	}
	return append(gitRebaseRules, gitCommitRules...)
}
