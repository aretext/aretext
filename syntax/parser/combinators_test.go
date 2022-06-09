package parser

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/aretext/aretext/text"
)

func TestMaybeBefore(t *testing.T) {
	// Parse consecutive numerals as numbers.
	firstParseFunc := func(iter TrackingRuneIter, state State) Result {
		var n uint64
		for {
			r, err := iter.NextRune()
			if err != nil || r < '0' || r > '9' {
				break
			}
			n++
		}
		return Result{
			NumConsumed: n,
			ComputedTokens: []ComputedToken{
				{
					Length: n,
					Role:   TokenRoleNumber,
				},
			},
			NextState: state,
		}
	}

	// Parse alpha characters as keywords.
	secondParseFunc := func(iter TrackingRuneIter, state State) Result {
		var n uint64
		for {
			r, err := iter.NextRune()
			if err != nil || r < 'A' || r > 'z' {
				break
			}
			n++
		}
		return Result{
			NumConsumed: n,
			ComputedTokens: []ComputedToken{
				{
					Length: n,
					Role:   TokenRoleKeyword,
				},
			},
			NextState: state,
		}
	}

	// Alpha characters, optionally prefixed with spaces.
	combinedParseFunc := Func(firstParseFunc).MaybeBefore(Func(secondParseFunc))

	testCases := []struct {
		name     string
		text     string
		expected []Token
	}{
		{
			name: "only second parse func",
			text: "abc",
			expected: []Token{
				{StartPos: 0, EndPos: 3, Role: TokenRoleKeyword},
			},
		},
		{
			name: "first and second parse func",
			text: "1234abc",
			expected: []Token{
				{StartPos: 0, EndPos: 4, Role: TokenRoleNumber},
				{StartPos: 4, EndPos: 7, Role: TokenRoleKeyword},
			},
		},
		{
			name:     "only first parse func",
			text:     "1234",
			expected: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tree, err := text.NewTreeFromString(tc.text)
			require.NoError(t, err)

			p := New(combinedParseFunc)
			p.ParseAll(tree)
			tokens := p.TokensIntersectingRange(0, math.MaxUint64)
			assert.Equal(t, tc.expected, tokens)
		})
	}

}

func TestThenCombinatorShiftTokens(t *testing.T) {
	// Parse up to ":" as a keyword.
	firstParseFunc := func(iter TrackingRuneIter, state State) Result {
		var n uint64
		for {
			r, err := iter.NextRune()
			if err != nil || r == ':' {
				break
			}
			n++
		}
		return Result{
			NumConsumed: n,
			NextState:   state,
			ComputedTokens: []ComputedToken{
				{
					Length: n,
					Role:   TokenRoleKeyword,
				},
			},
		}
	}

	// Parse rest of the string as a number.
	secondParseFunc := func(iter TrackingRuneIter, state State) Result {
		var n uint64
		for {
			_, err := iter.NextRune()
			if err != nil {
				break
			}
			n++
		}
		return Result{
			NumConsumed: n,
			NextState:   state,
			ComputedTokens: []ComputedToken{
				{
					Length: n,
					Role:   TokenRoleNumber,
				},
			},
		}
	}

	tree, err := text.NewTreeFromString("abc:123")
	require.NoError(t, err)

	combinedParseFunc := Func(firstParseFunc).Then(Func(secondParseFunc))
	p := New(combinedParseFunc)
	p.ParseAll(tree)
	tokens := p.TokensIntersectingRange(0, math.MaxUint64)
	expectedTokens := []Token{
		{StartPos: 0, EndPos: 3, Role: TokenRoleKeyword},
		{StartPos: 3, EndPos: 7, Role: TokenRoleNumber},
	}
	assert.Equal(t, expectedTokens, tokens)
}

func TestMapWithInput(t *testing.T) {
	// Parse func that consumes up to three runes in the input without producing any tokens.
	parseFunc := func(iter TrackingRuneIter, state State) Result {
		var n uint64
		for n < 3 {
			_, err := iter.NextRune()
			if err != nil {
				break
			}
			n++
		}
		return Result{
			NumConsumed: n,
			NextState:   state,
		}
	}

	// MapFn that produces a single token with length of runes read from the iterator.
	// This validates that the iterator returns only runes consumed by the parse func.
	mapFn := func(result Result, iter TrackingRuneIter, state State) Result {
		var n uint64
		for {
			_, err := iter.NextRune()
			if err != nil {
				break
			}
			n++
		}

		result.ComputedTokens = append(result.ComputedTokens, ComputedToken{
			Offset: 0,
			Length: n,
			Role:   TokenRoleNumber,
		})
		return result
	}

	tree, err := text.NewTreeFromString("abc123")
	require.NoError(t, err)

	combinedParseFunc := Func(parseFunc).MapWithInput(mapFn)
	p := New(combinedParseFunc)
	p.ParseAll(tree)
	tokens := p.TokensIntersectingRange(0, math.MaxUint64)
	expectedTokens := []Token{
		{StartPos: 0, EndPos: 3, Role: TokenRoleNumber},
		{StartPos: 3, EndPos: 6, Role: TokenRoleNumber},
	}
	assert.Equal(t, expectedTokens, tokens)
}
