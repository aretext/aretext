package parser

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestShiftAdd(t *testing.T) {
	testCases := []struct {
		name     string
		s1       Shift
		s2       Shift
		expected Shift
	}{
		{
			name:     "zero values",
			s1:       Shift{},
			s2:       Shift{},
			expected: Shift{},
		},
		{
			name:     "positive plus positive",
			s1:       Shift{Direction: ShiftDirectionForward, Offset: 1},
			s2:       Shift{Direction: ShiftDirectionForward, Offset: 2},
			expected: Shift{Direction: ShiftDirectionForward, Offset: 3},
		},
		{
			name:     "positive plus zero",
			s1:       Shift{Direction: ShiftDirectionForward, Offset: 1},
			s2:       Shift{},
			expected: Shift{Direction: ShiftDirectionForward, Offset: 1},
		},
		{
			name:     "zero plus positive",
			s1:       Shift{},
			s2:       Shift{Direction: ShiftDirectionForward, Offset: 1},
			expected: Shift{Direction: ShiftDirectionForward, Offset: 1},
		},
		{
			name:     "positive plus negative, sum is positive",
			s1:       Shift{Direction: ShiftDirectionForward, Offset: 3},
			s2:       Shift{Direction: ShiftDirectionBackward, Offset: 1},
			expected: Shift{Direction: ShiftDirectionForward, Offset: 2},
		},
		{
			name:     "positive plus negative, sum is zero",
			s1:       Shift{Direction: ShiftDirectionForward, Offset: 3},
			s2:       Shift{Direction: ShiftDirectionBackward, Offset: 3},
			expected: Shift{Direction: ShiftDirectionForward, Offset: 0},
		},
		{
			name:     "positive plus negative, sum is negative",
			s1:       Shift{Direction: ShiftDirectionForward, Offset: 3},
			s2:       Shift{Direction: ShiftDirectionBackward, Offset: 5},
			expected: Shift{Direction: ShiftDirectionBackward, Offset: 2},
		},
		{
			name:     "negative plus positive, sum is positive",
			s1:       Shift{Direction: ShiftDirectionBackward, Offset: 1},
			s2:       Shift{Direction: ShiftDirectionForward, Offset: 3},
			expected: Shift{Direction: ShiftDirectionForward, Offset: 2},
		},
		{
			name:     "negative plus positive, sum is zero",
			s1:       Shift{Direction: ShiftDirectionBackward, Offset: 3},
			s2:       Shift{Direction: ShiftDirectionForward, Offset: 3},
			expected: Shift{Direction: ShiftDirectionBackward, Offset: 0},
		},
		{
			name:     "negative plus positive, sum is negative",
			s1:       Shift{Direction: ShiftDirectionBackward, Offset: 5},
			s2:       Shift{Direction: ShiftDirectionForward, Offset: 3},
			expected: Shift{Direction: ShiftDirectionBackward, Offset: 2},
		},
		{
			name:     "negative plus negative",
			s1:       Shift{Direction: ShiftDirectionBackward, Offset: 5},
			s2:       Shift{Direction: ShiftDirectionBackward, Offset: 3},
			expected: Shift{Direction: ShiftDirectionBackward, Offset: 8},
		},
		{
			name:     "negative plus zero",
			s1:       Shift{Direction: ShiftDirectionBackward, Offset: 5},
			s2:       Shift{},
			expected: Shift{Direction: ShiftDirectionBackward, Offset: 5},
		},
		{
			name:     "zero plus negative",
			s1:       Shift{},
			s2:       Shift{Direction: ShiftDirectionBackward, Offset: 5},
			expected: Shift{Direction: ShiftDirectionBackward, Offset: 5},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			sum := tc.s1.Add(tc.s2)
			assert.Equal(t, tc.expected, sum)
		})
	}
}

func TestShiftResolve(t *testing.T) {
	testCases := []struct {
		name     string
		pos      uint64
		shift    Shift
		expected uint64
	}{
		{
			name:     "zero shift",
			pos:      9,
			shift:    Shift{},
			expected: 9,
		},
		{
			name:     "shift forward",
			pos:      5,
			shift:    Shift{Direction: ShiftDirectionForward, Offset: 2},
			expected: 7,
		},
		{
			name:     "shift backward, offset less than pos",
			pos:      5,
			shift:    Shift{Direction: ShiftDirectionBackward, Offset: 2},
			expected: 3,
		},
		{
			name:     "shift backward, offset equal pos",
			pos:      5,
			shift:    Shift{Direction: ShiftDirectionBackward, Offset: 5},
			expected: 0,
		},
		{
			name:     "shift backward, offset greater than pos",
			pos:      5,
			shift:    Shift{Direction: ShiftDirectionBackward, Offset: 6},
			expected: 0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.shift.Resolve(tc.pos)
			assert.Equal(t, tc.expected, result)
		})
	}
}
