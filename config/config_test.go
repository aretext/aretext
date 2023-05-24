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
				LineWrap:       "character",
				MenuCommands:   []MenuCommandConfig{},
				Styles:         map[string]StyleConfig{},
				LineNumberMode: "absolute",
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
				LineWrap:       "character",
				MenuCommands:   []MenuCommandConfig{},
				LineNumberMode: "absolute",
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

func TestValidateConfig(t *testing.T) {
	testCases := []struct {
		name         string
		updateFunc   func(c *Config)
		expectErrMsg string
	}{
		{
			name:         "default config is valid",
			updateFunc:   nil,
			expectErrMsg: "",
		},
		{
			name: "tabSize zero is invalid",
			updateFunc: func(c *Config) {
				c.TabSize = 0
			},
			expectErrMsg: "TabSize must be greater than zero",
		},
		{
			name: "lineWrap is invalid",
			updateFunc: func(c *Config) {
				c.LineWrap = "invalid"
			},
			expectErrMsg: `LineWrap must be either "character" or "word"`,
		},
		{
			name: "lineNumberMode is invalid",
			updateFunc: func(c *Config) {
				c.LineNumberMode = "invalid"
			},
			expectErrMsg: `LineNumberMode must be either "absolute" or "relative"`,
		},
		{
			name: "menu name is empty",
			updateFunc: func(c *Config) {
				c.MenuCommands = append(c.MenuCommands, MenuCommandConfig{
					Name:     "",
					ShellCmd: "echo 'hello'",
					Mode:     "silent",
				})
			},
			expectErrMsg: `Menu name cannot be empty`,
		},
		{
			name: "menu shellCmd is empty",
			updateFunc: func(c *Config) {
				c.MenuCommands = append(c.MenuCommands, MenuCommandConfig{
					Name:     "testcmd",
					ShellCmd: "",
					Mode:     "silent",
				})
			},
			expectErrMsg: `Menu command "testcmd" shellCmd cannot be empty`,
		},
		{
			name: "menu mode is invalid",
			updateFunc: func(c *Config) {
				c.MenuCommands = append(c.MenuCommands, MenuCommandConfig{
					Name:     "testcmd",
					ShellCmd: "echo 'hello'",
					Mode:     "invalid",
				})
			},
			expectErrMsg: `Menu command "testcmd" must have mode set to either "silent", "terminal", "insert", "insertChoice", "fileLocations", or "workingDir"`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := ConfigFromUntypedMap(nil)
			if tc.updateFunc != nil {
				tc.updateFunc(&config)
			}

			err := config.Validate()
			if tc.expectErrMsg == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.expectErrMsg)
			}
		})
	}
}
