package parser

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

type stubState struct{ x int }

func (s stubState) Equals(other State) bool {
	otherStubState, ok := other.(stubState)
	return ok && s.x == otherStubState.x
}

func TestComputationLargestMatchingSubComputation(t *testing.T) {
	testCases := []struct {
		name               string
		builder            func() *computation
		readStartPos       uint64
		readEndPos         uint64
		state              State
		expectedReadLength uint64
	}{
		{
			name: "single computation, start does not match",
			builder: func() *computation {
				return newComputation(2, 2, EmptyState{}, EmptyState{}, nil)
			},
			readStartPos:       2,
			readEndPos:         5,
			state:              EmptyState{},
			expectedReadLength: 0,
		},
		{
			name: "single computation, smaller than range",
			builder: func() *computation {
				return newComputation(2, 2, EmptyState{}, EmptyState{}, nil)
			},
			readStartPos:       0,
			readEndPos:         5,
			state:              EmptyState{},
			expectedReadLength: 2,
		},
		{
			name: "single computation, one less than end of range",
			builder: func() *computation {
				return newComputation(2, 2, EmptyState{}, EmptyState{}, nil)
			},
			readStartPos:       0,
			readEndPos:         3,
			state:              EmptyState{},
			expectedReadLength: 2,
		},
		{
			name: "single computation, equal to range",
			builder: func() *computation {
				return newComputation(2, 2, EmptyState{}, EmptyState{}, nil)
			},
			readStartPos:       0,
			readEndPos:         2,
			state:              EmptyState{},
			expectedReadLength: 2,
		},
		{
			name: "single computation, greater than range",
			builder: func() *computation {
				return newComputation(5, 5, EmptyState{}, EmptyState{}, nil)
			},
			readStartPos:       0,
			readEndPos:         4,
			state:              EmptyState{},
			expectedReadLength: 0,
		},
		{
			name: "multiple computations, match left child",
			builder: func() *computation {
				left := newComputation(3, 3, EmptyState{}, EmptyState{}, nil)
				right := newComputation(5, 5, EmptyState{}, EmptyState{}, nil)
				return left.Append(right)
			},
			readStartPos:       0,
			readEndPos:         4,
			state:              EmptyState{},
			expectedReadLength: 3,
		},
		{
			name: "multiple computations, match right child",
			builder: func() *computation {
				left := newComputation(3, 3, EmptyState{}, EmptyState{}, nil)
				right := newComputation(5, 5, EmptyState{}, EmptyState{}, nil)
				return left.Append(right)
			},
			readStartPos:       3,
			readEndPos:         9,
			state:              EmptyState{},
			expectedReadLength: 5,
		},
		{
			name: "multiple computations, match left child with lookahead",
			builder: func() *computation {
				left := newComputation(10, 3, EmptyState{}, EmptyState{}, nil)
				right := newComputation(5, 5, EmptyState{}, EmptyState{}, nil)
				return left.Append(right)
			},
			readStartPos:       0,
			readEndPos:         15,
			state:              EmptyState{},
			expectedReadLength: 10,
		},
		{
			name: "multiple computations, match right child with lookahead",
			builder: func() *computation {
				left := newComputation(10, 3, EmptyState{}, EmptyState{}, nil)
				right := newComputation(9, 8, EmptyState{}, EmptyState{}, nil)
				return left.Append(right)
			},
			readStartPos:       3,
			readEndPos:         20,
			state:              EmptyState{},
			expectedReadLength: 9,
		},
		{
			name: "match state left child",
			builder: func() *computation {
				left := newComputation(3, 3, stubState{1}, stubState{2}, nil)
				right := newComputation(5, 5, stubState{3}, stubState{4}, nil)
				return left.Append(right)
			},
			readStartPos:       0,
			readEndPos:         4,
			state:              stubState{1},
			expectedReadLength: 3,
		},
		{
			name: "mismatch state left child",
			builder: func() *computation {
				left := newComputation(3, 3, stubState{1}, stubState{2}, nil)
				right := newComputation(5, 5, stubState{3}, stubState{4}, nil)
				return left.Append(right)
			},
			readStartPos:       0,
			readEndPos:         3,
			state:              stubState{99},
			expectedReadLength: 0,
		},
		{
			name: "match state parent",
			builder: func() *computation {
				left := newComputation(3, 3, stubState{1}, stubState{2}, nil)
				right := newComputation(5, 5, stubState{3}, stubState{4}, nil)
				return left.Append(right)
			},
			readStartPos:       0,
			readEndPos:         9,
			state:              stubState{1},
			expectedReadLength: 8,
		},
		{
			name: "mismatch state parent",
			builder: func() *computation {
				left := newComputation(3, 3, stubState{1}, stubState{2}, nil)
				right := newComputation(5, 5, stubState{3}, stubState{4}, nil)
				return left.Append(right)
			},
			readStartPos:       0,
			readEndPos:         9,
			state:              stubState{99},
			expectedReadLength: 0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			c := tc.builder()
			sub := c.LargestMatchingSubComputation(tc.readStartPos, tc.readEndPos, tc.state)
			assert.Equal(t, tc.expectedReadLength, sub.ReadLength())
		})
	}
}

func TestComputationTokensIntersectingRange(t *testing.T) {
	testCases := []struct {
		name           string
		builder        func() *computation
		startPos       uint64
		endPos         uint64
		expectedTokens []Token
	}{
		{
			name: "single computation, no tokens",
			builder: func() *computation {
				return newComputation(1, 1, EmptyState{}, EmptyState{}, nil)
			},
			startPos:       0,
			endPos:         100,
			expectedTokens: nil,
		},
		{
			name: "single computation, single token equals range",
			builder: func() *computation {
				return newComputation(2, 2, EmptyState{}, EmptyState{}, []ComputedToken{
					{Offset: 0, Length: 2},
				})
			},
			startPos: 0,
			endPos:   2,
			expectedTokens: []Token{
				{StartPos: 0, EndPos: 2},
			},
		},
		{
			name: "single computation, multiple tokens in range",
			builder: func() *computation {
				return newComputation(4, 4, EmptyState{}, EmptyState{}, []ComputedToken{
					{Offset: 0, Length: 3},
					{Offset: 3, Length: 1},
				})
			},
			startPos: 0,
			endPos:   4,
			expectedTokens: []Token{
				{StartPos: 0, EndPos: 3},
				{StartPos: 3, EndPos: 4},
			},
		},
		{
			name: "single computation, token ending before range",
			builder: func() *computation {
				return newComputation(4, 4, EmptyState{}, EmptyState{}, []ComputedToken{
					{Offset: 0, Length: 1},
				})
			},
			startPos:       2,
			endPos:         4,
			expectedTokens: nil,
		},
		{
			name: "single computation, token ending at range start",
			builder: func() *computation {
				return newComputation(4, 4, EmptyState{}, EmptyState{}, []ComputedToken{
					{Offset: 0, Length: 1},
				})
			},
			startPos:       1,
			endPos:         4,
			expectedTokens: nil,
		},
		{
			name: "single computation, token starting at range end",
			builder: func() *computation {
				return newComputation(4, 4, EmptyState{}, EmptyState{}, []ComputedToken{
					{Offset: 2, Length: 1},
				})
			},
			startPos:       0,
			endPos:         2,
			expectedTokens: nil,
		},
		{
			name: "single computation, token starting after range end",
			builder: func() *computation {
				return newComputation(4, 4, EmptyState{}, EmptyState{}, []ComputedToken{
					{Offset: 3, Length: 1},
				})
			},
			startPos:       0,
			endPos:         2,
			expectedTokens: nil,
		},
		{
			name: "append two computations, all tokens intersect range",
			builder: func() *computation {
				return newComputation(4, 4, EmptyState{}, EmptyState{}, []ComputedToken{
					{Offset: 0, Length: 4},
				}).Append(
					newComputation(3, 3, EmptyState{}, EmptyState{}, []ComputedToken{
						{Offset: 0, Length: 3},
					}))
			},
			startPos: 0,
			endPos:   7,
			expectedTokens: []Token{
				{StartPos: 0, EndPos: 4},
				{StartPos: 4, EndPos: 7},
			},
		},
		{
			name: "append many computations in sequence, all tokens intersect range",
			builder: func() *computation {
				var c *computation
				for i := 0; i < 10; i++ {
					c = c.Append(newComputation(1, 1, EmptyState{}, EmptyState{}, []ComputedToken{
						{Offset: 0, Length: 1},
					}))
				}
				return c
			},
			startPos: 0,
			endPos:   10,
			expectedTokens: []Token{
				{StartPos: 0, EndPos: 1},
				{StartPos: 1, EndPos: 2},
				{StartPos: 2, EndPos: 3},
				{StartPos: 3, EndPos: 4},
				{StartPos: 4, EndPos: 5},
				{StartPos: 5, EndPos: 6},
				{StartPos: 6, EndPos: 7},
				{StartPos: 7, EndPos: 8},
				{StartPos: 8, EndPos: 9},
				{StartPos: 9, EndPos: 10},
			},
		},
		{
			name: "prepend many computations in sequence, all tokens intersect range",
			builder: func() *computation {
				var c *computation
				for i := 0; i < 10; i++ {
					c = newComputation(1, 1, EmptyState{}, EmptyState{}, []ComputedToken{
						{Offset: 0, Length: 1},
					}).Append(c)
				}
				return c
			},
			startPos: 0,
			endPos:   10,
			expectedTokens: []Token{
				{StartPos: 0, EndPos: 1},
				{StartPos: 1, EndPos: 2},
				{StartPos: 2, EndPos: 3},
				{StartPos: 3, EndPos: 4},
				{StartPos: 4, EndPos: 5},
				{StartPos: 5, EndPos: 6},
				{StartPos: 6, EndPos: 7},
				{StartPos: 7, EndPos: 8},
				{StartPos: 8, EndPos: 9},
				{StartPos: 9, EndPos: 10},
			},
		},
		{
			name: "append two computations each with many sub-computations, all tokens intersect range",
			builder: func() *computation {
				var c1, c2 *computation
				for i := 0; i < 5; i++ {
					c1 = c1.Append(newComputation(1, 1, EmptyState{}, EmptyState{}, []ComputedToken{
						{Offset: 0, Length: 1},
					}))
					c2 = c2.Append(newComputation(1, 1, EmptyState{}, EmptyState{}, []ComputedToken{
						{Offset: 0, Length: 1},
					}))
				}
				return c1.Append(c2)
			},
			startPos: 0,
			endPos:   10,
			expectedTokens: []Token{
				{StartPos: 0, EndPos: 1},
				{StartPos: 1, EndPos: 2},
				{StartPos: 2, EndPos: 3},
				{StartPos: 3, EndPos: 4},
				{StartPos: 4, EndPos: 5},
				{StartPos: 5, EndPos: 6},
				{StartPos: 6, EndPos: 7},
				{StartPos: 7, EndPos: 8},
				{StartPos: 8, EndPos: 9},
				{StartPos: 9, EndPos: 10},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			c := tc.builder()
			tokens := c.TokensIntersectingRange(tc.startPos, tc.endPos)
			assert.Equal(t, tc.expectedTokens, tokens)
		})
	}
}

func TestTokenAtPosition(t *testing.T) {
	testCases := []struct {
		name          string
		builder       func() *computation
		pos           uint64
		expectFound   bool
		expectedToken Token
	}{
		{
			name: "single computation, no tokens",
			builder: func() *computation {
				return newComputation(1, 1, EmptyState{}, EmptyState{}, nil)
			},
			pos:           0,
			expectedToken: Token{},
		},
		{
			name: "single computation, single token containing position at start",
			builder: func() *computation {
				return newComputation(3, 3, EmptyState{}, EmptyState{}, []ComputedToken{
					{Offset: 0, Length: 3},
				})
			},
			pos:           0,
			expectedToken: Token{StartPos: 0, EndPos: 3},
		},
		{
			name: "single computation, single token containing position in middle",
			builder: func() *computation {
				return newComputation(3, 3, EmptyState{}, EmptyState{}, []ComputedToken{
					{Offset: 0, Length: 3},
				})
			},
			pos:           1,
			expectedToken: Token{StartPos: 0, EndPos: 3},
		},
		{
			name: "single computation, single token containing position at end",
			builder: func() *computation {
				return newComputation(3, 3, EmptyState{}, EmptyState{}, []ComputedToken{
					{Offset: 0, Length: 3},
				})
			},
			pos:           2,
			expectedToken: Token{StartPos: 0, EndPos: 3},
		},
		{
			name: "single computation, single token position just past end",
			builder: func() *computation {
				return newComputation(3, 3, EmptyState{}, EmptyState{}, []ComputedToken{
					{Offset: 0, Length: 3},
				})
			},
			pos:           3,
			expectedToken: Token{},
		},
		{
			name: "single computation, multiple tokens, one contains position",
			builder: func() *computation {
				return newComputation(3, 3, EmptyState{}, EmptyState{}, []ComputedToken{
					{Offset: 0, Length: 1},
					{Offset: 1, Length: 1},
					{Offset: 2, Length: 1},
				})
			},
			pos:           1,
			expectedToken: Token{StartPos: 1, EndPos: 2},
		},
		{
			name: "single computation, multiple tokens, none contain position",
			builder: func() *computation {
				return newComputation(3, 3, EmptyState{}, EmptyState{}, []ComputedToken{
					{Offset: 0, Length: 1},
					{Offset: 2, Length: 1},
				})
			},
			pos:           1,
			expectedToken: Token{},
		},
		{
			name: "multiple computations, token in left child",
			builder: func() *computation {
				c := newComputation(3, 3, EmptyState{}, EmptyState{}, []ComputedToken{
					{Offset: 0, Length: 3},
				})
				return c.Append(newComputation(1, 1, EmptyState{}, EmptyState{}, []ComputedToken{}))
			},
			pos:           1,
			expectedToken: Token{StartPos: 0, EndPos: 3},
		},
		{
			name: "multiple computations, token in right child",
			builder: func() *computation {
				c := newComputation(3, 3, EmptyState{}, EmptyState{}, []ComputedToken{})
				return c.Append(newComputation(4, 4, EmptyState{}, EmptyState{}, []ComputedToken{
					{Offset: 0, Length: 4},
				}))
			},
			pos:           4,
			expectedToken: Token{StartPos: 3, EndPos: 7},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			c := tc.builder()
			token := c.TokenAtPosition(tc.pos)
			assert.Equal(t, tc.expectedToken, token)
		})
	}
}

func TestConcatLeafComputations(t *testing.T) {
	testCases := []struct {
		name         string
		computations []*computation
	}{
		{
			name:         "empty",
			computations: nil,
		},
		{
			name: "single computation",
			computations: []*computation{
				newComputation(5, 5, EmptyState{}, EmptyState{}, []ComputedToken{
					{Length: 3},
				}),
			},
		},
		{
			name: "two computations",
			computations: []*computation{
				newComputation(5, 5, EmptyState{}, EmptyState{}, []ComputedToken{
					{Length: 3},
				}),
				newComputation(8, 8, EmptyState{}, EmptyState{}, []ComputedToken{
					{Length: 8},
				}),
			},
		},
		{
			name: "many computations",
			computations: []*computation{
				newComputation(5, 5, EmptyState{}, EmptyState{}, []ComputedToken{
					{Length: 3},
				}),
				newComputation(8, 8, EmptyState{}, EmptyState{}, []ComputedToken{
					{Length: 8},
				}),
				newComputation(2, 2, EmptyState{}, EmptyState{}, []ComputedToken{
					{Length: 2},
				}),
				newComputation(7, 7, EmptyState{}, EmptyState{}, []ComputedToken{
					{Length: 7},
				}),
				newComputation(3, 3, EmptyState{}, EmptyState{}, []ComputedToken{
					{Length: 3},
				}),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			c1 := concatLeafComputations(tc.computations)

			var c2 *computation
			for _, leaf := range tc.computations {
				c2 = c2.Append(leaf)
			}

			assert.Equal(t, c1.ReadLength(), c2.ReadLength())
			assert.Equal(t, c1.ConsumedLength(), c2.ConsumedLength())
			assert.Equal(t, c1.TreeHeight(), c2.TreeHeight())

			actualTokens := c1.TokensIntersectingRange(0, math.MaxUint64)
			expectedTokens := c2.TokensIntersectingRange(0, math.MaxUint64)
			assert.Equal(t, actualTokens, expectedTokens)
		})
	}
}
