package rules

import "github.com/aretext/aretext/syntax/parser"

func YamlRules() []parser.TokenizerRule {
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
			TokenRole: parser.TokenRoleKey,
		},
		{
			Regexp:    "#[^\n]*",
			TokenRole: parser.TokenRoleComment,
		},
		{
			Regexp:    `[a-zA-Z][a-zA-Z0-9_]*\t*:`,
			TokenRole: parser.TokenRoleKey,
		},
	}
	return append(yamlRules, jsonRules...)
}
