package rules

import "github.com/aretext/aretext/syntax/parser"

func GitCommitRules() []parser.TokenizerRule {
	return []parser.TokenizerRule{
		{
			Regexp:    "(^|\n)#[^\n]*",
			TokenRole: parser.TokenRoleNone,
			SubRules: []parser.TokenizerRule{
				{
					Regexp:    `#[^\n]*`,
					TokenRole: parser.TokenRoleComment,
				},
			},
		},
	}
}
