package parser

import (
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTokenizeAll(t *testing.T) {
	tokenizer := constructTokenizer(t)

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

func TestRetokenizeAfterInsert(t *testing.T) {
	tokenizer := constructTokenizer(t)

	testCases := []struct {
		name        string
		initialText string
		insertPos   int64
		insertText  string
	}{
		{
			name:        "empty, insert keyword",
			initialText: "",
			insertPos:   0,
			insertText:  "def",
		},
		{
			name:        "empty, insert unrecognized chars",
			initialText: "",
			insertPos:   0,
			insertText:  "          ",
		},
		{
			name:        "empty, insert multiple",
			initialText: "",
			insertPos:   0,
			insertText:  "def foo():\n    return bar + 10",
		},
		{
			name:        "insert new tokens",
			initialText: "abcd + 123",
			insertPos:   5,
			insertText:  "- xyz ",
		},
		{
			name:        "modify existing token",
			initialText: "abcd + 123",
			insertPos:   8,
			insertText:  "789",
		},
		{
			name:        "replace existing token",
			initialText: "abcd + 123",
			insertPos:   8,
			insertText:  " * 5 + ",
		},
		{
			name:        "insert token at beginning",
			initialText: "abcd + 123",
			insertPos:   0,
			insertText:  "xyz + ",
		},
		{
			name:        "append token at end",
			initialText: "abcd + 123",
			insertPos:   10,
			insertText:  "+3",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assertRetokenizeMatchesFullTokenize(t, tokenizer, tc.initialText, func(s string) (string, Edit) {
				updatedText := fmt.Sprintf("%s%s%s", s[0:tc.insertPos], tc.insertText, s[tc.insertPos:len(s)])
				edit := Edit{
					Pos:         uint64(tc.insertPos),
					NumInserted: uint64(len(tc.insertText)),
				}
				return updatedText, edit
			})
		})
	}
}

func TestRetokenizeAfterDelete(t *testing.T) {
	tokenizer := constructTokenizer(t)

	testCases := []struct {
		name        string
		initialText string
		deletePos   int64
		numDeleted  uint64
	}{
		{
			name:        "delete all chars",
			initialText: "abcd",
			deletePos:   0,
			numDeleted:  4,
		},
		{
			name:        "delete part of token",
			initialText: "abcd",
			deletePos:   1,
			numDeleted:  1,
		},
		{
			name:        "delete single token",
			initialText: "abcd + 123 + 456",
			deletePos:   7,
			numDeleted:  3,
		},
		{
			name:        "delete multiple tokens",
			initialText: "abcd + 123 + xyz + 456",
			deletePos:   7,
			numDeleted:  12,
		},
		{
			name:        "delete last tokens",
			initialText: "abcd + 123",
			deletePos:   5,
			numDeleted:  5,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assertRetokenizeMatchesFullTokenize(t, tokenizer, tc.initialText, func(s string) (string, Edit) {
				updatedText := fmt.Sprintf("%s%s", s[0:tc.deletePos], s[tc.deletePos+int64(tc.numDeleted):len(s)])
				edit := Edit{
					Pos:        uint64(tc.deletePos),
					NumDeleted: tc.numDeleted,
				}
				return updatedText, edit
			})
		})
	}

}

func constructTokenizer(t *testing.T) *Tokenizer {
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
	return tokenizer
}

type editTextFunc func(s string) (string, Edit)

func assertRetokenizeMatchesFullTokenize(t *testing.T, tokenizer *Tokenizer, initialText string, editTextFunc editTextFunc) {
	textLen := uint64(len(initialText))
	r := strings.NewReader(initialText)
	initialTree, err := tokenizer.TokenizeAll(r, textLen)
	require.NoError(t, err)

	updatedText, edit := editTextFunc(initialText)
	updatedTextLen := uint64(len(updatedText))
	r = strings.NewReader(updatedText)

	expectedTree, err := tokenizer.TokenizeAll(r, updatedTextLen)
	require.NoError(t, err)

	retokenizedTree, err := tokenizer.RetokenizeAfterEdit(initialTree, edit, updatedTextLen, func(pos uint64) io.ReadSeeker {
		_, err := r.Seek(int64(pos), io.SeekStart)
		require.NoError(t, err)
		return r
	})
	require.NoError(t, err)

	expectedTokens := expectedTree.IterFromPosition(0).Collect()
	actualTokens := retokenizedTree.IterFromPosition(0).Collect()
	assert.Equal(t, expectedTokens, actualTokens)
}
