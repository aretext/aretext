package rules

import (
	"github.com/aretext/aretext/syntax/parser"
)

func JsonRules() []parser.TokenizerRule {
	const tokenRoleKey = parser.TokenRoleCustom1
	stringPattern := `""|"([^\"\n]|\\")*(\\\\|[^\\\n])"`
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
		},
		{
			Regexp:    stringPattern + `[ \t]*:`,
			TokenRole: tokenRoleKey,
		},

		// This prevents the number and keyword rules from matching substrings of a symbol.
		{
			Regexp:    `-?([a-zA-Z0-9._\-])+`,
			TokenRole: parser.TokenRoleNone,
		},
	}
}
