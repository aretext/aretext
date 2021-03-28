package state

import (
	"testing"

	"github.com/aretext/aretext/locate"
	"github.com/aretext/aretext/text"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMoveCursor(t *testing.T) {
	textTree, err := text.NewTreeFromString("abcd")
	require.NoError(t, err)
	state := NewEditorState(100, 100, nil)
	state.documentBuffer.textTree = textTree
	state.documentBuffer.cursor.position = 2
	MoveCursor(state, func(params LocatorParams) uint64 {
		return locate.NextCharInLine(params.TextTree, 1, false, params.CursorPos)
	})
	assert.Equal(t, uint64(3), state.documentBuffer.cursor.position)
}

func TestMoveCursorToLineAbove(t *testing.T) {
	testCases := []struct {
		name           string
		inputString    string
		count          uint64
		initialCursor  cursorState
		expectedCursor cursorState
	}{
		{
			name:           "empty string, move up one line",
			inputString:    "",
			count:          1,
			initialCursor:  cursorState{position: 0},
			expectedCursor: cursorState{position: 0},
		},
		{
			name:           "single line, move up one line",
			inputString:    "abcdefgh",
			count:          1,
			initialCursor:  cursorState{position: 3},
			expectedCursor: cursorState{position: 3},
		},
		{
			name:           "single line, move up one line with logical offset",
			inputString:    "abcdefgh",
			count:          1,
			initialCursor:  cursorState{position: 7, logicalOffset: 4},
			expectedCursor: cursorState{position: 7, logicalOffset: 4},
		},
		{
			name:           "two lines, move up one line at start of line",
			inputString:    "abcdefgh\nijklm\nopqrs",
			count:          1,
			initialCursor:  cursorState{position: 15},
			expectedCursor: cursorState{position: 9},
		},
		{
			name:           "two lines, move up at same offset",
			inputString:    "abcdefgh\nijklmnop",
			count:          1,
			initialCursor:  cursorState{position: 11},
			expectedCursor: cursorState{position: 2},
		},
		{
			name:           "two lines, move up from shorter line to longer line",
			inputString:    "abcdefgh\nijk",
			count:          1,
			initialCursor:  cursorState{position: 11},
			expectedCursor: cursorState{position: 2},
		},
		{
			name:           "two lines, move up from shorter line with logical offset to longer line",
			inputString:    "abcdefgh\nijk",
			count:          1,
			initialCursor:  cursorState{position: 11, logicalOffset: 2},
			expectedCursor: cursorState{position: 4},
		},
		{
			name:           "two lines, move up from longer line to shorter line",
			inputString:    "abc\nefghijkl",
			count:          1,
			initialCursor:  cursorState{position: 9},
			expectedCursor: cursorState{position: 2, logicalOffset: 3},
		},
		{
			name:           "two lines, move up from longer line with logical offset to shorter line",
			inputString:    "abc\nefghijkl",
			count:          1,
			initialCursor:  cursorState{position: 11, logicalOffset: 5},
			expectedCursor: cursorState{position: 2, logicalOffset: 10},
		},
		{
			name:           "two lines, move up with multi-char grapheme cluster",
			inputString:    "abcde\u0301fgh\nijklmnopqrstuv",
			count:          1,
			initialCursor:  cursorState{position: 15},
			expectedCursor: cursorState{position: 6},
		},
		{
			name:           "three lines, move up from longer line to empty line",
			inputString:    "abcd\n\nefghijkl",
			count:          1,
			initialCursor:  cursorState{position: 8},
			expectedCursor: cursorState{position: 5, logicalOffset: 2},
		},
		{
			name:           "move up multiple lines",
			inputString:    "abcd\nefgh\nijkl",
			count:          2,
			initialCursor:  cursorState{position: 12},
			expectedCursor: cursorState{position: 2},
		},
		{
			name:           "move up to tab",
			inputString:    "abcd\ne\tefg\nhijkl",
			count:          1,
			initialCursor:  cursorState{position: 13},
			expectedCursor: cursorState{position: 6, logicalOffset: 1},
		},
		{
			name:           "move up from tab",
			inputString:    "abcd\ne\tefg\nhijkl",
			count:          1,
			initialCursor:  cursorState{position: 6, logicalOffset: 2},
			expectedCursor: cursorState{position: 3},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			textTree, err := text.NewTreeFromString(tc.inputString)
			require.NoError(t, err)
			state := NewEditorState(100, 100, nil)
			state.documentBuffer.textTree = textTree
			state.documentBuffer.cursor = tc.initialCursor
			state.documentBuffer.tabSize = 4
			MoveCursorToLineAbove(state, tc.count)
			assert.Equal(t, tc.expectedCursor, state.documentBuffer.cursor)
		})
	}
}

func TestMoveCursorToLineBelow(t *testing.T) {
	testCases := []struct {
		name           string
		inputString    string
		count          uint64
		initialCursor  cursorState
		expectedCursor cursorState
	}{
		{
			name:           "empty string, move down one line",
			inputString:    "",
			count:          1,
			initialCursor:  cursorState{position: 0},
			expectedCursor: cursorState{position: 0},
		},
		{
			name:           "single line, move down one line",
			inputString:    "abcdefgh",
			count:          1,
			initialCursor:  cursorState{position: 3},
			expectedCursor: cursorState{position: 3},
		},
		{
			name:           "single line, move down one line with logical offset",
			inputString:    "abcdefgh",
			count:          1,
			initialCursor:  cursorState{position: 7, logicalOffset: 4},
			expectedCursor: cursorState{position: 7, logicalOffset: 4},
		},
		{
			name:           "two lines, move down one line at start of line",
			inputString:    "abcdefgh\nijklm\nopqrs",
			count:          1,
			initialCursor:  cursorState{position: 9},
			expectedCursor: cursorState{position: 15},
		},
		{
			name:           "two lines, move down at same offset",
			inputString:    "abcdefgh\nijklmnop",
			count:          1,
			initialCursor:  cursorState{position: 2},
			expectedCursor: cursorState{position: 11},
		},
		{
			name:           "two lines, move down from shorter line to longer line",
			inputString:    "abc\nefghijkl",
			count:          1,
			initialCursor:  cursorState{position: 2},
			expectedCursor: cursorState{position: 6},
		},
		{
			name:           "two lines, move down from shorter line with logical offset to longer line",
			inputString:    "abc\nefghijkl",
			count:          1,
			initialCursor:  cursorState{position: 2, logicalOffset: 3},
			expectedCursor: cursorState{position: 9},
		},
		{
			name:           "two lines, move down from longer line to shorter line",
			inputString:    "abcdefgh\nijkl",
			count:          1,
			initialCursor:  cursorState{position: 7},
			expectedCursor: cursorState{position: 12, logicalOffset: 4},
		},
		{
			name:           "two lines, move down from longer line with logical offset to shorter line",
			inputString:    "abcdefgh\nijkl",
			count:          1,
			initialCursor:  cursorState{position: 7, logicalOffset: 5},
			expectedCursor: cursorState{position: 12, logicalOffset: 9},
		},
		{
			name:           "two lines, move down with multi-char grapheme cluster",
			inputString:    "abcdefgh\nijklmno\u0301pqrstuv",
			count:          1,
			initialCursor:  cursorState{position: 7},
			expectedCursor: cursorState{position: 17},
		},
		{
			name:           "three lines, move down from longer line to empty line",
			inputString:    "abcdefgh\n\nijkl",
			count:          1,
			initialCursor:  cursorState{position: 2},
			expectedCursor: cursorState{position: 9, logicalOffset: 2},
		},
		{
			name:           "move down multiple lines",
			inputString:    "abcd\nefgh\nijkl",
			count:          2,
			initialCursor:  cursorState{position: 2},
			expectedCursor: cursorState{position: 12},
		},
		{
			name:           "move down past newline at end of text",
			inputString:    "abcd\nefgh\nijkl\n",
			count:          1,
			initialCursor:  cursorState{position: 12},
			expectedCursor: cursorState{position: 15, logicalOffset: 2},
		},
		{
			name:           "move down past single newline",
			inputString:    "\n",
			count:          1,
			initialCursor:  cursorState{position: 0},
			expectedCursor: cursorState{position: 1},
		},
		{
			name:           "move down to tab",
			inputString:    "abcd\ne\tefg\nhijkl",
			count:          1,
			initialCursor:  cursorState{position: 3},
			expectedCursor: cursorState{position: 6, logicalOffset: 2},
		},
		{
			name:           "move down from tab",
			inputString:    "abcd\ne\tefg\nhijkl",
			count:          1,
			initialCursor:  cursorState{position: 6, logicalOffset: 1},
			expectedCursor: cursorState{position: 13},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			textTree, err := text.NewTreeFromString(tc.inputString)
			require.NoError(t, err)
			state := NewEditorState(100, 100, nil)
			state.documentBuffer.textTree = textTree
			state.documentBuffer.cursor = tc.initialCursor
			state.documentBuffer.tabSize = 4
			MoveCursorToLineBelow(state, tc.count)
			assert.Equal(t, tc.expectedCursor, state.documentBuffer.cursor)
		})
	}
}
