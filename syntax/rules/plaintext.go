package rules

import "github.com/aretext/aretext/syntax/parser"

func PlaintextRules() []parser.TokenizerRule {
	return []parser.TokenizerRule{
		{
			Regexp:    `\p{P}`,
			TokenRole: parser.TokenRolePunctuation,
		},
		{
			Regexp:    `(\p{L}|\p{Nd})+`,
			TokenRole: parser.TokenRoleWord,
		},
	}
}
