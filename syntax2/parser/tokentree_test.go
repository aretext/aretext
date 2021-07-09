package parser

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTokenTreeInsert(t *testing.T) {
	testCases := []struct {
		name           string
		tokens         []Token
		expectedTokens []Token
	}{
		{
			name:           "empty",
			tokens:         []Token{},
			expectedTokens: []Token{},
		},
		{
			name: "single token",
			tokens: []Token{
				{StartPos: 1, EndPos: 2, LookaheadPos: 3},
			},
			expectedTokens: []Token{
				{StartPos: 1, EndPos: 2, LookaheadPos: 3},
			},
		},
		{
			name: "two tokens, in ascending order",
			tokens: []Token{
				{StartPos: 1, EndPos: 2, LookaheadPos: 3},
				{StartPos: 2, EndPos: 3, LookaheadPos: 4},
			},
			expectedTokens: []Token{
				{StartPos: 1, EndPos: 2, LookaheadPos: 3},
				{StartPos: 2, EndPos: 3, LookaheadPos: 4},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var tree *TokenTree
			for _, tok := range tc.tokens {
				tree = tree.Insert(tok)
			}
			tokens := tree.IterFromPosition(0).Collect()
			assert.Equal(t, tc.expectedTokens, tokens)
		})
	}
}
