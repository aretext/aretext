package exec

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wedaly/aretext/internal/pkg/text"
)

func TestNextCharInLine(t *testing.T) {
	testCases := []struct {
		name                   string
		inputString            string
		initialCursor          cursorState
		numChars               uint64
		includeEndOfLineOrFile bool
		expectedCursor         cursorState
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
		{
			name:                   "include end of line",
			inputString:            "abcd\nefgh",
			initialCursor:          cursorState{position: 2},
			numChars:               5,
			includeEndOfLineOrFile: true,
			expectedCursor:         cursorState{position: 4},
		},
		{
			name:                   "include end of file",
			inputString:            "abcd",
			initialCursor:          cursorState{position: 2},
			numChars:               5,
			includeEndOfLineOrFile: true,
			expectedCursor:         cursorState{position: 4},
		},
		{
			name:                   "first character when including end of line or file",
			inputString:            "abcdefgh",
			initialCursor:          cursorState{position: 0},
			numChars:               1,
			includeEndOfLineOrFile: true,
			expectedCursor:         cursorState{position: 1},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tree, err := text.NewTreeFromString(tc.inputString)
			require.NoError(t, err)
			state := State{tree, tc.initialCursor}
			loc := NewCharInLineLocator(text.ReadDirectionForward, tc.numChars, tc.includeEndOfLineOrFile)
			nextCursor := loc.Locate(&state)
			assert.Equal(t, tc.expectedCursor, nextCursor)
		})
	}
}

func TestPrevCharInLine(t *testing.T) {
	testCases := []struct {
		name                   string
		inputString            string
		initialCursor          cursorState
		numChars               uint64
		includeEndOfLineOrFile bool
		expectedCursor         cursorState
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
		{
			name:                   "include end of previous line",
			inputString:            "abcd\nefgh",
			initialCursor:          cursorState{position: 5},
			numChars:               1,
			includeEndOfLineOrFile: true,
			expectedCursor:         cursorState{position: 4},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tree, err := text.NewTreeFromString(tc.inputString)
			require.NoError(t, err)
			state := State{tree, tc.initialCursor}
			loc := NewCharInLineLocator(text.ReadDirectionBackward, tc.numChars, tc.includeEndOfLineOrFile)
			nextCursor := loc.Locate(&state)
			assert.Equal(t, tc.expectedCursor, nextCursor)
		})
	}
}

func TestOntoLineLocator(t *testing.T) {
	testCases := []struct {
		name           string
		inputString    string
		initialCursor  cursorState
		expectedCursor cursorState
	}{
		{
			name:           "empty string, cursor at origin",
			inputString:    "",
			initialCursor:  cursorState{position: 0},
			expectedCursor: cursorState{position: 0},
		},
		{
			name:           "empty string, cursor past origin",
			inputString:    "",
			initialCursor:  cursorState{position: 1},
			expectedCursor: cursorState{position: 0},
		},
		{
			name:           "cursor already on line at beginning of file",
			inputString:    "abcd",
			initialCursor:  cursorState{position: 0},
			expectedCursor: cursorState{position: 0},
		},
		{
			name:           "cursor already on line at in middle of line",
			inputString:    "abcd",
			initialCursor:  cursorState{position: 2},
			expectedCursor: cursorState{position: 2},
		},
		{
			name:           "cursor already on line at beginning of line",
			inputString:    "abcd\nefg",
			initialCursor:  cursorState{position: 5},
			expectedCursor: cursorState{position: 5},
		},
		{
			name:           "cursor already on line at end of line",
			inputString:    "abcd\nefg",
			initialCursor:  cursorState{position: 3},
			expectedCursor: cursorState{position: 3},
		},
		{
			name:           "cursor past end of file by single char",
			inputString:    "abcd",
			initialCursor:  cursorState{position: 4},
			expectedCursor: cursorState{position: 3},
		},
		{
			name:           "cursor past end of file by multiple chars",
			inputString:    "abcd",
			initialCursor:  cursorState{position: 10},
			expectedCursor: cursorState{position: 3},
		},
		{
			name:           "cursor on newline",
			inputString:    "abcd\nefgh",
			initialCursor:  cursorState{position: 4},
			expectedCursor: cursorState{position: 3},
		},
		{
			name:           "cursor on newline preceded by newline",
			inputString:    "abcd\n\nefgh",
			initialCursor:  cursorState{position: 5},
			expectedCursor: cursorState{position: 4},
		},
		{
			name:           "cursor at newline in file with only newline",
			inputString:    "\n",
			initialCursor:  cursorState{position: 0},
			expectedCursor: cursorState{position: 0},
		},
		{
			name:           "cursor at newline in file with multiple newlines",
			inputString:    "\n\n\n",
			initialCursor:  cursorState{position: 2},
			expectedCursor: cursorState{position: 1},
		},
		{
			name:           "cursor at newline with carriage return, on line feed",
			inputString:    "abcd\r\nefgh",
			initialCursor:  cursorState{position: 5},
			expectedCursor: cursorState{position: 3},
		},
		{
			name:           "cursor at newline with carriage return, on carriage return",
			inputString:    "abcd\r\nefgh",
			initialCursor:  cursorState{position: 4},
			expectedCursor: cursorState{position: 3},
		},
		{
			name:           "cursor on newline ending with multi-char grapheme cluster",
			inputString:    "abcde\u0301\nfgh",
			initialCursor:  cursorState{position: 6},
			expectedCursor: cursorState{position: 4},
		},
		{
			name:           "cursor on newline with carriage return ending with multi-char grapheme cluster",
			inputString:    "abcde\u0301\r\nfgh",
			initialCursor:  cursorState{position: 7},
			expectedCursor: cursorState{position: 4},
		},
		{
			name:           "cursor past end of text ending with multi-char grapheme cluster",
			inputString:    "abcde\u0301",
			initialCursor:  cursorState{position: 6},
			expectedCursor: cursorState{position: 4},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tree, err := text.NewTreeFromString(tc.inputString)
			require.NoError(t, err)
			state := State{tree, tc.initialCursor}
			loc := NewOntoLineLocator()
			nextCursor := loc.Locate(&state)
			assert.Equal(t, tc.expectedCursor, nextCursor)
		})
	}
}

func TestRelativeLineLocator(t *testing.T) {
	testCases := []struct {
		name           string
		inputString    string
		direction      text.ReadDirection
		count          uint64
		initialCursor  cursorState
		expectedCursor cursorState
	}{
		{
			name:           "empty string, move up one line",
			inputString:    "",
			direction:      text.ReadDirectionBackward,
			count:          1,
			initialCursor:  cursorState{position: 0},
			expectedCursor: cursorState{position: 0},
		},
		{
			name:           "empty string, move down one line",
			inputString:    "",
			direction:      text.ReadDirectionForward,
			count:          1,
			initialCursor:  cursorState{position: 0},
			expectedCursor: cursorState{position: 0},
		},
		{
			name:           "single line, move up one line",
			inputString:    "abcdefgh",
			direction:      text.ReadDirectionBackward,
			count:          1,
			initialCursor:  cursorState{position: 3},
			expectedCursor: cursorState{position: 3},
		},
		{
			name:           "single line, move up one line with logical offset",
			inputString:    "abcdefgh",
			direction:      text.ReadDirectionBackward,
			count:          1,
			initialCursor:  cursorState{position: 7, logicalOffset: 4},
			expectedCursor: cursorState{position: 7, logicalOffset: 4},
		},
		{
			name:           "single line, move down one line",
			inputString:    "abcdefgh",
			direction:      text.ReadDirectionForward,
			count:          1,
			initialCursor:  cursorState{position: 3},
			expectedCursor: cursorState{position: 3},
		},
		{
			name:           "single line, move down one line with logical offset",
			inputString:    "abcdefgh",
			direction:      text.ReadDirectionForward,
			count:          1,
			initialCursor:  cursorState{position: 7, logicalOffset: 4},
			expectedCursor: cursorState{position: 7, logicalOffset: 4},
		},
		{
			name:           "two lines, move up one line at start of line",
			inputString:    "abcdefgh\nijklm\nopqrs",
			direction:      text.ReadDirectionBackward,
			count:          1,
			initialCursor:  cursorState{position: 15},
			expectedCursor: cursorState{position: 9},
		},
		{
			name:           "two lines, move down one line at start of line",
			inputString:    "abcdefgh\nijklm\nopqrs",
			direction:      text.ReadDirectionForward,
			count:          1,
			initialCursor:  cursorState{position: 9},
			expectedCursor: cursorState{position: 15},
		},
		{
			name:           "two lines, move up at same offset",
			inputString:    "abcdefgh\nijklmnop",
			direction:      text.ReadDirectionBackward,
			count:          1,
			initialCursor:  cursorState{position: 11},
			expectedCursor: cursorState{position: 2},
		},
		{
			name:           "two lines, move down at same offset",
			inputString:    "abcdefgh\nijklmnop",
			direction:      text.ReadDirectionForward,
			count:          1,
			initialCursor:  cursorState{position: 2},
			expectedCursor: cursorState{position: 11},
		},
		{
			name:           "two lines, move up from shorter line to longer line",
			inputString:    "abcdefgh\nijk",
			direction:      text.ReadDirectionBackward,
			count:          1,
			initialCursor:  cursorState{position: 11},
			expectedCursor: cursorState{position: 2},
		},
		{
			name:           "two lines, move up from shorter line with logical offset to longer line",
			inputString:    "abcdefgh\nijk",
			direction:      text.ReadDirectionBackward,
			count:          1,
			initialCursor:  cursorState{position: 11, logicalOffset: 2},
			expectedCursor: cursorState{position: 4},
		},
		{
			name:           "two lines, move up from longer line to shorter line",
			inputString:    "abc\nefghijkl",
			direction:      text.ReadDirectionBackward,
			count:          1,
			initialCursor:  cursorState{position: 9},
			expectedCursor: cursorState{position: 2, logicalOffset: 3},
		},
		{
			name:           "two lines, move up from longer line with logical offset to shorter line",
			inputString:    "abc\nefghijkl",
			direction:      text.ReadDirectionBackward,
			count:          1,
			initialCursor:  cursorState{position: 11, logicalOffset: 5},
			expectedCursor: cursorState{position: 2, logicalOffset: 10},
		},
		{
			name:           "two lines, move up with multi-char grapheme cluster",
			inputString:    "abcde\u0301fgh\nijklmnopqrstuv",
			direction:      text.ReadDirectionBackward,
			count:          1,
			initialCursor:  cursorState{position: 15},
			expectedCursor: cursorState{position: 6},
		},
		{
			name:           "two lines, move down from shorter line to longer line",
			inputString:    "abc\nefghijkl",
			direction:      text.ReadDirectionForward,
			count:          1,
			initialCursor:  cursorState{position: 2},
			expectedCursor: cursorState{position: 6},
		},
		{
			name:           "two lines, move down from shorter line with logical offset to longer line",
			inputString:    "abc\nefghijkl",
			direction:      text.ReadDirectionForward,
			count:          1,
			initialCursor:  cursorState{position: 2, logicalOffset: 3},
			expectedCursor: cursorState{position: 9},
		},
		{
			name:           "two lines, move down from longer line to shorter line",
			inputString:    "abcdefgh\nijkl",
			direction:      text.ReadDirectionForward,
			count:          1,
			initialCursor:  cursorState{position: 7},
			expectedCursor: cursorState{position: 12, logicalOffset: 4},
		},
		{
			name:           "two lines, move down from longer line with logical offset to shorter line",
			inputString:    "abcdefgh\nijkl",
			direction:      text.ReadDirectionForward,
			count:          1,
			initialCursor:  cursorState{position: 7, logicalOffset: 5},
			expectedCursor: cursorState{position: 12, logicalOffset: 9},
		},
		{
			name:           "two lines, move down with multi-char grapheme cluster",
			inputString:    "abcdefgh\nijklmno\u0301pqrstuv",
			direction:      text.ReadDirectionForward,
			count:          1,
			initialCursor:  cursorState{position: 7},
			expectedCursor: cursorState{position: 17},
		},
		{
			name:           "three lines, move up from longer line to empty line",
			inputString:    "abcd\n\nefghijkl",
			direction:      text.ReadDirectionBackward,
			count:          1,
			initialCursor:  cursorState{position: 8},
			expectedCursor: cursorState{position: 5, logicalOffset: 2},
		},
		{
			name:           "three lines, move down from longer line to empty line",
			inputString:    "abcdefgh\n\nijkl",
			direction:      text.ReadDirectionForward,
			count:          1,
			initialCursor:  cursorState{position: 2},
			expectedCursor: cursorState{position: 9, logicalOffset: 2},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tree, err := text.NewTreeFromString(tc.inputString)
			require.NoError(t, err)
			state := State{tree, tc.initialCursor}
			loc := NewRelativeLineLocator(tc.direction, tc.count)
			nextCursor := loc.Locate(&state)
			assert.Equal(t, tc.expectedCursor, nextCursor)
		})
	}
}
