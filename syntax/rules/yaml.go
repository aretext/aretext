package rules

import "github.com/aretext/aretext/syntax/parser"

func YamlRules() []parser.TokenizerRule {
	// Most YAML syntax cannot be recognized by a regular language.
	// Highlight the tokens that we can recognize accurately and ignore the rest.
	yamlRules := []parser.TokenizerRule{
		{
			Regexp:    "#[^\n]*",
			TokenRole: parser.TokenRoleComment,
			SubRules: []parser.TokenizerRule{
				{
					Regexp:    `^#`,
					TokenRole: parser.TokenRoleCommentDelimiter,
				},
			},
		},
	}
	return append(yamlRules, PlaintextRules()...)
}
