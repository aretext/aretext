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
			name:           "no rules, default config",
			rules:          nil,
			path:           "test.go",
			expectedConfig: DefaultConfig(),
		},
		{
			name: "rule matches, set syntax language",
			rules: []Rule{
				Rule{
					Name:    "json",
					Pattern: "**/*.json",
					Config:  Config{SyntaxLanguage: "json"},
				},
				Rule{
					Name:    "mismatched rule",
					Pattern: "**/*.txt",
					Config:  Config{SyntaxLanguage: "undefined"},
				},
			},
			path:           "test.json",
			expectedConfig: Config{SyntaxLanguage: "json"},
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
