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
				{StartPos: 0, EndPos: 2, LookaheadPos: 2, Role: TokenRoleOperator},
			},
			position: 0,
			expectedTokens: []Token{
				{StartPos: 0, EndPos: 2, LookaheadPos: 2, Role: TokenRoleOperator},
			},
		},
		{
			name: "single token, position one, intersects",
			tokens: []Token{
				{StartPos: 0, EndPos: 2, LookaheadPos: 2, Role: TokenRoleOperator},
			},
			position: 1,
			expectedTokens: []Token{
				{StartPos: 0, EndPos: 2, LookaheadPos: 2, Role: TokenRoleOperator},
			},
		},
		{
			name: "two tokens, position before token",
			tokens: []Token{
				{StartPos: 0, EndPos: 1, LookaheadPos: 1, Role: TokenRoleOperator},
				{StartPos: 1, EndPos: 2, LookaheadPos: 2, Role: TokenRoleOperator},
			},
			position: 0,
			expectedTokens: []Token{
				{StartPos: 0, EndPos: 1, LookaheadPos: 1, Role: TokenRoleOperator},
				{StartPos: 1, EndPos: 2, LookaheadPos: 2, Role: TokenRoleOperator},
			},
		},
		{
			name: "two tokens, position at token end",
			tokens: []Token{
				{StartPos: 0, EndPos: 1, LookaheadPos: 1, Role: TokenRoleOperator},
				{StartPos: 1, EndPos: 2, LookaheadPos: 2, Role: TokenRoleOperator},
			},
			position:       2,
			expectedTokens: []Token{},
		},
		{
			name: "two tokens, position after token end",
			tokens: []Token{
				{StartPos: 0, EndPos: 1, LookaheadPos: 1, Role: TokenRoleOperator},
				{StartPos: 1, EndPos: 2, LookaheadPos: 2, Role: TokenRoleOperator},
			},
			position:       3,
			expectedTokens: []Token{},
		},
		{
			name: "multiple tokens, iter from start",
			tokens: []Token{
				{StartPos: 0, EndPos: 1, LookaheadPos: 1, Role: TokenRoleOperator},
				{StartPos: 1, EndPos: 2, LookaheadPos: 2, Role: TokenRoleOperator},
				{StartPos: 2, EndPos: 3, LookaheadPos: 3, Role: TokenRoleIdentifier},
				{StartPos: 3, EndPos: 6, LookaheadPos: 6, Role: TokenRoleComment},
			},
			position: 0,
			expectedTokens: []Token{
				{StartPos: 0, EndPos: 1, LookaheadPos: 1, Role: TokenRoleOperator},
				{StartPos: 1, EndPos: 2, LookaheadPos: 2, Role: TokenRoleOperator},
				{StartPos: 2, EndPos: 3, LookaheadPos: 3, Role: TokenRoleIdentifier},
				{StartPos: 3, EndPos: 6, LookaheadPos: 6, Role: TokenRoleComment},
			},
		},
		{
			name: "multiple tokens, iter from middle",
			tokens: []Token{
				{StartPos: 0, EndPos: 1, LookaheadPos: 1, Role: TokenRoleOperator},
				{StartPos: 1, EndPos: 2, LookaheadPos: 2, Role: TokenRoleOperator},
				{StartPos: 2, EndPos: 3, LookaheadPos: 3, Role: TokenRoleIdentifier},
				{StartPos: 3, EndPos: 6, LookaheadPos: 6, Role: TokenRoleComment},
			},
			position: 2,
			expectedTokens: []Token{
				{StartPos: 2, EndPos: 3, LookaheadPos: 3, Role: TokenRoleIdentifier},
				{StartPos: 3, EndPos: 6, LookaheadPos: 6, Role: TokenRoleComment},
			},
		},
		{
			name: "multiple tokens, iter from last",
			tokens: []Token{
				{StartPos: 0, EndPos: 1, LookaheadPos: 1, Role: TokenRoleOperator},
				{StartPos: 1, EndPos: 2, LookaheadPos: 2, Role: TokenRoleOperator},
				{StartPos: 2, EndPos: 3, LookaheadPos: 3, Role: TokenRoleIdentifier},
				{StartPos: 3, EndPos: 6, LookaheadPos: 6, Role: TokenRoleComment},
			},
			position: 3,
			expectedTokens: []Token{
				{StartPos: 3, EndPos: 6, LookaheadPos: 6, Role: TokenRoleComment},
			},
		},
		{
			name: "multiple tokens, iter from end",
			tokens: []Token{
				{StartPos: 0, EndPos: 1, LookaheadPos: 1, Role: TokenRoleOperator},
				{StartPos: 1, EndPos: 2, LookaheadPos: 2, Role: TokenRoleOperator},
				{StartPos: 2, EndPos: 3, LookaheadPos: 3, Role: TokenRoleIdentifier},
				{StartPos: 3, EndPos: 6, LookaheadPos: 6, Role: TokenRoleComment},
			},
			position:       6,
			expectedTokens: []Token{},
		},
		{
			name:           "many tokens, iter from beginning",
			tokens:         generateTokens(1000),
			position:       0,
			expectedTokens: generateTokens(1000),
		},
		{
			name:           "many tokens, iter from middle",
			tokens:         generateTokens(1000),
			position:       500,
			expectedTokens: generateTokens(1000)[500:1000],
		},
		{
			name:           "many tokens, iter from end",
			tokens:         generateTokens(1000),
			position:       1000,
			expectedTokens: []Token{},
		},
		{
			name:           "very large tree, iter from beginning",
			tokens:         generateTokens(maxEntriesPerLeafNode * maxEntriesPerInnerNode * 2),
			position:       0,
			expectedTokens: generateTokens(maxEntriesPerLeafNode * maxEntriesPerInnerNode * 2),
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

func TestTokenTreeIterFromFirstAffected(t *testing.T) {
	testCases := []struct {
		name          string
		tokens        []Token
		editPos       uint64
		expectedToken Token
	}{
		{
			name:          "empty tree",
			tokens:        []Token{},
			editPos:       0,
			expectedToken: Token{},
		},
		{
			name: "single token, edit at start pos",
			tokens: []Token{
				Token{StartPos: 0, EndPos: 6, LookaheadPos: 7},
			},
			editPos:       0,
			expectedToken: Token{StartPos: 0, EndPos: 6, LookaheadPos: 7},
		},
		{
			name: "single token, edit after start pos, before end and lookahead pos",
			tokens: []Token{
				Token{StartPos: 0, EndPos: 6, LookaheadPos: 7},
			},
			editPos:       5,
			expectedToken: Token{StartPos: 0, EndPos: 6, LookaheadPos: 7},
		},
		{
			name: "single token, edit before lookahead pos",
			tokens: []Token{
				Token{StartPos: 0, EndPos: 6, LookaheadPos: 7},
			},
			editPos:       6,
			expectedToken: Token{StartPos: 0, EndPos: 6, LookaheadPos: 7},
		},
		{
			name: "single token, edit at lookahead pos",
			tokens: []Token{
				Token{StartPos: 0, EndPos: 6, LookaheadPos: 7},
			},
			editPos:       7,
			expectedToken: Token{StartPos: 0, EndPos: 6, LookaheadPos: 7},
		},
		{
			name: "two tokens with overlapping lookahead regions",
			tokens: []Token{
				Token{StartPos: 0, EndPos: 1, LookaheadPos: 7},
				Token{StartPos: 1, EndPos: 6, LookaheadPos: 7},
			},
			editPos:       3,
			expectedToken: Token{StartPos: 0, EndPos: 1, LookaheadPos: 7},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tree := NewTokenTree(tc.tokens)
			iter := tree.iterFromFirstAffected(tc.editPos)

			var tok Token
			iter.Get(&tok)
			assert.Equal(t, tc.expectedToken, tok)
		})
	}
}

func TestTokenTreeInsertToken(t *testing.T) {
	testCases := []struct {
		name           string
		initialTokens  []Token
		insertTokens   []Token
		expectedTokens []Token
	}{
		{
			name:          "empty tree, insert single token",
			initialTokens: []Token{},
			insertTokens: []Token{
				Token{StartPos: 0, EndPos: 2, LookaheadPos: 2, Role: TokenRoleOperator},
			},
			expectedTokens: []Token{
				Token{StartPos: 0, EndPos: 2, LookaheadPos: 2, Role: TokenRoleOperator},
			},
		},
		{
			name:           "empty tree, insert multiple tokens in ascending order",
			initialTokens:  []Token{},
			insertTokens:   generateTokens(10),
			expectedTokens: generateTokens(10),
		},
		{
			name:           "empty tree, insert maxEntriesPerLeafNode plus one tokens in ascending order",
			initialTokens:  []Token{},
			insertTokens:   generateTokens(maxEntriesPerLeafNode + 1),
			expectedTokens: generateTokens(maxEntriesPerLeafNode + 1),
		},
		{
			name:           "empty tree, insert maxEntriesPerInnerNode times maxEntriesPerLeafNode plus one tokens in ascending order",
			initialTokens:  []Token{},
			insertTokens:   generateTokens(maxEntriesPerInnerNode*maxEntriesPerLeafNode + 1),
			expectedTokens: generateTokens(maxEntriesPerInnerNode*maxEntriesPerLeafNode + 1),
		},
		{
			name: "small tree, insert single token at beginning",
			initialTokens: []Token{
				Token{StartPos: 0, EndPos: 2, LookaheadPos: 2, Role: TokenRoleOperator},
			},
			insertTokens: []Token{
				Token{StartPos: 0, EndPos: 1, LookaheadPos: 1, Role: TokenRoleComment},
			},
			expectedTokens: []Token{
				Token{StartPos: 0, EndPos: 1, LookaheadPos: 1, Role: TokenRoleComment},
				Token{StartPos: 1, EndPos: 3, LookaheadPos: 3, Role: TokenRoleOperator},
			},
		},
		{
			name: "small tree, insert single token in middle",
			initialTokens: []Token{
				Token{StartPos: 0, EndPos: 1, LookaheadPos: 1, Role: TokenRoleComment},
				Token{StartPos: 1, EndPos: 3, LookaheadPos: 3, Role: TokenRoleOperator},
				Token{StartPos: 3, EndPos: 5, LookaheadPos: 5, Role: TokenRoleString},
			},
			insertTokens: []Token{
				Token{StartPos: 1, EndPos: 2, LookaheadPos: 2, Role: TokenRoleNumber},
			},
			expectedTokens: []Token{
				Token{StartPos: 0, EndPos: 1, LookaheadPos: 1, Role: TokenRoleComment},
				Token{StartPos: 1, EndPos: 2, LookaheadPos: 2, Role: TokenRoleNumber},
				Token{StartPos: 2, EndPos: 4, LookaheadPos: 4, Role: TokenRoleOperator},
				Token{StartPos: 4, EndPos: 6, LookaheadPos: 6, Role: TokenRoleString},
			},
		},
		{
			name:          "very large tree, insert single token at beginning",
			initialTokens: generateTokens(maxEntriesPerLeafNode * maxEntriesPerInnerNode * 2),
			insertTokens: []Token{
				Token{StartPos: 0, EndPos: 2, LookaheadPos: 2, Role: TokenRoleOperator},
			},
			expectedTokens: append([]Token{
				Token{StartPos: 0, EndPos: 2, LookaheadPos: 2, Role: TokenRoleOperator},
			}, shiftPositionsForward(generateTokens(maxEntriesPerLeafNode*maxEntriesPerInnerNode*2), 2)...),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tree := NewTokenTree(tc.initialTokens)
			for _, token := range tc.insertTokens {
				tree.insertToken(token)
			}
			actualTokens := tree.IterFromPosition(0).Collect()
			assert.Equal(t, tc.expectedTokens, actualTokens)
		})
	}
}

func TestTokenTreeDeleteRange(t *testing.T) {
	testCases := []struct {
		name           string
		initialTokens  []Token
		startPos       uint64
		numDeleted     uint64
		expectedTokens []Token
	}{
		{
			name: "delete within token",
			initialTokens: []Token{
				Token{StartPos: 0, EndPos: 5, LookaheadPos: 6, Role: TokenRoleComment},
			},
			startPos:   1,
			numDeleted: 2,
			expectedTokens: []Token{
				Token{StartPos: 0, EndPos: 3, LookaheadPos: 4, Role: TokenRoleComment},
			},
		},
		{
			name: "delete truncates end of first token, removes second token, truncates beginning of third token",
			initialTokens: []Token{
				Token{StartPos: 0, EndPos: 5, LookaheadPos: 6, Role: TokenRoleComment},
				Token{StartPos: 5, EndPos: 6, LookaheadPos: 7, Role: TokenRoleOperator},
				Token{StartPos: 6, EndPos: 10, LookaheadPos: 10, Role: TokenRoleString},
			},
			startPos:   3,
			numDeleted: 5,
			expectedTokens: []Token{
				Token{StartPos: 0, EndPos: 3, LookaheadPos: 4, Role: TokenRoleComment},
				Token{StartPos: 3, EndPos: 5, LookaheadPos: 5, Role: TokenRoleString},
			},
		},
		{
			name: "delete from second token onward",
			initialTokens: []Token{
				Token{StartPos: 0, EndPos: 4, LookaheadPos: 5, Role: TokenRoleComment},
				Token{StartPos: 4, EndPos: 6, LookaheadPos: 7, Role: TokenRoleOperator},
				Token{StartPos: 6, EndPos: 10, LookaheadPos: 10, Role: TokenRoleString},
			},
			startPos:   4,
			numDeleted: 6,
			expectedTokens: []Token{
				Token{StartPos: 0, EndPos: 4, LookaheadPos: 5, Role: TokenRoleComment},
			},
		},
		{
			name:           "delete all tokens in very large tree",
			initialTokens:  generateTokens(maxEntriesPerLeafNode * maxEntriesPerInnerNode * 2),
			startPos:       0,
			numDeleted:     maxEntriesPerLeafNode * maxEntriesPerInnerNode * 2,
			expectedTokens: []Token{},
		},
		{
			name:           "delete from second inner node onward in very large tree",
			initialTokens:  generateTokens(maxEntriesPerLeafNode * maxEntriesPerInnerNode * 2),
			startPos:       maxEntriesPerLeafNode + 1,
			numDeleted:     maxEntriesPerLeafNode*maxEntriesPerInnerNode*2 - maxEntriesPerLeafNode,
			expectedTokens: generateTokens(maxEntriesPerLeafNode + 1),
		},
		{
			name:          "delete last token in very large tree",
			initialTokens: generateTokens(maxEntriesPerLeafNode * maxEntriesPerInnerNode * 2),
			startPos:      0,
			numDeleted:    maxEntriesPerLeafNode*maxEntriesPerInnerNode*2 - 1,
			expectedTokens: []Token{
				Token{StartPos: 0, EndPos: 1, LookaheadPos: 1, Role: TokenRoleString},
			},
		},
		{
			name:          "delete middle tokens in very large tree",
			initialTokens: generateTokens(maxEntriesPerLeafNode * maxEntriesPerInnerNode * 2),
			startPos:      1,
			numDeleted:    maxEntriesPerLeafNode*maxEntriesPerInnerNode*2 - 2,
			expectedTokens: []Token{
				Token{StartPos: 0, EndPos: 1, LookaheadPos: 1, Role: TokenRoleNone},
				Token{StartPos: 1, EndPos: 2, LookaheadPos: 2, Role: TokenRoleString},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tree := NewTokenTree(tc.initialTokens)
			tree.deleteRange(tc.startPos, tc.numDeleted)
			actualTokens := tree.IterFromPosition(0).Collect()
			assert.Equal(t, tc.expectedTokens, actualTokens)
		})
	}
}

func TestTokenTreeExtendTokenIntersectingPos(t *testing.T) {
	testCases := []struct {
		name          string
		initialTokens []Token
		pos           uint64
		extendLen     uint64
		expectedToken Token
	}{
		{
			name: "single token, position at start",
			initialTokens: []Token{
				Token{StartPos: 0, EndPos: 5, LookaheadPos: 10, Role: TokenRoleString},
			},
			pos:           0,
			extendLen:     2,
			expectedToken: Token{StartPos: 0, EndPos: 7, LookaheadPos: 12, Role: TokenRoleString},
		},
		{
			name: "single token, position in middle",
			initialTokens: []Token{
				Token{StartPos: 0, EndPos: 5, LookaheadPos: 10, Role: TokenRoleString},
			},
			pos:           2,
			extendLen:     2,
			expectedToken: Token{StartPos: 0, EndPos: 7, LookaheadPos: 12, Role: TokenRoleString},
		},
		{
			name: "single token, position at end middle",
			initialTokens: []Token{
				Token{StartPos: 0, EndPos: 5, LookaheadPos: 10, Role: TokenRoleString},
			},
			pos:           4,
			extendLen:     2,
			expectedToken: Token{StartPos: 0, EndPos: 7, LookaheadPos: 12, Role: TokenRoleString},
		},
		{
			name:          "many tokens, extend first token",
			initialTokens: generateTokens(maxEntriesPerLeafNode * maxEntriesPerInnerNode * 2),
			pos:           0,
			extendLen:     2,
			expectedToken: Token{StartPos: 0, EndPos: 3, LookaheadPos: 3, Role: TokenRoleNone},
		},
		{
			name:          "many tokens, extend token near end",
			initialTokens: generateTokens(maxEntriesPerLeafNode * maxEntriesPerInnerNode * 2),
			pos:           4000,
			extendLen:     2,
			expectedToken: Token{StartPos: 4000, EndPos: 4003, LookaheadPos: 4003, Role: TokenRoleNumber},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tree := NewTokenTree(tc.initialTokens)
			tree.extendTokenIntersectingPos(tc.pos, tc.extendLen)
			iter := tree.IterFromPosition(tc.pos)
			var token Token
			assert.True(t, iter.Get(&token))
			assert.Equal(t, tc.expectedToken, token)
		})
	}
}

func generateTokens(n int) []Token {
	tokens := make([]Token, 0, n)
	for i := 0; i < n; i++ {
		tokens = append(tokens, Token{
			StartPos:     uint64(i),
			EndPos:       uint64(i + 1),
			LookaheadPos: uint64(i + 1),
			Role:         TokenRole(i % 6),
		})
	}
	return tokens
}

func shiftPositionsForward(tokens []Token, shift uint64) []Token {
	updatedTokens := make([]Token, 0, len(tokens))
	for _, token := range tokens {
		updatedTokens = append(updatedTokens, Token{
			StartPos:     token.StartPos + shift,
			EndPos:       token.EndPos + shift,
			LookaheadPos: token.LookaheadPos + shift,
			Role:         token.Role,
		})
	}
	return updatedTokens
}
