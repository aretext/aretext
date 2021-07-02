package rules

import (
	"github.com/aretext/aretext/syntax/parser"
)

func DevlogRules() []parser.TokenizerRule {
	// Rules for the devlog file format.
	// https://devlog-cli.org/
	return []parser.TokenizerRule{
		// "To do" task
		{
			Regexp:    `(^|\n)\*`,
			TokenRole: parser.TokenRoleCustom1,
		},

		// "In progress" task
		{
			Regexp:    `(^|\n)\^[^\n]*`,
			TokenRole: parser.TokenRoleCustom2,
		},

		// "Completed" task
		{
			Regexp:    `(^|\n)\+[^\n]*`,
			TokenRole: parser.TokenRoleCustom3,
		},

		// "Blocked" task
		{
			Regexp:    `(^|\n)\-[^\n]*`,
			TokenRole: parser.TokenRoleCustom4,
		},

		// Code block
		{
			Regexp:    "```([^`]|`[^`]|``[^`])*```",
			TokenRole: parser.TokenRoleCustom5,
		},

		// Tilde separator
		{
			Regexp:    `(^|\n)~~~~*`,
			TokenRole: parser.TokenRoleCustom6,
		},
	}
}
