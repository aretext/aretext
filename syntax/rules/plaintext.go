package rules

import "github.com/aretext/aretext/syntax/parser"

var PlaintextRules []parser.TokenizerRule

func init() {
	PlaintextRules = []parser.TokenizerRule{
		{
			Regexp:    `\p{P}`,
			TokenRole: parser.TokenRolePunctuation,
		},
		{
			Regexp:    `(\p{L}|\p{Nd})+`,
			TokenRole: parser.TokenRoleIdentifier,
		},
	}
}
