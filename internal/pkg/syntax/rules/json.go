package rules

import (
	"github.com/wedaly/aretext/internal/pkg/syntax/parser"
)

var JsonRules []parser.TokenizerRule

func init() {
	JsonRules = []parser.TokenizerRule{
		{
			Regexp:    `true|false|null`,
			TokenRole: parser.TokenRoleKeyword,
		},
		{
			Regexp:    `-?[0-9]+(\.[0-9]+)?((e|E)-?[0-9]+)?`,
			TokenRole: parser.TokenRoleNumber,
		},
		{
			Regexp:    `"([^\"\n]|\\")*"`,
			TokenRole: parser.TokenRoleString,
		},
	}
}
