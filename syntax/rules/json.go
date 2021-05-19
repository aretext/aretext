package rules

import (
	"strings"

	"github.com/aretext/aretext/syntax/parser"
)

func JsonRules() []parser.TokenizerRule {
	stringPattern := `""|"([^\"\n]|\\")*[^\\\n]"`
	return []parser.TokenizerRule{
		{
			Regexp:    `true|false|null`,
			TokenRole: parser.TokenRoleKeyword,
		},
		{
			Regexp:    `-?[0-9]+(\.[0-9]+)?((e|E)-?[0-9]+)?`,
			TokenRole: parser.TokenRoleNumber,
		},
		{
			Regexp:    stringPattern,
			TokenRole: parser.TokenRoleString,
			SubRules: []parser.TokenizerRule{
				{
					Regexp:    `^"`,
					TokenRole: parser.TokenRoleStringQuote,
				},
				{
					Regexp:    `"$`,
					TokenRole: parser.TokenRoleStringQuote,
				},
			},
		},
		{
			Regexp:    stringPattern + `[ \t]*:`,
			TokenRole: parser.TokenRoleKey,
		},
		{
			Regexp:    strings.Join([]string{`\{`, `\}`, `\[`, `\]`, `,`}, "|"),
			TokenRole: parser.TokenRolePunctuation,
		},

		// This prevents the number and keyword rules from matching substrings of a symbol.
		{
			Regexp:    `-?[a-zA-Z0-9_]+`,
			TokenRole: parser.TokenRoleNone,
		},
	}
}
