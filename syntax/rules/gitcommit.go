package rules

import "github.com/aretext/aretext/syntax/parser"

func GitCommitRules() []parser.TokenizerRule {
	plaintextRules := PlaintextRules()
	gitCommitRules := []parser.TokenizerRule{
		{
			Regexp:    "(^|\n)#[^\n]*",
			TokenRole: parser.TokenRoleNone,
			SubRules: []parser.TokenizerRule{
				{
					Regexp:    `#[^\n]*`,
					TokenRole: parser.TokenRoleComment,
					SubRules: []parser.TokenizerRule{
						{
							Regexp:    `^#`,
							TokenRole: parser.TokenRoleCommentDelimiter,
						},
					},
				},
			},
		},
	}
	return append(gitCommitRules, plaintextRules...)
}
