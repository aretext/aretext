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
			},
		},
		{
			name: "doc with keys, number, and comment",
			inputString: `- foo: # this is a test
							bar: 123`,
			expectedTokens: []TokenWithText{
				{Text: `foo:`, Role: parser.TokenRoleKey},
				{Text: `# this is a test`, Role: parser.TokenRoleComment},
				{Text: `bar:`, Role: parser.TokenRoleKey},
				{Text: `123`, Role: parser.TokenRoleNumber},
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
