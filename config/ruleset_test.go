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
					Config:  Config{SyntaxLanguage: "json"}.ToPartial(),
				},
				Rule{
					Name:    "mismatched rule",
					Pattern: "**/*.txt",
					Config:  Config{SyntaxLanguage: "undefined"}.ToPartial(),
				},
			},
			path: "test.json",
			expectedConfig: Config{
				SyntaxLanguage: "json",
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

func TestValidateRuleSet(t *testing.T) {
	validTabSize := 8
	invalidTabSize := 0

	testCases := []struct {
		name              string
		ruleSet           *RuleSet
		expectValid       bool
		expectErrContains string
	}{
		{
			name: "valid",
			ruleSet: &RuleSet{
				Rules: []Rule{
					Rule{
						Name:    "test",
						Pattern: "**",
						Config: PartialConfig{
							TabSize: &validTabSize,
						},
					},
				},
			},
			expectValid: true,
		},
		{
			name: "invalid",
			ruleSet: &RuleSet{
				Rules: []Rule{
					Rule{
						Name:    "test",
						Pattern: "**",
						Config: PartialConfig{
							TabSize: &invalidTabSize,
						},
					},
				},
			},
			expectValid:       false,
			expectErrContains: "Validation error in config rule test: field TabSize failed validator IntGreaterThan(0)",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.ruleSet.Validate()
			if tc.expectValid {
				assert.NoError(t, err)
			} else {
				assert.Contains(t, err.Error(), tc.expectErrContains)
			}
		})
	}
}
