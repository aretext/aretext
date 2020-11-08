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
			iter := tree.IterFromPosition(tc.position)
			tokens := iter.Collect()
			assert.Equal(t, tc.expectedTokens, tokens)
		})
	}
}
