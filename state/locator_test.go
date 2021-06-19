package state

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/aretext/aretext/selection"
	"github.com/aretext/aretext/text"
)

func TestSelectionEndLocator(t *testing.T) {
	testCases := []struct {
		name              string
		originalText      string
		selectionMode     selection.Mode
		selectionStartPos uint64
		selectionEndPos   uint64
		newText           string
		cursorPos         uint64
		expectedEndPos    uint64
	}{
		{
			name:              "charwise, empty document to non-empty document",
			originalText:      "",
			selectionMode:     selection.ModeChar,
			selectionStartPos: 0,
			selectionEndPos:   0,
			newText:           "ijklmnop",
			cursorPos:         0,
			expectedEndPos:    0,
		},
		{
			name:              "charwise, non-empty document to empty document",
			originalText:      "abcdefgh",
			selectionMode:     selection.ModeChar,
			selectionStartPos: 0,
			selectionEndPos:   2,
			newText:           "",
			cursorPos:         0,
			expectedEndPos:    0,
		},
		{
			name:              "charwise, same line",
			originalText:      "abcdefgh",
			selectionMode:     selection.ModeChar,
			selectionStartPos: 1,
			selectionEndPos:   3,
			newText:           "ijklmnop",
			cursorPos:         4,
			expectedEndPos:    7,
		},
		{
			name:              "charwise, same line, next line shorter",
			originalText:      "abcdefgh",
			selectionMode:     selection.ModeChar,
			selectionStartPos: 1,
			selectionEndPos:   3,
			newText:           "ijklmn",
			cursorPos:         4,
			expectedEndPos:    6,
		},
		{
			name:              "charwise, to end of first line",
			originalText:      "abcd\nefgh\nijkl",
			selectionMode:     selection.ModeChar,
			selectionStartPos: 1,
			selectionEndPos:   3,
			newText:           "ijklm\nnopqrstuv\nwxyz",
			cursorPos:         7,
			expectedEndPos:    10,
		},
		{
			name:              "charwise, past end of first line",
			originalText:      "abcd\nefgh\nijkl",
			selectionMode:     selection.ModeChar,
			selectionStartPos: 1,
			selectionEndPos:   4,
			newText:           "ijklm\nnopqrstuv\nwxyz",
			cursorPos:         7,
			expectedEndPos:    16,
		},
		{
			name:              "charwise, to start of next line",
			originalText:      "abcd\nefgh\nijkl",
			selectionMode:     selection.ModeChar,
			selectionStartPos: 1,
			selectionEndPos:   5,
			newText:           "ijklm\nnopqrstuv\nwxyz",
			cursorPos:         7,
			expectedEndPos:    17,
		},
		{
			name:              "charwise, past start of next line",
			originalText:      "abcd\nefgh\nijkl",
			selectionMode:     selection.ModeChar,
			selectionStartPos: 1,
			selectionEndPos:   7,
			newText:           "ijklm\nnopqrstuv\nwxyz",
			cursorPos:         7,
			expectedEndPos:    19,
		},
		{
			name:              "charwise, past start of next line, next line shorter",
			originalText:      "abcd\nefgh\nijkl",
			selectionMode:     selection.ModeChar,
			selectionStartPos: 1,
			selectionEndPos:   7,
			newText:           "ijklm\nnopqrstuv\nwx\nyz",
			cursorPos:         7,
			expectedEndPos:    18,
		},
		{
			name:              "charwise, past start of next line, next line eof",
			originalText:      "abcd\nefgh\nijkl",
			selectionMode:     selection.ModeChar,
			selectionStartPos: 1,
			selectionEndPos:   7,
			newText:           "ijklm\nnopqrstuv\nw",
			cursorPos:         7,
			expectedEndPos:    17,
		},
		{
			name:              "charwise, past start of next line, new line is last line",
			originalText:      "abcd\nefgh\nijkl",
			selectionMode:     selection.ModeChar,
			selectionStartPos: 3,
			selectionEndPos:   5,
			newText:           "ijklm\nnopq",
			cursorPos:         8,
			expectedEndPos:    10,
		},
		{
			name:              "linewise, empty to non-empty document",
			originalText:      "",
			selectionMode:     selection.ModeLine,
			selectionStartPos: 0,
			selectionEndPos:   0,
			newText:           "ijklm\nnopq",
			cursorPos:         1,
			expectedEndPos:    5,
		},
		{
			name:              "linewise, non-empty to empty document",
			originalText:      "abcdefgh\nhijkl",
			selectionMode:     selection.ModeLine,
			selectionStartPos: 0,
			selectionEndPos:   8,
			newText:           "",
			cursorPos:         0,
			expectedEndPos:    0,
		},
		{
			name:              "linewise, single line to longer single line",
			originalText:      "abc",
			selectionMode:     selection.ModeLine,
			selectionStartPos: 0,
			selectionEndPos:   3,
			newText:           "efghijkl",
			cursorPos:         0,
			expectedEndPos:    8,
		},
		{
			name:              "linewise, single line to shorter single line",
			originalText:      "abcdefgh",
			selectionMode:     selection.ModeLine,
			selectionStartPos: 0,
			selectionEndPos:   8,
			newText:           "efg",
			cursorPos:         0,
			expectedEndPos:    3,
		},
		{
			name:              "linewise, select to line below",
			originalText:      "abcd\nefgh\nijkl",
			selectionMode:     selection.ModeLine,
			selectionStartPos: 0,
			selectionEndPos:   5,
			newText:           "mn\nop\nqr",
			cursorPos:         0,
			expectedEndPos:    5,
		},
		{
			name:              "linewise, select past end of document",
			originalText:      "abcd\nefgh\nijkl",
			selectionMode:     selection.ModeLine,
			selectionStartPos: 0,
			selectionEndPos:   5,
			newText:           "mn\nop",
			cursorPos:         3,
			expectedEndPos:    5,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			textTree, err := text.NewTreeFromString(tc.originalText)
			require.NoError(t, err)
			selector := &selection.Selector{}
			selector.Start(tc.selectionMode, tc.selectionStartPos)
			cursorPos := tc.selectionEndPos
			endLoc := SelectionEndLocator(textTree, cursorPos, selector)

			newTextTree, err := text.NewTreeFromString(tc.newText)
			require.NoError(t, err)
			endPos := endLoc(LocatorParams{
				TextTree:  newTextTree,
				CursorPos: tc.cursorPos,
			})
			assert.Equal(t, tc.expectedEndPos, endPos)
		})
	}
}
