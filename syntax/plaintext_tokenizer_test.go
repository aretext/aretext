package syntax

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/aretext/aretext/syntax/parser"
)

func TestPlaintextTokenizer(t *testing.T) {
	testCases := []struct {
		name           string
		inputString    string
		expectedTokens []TokenWithText
	}{
		{
			name:        "words separated by spaces",
			inputString: "hello world",
			expectedTokens: []TokenWithText{
				{Text: "hello", Role: parser.TokenRoleIdentifier},
				{Text: "world", Role: parser.TokenRoleIdentifier},
			},
		},
		{
			name:        "punctuation",
			inputString: "\"hello, world!\"",
			expectedTokens: []TokenWithText{
				{Text: "\"", Role: parser.TokenRolePunctuation},
				{Text: "hello", Role: parser.TokenRoleIdentifier},
				{Text: ",", Role: parser.TokenRolePunctuation},
				{Text: "world", Role: parser.TokenRoleIdentifier},
				{Text: "!", Role: parser.TokenRolePunctuation},
				{Text: "\"", Role: parser.TokenRolePunctuation},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tokens, err := ParseTokensWithText(LanguagePlaintext, tc.inputString)
			require.NoError(t, err)
			assert.Equal(t, tc.expectedTokens, tokens)
		})
	}
}
