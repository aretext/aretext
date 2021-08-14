package parser

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/aretext/aretext/text"
)

// simpleParseFunc recognizes strings prefixed and suffixed with a double-quote.
func simpleParseFunc(iter text.CloneableRuneIter, state State) Result {
	// Consume the first character in the text.
	r, err := iter.NextRune()
	if err != nil {
		return FailedResult
	}
	n := uint64(1)

	if r == '"' {
		// Text starts with a double-quote, so search for a matching double-quote.
		for {
			r, err = iter.NextRune()
			if err != nil {
				// No double-quote found before EOF, so consume without producing any tokens.
				return Result{NumConsumed: n, NextState: state}
			} else if r == '"' {
				// Found matching double-quote at end, so produce a string token.
				token := ComputedToken{
					Length: n + 1,
					Role:   TokenRoleString,
				}
				return Result{
					NumConsumed:    token.Length,
					ComputedTokens: []ComputedToken{token},
					NextState:      state,
				}
			}
			n++
		}
	} else {
		// Text does not start with a double-quote, so consume up to the start
		// of the next double-quote or EOF.
		for {
			r, err = iter.NextRune()
			if err != nil || r == '"' {
				return Result{NumConsumed: n, NextState: state}
			}
			n++
		}
	}
}

func TestParseAll(t *testing.T) {
	testCases := []struct {
		name           string
		text           string
		expectedTokens []Token
	}{
		{
			name:           "empty",
			text:           "",
			expectedTokens: nil,
		},
		{
			name: "recognize single token",
			text: `"test"`,
			expectedTokens: []Token{
				{StartPos: 0, EndPos: 6, Role: TokenRoleString},
			},
		},
		{
			name: "recognize multiple tokens",
			text: `"foo""bar"`,
			expectedTokens: []Token{
				{StartPos: 0, EndPos: 5, Role: TokenRoleString},
				{StartPos: 5, EndPos: 10, Role: TokenRoleString},
			},
		},
		{
			name:           "consume all without recognizing tokens",
			text:           `foobar`,
			expectedTokens: nil,
		},
		{
			name: "token in middle",
			text: `    "test"    `,
			expectedTokens: []Token{
				{StartPos: 4, EndPos: 10, Role: TokenRoleString},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tree, err := text.NewTreeFromString(tc.text)
			require.NoError(t, err)
			p := New(simpleParseFunc)
			c := p.ParseAll(tree)
			tokens := c.TokensIntersectingRange(0, math.MaxUint64)
			assert.Equal(t, tc.expectedTokens, tokens)
		})
	}
}

func TestReparseAfterEditInsertion(t *testing.T) {
	testCases := []struct {
		name           string
		text           string
		editPos        uint64
		insertString   string
		expectedTokens []Token
	}{
		{
			name:           "empty, insert empty",
			text:           "",
			editPos:        0,
			insertString:   "",
			expectedTokens: nil,
		},
		{
			name:         "empty, insert token",
			text:         "",
			editPos:      0,
			insertString: `"test"`,
			expectedTokens: []Token{
				{StartPos: 0, EndPos: 6, Role: TokenRoleString},
			},
		},
		{
			name:           "empty, insert text but no tokens",
			text:           "",
			editPos:        0,
			insertString:   `test`,
			expectedTokens: nil,
		},
		{
			name:         "change tokens",
			text:         `"this is a test"`,
			editPos:      5,
			insertString: `"`,
			expectedTokens: []Token{
				{StartPos: 0, EndPos: 6, Role: TokenRoleString},
			},
		},
		{
			name:         "affect multiple tokens",
			text:         `"this is a test"`,
			editPos:      5,
			insertString: `" "`,
			expectedTokens: []Token{
				{StartPos: 0, EndPos: 6, Role: TokenRoleString},
				{StartPos: 7, EndPos: 19, Role: TokenRoleString},
			},
		},
		{
			name:         "affect some tokens but not others",
			text:         `"foo" "bar" "baz"`,
			editPos:      7,
			insertString: `x`,
			expectedTokens: []Token{
				{StartPos: 0, EndPos: 5, Role: TokenRoleString},
				{StartPos: 6, EndPos: 12, Role: TokenRoleString},
				{StartPos: 13, EndPos: 18, Role: TokenRoleString},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tree, err := text.NewTreeFromString(tc.text)
			require.NoError(t, err)
			p := New(simpleParseFunc)
			p.ParseAll(tree)

			var n uint64
			for _, r := range tc.insertString {
				err = tree.InsertAtPosition(tc.editPos+n, r)
				require.NoError(t, err)
				n++
			}

			edit := NewInsertEdit(tc.editPos, n)
			c := p.ReparseAfterEdit(tree, edit)
			tokens := c.TokensIntersectingRange(0, math.MaxUint64)
			assert.Equal(t, tc.expectedTokens, tokens)
		})
	}
}

func TestReparseAfterEditDeletion(t *testing.T) {
	testCases := []struct {
		name           string
		text           string
		editPos        uint64
		numDeleted     uint64
		expectedTokens []Token
	}{
		{
			name:           "no tokens, delete a few characters",
			text:           "abcdefghijk",
			editPos:        2,
			numDeleted:     3,
			expectedTokens: nil,
		},
		{
			name:       "delete affects tokens",
			text:       `"foo"bar"`,
			editPos:    4,
			numDeleted: 1,
			expectedTokens: []Token{
				{StartPos: 0, EndPos: 8, Role: TokenRoleString},
			},
		},
		{
			name:       "delete changes length of existing token",
			text:       `"foobar"`,
			editPos:    4,
			numDeleted: 2,
			expectedTokens: []Token{
				{StartPos: 0, EndPos: 6, Role: TokenRoleString},
			},
		},
		{
			name:       "delete affects multiple tokens",
			text:       `"foo" "bar" "baz"`,
			editPos:    4,
			numDeleted: 1,
			expectedTokens: []Token{
				{StartPos: 0, EndPos: 6, Role: TokenRoleString},
				{StartPos: 9, EndPos: 12, Role: TokenRoleString},
			},
		},
		{
			name:       "delete affects some tokens but not others",
			text:       `"foo" "bar" "baz"`,
			editPos:    8,
			numDeleted: 1,
			expectedTokens: []Token{
				{StartPos: 0, EndPos: 5, Role: TokenRoleString},
				{StartPos: 6, EndPos: 10, Role: TokenRoleString},
				{StartPos: 11, EndPos: 16, Role: TokenRoleString},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tree, err := text.NewTreeFromString(tc.text)
			require.NoError(t, err)
			p := New(simpleParseFunc)
			p.ParseAll(tree)

			for i := uint64(0); i < tc.numDeleted; i++ {
				tree.DeleteAtPosition(tc.editPos + i)
			}

			edit := NewDeleteEdit(tc.editPos, tc.numDeleted)
			c := p.ReparseAfterEdit(tree, edit)
			tokens := c.TokensIntersectingRange(0, math.MaxUint64)
			assert.Equal(t, tc.expectedTokens, tokens)
		})
	}
}
