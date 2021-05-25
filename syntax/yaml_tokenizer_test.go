package syntax

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/aretext/aretext/syntax/parser"
)

func TestYamlTokenizer(t *testing.T) {
	testCases := []struct {
		name           string
		inputString    string
		expectedTokens []TokenWithText
	}{
		{
			name:        "single line comment",
			inputString: `# abc`,
			expectedTokens: []TokenWithText{
				{Text: `# abc`, Role: parser.TokenRoleComment},
			},
		},
		{
			name:        "key without quotes",
			inputString: "abc: xyz",
			expectedTokens: []TokenWithText{
				{Text: `abc:`, Role: parser.TokenRoleKey},
				{Text: `xyz`, Role: parser.TokenRoleWord},
			},
		},
		{
			name:        "key with single-quoted string",
			inputString: "'abc': xyz",
			expectedTokens: []TokenWithText{
				{Text: `'abc':`, Role: parser.TokenRoleKey},
				{Text: `xyz`, Role: parser.TokenRoleWord},
			},
		},
		{
			name:        "empty single-quoted string",
			inputString: `''`,
			expectedTokens: []TokenWithText{
				{Text: `'`, Role: parser.TokenRoleStringQuote},
				{Text: `'`, Role: parser.TokenRoleStringQuote},
			},
		},
		{
			name:        "single-quoted string",
			inputString: `'abc'`,
			expectedTokens: []TokenWithText{
				{Text: `'`, Role: parser.TokenRoleStringQuote},
				{Text: `abc`, Role: parser.TokenRoleString},
				{Text: `'`, Role: parser.TokenRoleStringQuote},
			},
		},
		{
			name:        "single-quoted string with escaped quote",
			inputString: `'ab''c'`,
			expectedTokens: []TokenWithText{
				{Text: `'`, Role: parser.TokenRoleStringQuote},
				{Text: `ab''c`, Role: parser.TokenRoleString},
				{Text: `'`, Role: parser.TokenRoleStringQuote},
			},
		},
		{
			name: "doc with keys, number, and comment",
			inputString: `- foo: # this is a test
							bar: 123`,
			expectedTokens: []TokenWithText{
				{Text: `-`, Role: parser.TokenRolePunctuation},
				{Text: `foo:`, Role: parser.TokenRoleKey},
				{Text: `# this is a test`, Role: parser.TokenRoleComment},
				{Text: `bar:`, Role: parser.TokenRoleKey},
				{Text: `123`, Role: parser.TokenRoleNumber},
			},
		},
		{
			name:        "unquoted string with number suffix",
			inputString: "foo: v0.1.2",
			expectedTokens: []TokenWithText{
				{Text: `foo:`, Role: parser.TokenRoleKey},
				{Text: `v0`, Role: parser.TokenRoleWord},
				{Text: `.`, Role: parser.TokenRolePunctuation},
				{Text: `1`, Role: parser.TokenRoleWord},
				{Text: `.`, Role: parser.TokenRolePunctuation},
				{Text: `2`, Role: parser.TokenRoleWord},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tokens, err := ParseTokensWithText(LanguageYaml, tc.inputString)
			require.NoError(t, err)
			assert.Equal(t, tc.expectedTokens, tokens)
		})
	}
}
