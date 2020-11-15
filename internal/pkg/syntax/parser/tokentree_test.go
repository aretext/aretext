package parser

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTokenTreeIterFromPosition(t *testing.T) {
	testCases := []struct {
		name           string
		tokens         []Token
		position       uint64
		expectedTokens []Token
	}{
		{
			name:           "empty tree, position zero",
			tokens:         []Token{},
			position:       0,
			expectedTokens: []Token{},
		},
		{
			name:           "empty tree, position greater than zero",
			tokens:         []Token{},
			position:       10,
			expectedTokens: []Token{},
		},
		{
			name: "single token, position zero, intersects",
			tokens: []Token{
				{StartPos: 0, EndPos: 2, Role: TokenRoleOperator},
			},
			position: 0,
			expectedTokens: []Token{
				{StartPos: 0, EndPos: 2, Role: TokenRoleOperator},
			},
		},
		{
			name: "single token, position one, intersects",
			tokens: []Token{
				{StartPos: 0, EndPos: 2, Role: TokenRoleOperator},
			},
			position: 1,
			expectedTokens: []Token{
				{StartPos: 0, EndPos: 2, Role: TokenRoleOperator},
			},
		},
		{
			name: "single token, position before token",
			tokens: []Token{
				{StartPos: 1, EndPos: 2, Role: TokenRoleOperator},
			},
			position: 0,
			expectedTokens: []Token{
				{StartPos: 1, EndPos: 2, Role: TokenRoleOperator},
			},
		},
		{
			name: "single token, position at token end",
			tokens: []Token{
				{StartPos: 1, EndPos: 2, Role: TokenRoleOperator},
			},
			position:       2,
			expectedTokens: []Token{},
		},
		{
			name: "single token, position after token end",
			tokens: []Token{
				{StartPos: 1, EndPos: 2, Role: TokenRoleOperator},
			},
			position:       3,
			expectedTokens: []Token{},
		},
		{
			name: "single token, zero length, position at start",
			tokens: []Token{
				{StartPos: 1, EndPos: 1, Role: TokenRoleOperator},
			},
			position: 1,
			expectedTokens: []Token{
				{StartPos: 1, EndPos: 1, Role: TokenRoleOperator},
			},
		},
		{
			name: "multiple tokens, iter from start",
			tokens: []Token{
				{StartPos: 1, EndPos: 2, Role: TokenRoleOperator},
				{StartPos: 2, EndPos: 3, Role: TokenRoleIdentifier},
				{StartPos: 4, EndPos: 6, Role: TokenRoleComment},
			},
			position: 0,
			expectedTokens: []Token{
				{StartPos: 1, EndPos: 2, Role: TokenRoleOperator},
				{StartPos: 2, EndPos: 3, Role: TokenRoleIdentifier},
				{StartPos: 4, EndPos: 6, Role: TokenRoleComment},
			},
		},
		{
			name: "multiple tokens, iter from middle",
			tokens: []Token{
				{StartPos: 1, EndPos: 2, Role: TokenRoleOperator},
				{StartPos: 2, EndPos: 3, Role: TokenRoleIdentifier},
				{StartPos: 4, EndPos: 6, Role: TokenRoleComment},
			},
			position: 2,
			expectedTokens: []Token{
				{StartPos: 2, EndPos: 3, Role: TokenRoleIdentifier},
				{StartPos: 4, EndPos: 6, Role: TokenRoleComment},
			},
		},
		{
			name: "multiple tokens, iter from last",
			tokens: []Token{
				{StartPos: 1, EndPos: 2, Role: TokenRoleOperator},
				{StartPos: 2, EndPos: 3, Role: TokenRoleIdentifier},
				{StartPos: 4, EndPos: 6, Role: TokenRoleComment},
			},
			position: 3,
			expectedTokens: []Token{
				{StartPos: 4, EndPos: 6, Role: TokenRoleComment},
			},
		},
		{
			name: "multiple tokens, iter from end",
			tokens: []Token{
				{StartPos: 1, EndPos: 2, Role: TokenRoleOperator},
				{StartPos: 2, EndPos: 3, Role: TokenRoleIdentifier},
				{StartPos: 4, EndPos: 6, Role: TokenRoleComment},
			},
			position:       6,
			expectedTokens: []Token{},
		},
		{
			name: "non-full tree, iter from start",
			tokens: []Token{
				{StartPos: 1, EndPos: 2, Role: TokenRoleOperator},
				{StartPos: 2, EndPos: 3, Role: TokenRoleIdentifier},
				{StartPos: 4, EndPos: 6, Role: TokenRoleComment},
				{StartPos: 7, EndPos: 9, Role: TokenRoleNumber},
			},
			position: 0,
			expectedTokens: []Token{
				{StartPos: 1, EndPos: 2, Role: TokenRoleOperator},
				{StartPos: 2, EndPos: 3, Role: TokenRoleIdentifier},
				{StartPos: 4, EndPos: 6, Role: TokenRoleComment},
				{StartPos: 7, EndPos: 9, Role: TokenRoleNumber},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tree := NewTokenTree(tc.tokens)
			tokens := tree.IterFromPosition(tc.position).Collect()
			assert.Equal(t, tc.expectedTokens, tokens)
		})
	}
}

func TestTokenTreeShiftPositionsAfterEdit(t *testing.T) {
	testCases := []struct {
		name           string
		tokens         []Token
		edit           Edit
		expectedTokens []Token
	}{
		{
			name:           "insert into empty",
			tokens:         []Token{},
			edit:           Edit{Pos: 0, NumInserted: 1},
			expectedTokens: []Token{},
		},
		{
			name:           "delete from empty",
			tokens:         []Token{},
			edit:           Edit{Pos: 0, NumDeleted: 1},
			expectedTokens: []Token{},
		},
		{
			name: "insert one before",
			tokens: []Token{
				{StartPos: 0, EndPos: 1, LookaheadPos: 2},
				{StartPos: 3, EndPos: 4, LookaheadPos: 5},
				{StartPos: 4, EndPos: 5, LookaheadPos: 6},
			},
			edit: Edit{Pos: 2, NumInserted: 1},
			expectedTokens: []Token{
				{StartPos: 0, EndPos: 1, LookaheadPos: 2},
				{StartPos: 4, EndPos: 5, LookaheadPos: 6},
				{StartPos: 5, EndPos: 6, LookaheadPos: 7},
			},
		},
		{
			name: "insert one intersecting start",
			tokens: []Token{
				{StartPos: 0, EndPos: 1, LookaheadPos: 2},
				{StartPos: 3, EndPos: 4, LookaheadPos: 5},
				{StartPos: 4, EndPos: 5, LookaheadPos: 6},
			},
			edit: Edit{Pos: 3, NumInserted: 1},
			expectedTokens: []Token{
				{StartPos: 0, EndPos: 1, LookaheadPos: 2},
				{StartPos: 4, EndPos: 5, LookaheadPos: 6},
				{StartPos: 5, EndPos: 6, LookaheadPos: 7},
			},
		},
		{
			name: "insert one before with lazy update",
			tokens: []Token{
				{StartPos: 1, EndPos: 2, LookaheadPos: 3},
				{StartPos: 3, EndPos: 4, LookaheadPos: 5},
				{StartPos: 4, EndPos: 5, LookaheadPos: 6},
				{StartPos: 6, EndPos: 7, LookaheadPos: 8},
				{StartPos: 8, EndPos: 9, LookaheadPos: 10},
			},
			edit: Edit{Pos: 0, NumInserted: 1},
			expectedTokens: []Token{
				{StartPos: 2, EndPos: 3, LookaheadPos: 4},
				{StartPos: 4, EndPos: 5, LookaheadPos: 6},
				{StartPos: 5, EndPos: 6, LookaheadPos: 7},
				{StartPos: 7, EndPos: 8, LookaheadPos: 9},
				{StartPos: 9, EndPos: 10, LookaheadPos: 11},
			},
		},
		{
			name: "insert one before full tree",
			tokens: []Token{
				{StartPos: 1, EndPos: 2, LookaheadPos: 3},
				{StartPos: 3, EndPos: 4, LookaheadPos: 5},
				{StartPos: 4, EndPos: 5, LookaheadPos: 6},
				{StartPos: 6, EndPos: 7, LookaheadPos: 8},
				{StartPos: 8, EndPos: 9, LookaheadPos: 10},
				{StartPos: 10, EndPos: 11, LookaheadPos: 12},
				{StartPos: 11, EndPos: 12, LookaheadPos: 13},
				{StartPos: 13, EndPos: 14, LookaheadPos: 15},
			},
			edit: Edit{Pos: 0, NumInserted: 1},
			expectedTokens: []Token{
				{StartPos: 2, EndPos: 3, LookaheadPos: 4},
				{StartPos: 4, EndPos: 5, LookaheadPos: 6},
				{StartPos: 5, EndPos: 6, LookaheadPos: 7},
				{StartPos: 7, EndPos: 8, LookaheadPos: 9},
				{StartPos: 9, EndPos: 10, LookaheadPos: 11},
				{StartPos: 11, EndPos: 12, LookaheadPos: 13},
				{StartPos: 12, EndPos: 13, LookaheadPos: 14},
				{StartPos: 14, EndPos: 15, LookaheadPos: 16},
			},
		},
		{
			name: "insert one in right node of full tree",
			tokens: []Token{
				{StartPos: 1, EndPos: 2, LookaheadPos: 3},
				{StartPos: 3, EndPos: 4, LookaheadPos: 5},
				{StartPos: 4, EndPos: 5, LookaheadPos: 6},
				{StartPos: 6, EndPos: 7, LookaheadPos: 8},
				{StartPos: 8, EndPos: 9, LookaheadPos: 10},
				{StartPos: 10, EndPos: 11, LookaheadPos: 12},
				{StartPos: 11, EndPos: 12, LookaheadPos: 13},
				{StartPos: 13, EndPos: 14, LookaheadPos: 15},
			},
			edit: Edit{Pos: 11, NumInserted: 1},
			expectedTokens: []Token{
				{StartPos: 1, EndPos: 2, LookaheadPos: 3},
				{StartPos: 3, EndPos: 4, LookaheadPos: 5},
				{StartPos: 4, EndPos: 5, LookaheadPos: 6},
				{StartPos: 6, EndPos: 7, LookaheadPos: 8},
				{StartPos: 8, EndPos: 9, LookaheadPos: 10},
				{StartPos: 10, EndPos: 11, LookaheadPos: 12},
				{StartPos: 12, EndPos: 13, LookaheadPos: 14},
				{StartPos: 14, EndPos: 15, LookaheadPos: 16},
			},
		},
		{
			name: "delete one before",
			tokens: []Token{
				{StartPos: 1, EndPos: 2, LookaheadPos: 3},
				{StartPos: 3, EndPos: 4, LookaheadPos: 5},
				{StartPos: 4, EndPos: 5, LookaheadPos: 6},
				{StartPos: 6, EndPos: 7, LookaheadPos: 8},
				{StartPos: 8, EndPos: 9, LookaheadPos: 10},
				{StartPos: 10, EndPos: 11, LookaheadPos: 12},
				{StartPos: 11, EndPos: 12, LookaheadPos: 13},
				{StartPos: 13, EndPos: 14, LookaheadPos: 15},
			},
			edit: Edit{Pos: 0, NumDeleted: 1},
			expectedTokens: []Token{
				{StartPos: 0, EndPos: 1, LookaheadPos: 2},
				{StartPos: 2, EndPos: 3, LookaheadPos: 4},
				{StartPos: 3, EndPos: 4, LookaheadPos: 5},
				{StartPos: 5, EndPos: 6, LookaheadPos: 7},
				{StartPos: 7, EndPos: 8, LookaheadPos: 9},
				{StartPos: 9, EndPos: 10, LookaheadPos: 11},
				{StartPos: 10, EndPos: 11, LookaheadPos: 12},
				{StartPos: 12, EndPos: 13, LookaheadPos: 14},
			},
		},
		{
			name: "delete in right node of full tree",
			tokens: []Token{
				{StartPos: 1, EndPos: 2, LookaheadPos: 3},
				{StartPos: 3, EndPos: 4, LookaheadPos: 5},
				{StartPos: 4, EndPos: 5, LookaheadPos: 6},
				{StartPos: 6, EndPos: 7, LookaheadPos: 8},
				{StartPos: 8, EndPos: 9, LookaheadPos: 10},
				{StartPos: 10, EndPos: 11, LookaheadPos: 12},
				{StartPos: 11, EndPos: 12, LookaheadPos: 13},
				{StartPos: 13, EndPos: 14, LookaheadPos: 15},
			},
			edit: Edit{Pos: 11, NumDeleted: 1},
			expectedTokens: []Token{
				{StartPos: 1, EndPos: 2, LookaheadPos: 3},
				{StartPos: 3, EndPos: 4, LookaheadPos: 5},
				{StartPos: 4, EndPos: 5, LookaheadPos: 6},
				{StartPos: 6, EndPos: 7, LookaheadPos: 8},
				{StartPos: 8, EndPos: 9, LookaheadPos: 10},
				{StartPos: 10, EndPos: 11, LookaheadPos: 12},
				{StartPos: 11, EndPos: 11, LookaheadPos: 12},
				{StartPos: 12, EndPos: 13, LookaheadPos: 14},
			},
		},
		{
			name: "insert multiple",
			tokens: []Token{
				{StartPos: 1, EndPos: 2, LookaheadPos: 3},
				{StartPos: 2, EndPos: 3, LookaheadPos: 4},
			},
			edit: Edit{Pos: 0, NumInserted: 10},
			expectedTokens: []Token{
				{StartPos: 11, EndPos: 12, LookaheadPos: 13},
				{StartPos: 12, EndPos: 13, LookaheadPos: 14},
			},
		},
		{
			name: "insert multiple with overflow",
			tokens: []Token{
				{StartPos: 1, EndPos: 2, LookaheadPos: 3},
				{StartPos: 2, EndPos: 3, LookaheadPos: 4},
			},
			edit: Edit{Pos: 0, NumInserted: uint64(0xFFFFFFFFFFFFFFFF)},
			expectedTokens: []Token{
				{
					StartPos:     uint64(0xFFFFFFFFFFFFFFFF),
					EndPos:       uint64(0xFFFFFFFFFFFFFFFF),
					LookaheadPos: uint64(0xFFFFFFFFFFFFFFFF),
				},
				{
					StartPos:     uint64(0xFFFFFFFFFFFFFFFF),
					EndPos:       uint64(0xFFFFFFFFFFFFFFFF),
					LookaheadPos: uint64(0xFFFFFFFFFFFFFFFF),
				},
			},
		},
		{
			name: "delete multiple",
			tokens: []Token{
				{StartPos: 10, EndPos: 12, LookaheadPos: 13},
				{StartPos: 12, EndPos: 13, LookaheadPos: 14},
			},
			edit: Edit{Pos: 4, NumDeleted: 3},
			expectedTokens: []Token{
				{StartPos: 7, EndPos: 9, LookaheadPos: 10},
				{StartPos: 9, EndPos: 10, LookaheadPos: 11},
			},
		},
		{
			name: "delete multiple with underflow",
			tokens: []Token{
				{StartPos: 1, EndPos: 2, LookaheadPos: 3},
				{StartPos: 2, EndPos: 3, LookaheadPos: 4},
			},
			edit: Edit{Pos: 1, NumDeleted: 10},
			expectedTokens: []Token{
				{StartPos: 1, EndPos: 1, LookaheadPos: 1},
				{StartPos: 1, EndPos: 1, LookaheadPos: 1},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tree := NewTokenTree(tc.tokens)
			tree.ShiftPositionsAfterEdit(tc.edit)
			tokens := tree.IterFromPosition(0).Collect()
			assert.Equal(t, tc.expectedTokens, tokens)
		})
	}
}

func TestTokenTreeInsertToken(t *testing.T) {
	testCases := []struct {
		name           string
		buildTree      func() *TokenTree
		insertToken    Token
		expectedTokens []Token
	}{
		{
			name: "insert into empty tree",
			buildTree: func() *TokenTree {
				return NewTokenTree(nil)
			},
			insertToken: Token{StartPos: 0, EndPos: 1},
			expectedTokens: []Token{
				Token{StartPos: 0, EndPos: 1},
			},
		},
		{
			name: "insert before single node",
			buildTree: func() *TokenTree {
				return NewTokenTree([]Token{
					Token{StartPos: 1, EndPos: 2},
				})
			},
			insertToken: Token{StartPos: 0, EndPos: 1},
			expectedTokens: []Token{
				Token{StartPos: 0, EndPos: 1},
				Token{StartPos: 1, EndPos: 2},
			},
		},
		{
			name: "insert after single node",
			buildTree: func() *TokenTree {
				return NewTokenTree([]Token{
					Token{StartPos: 1, EndPos: 2},
				})
			},
			insertToken: Token{StartPos: 2, EndPos: 3},
			expectedTokens: []Token{
				Token{StartPos: 1, EndPos: 2},
				Token{StartPos: 2, EndPos: 3},
			},
		},
		{
			name: "insert overlapping start position",
			buildTree: func() *TokenTree {
				return NewTokenTree([]Token{
					Token{StartPos: 1, EndPos: 3},
				})
			},
			insertToken: Token{StartPos: 1, EndPos: 2},
			expectedTokens: []Token{
				Token{StartPos: 1, EndPos: 3},
				Token{StartPos: 1, EndPos: 2},
			},
		},
		{
			name: "insert with shifted positions",
			buildTree: func() *TokenTree {
				tree := NewTokenTree([]Token{
					Token{StartPos: 0, EndPos: 1},
					Token{StartPos: 1, EndPos: 2},
					Token{StartPos: 2, EndPos: 3},
					Token{StartPos: 3, EndPos: 4},
					Token{StartPos: 4, EndPos: 5},
					Token{StartPos: 5, EndPos: 6},
					Token{StartPos: 6, EndPos: 7},
				})

				tree.ShiftPositionsAfterEdit(Edit{Pos: 2, NumInserted: 10})
				return tree
			},
			insertToken: Token{StartPos: 3, EndPos: 5},
			expectedTokens: []Token{
				Token{StartPos: 0, EndPos: 1},
				Token{StartPos: 1, EndPos: 2},
				Token{StartPos: 3, EndPos: 5},
				Token{StartPos: 12, EndPos: 13},
				Token{StartPos: 13, EndPos: 14},
				Token{StartPos: 14, EndPos: 15},
				Token{StartPos: 15, EndPos: 16},
				Token{StartPos: 16, EndPos: 17},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tree := tc.buildTree()
			tree.InsertToken(tc.insertToken)
			tokens := tree.IterFromPosition(0).Collect()
			assert.Equal(t, tc.expectedTokens, tokens)
		})
	}
}

func TestTokenTreeDeleteToken(t *testing.T) {
	testCases := []struct {
		name            string
		buildTree       func() *TokenTree
		position        uint64
		expectHasNext   bool
		expectNextToken Token
		expectTokens    []Token
	}{
		{
			name: "delete from empty tree",
			buildTree: func() *TokenTree {
				return NewTokenTree(nil)
			},
			position:      0,
			expectHasNext: false,
			expectTokens:  []Token{},
		},
		{
			name: "delete single token",
			buildTree: func() *TokenTree {
				return NewTokenTree([]Token{
					Token{StartPos: 0, EndPos: 1},
				})
			},
			position:      0,
			expectHasNext: false,
			expectTokens:  []Token{},
		},
		{
			name: "delete root with two children",
			buildTree: func() *TokenTree {
				return NewTokenTree([]Token{
					Token{StartPos: 0, EndPos: 1},
					Token{StartPos: 1, EndPos: 2},
					Token{StartPos: 2, EndPos: 3},
				})
			},
			position:        1,
			expectHasNext:   true,
			expectNextToken: Token{StartPos: 2, EndPos: 3},
			expectTokens: []Token{
				Token{StartPos: 0, EndPos: 1},
				Token{StartPos: 2, EndPos: 3},
			},
		},
		{
			name: "delete subtree with no children",
			buildTree: func() *TokenTree {
				return NewTokenTree([]Token{
					Token{StartPos: 0, EndPos: 1},
					Token{StartPos: 1, EndPos: 2},
					Token{StartPos: 2, EndPos: 3},
				})
			},
			position:        0,
			expectHasNext:   true,
			expectNextToken: Token{StartPos: 1, EndPos: 2},
			expectTokens: []Token{
				Token{StartPos: 1, EndPos: 2},
				Token{StartPos: 2, EndPos: 3},
			},
		},
		{
			name: "delete left child with children",
			buildTree: func() *TokenTree {
				return NewTokenTree([]Token{
					Token{StartPos: 0, EndPos: 1},
					Token{StartPos: 1, EndPos: 2},
					Token{StartPos: 2, EndPos: 3},
					Token{StartPos: 3, EndPos: 4},
					Token{StartPos: 5, EndPos: 6},
					Token{StartPos: 7, EndPos: 8},
					Token{StartPos: 8, EndPos: 9},
				})
			},
			position:        1,
			expectHasNext:   true,
			expectNextToken: Token{StartPos: 2, EndPos: 3},
			expectTokens: []Token{
				Token{StartPos: 0, EndPos: 1},
				Token{StartPos: 2, EndPos: 3},
				Token{StartPos: 3, EndPos: 4},
				Token{StartPos: 5, EndPos: 6},
				Token{StartPos: 7, EndPos: 8},
				Token{StartPos: 8, EndPos: 9},
			},
		},
		{
			name: "delete right child with children",
			buildTree: func() *TokenTree {
				return NewTokenTree([]Token{
					Token{StartPos: 0, EndPos: 1},
					Token{StartPos: 1, EndPos: 2},
					Token{StartPos: 2, EndPos: 3},
					Token{StartPos: 3, EndPos: 4},
					Token{StartPos: 4, EndPos: 5},
					Token{StartPos: 5, EndPos: 6},
					Token{StartPos: 7, EndPos: 8},
					Token{StartPos: 8, EndPos: 9},
				})
			},
			position:        7,
			expectHasNext:   true,
			expectNextToken: Token{StartPos: 8, EndPos: 9},
			expectTokens: []Token{
				Token{StartPos: 0, EndPos: 1},
				Token{StartPos: 1, EndPos: 2},
				Token{StartPos: 2, EndPos: 3},
				Token{StartPos: 3, EndPos: 4},
				Token{StartPos: 4, EndPos: 5},
				Token{StartPos: 5, EndPos: 6},
				Token{StartPos: 8, EndPos: 9},
			},
		},
		{
			name: "delete node with only left child",
			buildTree: func() *TokenTree {
				tree := NewTokenTree([]Token{
					Token{StartPos: 0, EndPos: 1},
					Token{StartPos: 1, EndPos: 2},
					Token{StartPos: 2, EndPos: 3},
					Token{StartPos: 3, EndPos: 4},
					Token{StartPos: 4, EndPos: 5},
					Token{StartPos: 5, EndPos: 6},
					Token{StartPos: 6, EndPos: 7},
				})
				tree.IterFromPosition(2).Delete()
				return tree
			},
			position:        1,
			expectHasNext:   true,
			expectNextToken: Token{StartPos: 3, EndPos: 4},
			expectTokens: []Token{
				Token{StartPos: 0, EndPos: 1},
				Token{StartPos: 3, EndPos: 4},
				Token{StartPos: 4, EndPos: 5},
				Token{StartPos: 5, EndPos: 6},
				Token{StartPos: 6, EndPos: 7},
			},
		},
		{
			name: "delete node with only right child",
			buildTree: func() *TokenTree {
				tree := NewTokenTree([]Token{
					Token{StartPos: 0, EndPos: 1},
					Token{StartPos: 1, EndPos: 2},
					Token{StartPos: 2, EndPos: 3},
					Token{StartPos: 3, EndPos: 4},
					Token{StartPos: 4, EndPos: 5},
					Token{StartPos: 5, EndPos: 6},
					Token{StartPos: 7, EndPos: 8},
					Token{StartPos: 8, EndPos: 9},
				})
				tree.IterFromPosition(5).Delete()
				return tree
			},
			position:        7,
			expectHasNext:   true,
			expectNextToken: Token{StartPos: 8, EndPos: 9},
			expectTokens: []Token{
				Token{StartPos: 0, EndPos: 1},
				Token{StartPos: 1, EndPos: 2},
				Token{StartPos: 2, EndPos: 3},
				Token{StartPos: 3, EndPos: 4},
				Token{StartPos: 4, EndPos: 5},
				Token{StartPos: 8, EndPos: 9},
			},
		},
		{
			name: "delete root of full tree",
			buildTree: func() *TokenTree {
				return NewTokenTree([]Token{
					Token{StartPos: 0, EndPos: 1},
					Token{StartPos: 1, EndPos: 2},
					Token{StartPos: 2, EndPos: 3},
					Token{StartPos: 3, EndPos: 4},
					Token{StartPos: 4, EndPos: 5},
					Token{StartPos: 5, EndPos: 6},
					Token{StartPos: 6, EndPos: 7},
				})
			},
			position:        3,
			expectHasNext:   true,
			expectNextToken: Token{StartPos: 4, EndPos: 5},
			expectTokens: []Token{
				Token{StartPos: 0, EndPos: 1},
				Token{StartPos: 1, EndPos: 2},
				Token{StartPos: 2, EndPos: 3},
				Token{StartPos: 4, EndPos: 5},
				Token{StartPos: 5, EndPos: 6},
				Token{StartPos: 6, EndPos: 7},
			},
		},
		{
			name: "delete with shifted tokens",
			buildTree: func() *TokenTree {
				tree := NewTokenTree([]Token{
					Token{StartPos: 0, EndPos: 1},
					Token{StartPos: 1, EndPos: 2},
					Token{StartPos: 2, EndPos: 3},
					Token{StartPos: 3, EndPos: 4},
					Token{StartPos: 4, EndPos: 5},
					Token{StartPos: 5, EndPos: 6},
					Token{StartPos: 6, EndPos: 7},
				})

				tree.ShiftPositionsAfterEdit(Edit{Pos: 3, NumInserted: 10})
				return tree
			},
			position:        15,
			expectHasNext:   true,
			expectNextToken: Token{StartPos: 16, EndPos: 17},
			expectTokens: []Token{
				Token{StartPos: 0, EndPos: 1},
				Token{StartPos: 1, EndPos: 2},
				Token{StartPos: 2, EndPos: 3},
				Token{StartPos: 13, EndPos: 14},
				Token{StartPos: 14, EndPos: 15},
				Token{StartPos: 16, EndPos: 17},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tree := tc.buildTree()
			iter := tree.IterFromPosition(tc.position)
			iter.Delete()

			var tok Token
			assert.Equal(t, tc.expectHasNext, iter.Get(&tok))
			assert.Equal(t, tc.expectNextToken, tok)
			assert.Equal(t, tc.expectTokens, tree.IterFromPosition(0).Collect())
		})
	}
}
