package parser

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestComputationLargestSubComputationInRange(t *testing.T) {
	testCases := []struct {
		name               string
		builder            func() *Computation
		readStartPos       uint64
		readEndPos         uint64
		expectedReadLength uint64
	}{
		{
			name: "single computation, start does not match",
			builder: func() *Computation {
				return NewComputation(2, 2, nil)
			},
			readStartPos:       2,
			readEndPos:         5,
			expectedReadLength: 0,
		},
		{
			name: "single computation, smaller than range",
			builder: func() *Computation {
				return NewComputation(2, 2, nil)
			},
			readStartPos:       0,
			readEndPos:         5,
			expectedReadLength: 2,
		},
		{
			name: "single computation, equal to range",
			builder: func() *Computation {
				return NewComputation(2, 2, nil)
			},
			readStartPos:       0,
			readEndPos:         2,
			expectedReadLength: 2,
		},
		{
			name: "single computation, greater than range",
			builder: func() *Computation {
				return NewComputation(5, 5, nil)
			},
			readStartPos:       0,
			readEndPos:         4,
			expectedReadLength: 0,
		},
		{
			name: "multiple computations, match left child",
			builder: func() *Computation {
				left := NewComputation(3, 3, nil)
				right := NewComputation(5, 5, nil)
				return left.Append(right)
			},
			readStartPos:       0,
			readEndPos:         3,
			expectedReadLength: 3,
		},
		{
			name: "multiple computations, match right child",
			builder: func() *Computation {
				left := NewComputation(3, 3, nil)
				right := NewComputation(5, 5, nil)
				return left.Append(right)
			},
			readStartPos:       3,
			readEndPos:         8,
			expectedReadLength: 5,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			c := tc.builder()
			sub := c.LargestSubComputationInRange(tc.readStartPos, tc.readEndPos)
			assert.Equal(t, tc.expectedReadLength, sub.ReadLength())
		})
	}
}

func TestComputationTokensIntersectingRange(t *testing.T) {
	testCases := []struct {
		name           string
		builder        func() *Computation
		startPos       uint64
		endPos         uint64
		expectedTokens []Token
	}{
		{
			name: "single computation, no tokens",
			builder: func() *Computation {
				return NewComputation(1, 1, nil)
			},
			startPos:       0,
			endPos:         100,
			expectedTokens: nil,
		},
		{
			name: "single computation, single token equals range",
			builder: func() *Computation {
				return NewComputation(2, 2, []ComputedToken{
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
			builder: func() *Computation {
				return NewComputation(4, 4, []ComputedToken{
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
			builder: func() *Computation {
				return NewComputation(4, 4, []ComputedToken{
					{Offset: 0, Length: 1},
				})
			},
			startPos:       2,
			endPos:         4,
			expectedTokens: nil,
		},
		{
			name: "single computation, token ending at range start",
			builder: func() *Computation {
				return NewComputation(4, 4, []ComputedToken{
					{Offset: 0, Length: 1},
				})
			},
			startPos:       1,
			endPos:         4,
			expectedTokens: nil,
		},
		{
			name: "single computation, token starting at range end",
			builder: func() *Computation {
				return NewComputation(4, 4, []ComputedToken{
					{Offset: 2, Length: 1},
				})
			},
			startPos:       0,
			endPos:         2,
			expectedTokens: nil,
		},
		{
			name: "single computation, token starting after range end",
			builder: func() *Computation {
				return NewComputation(4, 4, []ComputedToken{
					{Offset: 3, Length: 1},
				})
			},
			startPos:       0,
			endPos:         2,
			expectedTokens: nil,
		},
		{
			name: "append two computations, all tokens intersect range",
			builder: func() *Computation {
				return NewComputation(4, 4, []ComputedToken{
					{Offset: 0, Length: 4},
				}).Append(
					NewComputation(3, 3, []ComputedToken{
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
			builder: func() *Computation {
				var c *Computation
				for i := 0; i < 10; i++ {
					c = c.Append(NewComputation(1, 1, []ComputedToken{
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
			builder: func() *Computation {
				var c *Computation
				for i := 0; i < 10; i++ {
					c = NewComputation(1, 1, []ComputedToken{
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
			builder: func() *Computation {
				var c1, c2 *Computation
				for i := 0; i < 5; i++ {
					c1 = c1.Append(NewComputation(1, 1, []ComputedToken{
						{Offset: 0, Length: 1},
					}))
					c2 = c2.Append(NewComputation(1, 1, []ComputedToken{
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