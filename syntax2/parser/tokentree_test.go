package parser

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

func sortedTokens(tokens []Token) []Token {
	result := make([]Token, len(tokens))
	copy(result, tokens)
	sort.Slice(result, func(i, j int) bool {
		return result[i].StartPos < result[j].StartPos
	})
	return result
}

func TestTokenTreeInsert(t *testing.T) {
	testCases := []struct {
		name   string
		tokens []Token
	}{
		{
			name:   "empty",
			tokens: []Token{},
		},
		{
			name: "single token",
			tokens: []Token{
				{StartPos: 1, EndPos: 2, LookaheadPos: 3},
			},
		},
		{
			name: "single token from position zero",
			tokens: []Token{
				{StartPos: 0, EndPos: 1, LookaheadPos: 1},
			},
		},
		{
			name: "two tokens, in ascending order",
			tokens: []Token{
				{StartPos: 1, EndPos: 2, LookaheadPos: 3},
				{StartPos: 2, EndPos: 3, LookaheadPos: 4},
			},
		},
		{
			name: "two tokens, in descending order",
			tokens: []Token{
				{StartPos: 2, EndPos: 3, LookaheadPos: 4},
				{StartPos: 1, EndPos: 2, LookaheadPos: 3},
			},
		},
		{
			name: "many tokens, random order",
			tokens: []Token{
				{StartPos: 2, EndPos: 3, LookaheadPos: 3},
				{StartPos: 8, EndPos: 9, LookaheadPos: 9},
				{StartPos: 4, EndPos: 5, LookaheadPos: 5},
				{StartPos: 1, EndPos: 2, LookaheadPos: 2},
				{StartPos: 6, EndPos: 7, LookaheadPos: 7},
				{StartPos: 0, EndPos: 1, LookaheadPos: 1},
				{StartPos: 9, EndPos: 10, LookaheadPos: 10},
				{StartPos: 7, EndPos: 8, LookaheadPos: 8},
				{StartPos: 5, EndPos: 6, LookaheadPos: 6},
				{StartPos: 3, EndPos: 4, LookaheadPos: 4},
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
			expectedTokens := sortedTokens(tc.tokens)
			assert.Equal(t, expectedTokens, tokens)
		})
	}
}
