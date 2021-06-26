package rules

import "github.com/aretext/aretext/syntax/parser"

func YamlRules() []parser.TokenizerRule {
	const tokenRoleKey = parser.TokenRoleCustom1
	jsonRules := JsonRules() // YAML is a superset of JSON.
	singleQuoteStringPattern := `'([^']|'')*'`
	yamlRules := []parser.TokenizerRule{
		{
			Regexp:    singleQuoteStringPattern,
			TokenRole: parser.TokenRoleString,
			SubRules: []parser.TokenizerRule{
				{
					Regexp:    `^'`,
					TokenRole: parser.TokenRoleStringQuote,
				},
				{
					Regexp:    `'$`,
					TokenRole: parser.TokenRoleStringQuote,
				},
			},
		},
		{
			Regexp:    singleQuoteStringPattern + `[ \t]*:`,
			TokenRole: tokenRoleKey,
		},
		{
			Regexp:    "#[^\n]*",
			TokenRole: parser.TokenRoleComment,
		},
		{
			Regexp:    `[a-zA-Z]([a-zA-Z0-9._\-])*\t*:`,
			TokenRole: tokenRoleKey,
		},
	}
	return append(yamlRules, jsonRules...)
}
