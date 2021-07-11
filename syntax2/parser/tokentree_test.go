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
				{StartPos: 1, EndPos: 2},
			},
		},
		{
			name: "single token from position zero",
			tokens: []Token{
				{StartPos: 0, EndPos: 1},
			},
		},
		{
			name: "two tokens, in ascending order",
			tokens: []Token{
				{StartPos: 1, EndPos: 2},
				{StartPos: 2, EndPos: 3},
			},
		},
		{
			name: "two tokens, in descending order",
			tokens: []Token{
				{StartPos: 2, EndPos: 3},
				{StartPos: 1, EndPos: 2},
			},
		},
		{
			name: "ten tokens, random order",
			tokens: []Token{
				{StartPos: 2, EndPos: 3},
				{StartPos: 8, EndPos: 9},
				{StartPos: 4, EndPos: 5},
				{StartPos: 1, EndPos: 2},
				{StartPos: 6, EndPos: 7},
				{StartPos: 0, EndPos: 1},
				{StartPos: 9, EndPos: 10},
				{StartPos: 7, EndPos: 8},
				{StartPos: 5, EndPos: 6},
				{StartPos: 3, EndPos: 4},
			},
		},
		{
			name: "twenty-five tokens, random order",
			tokens: []Token{
				{StartPos: 2, EndPos: 3},
				{StartPos: 1, EndPos: 2},
				{StartPos: 0, EndPos: 1},
				{StartPos: 8, EndPos: 9},
				{StartPos: 3, EndPos: 4},
				{StartPos: 24, EndPos: 25},
				{StartPos: 14, EndPos: 15},
				{StartPos: 5, EndPos: 6},
				{StartPos: 7, EndPos: 8},
				{StartPos: 10, EndPos: 11},
				{StartPos: 19, EndPos: 20},
				{StartPos: 21, EndPos: 22},
				{StartPos: 15, EndPos: 16},
				{StartPos: 20, EndPos: 21},
				{StartPos: 6, EndPos: 7},
				{StartPos: 11, EndPos: 12},
				{StartPos: 12, EndPos: 13},
				{StartPos: 23, EndPos: 24},
				{StartPos: 9, EndPos: 10},
				{StartPos: 16, EndPos: 17},
				{StartPos: 17, EndPos: 18},
				{StartPos: 22, EndPos: 23},
				{StartPos: 4, EndPos: 5},
				{StartPos: 18, EndPos: 19},
				{StartPos: 13, EndPos: 14},
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
				{StartPos: 1, EndPos: 2},
				{StartPos: 2, EndPos: 3},
			},
			pos:             0,
			expectHasNext:   true,
			expectNextToken: Token{StartPos: 1, EndPos: 2},
		},
		{
			name: "in-order, start of first token",
			tokens: []Token{
				{StartPos: 1, EndPos: 2},
				{StartPos: 2, EndPos: 3},
			},
			pos:             1,
			expectHasNext:   true,
			expectNextToken: Token{StartPos: 1, EndPos: 2},
		},
		{
			name: "in-order, middle of first token",
			tokens: []Token{
				{StartPos: 1, EndPos: 3},
				{StartPos: 3, EndPos: 4},
			},
			pos:             2,
			expectHasNext:   true,
			expectNextToken: Token{StartPos: 1, EndPos: 3},
		},
		{
			name: "in-order, end of first token",
			tokens: []Token{
				{StartPos: 1, EndPos: 3},
				{StartPos: 3, EndPos: 4},
			},
			pos:             3,
			expectHasNext:   true,
			expectNextToken: Token{StartPos: 3, EndPos: 4},
		},
		{
			name: "in-order, end of last token",
			tokens: []Token{
				{StartPos: 1, EndPos: 3},
				{StartPos: 3, EndPos: 4},
			},
			pos:           4,
			expectHasNext: false,
		},
		{
			name: "in-order, past end of last token",
			tokens: []Token{
				{StartPos: 1, EndPos: 3},
				{StartPos: 3, EndPos: 4},
			},
			pos:           5,
			expectHasNext: false,
		},
		{
			name: "reverse-order, before first token",
			tokens: []Token{
				{StartPos: 2, EndPos: 3},
				{StartPos: 1, EndPos: 2},
			},
			pos:             0,
			expectHasNext:   true,
			expectNextToken: Token{StartPos: 1, EndPos: 2},
		},
		{
			name: "reverse-order, start of first token",
			tokens: []Token{
				{StartPos: 2, EndPos: 3},
				{StartPos: 1, EndPos: 2},
			},
			pos:             1,
			expectHasNext:   true,
			expectNextToken: Token{StartPos: 1, EndPos: 2},
		},
		{
			name: "reverse-order, middle of first token",
			tokens: []Token{
				{StartPos: 3, EndPos: 4},
				{StartPos: 1, EndPos: 3},
			},
			pos:             2,
			expectHasNext:   true,
			expectNextToken: Token{StartPos: 1, EndPos: 3},
		},
		{
			name: "reverse-order, end of first token",
			tokens: []Token{
				{StartPos: 3, EndPos: 4},
				{StartPos: 1, EndPos: 3},
			},
			pos:             3,
			expectHasNext:   true,
			expectNextToken: Token{StartPos: 3, EndPos: 4},
		},
		{
			name: "reverse-order, end of last token",
			tokens: []Token{
				{StartPos: 3, EndPos: 4},
				{StartPos: 1, EndPos: 3},
			},
			pos:           4,
			expectHasNext: false,
		},
		{
			name: "reverse-order, past end of last token",
			tokens: []Token{
				{StartPos: 3, EndPos: 4},
				{StartPos: 1, EndPos: 3},
			},
			pos:           5,
			expectHasNext: false,
		},
		{
			name: "many tokens, random order",
			tokens: []Token{
				{StartPos: 2, EndPos: 3},
				{StartPos: 1, EndPos: 2},
				{StartPos: 0, EndPos: 1},
				{StartPos: 8, EndPos: 9},
				{StartPos: 3, EndPos: 4},
				{StartPos: 24, EndPos: 25},
				{StartPos: 14, EndPos: 15},
				{StartPos: 5, EndPos: 6},
				{StartPos: 7, EndPos: 8},
				{StartPos: 10, EndPos: 11},
				{StartPos: 19, EndPos: 20},
				{StartPos: 21, EndPos: 22},
				{StartPos: 15, EndPos: 16},
				{StartPos: 20, EndPos: 21},
				{StartPos: 6, EndPos: 7},
				{StartPos: 11, EndPos: 12},
				{StartPos: 12, EndPos: 13},
				{StartPos: 23, EndPos: 24},
				{StartPos: 9, EndPos: 10},
				{StartPos: 16, EndPos: 17},
				{StartPos: 17, EndPos: 18},
				{StartPos: 22, EndPos: 23},
				{StartPos: 4, EndPos: 5},
				{StartPos: 18, EndPos: 19},
				{StartPos: 13, EndPos: 14},
			},
			pos:             12,
			expectHasNext:   true,
			expectNextToken: Token{StartPos: 12, EndPos: 13},
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

func TestTokenTreeJoin(t *testing.T) {
	testCases := []struct {
		name             string
		firstTreeTokens  []Token
		secondTreeTokens []Token
	}{
		{
			name:             "both empty",
			firstTreeTokens:  nil,
			secondTreeTokens: nil,
		},
		{
			name: "first non-empty, second empty",
			firstTreeTokens: []Token{
				{StartPos: 1, EndPos: 2},
				{StartPos: 2, EndPos: 3},
			},
			secondTreeTokens: nil,
		},
		{
			name:            "first empty, second non-empty",
			firstTreeTokens: nil,
			secondTreeTokens: []Token{
				{StartPos: 1, EndPos: 2},
				{StartPos: 2, EndPos: 3},
			},
		},
		{
			name: "first before second",
			firstTreeTokens: []Token{
				{StartPos: 1, EndPos: 2},
				{StartPos: 2, EndPos: 3},
			},
			secondTreeTokens: []Token{
				{StartPos: 3, EndPos: 4},
			},
		},
		{
			name: "first after second",
			firstTreeTokens: []Token{
				{StartPos: 3, EndPos: 4},
			},
			secondTreeTokens: []Token{
				{StartPos: 1, EndPos: 2},
				{StartPos: 2, EndPos: 3},
			},
		},
		{
			name: "larger trees, first before second",
			firstTreeTokens: []Token{
				{StartPos: 2, EndPos: 3},
				{StartPos: 8, EndPos: 9},
				{StartPos: 4, EndPos: 5},
				{StartPos: 1, EndPos: 2},
				{StartPos: 6, EndPos: 7},
				{StartPos: 0, EndPos: 1},
				{StartPos: 9, EndPos: 10},
				{StartPos: 7, EndPos: 8},
				{StartPos: 5, EndPos: 6},
				{StartPos: 3, EndPos: 4},
			},
			secondTreeTokens: []Token{
				{StartPos: 103, EndPos: 104},
				{StartPos: 102, EndPos: 103},
				{StartPos: 105, EndPos: 106},
				{StartPos: 101, EndPos: 102},
				{StartPos: 100, EndPos: 101},
				{StartPos: 104, EndPos: 105},
			},
		},
		{
			name: "larger trees, first after second",
			firstTreeTokens: []Token{
				{StartPos: 103, EndPos: 104},
				{StartPos: 102, EndPos: 103},
				{StartPos: 105, EndPos: 106},
				{StartPos: 101, EndPos: 102},
				{StartPos: 100, EndPos: 101},
				{StartPos: 104, EndPos: 105},
			},
			secondTreeTokens: []Token{
				{StartPos: 2, EndPos: 3},
				{StartPos: 8, EndPos: 9},
				{StartPos: 4, EndPos: 5},
				{StartPos: 1, EndPos: 2},
				{StartPos: 6, EndPos: 7},
				{StartPos: 0, EndPos: 1},
				{StartPos: 9, EndPos: 10},
				{StartPos: 7, EndPos: 8},
				{StartPos: 5, EndPos: 6},
				{StartPos: 3, EndPos: 4},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			firstTree := treeForTokens(tc.firstTreeTokens)
			secondTree := treeForTokens(tc.secondTreeTokens)
			tree := firstTree.Join(secondTree)
			tokens := tree.IterFromPosition(0).Collect()
			expectedTokens := sortedTokens(append(tc.firstTreeTokens, tc.secondTreeTokens...))
			assert.Equal(t, expectedTokens, tokens)
		})
	}
}
