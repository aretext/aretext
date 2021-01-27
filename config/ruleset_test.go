package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfigForPath(t *testing.T) {
	testCases := []struct {
		name           string
		rules          []Rule
		path           string
		expectedConfig Config
	}{
		{
			name:  "no rules, default config",
			rules: nil,
			path:  "test.go",
			expectedConfig: Config{
				SyntaxLanguage: DefaultSyntaxLanguage,
				TabSize:        DefaultTabSize,
				TabExpand:      DefaultTabExpand,
				AutoIndent:     DefaultAutoIndent,
				MenuCommands:   []MenuCommandConfig{},
			},
		},
		{
			name: "rule matches, set syntax language",
			rules: []Rule{
				Rule{
					Name:    "json",
					Pattern: "**/*.json",
					Config: map[string]interface{}{
						"syntaxLanguage": "json",
					},
				},
				Rule{
					Name:    "mismatched rule",
					Pattern: "**/*.txt",
					Config: map[string]interface{}{
						"syntaxLanguage": "undefined",
					},
				},
			},
			path: "test.json",
			expectedConfig: Config{
				SyntaxLanguage: "json",
				TabSize:        DefaultTabSize,
				TabExpand:      DefaultTabExpand,
				AutoIndent:     DefaultAutoIndent,
				MenuCommands:   []MenuCommandConfig{},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rs := RuleSet{Rules: tc.rules}
			c := rs.ConfigForPath(tc.path)
			assert.Equal(t, tc.expectedConfig, c)
		})
	}
}
