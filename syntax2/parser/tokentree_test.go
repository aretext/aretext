package parser

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

func treeForTokens(tokens []Token) *TokenTree {
	var tree *TokenTree
	for _, tok := range tokens {
		tree = tree.Insert(tok)
	}
	return tree
}

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
			name: "ten tokens, random order",
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
		{
			name: "twenty-five tokens, random order",
			tokens: []Token{
				{StartPos: 2, EndPos: 3, LookaheadPos: 3},
				{StartPos: 1, EndPos: 2, LookaheadPos: 2},
				{StartPos: 0, EndPos: 1, LookaheadPos: 1},
				{StartPos: 8, EndPos: 9, LookaheadPos: 9},
				{StartPos: 3, EndPos: 4, LookaheadPos: 4},
				{StartPos: 24, EndPos: 25, LookaheadPos: 25},
				{StartPos: 14, EndPos: 15, LookaheadPos: 15},
				{StartPos: 5, EndPos: 6, LookaheadPos: 6},
				{StartPos: 7, EndPos: 8, LookaheadPos: 8},
				{StartPos: 10, EndPos: 11, LookaheadPos: 11},
				{StartPos: 19, EndPos: 20, LookaheadPos: 20},
				{StartPos: 21, EndPos: 22, LookaheadPos: 22},
				{StartPos: 15, EndPos: 16, LookaheadPos: 16},
				{StartPos: 20, EndPos: 21, LookaheadPos: 21},
				{StartPos: 6, EndPos: 7, LookaheadPos: 7},
				{StartPos: 11, EndPos: 12, LookaheadPos: 12},
				{StartPos: 12, EndPos: 13, LookaheadPos: 13},
				{StartPos: 23, EndPos: 24, LookaheadPos: 24},
				{StartPos: 9, EndPos: 10, LookaheadPos: 10},
				{StartPos: 16, EndPos: 17, LookaheadPos: 17},
				{StartPos: 17, EndPos: 18, LookaheadPos: 18},
				{StartPos: 22, EndPos: 23, LookaheadPos: 23},
				{StartPos: 4, EndPos: 5, LookaheadPos: 5},
				{StartPos: 18, EndPos: 19, LookaheadPos: 19},
				{StartPos: 13, EndPos: 14, LookaheadPos: 14},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tree := treeForTokens(tc.tokens)
			tokens := tree.IterFromPosition(0).Collect()
			expectedTokens := sortedTokens(tc.tokens)
			assert.Equal(t, expectedTokens, tokens)
		})
	}
}

func TestTokenTreeIterFromPosition(t *testing.T) {
	testCases := []struct {
		name            string
		tokens          []Token
		pos             uint64
		expectHasNext   bool
		expectNextToken Token
	}{
		{
			name:          "empty",
			tokens:        nil,
			pos:           0,
			expectHasNext: false,
		},
		{
			name: "in-order, before first token",
			tokens: []Token{
				{StartPos: 1, EndPos: 2, LookaheadPos: 2},
				{StartPos: 2, EndPos: 3, LookaheadPos: 3},
			},
			pos:             0,
			expectHasNext:   true,
			expectNextToken: Token{StartPos: 1, EndPos: 2, LookaheadPos: 2},
		},
		{
			name: "in-order, start of first token",
			tokens: []Token{
				{StartPos: 1, EndPos: 2, LookaheadPos: 2},
				{StartPos: 2, EndPos: 3, LookaheadPos: 3},
			},
			pos:             1,
			expectHasNext:   true,
			expectNextToken: Token{StartPos: 1, EndPos: 2, LookaheadPos: 2},
		},
		{
			name: "in-order, middle of first token",
			tokens: []Token{
				{StartPos: 1, EndPos: 3, LookaheadPos: 3},
				{StartPos: 3, EndPos: 4, LookaheadPos: 4},
			},
			pos:             2,
			expectHasNext:   true,
			expectNextToken: Token{StartPos: 1, EndPos: 3, LookaheadPos: 3},
		},
		{
			name: "in-order, end of first token",
			tokens: []Token{
				{StartPos: 1, EndPos: 3, LookaheadPos: 3},
				{StartPos: 3, EndPos: 4, LookaheadPos: 4},
			},
			pos:             3,
			expectHasNext:   true,
			expectNextToken: Token{StartPos: 3, EndPos: 4, LookaheadPos: 4},
		},
		{
			name: "in-order, end of last token",
			tokens: []Token{
				{StartPos: 1, EndPos: 3, LookaheadPos: 3},
				{StartPos: 3, EndPos: 4, LookaheadPos: 4},
			},
			pos:           4,
			expectHasNext: false,
		},
		{
			name: "in-order, past end of last token",
			tokens: []Token{
				{StartPos: 1, EndPos: 3, LookaheadPos: 3},
				{StartPos: 3, EndPos: 4, LookaheadPos: 4},
			},
			pos:           5,
			expectHasNext: false,
		},
		{
			name: "reverse-order, before first token",
			tokens: []Token{
				{StartPos: 2, EndPos: 3, LookaheadPos: 3},
				{StartPos: 1, EndPos: 2, LookaheadPos: 2},
			},
			pos:             0,
			expectHasNext:   true,
			expectNextToken: Token{StartPos: 1, EndPos: 2, LookaheadPos: 2},
		},
		{
			name: "reverse-order, start of first token",
			tokens: []Token{
				{StartPos: 2, EndPos: 3, LookaheadPos: 3},
				{StartPos: 1, EndPos: 2, LookaheadPos: 2},
			},
			pos:             1,
			expectHasNext:   true,
			expectNextToken: Token{StartPos: 1, EndPos: 2, LookaheadPos: 2},
		},
		{
			name: "reverse-order, middle of first token",
			tokens: []Token{
				{StartPos: 3, EndPos: 4, LookaheadPos: 4},
				{StartPos: 1, EndPos: 3, LookaheadPos: 3},
			},
			pos:             2,
			expectHasNext:   true,
			expectNextToken: Token{StartPos: 1, EndPos: 3, LookaheadPos: 3},
		},
		{
			name: "reverse-order, end of first token",
			tokens: []Token{
				{StartPos: 3, EndPos: 4, LookaheadPos: 4},
				{StartPos: 1, EndPos: 3, LookaheadPos: 3},
			},
			pos:             3,
			expectHasNext:   true,
			expectNextToken: Token{StartPos: 3, EndPos: 4, LookaheadPos: 4},
		},
		{
			name: "reverse-order, end of last token",
			tokens: []Token{
				{StartPos: 3, EndPos: 4, LookaheadPos: 4},
				{StartPos: 1, EndPos: 3, LookaheadPos: 3},
			},
			pos:           4,
			expectHasNext: false,
		},
		{
			name: "reverse-order, past end of last token",
			tokens: []Token{
				{StartPos: 3, EndPos: 4, LookaheadPos: 4},
				{StartPos: 1, EndPos: 3, LookaheadPos: 3},
			},
			pos:           5,
			expectHasNext: false,
		},
		{
			name: "many tokens, random order",
			tokens: []Token{
				{StartPos: 2, EndPos: 3, LookaheadPos: 3},
				{StartPos: 1, EndPos: 2, LookaheadPos: 2},
				{StartPos: 0, EndPos: 1, LookaheadPos: 1},
				{StartPos: 8, EndPos: 9, LookaheadPos: 9},
				{StartPos: 3, EndPos: 4, LookaheadPos: 4},
				{StartPos: 24, EndPos: 25, LookaheadPos: 25},
				{StartPos: 14, EndPos: 15, LookaheadPos: 15},
				{StartPos: 5, EndPos: 6, LookaheadPos: 6},
				{StartPos: 7, EndPos: 8, LookaheadPos: 8},
				{StartPos: 10, EndPos: 11, LookaheadPos: 11},
				{StartPos: 19, EndPos: 20, LookaheadPos: 20},
				{StartPos: 21, EndPos: 22, LookaheadPos: 22},
				{StartPos: 15, EndPos: 16, LookaheadPos: 16},
				{StartPos: 20, EndPos: 21, LookaheadPos: 21},
				{StartPos: 6, EndPos: 7, LookaheadPos: 7},
				{StartPos: 11, EndPos: 12, LookaheadPos: 12},
				{StartPos: 12, EndPos: 13, LookaheadPos: 13},
				{StartPos: 23, EndPos: 24, LookaheadPos: 24},
				{StartPos: 9, EndPos: 10, LookaheadPos: 10},
				{StartPos: 16, EndPos: 17, LookaheadPos: 17},
				{StartPos: 17, EndPos: 18, LookaheadPos: 18},
				{StartPos: 22, EndPos: 23, LookaheadPos: 23},
				{StartPos: 4, EndPos: 5, LookaheadPos: 5},
				{StartPos: 18, EndPos: 19, LookaheadPos: 19},
				{StartPos: 13, EndPos: 14, LookaheadPos: 14},
			},
			pos:             12,
			expectHasNext:   true,
			expectNextToken: Token{StartPos: 12, EndPos: 13, LookaheadPos: 13},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tree := treeForTokens(tc.tokens)
			iter := tree.IterFromPosition(tc.pos)

			var tok Token
			hasNext := iter.Get(&tok)
			assert.Equal(t, tc.expectHasNext, hasNext)
			if hasNext && tc.expectHasNext {
				assert.Equal(t, tc.expectNextToken, tok)
			}
		})
	}
}
