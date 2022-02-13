package parser

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/aretext/aretext/text"
)

func TestThenCombinatorShiftTokens(t *testing.T) {
	// Parse up to ":" as an identifier.
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
					Role:   TokenRoleIdentifier,
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
		{StartPos: 0, EndPos: 3, Role: TokenRoleIdentifier},
		{StartPos: 3, EndPos: 7, Role: TokenRoleNumber},
	}
	assert.Equal(t, expectedTokens, tokens)
}
