package parser

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTokenizeAll(t *testing.T) {
	tokenizer, err := GenerateTokenizer([]TokenizerRule{
		{
			Regexp:    `\+`,
			TokenRole: TokenRoleOperator,
		},
		{
			Regexp:    "def|return",
			TokenRole: TokenRoleKeyword,
		},
		{
			Regexp:    "[a-zA-Z][a-zA-Z0-9]+",
			TokenRole: TokenRoleIdentifier,
		},
		{
			Regexp:    "[0-9]+",
			TokenRole: TokenRoleNumber,
		},
	})
	require.NoError(t, err)

	testCases := []struct {
		name           string
		inputText      string
		expectedTokens []Token
	}{
		{
			name:           "empty string",
			inputText:      "",
			expectedTokens: []Token{},
		},
		{
			name:      "multiple tokens",
			inputText: "def foo():\n    return bar + 10",
			expectedTokens: []Token{
				// Whitespace and punctuation are skipped because they don't match any rules.
				// "def"
				{StartPos: 0, EndPos: 3, Role: TokenRoleKeyword},

				// "foo"
				{StartPos: 4, EndPos: 7, Role: TokenRoleIdentifier},

				// "return"
				{StartPos: 15, EndPos: 21, Role: TokenRoleKeyword},

				// "bar"
				{StartPos: 22, EndPos: 25, Role: TokenRoleIdentifier},

				// "+"
				{StartPos: 26, EndPos: 27, Role: TokenRoleOperator},

				// "10"
				{StartPos: 28, EndPos: 30, Role: TokenRoleNumber},
			},
		},
		{
			name:      "identifier with keyword prefix",
			inputText: "defIdentifier",
			expectedTokens: []Token{
				{StartPos: 0, EndPos: 13, Role: TokenRoleIdentifier},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			textLen := uint64(len(tc.inputText))
			r := strings.NewReader(tc.inputText)
			tokenTree, err := tokenizer.TokenizeAll(r, textLen)
			require.NoError(t, err)
			tokens := tokenTree.IterFromPosition(0).Collect()
			assert.Equal(t, tc.expectedTokens, tokens)
		})
	}
}
