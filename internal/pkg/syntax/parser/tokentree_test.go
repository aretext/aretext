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
				{StartPos: 0, EndPos: 1},
				{StartPos: 3, EndPos: 4},
				{StartPos: 4, EndPos: 5},
			},
			edit: Edit{Pos: 2, NumInserted: 1},
			expectedTokens: []Token{
				{StartPos: 0, EndPos: 1},
				{StartPos: 4, EndPos: 5},
				{StartPos: 5, EndPos: 6},
			},
		},
		{
			name: "insert one intersecting start",
			tokens: []Token{
				{StartPos: 0, EndPos: 1},
				{StartPos: 3, EndPos: 4},
				{StartPos: 4, EndPos: 5},
			},
			edit: Edit{Pos: 3, NumInserted: 1},
			expectedTokens: []Token{
				{StartPos: 0, EndPos: 1},
				{StartPos: 4, EndPos: 5},
				{StartPos: 5, EndPos: 6},
			},
		},
		{
			name: "insert one before with lazy update",
			tokens: []Token{
				{StartPos: 1, EndPos: 2},
				{StartPos: 3, EndPos: 4},
				{StartPos: 4, EndPos: 5},
				{StartPos: 6, EndPos: 7},
				{StartPos: 8, EndPos: 9},
			},
			edit: Edit{Pos: 0, NumInserted: 1},
			expectedTokens: []Token{
				{StartPos: 2, EndPos: 3},
				{StartPos: 4, EndPos: 5},
				{StartPos: 5, EndPos: 6},
				{StartPos: 7, EndPos: 8},
				{StartPos: 9, EndPos: 10},
			},
		},
		{
			name: "insert one before full tree",
			tokens: []Token{
				{StartPos: 1, EndPos: 2},
				{StartPos: 3, EndPos: 4},
				{StartPos: 4, EndPos: 5},
				{StartPos: 6, EndPos: 7},
				{StartPos: 8, EndPos: 9},
				{StartPos: 10, EndPos: 11},
				{StartPos: 11, EndPos: 12},
				{StartPos: 13, EndPos: 14},
			},
			edit: Edit{Pos: 0, NumInserted: 1},
			expectedTokens: []Token{
				{StartPos: 2, EndPos: 3},
				{StartPos: 4, EndPos: 5},
				{StartPos: 5, EndPos: 6},
				{StartPos: 7, EndPos: 8},
				{StartPos: 9, EndPos: 10},
				{StartPos: 11, EndPos: 12},
				{StartPos: 12, EndPos: 13},
				{StartPos: 14, EndPos: 15},
			},
		},
		{
			name: "insert one in right node of full tree",
			tokens: []Token{
				{StartPos: 1, EndPos: 2},
				{StartPos: 3, EndPos: 4},
				{StartPos: 4, EndPos: 5},
				{StartPos: 6, EndPos: 7},
				{StartPos: 8, EndPos: 9},
				{StartPos: 10, EndPos: 11},
				{StartPos: 11, EndPos: 12},
				{StartPos: 13, EndPos: 14},
			},
			edit: Edit{Pos: 11, NumInserted: 1},
			expectedTokens: []Token{
				{StartPos: 1, EndPos: 2},
				{StartPos: 3, EndPos: 4},
				{StartPos: 4, EndPos: 5},
				{StartPos: 6, EndPos: 7},
				{StartPos: 8, EndPos: 9},
				{StartPos: 10, EndPos: 11},
				{StartPos: 12, EndPos: 13},
				{StartPos: 14, EndPos: 15},
			},
		},
		{
			name: "delete one before",
			tokens: []Token{
				{StartPos: 1, EndPos: 2},
				{StartPos: 3, EndPos: 4},
				{StartPos: 4, EndPos: 5},
				{StartPos: 6, EndPos: 7},
				{StartPos: 8, EndPos: 9},
				{StartPos: 10, EndPos: 11},
				{StartPos: 11, EndPos: 12},
				{StartPos: 13, EndPos: 14},
			},
			edit: Edit{Pos: 0, NumDeleted: 1},
			expectedTokens: []Token{
				{StartPos: 0, EndPos: 1},
				{StartPos: 2, EndPos: 3},
				{StartPos: 3, EndPos: 4},
				{StartPos: 5, EndPos: 6},
				{StartPos: 7, EndPos: 8},
				{StartPos: 9, EndPos: 10},
				{StartPos: 10, EndPos: 11},
				{StartPos: 12, EndPos: 13},
			},
		},
		{
			name: "delete in right node of full tree",
			tokens: []Token{
				{StartPos: 1, EndPos: 2},
				{StartPos: 3, EndPos: 4},
				{StartPos: 4, EndPos: 5},
				{StartPos: 6, EndPos: 7},
				{StartPos: 8, EndPos: 9},
				{StartPos: 10, EndPos: 11},
				{StartPos: 11, EndPos: 12},
				{StartPos: 13, EndPos: 14},
			},
			edit: Edit{Pos: 11, NumDeleted: 1},
			expectedTokens: []Token{
				{StartPos: 1, EndPos: 2},
				{StartPos: 3, EndPos: 4},
				{StartPos: 4, EndPos: 5},
				{StartPos: 6, EndPos: 7},
				{StartPos: 8, EndPos: 9},
				{StartPos: 10, EndPos: 11},
				{StartPos: 11, EndPos: 11},
				{StartPos: 12, EndPos: 13},
			},
		},
		{
			name: "insert multiple",
			tokens: []Token{
				{StartPos: 1, EndPos: 2},
				{StartPos: 2, EndPos: 3},
			},
			edit: Edit{Pos: 0, NumInserted: 10},
			expectedTokens: []Token{
				{StartPos: 11, EndPos: 12},
				{StartPos: 12, EndPos: 13},
			},
		},
		{
			name: "insert multiple with overflow",
			tokens: []Token{
				{StartPos: 1, EndPos: 2},
				{StartPos: 2, EndPos: 3},
			},
			edit: Edit{Pos: 0, NumInserted: uint64(0xFFFFFFFFFFFFFFFF)},
			expectedTokens: []Token{
				{StartPos: uint64(0xFFFFFFFFFFFFFFFF), EndPos: uint64(0xFFFFFFFFFFFFFFFF)},
				{StartPos: uint64(0xFFFFFFFFFFFFFFFF), EndPos: uint64(0xFFFFFFFFFFFFFFFF)},
			},
		},
		{
			name: "delete multiple",
			tokens: []Token{
				{StartPos: 10, EndPos: 12},
				{StartPos: 12, EndPos: 13},
			},
			edit: Edit{Pos: 4, NumDeleted: 3},
			expectedTokens: []Token{
				{StartPos: 7, EndPos: 9},
				{StartPos: 9, EndPos: 10},
			},
		},
		{
			name: "delete multiple with underflow",
			tokens: []Token{
				{StartPos: 1, EndPos: 2},
				{StartPos: 2, EndPos: 3},
			},
			edit: Edit{Pos: 1, NumDeleted: 10},
			expectedTokens: []Token{
				{StartPos: 1, EndPos: 1},
				{StartPos: 1, EndPos: 1},
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
