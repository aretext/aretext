package exec

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wedaly/aretext/internal/pkg/text"
)

func TestNextCharInLine(t *testing.T) {
	testCases := []struct {
		name           string
		inputString    string
		initialCursor  cursorState
		numChars       uint64
		expectedCursor cursorState
	}{
		{
			name:           "empty string",
			inputString:    "",
			initialCursor:  cursorState{position: 0},
			numChars:       1,
			expectedCursor: cursorState{position: 0},
		},
		{
			name:           "first char, ASCII string",
			inputString:    "abcd",
			initialCursor:  cursorState{position: 0},
			numChars:       1,
			expectedCursor: cursorState{position: 1},
		},
		{
			name:           "second char, ASCII string",
			inputString:    "abcd",
			initialCursor:  cursorState{position: 1},
			numChars:       1,
			expectedCursor: cursorState{position: 2},
		},
		{
			name:           "last char, ASCII string",
			inputString:    "abcd",
			initialCursor:  cursorState{position: 3},
			numChars:       1,
			expectedCursor: cursorState{position: 3},
		},
		{
			name:           "multi-char grapheme cluster",
			inputString:    "e\u0301xyz",
			initialCursor:  cursorState{position: 0},
			numChars:       1,
			expectedCursor: cursorState{position: 2},
		},
		{
			name:           "up to end of line",
			inputString:    "ab\ncd",
			initialCursor:  cursorState{position: 1},
			numChars:       1,
			expectedCursor: cursorState{position: 1},
		},
		{
			name:           "at end of line",
			inputString:    "ab\ncd",
			initialCursor:  cursorState{position: 2},
			numChars:       1,
			expectedCursor: cursorState{position: 2},
		},
		{
			name:           "end of line with carriage return",
			inputString:    "ab\r\ncd",
			initialCursor:  cursorState{position: 1},
			numChars:       1,
			expectedCursor: cursorState{position: 1},
		},
		{
			name:           "move cursor multiple chars within line",
			inputString:    "abcdefgh",
			initialCursor:  cursorState{position: 2},
			numChars:       3,
			expectedCursor: cursorState{position: 5},
		},
		{
			name:           "move cursor multiple chars to end of line",
			inputString:    "abcd\nefgh",
			initialCursor:  cursorState{position: 1},
			numChars:       2,
			expectedCursor: cursorState{position: 3},
		},
		{
			name:           "move cursor multiple chars past end of line",
			inputString:    "abcd\nefgh",
			initialCursor:  cursorState{position: 1},
			numChars:       5,
			expectedCursor: cursorState{position: 3},
		},
		{
			name:           "move cursor multiple chars past end of string",
			inputString:    "abcd",
			initialCursor:  cursorState{position: 0},
			numChars:       100,
			expectedCursor: cursorState{position: 3},
		},
		{
			name:           "logical offset reset if moved",
			inputString:    "abcd",
			initialCursor:  cursorState{position: 1, logicalOffset: 2},
			numChars:       1,
			expectedCursor: cursorState{position: 2, logicalOffset: 0},
		},
		{
			name:           "logical offset preserved if not moved",
			inputString:    "abcd",
			initialCursor:  cursorState{position: 3, logicalOffset: 2},
			numChars:       1,
			expectedCursor: cursorState{position: 3, logicalOffset: 2},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tree, err := text.NewTreeFromString(tc.inputString)
			require.NoError(t, err)
			state := State{tree, tc.initialCursor}
			loc := NewCharInLineLocator(text.ReadDirectionForward, tc.numChars)
			nextCursor := loc.Locate(&state)
			assert.Equal(t, tc.expectedCursor, nextCursor)
		})
	}
}

func TestPrevCharInLine(t *testing.T) {
	testCases := []struct {
		name           string
		inputString    string
		initialCursor  cursorState
		numChars       uint64
		expectedCursor cursorState
	}{
		{
			name:           "empty",
			inputString:    "",
			initialCursor:  cursorState{position: 0},
			numChars:       1,
			expectedCursor: cursorState{position: 0},
		},
		{
			name:           "last char, ASCII string",
			inputString:    "abcd",
			initialCursor:  cursorState{position: 3},
			numChars:       1,
			expectedCursor: cursorState{position: 2},
		},
		{
			name:           "second-to-last char, ASCII string",
			inputString:    "abcd",
			initialCursor:  cursorState{position: 2},
			numChars:       1,
			expectedCursor: cursorState{position: 1},
		},
		{
			name:           "first char, ASCII string",
			inputString:    "abcd",
			initialCursor:  cursorState{position: 0},
			numChars:       1,
			expectedCursor: cursorState{position: 0},
		},
		{
			name:           "first char in line",
			inputString:    "ab\ncd",
			initialCursor:  cursorState{position: 3},
			numChars:       1,
			expectedCursor: cursorState{position: 3},
		},
		{
			name:           "multi-char grapheme cluster",
			inputString:    "abe\u0301xyz",
			initialCursor:  cursorState{position: 4},
			numChars:       1,
			expectedCursor: cursorState{position: 2},
		},
		{
			name:           "move cursor multiple chars within line",
			inputString:    "abcdefgh",
			initialCursor:  cursorState{position: 4},
			numChars:       3,
			expectedCursor: cursorState{position: 1},
		},
		{
			name:           "move cursor multiple chars to beginning of line",
			inputString:    "ab\ncdefgh",
			initialCursor:  cursorState{position: 5},
			numChars:       2,
			expectedCursor: cursorState{position: 3},
		},
		{
			name:           "move cursor multiple chars past beginning of line",
			inputString:    "ab\ncdefgh",
			initialCursor:  cursorState{position: 5},
			numChars:       4,
			expectedCursor: cursorState{position: 3},
		},
		{
			name:           "logical offset reset if moved",
			inputString:    "abcd",
			initialCursor:  cursorState{position: 3, logicalOffset: 2},
			numChars:       1,
			expectedCursor: cursorState{position: 2, logicalOffset: 0},
		},
		{
			name:           "logical offset preserved if not moved",
			inputString:    "abcd",
			initialCursor:  cursorState{position: 0, logicalOffset: 2},
			numChars:       1,
			expectedCursor: cursorState{position: 0, logicalOffset: 2},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tree, err := text.NewTreeFromString(tc.inputString)
			require.NoError(t, err)
			state := State{tree, tc.initialCursor}
			loc := NewCharInLineLocator(text.ReadDirectionBackward, tc.numChars)
			nextCursor := loc.Locate(&state)
			assert.Equal(t, tc.expectedCursor, nextCursor)
		})
	}
}
