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
				// "def"
				{StartPos: 0, EndPos: 3, LookaheadPos: 4, Role: TokenRoleKeyword},
				{StartPos: 3, EndPos: 4, LookaheadPos: 5, Role: TokenRoleNone},

				// "foo"
				{StartPos: 4, EndPos: 7, LookaheadPos: 8, Role: TokenRoleIdentifier},
				{StartPos: 7, EndPos: 15, LookaheadPos: 16, Role: TokenRoleNone},

				// "return"
				{StartPos: 15, EndPos: 21, LookaheadPos: 22, Role: TokenRoleKeyword},
				{StartPos: 21, EndPos: 22, LookaheadPos: 23, Role: TokenRoleNone},

				// "bar"
				{StartPos: 22, EndPos: 25, LookaheadPos: 26, Role: TokenRoleIdentifier},
				{StartPos: 25, EndPos: 26, LookaheadPos: 27, Role: TokenRoleNone},

				// "+"
				{StartPos: 26, EndPos: 27, LookaheadPos: 28, Role: TokenRoleOperator},
				{StartPos: 27, EndPos: 28, LookaheadPos: 29, Role: TokenRoleNone},

				// "10"
				{StartPos: 28, EndPos: 30, LookaheadPos: 30, Role: TokenRoleNumber},
			},
		},
		{
			name:      "identifier with keyword prefix",
			inputText: "defIdentifier",
			expectedTokens: []Token{
				{StartPos: 0, EndPos: 13, LookaheadPos: 13, Role: TokenRoleIdentifier},
			},
		},
		{
			name:      "all whitespace, no matches",
			inputText: "  ",
			expectedTokens: []Token{
				{StartPos: 0, EndPos: 2, LookaheadPos: 2, Role: TokenRoleNone},
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
