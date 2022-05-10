package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfigFromUntypedMap(t *testing.T) {
	testCases := []struct {
		name     string
		input    map[string]any
		expected Config
	}{
		{
			name:  "empty map",
			input: map[string]any{},
			expected: Config{
				SyntaxLanguage: "plaintext",
				TabSize:        4,
				MenuCommands:   []MenuCommandConfig{},
				Styles:         map[string]StyleConfig{},
			},
		},
		{
			name: "custom styles",
			input: map[string]any{
				"syntaxLanguage": "customLang",
				"styles": map[string]any{
					"lineNum": map[string]any{
						"color": "olive",
					},
					"tokenCustom1": map[string]any{
						"color":  "teal",
						"bold":   true,
						"italic": true,
					},
					"tokenCustom2": map[string]any{
						"color":     "fuchsia",
						"underline": true,
					},
					"tokenCustom3": map[string]any{
						"color":         "red",
						"strikethrough": true,
					},
					"tokenCustom4": map[string]any{
						"backgroundColor": "black",
					},
				},
			},
			expected: Config{
				SyntaxLanguage: "customLang",
				TabSize:        4,
				MenuCommands:   []MenuCommandConfig{},
				Styles: map[string]StyleConfig{
					"lineNum": {
						Color: "olive",
					},
					"tokenCustom1": {
						Color:  "teal",
						Bold:   true,
						Italic: true,
					},
					"tokenCustom2": {
						Color:     "fuchsia",
						Underline: true,
					},
					"tokenCustom3": {
						Color:         "red",
						StrikeThrough: true,
					},
					"tokenCustom4": {
						BackgroundColor: "black",
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := ConfigFromUntypedMap(tc.input)
			assert.Equal(t, tc.expected, config)
		})
	}
}
