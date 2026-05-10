package config

import (
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
					Pattern: "**/*.json",
					Config: map[string]any{
						"syntaxLanguage": "json",
					},
				},
				{
					Name:    "mismatched rule",
					Pattern: "**/*.txt",
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
		{
			name: "matching rules merge system clipboard config",
			ruleSet: []Rule{
				{
					Name:    "clipboard",
					Pattern: "**/*.txt",
					Config: map[string]any{
						"systemClipboard": map[string]any{
							"copyCmd":  "wl-copy",
							"pasteCmd": "wl-paste --no-newline",
						},
					},
				},
				{
					Name:    "clipboard use by default",
					Pattern: "**/*.txt",
					Config: map[string]any{
						"systemClipboard": map[string]any{
							"useByDefault": true,
							"copyCmd":      "wl-copy",
							"pasteCmd":     "wl-paste --no-newline",
						},
					},
				},
			},
			path: "test.txt",
			expectedConfig: Config{
				SyntaxLanguage: DefaultSyntaxLanguage,
				TabSize:        DefaultTabSize,
				TabExpand:      DefaultTabExpand,
				AutoIndent:     DefaultAutoIndent,
				LineWrap:       DefaultLineWrap,
				LineNumberMode: string(DefaultLineNumberMode),
				MenuCommands:   []MenuCommandConfig{},
				SystemClipboard: &SystemClipboardConfig{
					UseByDefault: true,
					CopyCmd:      "wl-copy",
					PasteCmd:     "wl-paste --no-newline",
				},
				Styles: map[string]StyleConfig{},
			},
		},
		{
			name: "matching rule without system clipboard preserves prior system clipboard config",
			ruleSet: []Rule{
				{
					Name:    "clipboard",
					Pattern: "**/*.txt",
					Config: map[string]any{
						"systemClipboard": map[string]any{
							"useByDefault": true,
							"copyCmd":      "wl-copy",
							"pasteCmd":     "wl-paste --no-newline",
						},
					},
				},
				{
					Name:    "plain text",
					Pattern: "**/*.txt",
					Config: map[string]any{
						"syntaxLanguage": "plaintext",
					},
				},
			},
			path: "test.txt",
			expectedConfig: Config{
				SyntaxLanguage: "plaintext",
				TabSize:        DefaultTabSize,
				TabExpand:      DefaultTabExpand,
				AutoIndent:     DefaultAutoIndent,
				LineWrap:       DefaultLineWrap,
				LineNumberMode: string(DefaultLineNumberMode),
				MenuCommands:   []MenuCommandConfig{},
				SystemClipboard: &SystemClipboardConfig{
					UseByDefault: true,
					CopyCmd:      "wl-copy",
					PasteCmd:     "wl-paste --no-newline",
				},
				Styles: map[string]StyleConfig{},
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
