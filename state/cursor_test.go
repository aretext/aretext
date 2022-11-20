package state

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/aretext/aretext/locate"
	"github.com/aretext/aretext/selection"
	"github.com/aretext/aretext/text"
)

func TestMoveCursor(t *testing.T) {
	textTree, err := text.NewTreeFromString("abcd")
	require.NoError(t, err)
	state := NewEditorState(100, 100, nil, nil)
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
			state := NewEditorState(100, 100, nil, nil)
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
			state := NewEditorState(100, 100, nil, nil)
			state.documentBuffer.textTree = textTree
			state.documentBuffer.cursor = tc.initialCursor
			state.documentBuffer.tabSize = 4
			MoveCursorToLineBelow(state, tc.count)
			assert.Equal(t, tc.expectedCursor, state.documentBuffer.cursor)
		})
	}
}

func TestMoveCursorToStartOfSelection(t *testing.T) {
	testCases := []struct {
		name              string
		inputString       string
		selectionMode     selection.Mode
		selectionStartPos uint64
		selectionEndPos   uint64
		expectedCursor    cursorState
	}{
		{
			name:              "empty",
			inputString:       "",
			selectionMode:     selection.ModeChar,
			selectionStartPos: 0,
			selectionEndPos:   0,
			expectedCursor:    cursorState{position: 0},
		},
		{
			name:              "no selection",
			inputString:       "abcdefh",
			selectionMode:     selection.ModeNone,
			selectionStartPos: 1,
			selectionEndPos:   3,
			expectedCursor:    cursorState{position: 3},
		},
		{
			name:              "charwise, select forward",
			inputString:       "abcdefgh",
			selectionMode:     selection.ModeChar,
			selectionStartPos: 1,
			selectionEndPos:   3,
			expectedCursor:    cursorState{position: 1},
		},
		{
			name:              "charwise, select backward",
			inputString:       "abcdefgh",
			selectionMode:     selection.ModeChar,
			selectionStartPos: 3,
			selectionEndPos:   1,
			expectedCursor:    cursorState{position: 1},
		},
		{
			name:              "linewise select forward",
			inputString:       "abcd\nef\ngh\nijkl",
			selectionMode:     selection.ModeLine,
			selectionStartPos: 6,
			selectionEndPos:   9,
			expectedCursor:    cursorState{position: 5},
		},
		{
			name:              "linewise select backward",
			inputString:       "abcd\nef\ngh\nijkl",
			selectionMode:     selection.ModeLine,
			selectionStartPos: 9,
			selectionEndPos:   6,
			expectedCursor:    cursorState{position: 5},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			textTree, err := text.NewTreeFromString(tc.inputString)
			require.NoError(t, err)
			state := NewEditorState(100, 100, nil, nil)
			state.documentBuffer.textTree = textTree
			state.documentBuffer.selector.Start(tc.selectionMode, tc.selectionStartPos)
			state.documentBuffer.cursor = cursorState{position: tc.selectionEndPos}
			MoveCursorToStartOfSelection(state)
			assert.Equal(t, tc.expectedCursor, state.documentBuffer.cursor)
		})
	}
}

func TestSelectRange(t *testing.T) {
	textTree, err := text.NewTreeFromString("abc def ghi")
	require.NoError(t, err)
	state := NewEditorState(100, 100, nil, nil)
	state.documentBuffer.textTree = textTree
	state.documentBuffer.selector.Start(selection.ModeLine, 0)
	state.documentBuffer.cursor = cursorState{position: 3}

	SelectRange(state, func(LocatorParams) (uint64, uint64) {
		return 5, 7
	})
	assert.Equal(t, selection.ModeChar, state.documentBuffer.SelectionMode())
	assert.Equal(t, selection.Region{StartPos: 5, EndPos: 7}, state.documentBuffer.SelectedRegion())
	assert.Equal(t, cursorState{position: 6}, state.documentBuffer.cursor)
}
