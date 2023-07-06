package config

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfigForPath(t *testing.T) {
	testCases := []struct {
		name           string
		ruleSet        RuleSet
		path           string
		expectedConfig Config
	}{
		{
			name:    "no rules, default config",
			ruleSet: nil,
			path:    "test.go",
			expectedConfig: Config{
				SyntaxLanguage: DefaultSyntaxLanguage,
				TabSize:        DefaultTabSize,
				TabExpand:      DefaultTabExpand,
				AutoIndent:     DefaultAutoIndent,
				LineWrap:       DefaultLineWrap,
				LineNumberMode: string(DefaultLineNumberMode),
				MenuCommands:   []MenuCommandConfig{},
				Styles:         map[string]StyleConfig{},
			},
		},
		{
			name: "rule matches, set syntax language",
			ruleSet: []Rule{
				{
					Name:    "json",
					Pattern: filepath.FromSlash("**/*.json"),
					Config: map[string]any{
						"syntaxLanguage": "json",
					},
				},
				{
					Name:    "mismatched rule",
					Pattern: filepath.FromSlash("**/*.txt"),
					Config: map[string]any{
						"syntaxLanguage": "undefined",
					},
				},
			},
			path: "test.json",
			expectedConfig: Config{
				SyntaxLanguage: "json",
				TabSize:        DefaultTabSize,
				TabExpand:      DefaultTabExpand,
				LineWrap:       DefaultLineWrap,
				AutoIndent:     DefaultAutoIndent,
				LineNumberMode: string(DefaultLineNumberMode),
				MenuCommands:   []MenuCommandConfig{},
				Styles:         map[string]StyleConfig{},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			c := tc.ruleSet.ConfigForPath(tc.path)
			assert.Equal(t, tc.expectedConfig, c)
		})
	}
}
