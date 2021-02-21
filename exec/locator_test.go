package exec

import (
	"testing"

	"github.com/aretext/aretext/syntax"
	"github.com/aretext/aretext/text"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
			textTree, err := text.NewTreeFromString(tc.inputString)
			require.NoError(t, err)
			state := BufferState{
				textTree: textTree,
				cursor:   tc.initialCursor,
			}
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
			textTree, err := text.NewTreeFromString(tc.inputString)
			require.NoError(t, err)
			state := BufferState{
				textTree: textTree,
				cursor:   tc.initialCursor,
			}
			loc := NewCharInLineLocator(text.ReadDirectionBackward, tc.numChars, tc.includeEndOfLineOrFile)
			nextCursor := loc.Locate(&state)
			assert.Equal(t, tc.expectedCursor, nextCursor)
		})
	}
}

func TestPrevCharLocator(t *testing.T) {
	testCases := []struct {
		name           string
		inputString    string
		initialCursor  cursorState
		count          uint64
		expectedCursor cursorState
	}{
		{
			name:           "empty string",
			inputString:    "",
			initialCursor:  cursorState{position: 0},
			count:          1,
			expectedCursor: cursorState{position: 0},
		},
		{
			name:           "back single char, same line",
			inputString:    "abc\ndef",
			initialCursor:  cursorState{position: 5},
			count:          1,
			expectedCursor: cursorState{position: 4},
		},
		{
			name:           "back single char, prev line",
			inputString:    "abc\ndef",
			initialCursor:  cursorState{position: 3},
			count:          1,
			expectedCursor: cursorState{position: 2},
		},
		{
			name:           "back multi-char grapheme cluster",
			inputString:    "e\u0301xyz",
			initialCursor:  cursorState{position: 2},
			count:          1,
			expectedCursor: cursorState{position: 0},
		},
		{
			name:           "back multiple chars, within document",
			inputString:    "abc\ndef",
			initialCursor:  cursorState{position: 5},
			count:          3,
			expectedCursor: cursorState{position: 2},
		},
		{
			name:           "back multiple chars, outside document",
			inputString:    "abc\ndef",
			initialCursor:  cursorState{position: 5},
			count:          100,
			expectedCursor: cursorState{position: 0},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			textTree, err := text.NewTreeFromString(tc.inputString)
			require.NoError(t, err)
			state := BufferState{
				textTree: textTree,
				cursor:   tc.initialCursor,
			}
			loc := NewPrevCharLocator(tc.count)
			nextCursor := loc.Locate(&state)
			assert.Equal(t, tc.expectedCursor, nextCursor)
		})
	}
}

func TestPrevAutoIndentLocator(t *testing.T) {
	testCases := []struct {
		name              string
		inputString       string
		autoIndentEnabled bool
		initialCursor     cursorState
		expectedCursor    cursorState
	}{
		{
			name:           "empty string",
			inputString:    "",
			initialCursor:  cursorState{position: 0},
			expectedCursor: cursorState{position: 0},
		},
		{
			name:           "multiple tabs, autoindent disabled",
			inputString:    "\t\t",
			initialCursor:  cursorState{position: 2},
			expectedCursor: cursorState{position: 2},
		},
		{
			name:              "single space, autoindent enabled",
			inputString:       " ",
			autoIndentEnabled: true,
			initialCursor:     cursorState{position: 1},
			expectedCursor:    cursorState{position: 0},
		},
		{
			name:              "multiple spaces, autoindent enabled",
			inputString:       "        ",
			autoIndentEnabled: true,
			initialCursor:     cursorState{position: 8},
			expectedCursor:    cursorState{position: 4},
		},
		{
			name:              "multiple tabs, autoindent enabled",
			inputString:       "\t\t",
			autoIndentEnabled: true,
			initialCursor:     cursorState{position: 2},
			expectedCursor:    cursorState{position: 1},
		},
		{
			name:              "mixed tabs and spaces, autoindent enabled",
			inputString:       " \t",
			autoIndentEnabled: true,
			initialCursor:     cursorState{position: 2},
			expectedCursor:    cursorState{position: 0},
		},
		{
			name:              "no tabs or spaces, autoindent enabled",
			inputString:       "ab",
			autoIndentEnabled: true,
			initialCursor:     cursorState{position: 2},
			expectedCursor:    cursorState{position: 2},
		},
		{
			name:              "start of line, autoindent enabled",
			inputString:       "ab\ncd",
			autoIndentEnabled: true,
			initialCursor:     cursorState{position: 2},
			expectedCursor:    cursorState{position: 2},
		},
		{
			name:              "end of document, autoindent enabled",
			inputString:       "ab\n\n",
			autoIndentEnabled: true,
			initialCursor:     cursorState{position: 3},
			expectedCursor:    cursorState{position: 3},
		},
		{
			name:              "spaces within line aligned, autoindent enabled",
			inputString:       "abcd    ef",
			autoIndentEnabled: true,
			initialCursor:     cursorState{position: 8},
			expectedCursor:    cursorState{position: 4},
		},
		{
			name:              "spaces within line misaligned, autoindent enabled",
			inputString:       "ab    cd",
			autoIndentEnabled: true,
			initialCursor:     cursorState{position: 6},
			expectedCursor:    cursorState{position: 4},
		},
		{
			name:              "tabs within line, autoindent enabled",
			inputString:       "ab\t\tcd",
			autoIndentEnabled: true,
			initialCursor:     cursorState{position: 4},
			expectedCursor:    cursorState{position: 3},
		},
		{
			name:              "spaces within line but not before cursor, autoindent enabled",
			inputString:       "ab    cdef",
			autoIndentEnabled: true,
			initialCursor:     cursorState{position: 7},
			expectedCursor:    cursorState{position: 7},
		},
		{
			name:              "spaces at end of line less than tab size, autoident enabled",
			inputString:       "abcdef  ",
			autoIndentEnabled: true,
			initialCursor:     cursorState{position: 8},
			expectedCursor:    cursorState{position: 6},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			textTree, err := text.NewTreeFromString(tc.inputString)
			require.NoError(t, err)
			state := BufferState{
				textTree:   textTree,
				cursor:     tc.initialCursor,
				autoIndent: tc.autoIndentEnabled,
				tabSize:    4,
			}
			loc := NewPrevAutoIndentLocator()
			nextCursor := loc.Locate(&state)
			assert.Equal(t, tc.expectedCursor, nextCursor)
		})
	}
}

func TestOntoDocumentLocator(t *testing.T) {
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
			name:           "empty string, cursor past end",
			inputString:    "",
			initialCursor:  cursorState{position: 10},
			expectedCursor: cursorState{position: 0},
		},
		{
			name:           "past end of file ending with newline",
			inputString:    "abcd\n",
			initialCursor:  cursorState{position: 20},
			expectedCursor: cursorState{position: 4},
		},
		{
			name:           "past end of file ending with character",
			inputString:    "abcd",
			initialCursor:  cursorState{position: 20},
			expectedCursor: cursorState{position: 3},
		},
		{
			name:           "discard logical offset",
			inputString:    "abcd\nxyz",
			initialCursor:  cursorState{position: 3, logicalOffset: 5},
			expectedCursor: cursorState{position: 3},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			textTree, err := text.NewTreeFromString(tc.inputString)
			require.NoError(t, err)
			state := BufferState{
				textTree: textTree,
				cursor:   tc.initialCursor,
			}
			loc := NewOntoDocumentLocator()
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
			expectedCursor: cursorState{position: 5},
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
			expectedCursor: cursorState{position: 2},
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
			textTree, err := text.NewTreeFromString(tc.inputString)
			require.NoError(t, err)
			state := BufferState{
				textTree: textTree,
				cursor:   tc.initialCursor,
			}
			loc := NewOntoLineLocator()
			nextCursor := loc.Locate(&state)
			assert.Equal(t, tc.expectedCursor, nextCursor)
		})
	}
}

func TestRelativeLineStartLocator(t *testing.T) {
	testCases := []struct {
		name           string
		inputString    string
		direction      text.ReadDirection
		count          uint64
		initialCursor  cursorState
		expectedCursor cursorState
	}{
		{
			name:           "empty, read forward",
			inputString:    "",
			direction:      text.ReadDirectionForward,
			count:          1,
			initialCursor:  cursorState{position: 0},
			expectedCursor: cursorState{position: 0},
		},
		{
			name:           "empty, read backward",
			inputString:    "",
			direction:      text.ReadDirectionBackward,
			count:          1,
			initialCursor:  cursorState{position: 0},
			expectedCursor: cursorState{position: 0},
		},
		{
			name:           "start of line above",
			inputString:    "abcd\nefgh\nijkl\nmnop",
			direction:      text.ReadDirectionBackward,
			count:          2,
			initialCursor:  cursorState{position: 17},
			expectedCursor: cursorState{position: 5},
		},
		{
			name:           "start of line below",
			inputString:    "abcd\nefgh\nijkl\nmnop",
			direction:      text.ReadDirectionForward,
			count:          2,
			initialCursor:  cursorState{position: 3},
			expectedCursor: cursorState{position: 10},
		},
		{
			name:           "ends with newline",
			inputString:    "abcd\nefgh\nijkl\n",
			direction:      text.ReadDirectionForward,
			count:          5,
			initialCursor:  cursorState{position: 1},
			expectedCursor: cursorState{position: 15},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			textTree, err := text.NewTreeFromString(tc.inputString)
			require.NoError(t, err)
			state := BufferState{
				textTree: textTree,
				cursor:   tc.initialCursor,
			}
			loc := NewRelativeLineStartLocator(tc.direction, tc.count)
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
		{
			name:           "move down multiple lines",
			inputString:    "abcd\nefgh\nijkl",
			direction:      text.ReadDirectionForward,
			count:          2,
			initialCursor:  cursorState{position: 2},
			expectedCursor: cursorState{position: 12},
		},
		{
			name:           "move up multiple lines",
			inputString:    "abcd\nefgh\nijkl",
			direction:      text.ReadDirectionBackward,
			count:          2,
			initialCursor:  cursorState{position: 12},
			expectedCursor: cursorState{position: 2},
		},
		{
			name:           "move down past newline at end of text",
			inputString:    "abcd\nefgh\nijkl\n",
			direction:      text.ReadDirectionForward,
			count:          1,
			initialCursor:  cursorState{position: 12},
			expectedCursor: cursorState{position: 15, logicalOffset: 2},
		},
		{
			name:           "move down past single newline",
			inputString:    "\n",
			direction:      text.ReadDirectionForward,
			count:          1,
			initialCursor:  cursorState{position: 0},
			expectedCursor: cursorState{position: 1},
		},
		{
			name:           "move up to tab",
			inputString:    "abcd\ne\tefg\nhijkl",
			direction:      text.ReadDirectionBackward,
			count:          1,
			initialCursor:  cursorState{position: 13},
			expectedCursor: cursorState{position: 6, logicalOffset: 1},
		},
		{
			name:           "move down to tab",
			inputString:    "abcd\ne\tefg\nhijkl",
			direction:      text.ReadDirectionForward,
			count:          1,
			initialCursor:  cursorState{position: 3},
			expectedCursor: cursorState{position: 6, logicalOffset: 2},
		},
		{
			name:           "move up from tab",
			inputString:    "abcd\ne\tefg\nhijkl",
			direction:      text.ReadDirectionBackward,
			count:          1,
			initialCursor:  cursorState{position: 6, logicalOffset: 2},
			expectedCursor: cursorState{position: 3},
		},
		{
			name:           "move down from tab",
			inputString:    "abcd\ne\tefg\nhijkl",
			direction:      text.ReadDirectionForward,
			count:          1,
			initialCursor:  cursorState{position: 6, logicalOffset: 1},
			expectedCursor: cursorState{position: 13},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			textTree, err := text.NewTreeFromString(tc.inputString)
			require.NoError(t, err)
			state := BufferState{
				textTree: textTree,
				cursor:   tc.initialCursor,
				tabSize:  4,
			}
			loc := NewRelativeLineLocator(tc.direction, tc.count)
			nextCursor := loc.Locate(&state)
			assert.Equal(t, tc.expectedCursor, nextCursor)
		})
	}
}

func TestLineBoundaryLocator(t *testing.T) {
	testCases := []struct {
		name                   string
		inputString            string
		initialCursor          cursorState
		direction              text.ReadDirection
		includeEndOfLineOrFile bool
		expectedCursor         cursorState
	}{
		{
			name:           "empty, read forward",
			inputString:    "",
			initialCursor:  cursorState{position: 0},
			direction:      text.ReadDirectionForward,
			expectedCursor: cursorState{position: 0},
		},
		{
			name:           "empty, read backward",
			inputString:    "",
			initialCursor:  cursorState{position: 0},
			direction:      text.ReadDirectionBackward,
			expectedCursor: cursorState{position: 0},
		},
		{
			name:           "read backward, first line",
			inputString:    "abcd\nefgh",
			initialCursor:  cursorState{position: 2},
			direction:      text.ReadDirectionBackward,
			expectedCursor: cursorState{position: 0},
		},
		{
			name:           "read backward to line break",
			inputString:    "abcd\nefgh",
			initialCursor:  cursorState{position: 8},
			direction:      text.ReadDirectionBackward,
			expectedCursor: cursorState{position: 5},
		},
		{
			name:           "read forward to line break",
			inputString:    "abcd\nefgh",
			initialCursor:  cursorState{position: 2},
			direction:      text.ReadDirectionForward,
			expectedCursor: cursorState{position: 3},
		},
		{
			name:           "read forward, last line",
			inputString:    "abcd\nefgh",
			initialCursor:  cursorState{position: 6},
			direction:      text.ReadDirectionForward,
			expectedCursor: cursorState{position: 8},
		},
		{
			name:                   "read forward, include end of line",
			inputString:            "abcd\nefgh",
			initialCursor:          cursorState{position: 2},
			direction:              text.ReadDirectionForward,
			includeEndOfLineOrFile: true,
			expectedCursor:         cursorState{position: 4},
		},
		{
			name:           "read forward, include end of file",
			inputString:    "abcd\nefgh",
			initialCursor:  cursorState{position: 6},
			direction:      text.ReadDirectionForward,
			expectedCursor: cursorState{position: 8},
		},
		{
			name:           "read backward with movement resets logical offset",
			inputString:    "abcd\nefgh",
			initialCursor:  cursorState{position: 8, logicalOffset: 2},
			direction:      text.ReadDirectionBackward,
			expectedCursor: cursorState{position: 5},
		},
		{
			name:           "read forward at end of line preserves logical offset",
			inputString:    "abcd\nefgh",
			initialCursor:  cursorState{position: 8, logicalOffset: 2},
			direction:      text.ReadDirectionForward,
			expectedCursor: cursorState{position: 8, logicalOffset: 2},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			textTree, err := text.NewTreeFromString(tc.inputString)
			require.NoError(t, err)
			state := BufferState{
				textTree: textTree,
				cursor:   tc.initialCursor,
			}
			loc := NewLineBoundaryLocator(tc.direction, tc.includeEndOfLineOrFile)
			nextCursor := loc.Locate(&state)
			assert.Equal(t, tc.expectedCursor, nextCursor)
		})
	}
}

func TestNonWhitespaceOrNewlineLocator(t *testing.T) {
	testCases := []struct {
		name           string
		inputString    string
		initialCursor  cursorState
		expectedCursor cursorState
	}{
		{
			name:           "empty",
			inputString:    "",
			initialCursor:  cursorState{position: 0},
			expectedCursor: cursorState{position: 0},
		},
		{
			name:           "no movement",
			inputString:    "   abcd   ",
			initialCursor:  cursorState{position: 4},
			expectedCursor: cursorState{position: 4},
		},
		{
			name:           "movement",
			inputString:    "   abcd   ",
			initialCursor:  cursorState{position: 1},
			expectedCursor: cursorState{position: 3},
		},
		{
			name:           "movement resets logical offset",
			inputString:    "   abcd   ",
			initialCursor:  cursorState{position: 1, logicalOffset: 10},
			expectedCursor: cursorState{position: 3},
		},
		{
			name:           "no movement preserves logical offset",
			inputString:    "abcd\nefgh",
			initialCursor:  cursorState{position: 3, logicalOffset: 10},
			expectedCursor: cursorState{position: 3, logicalOffset: 10},
		},
		{
			name:           "stop before newline on empty line",
			inputString:    "abcd\n\n\nefgh",
			initialCursor:  cursorState{position: 5},
			expectedCursor: cursorState{position: 5},
		},
		{
			name:           "stop before newline at end of line",
			inputString:    "abcd\nefghi",
			initialCursor:  cursorState{position: 3},
			expectedCursor: cursorState{position: 3},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			textTree, err := text.NewTreeFromString(tc.inputString)
			require.NoError(t, err)
			state := BufferState{
				textTree: textTree,
				cursor:   tc.initialCursor,
			}
			loc := NewNonWhitespaceOrNewlineLocator(NewCurrentCursorLocator())
			nextCursor := loc.Locate(&state)
			assert.Equal(t, tc.expectedCursor, nextCursor)
		})
	}
}

func TestLineNumLocator(t *testing.T) {
	testCases := []struct {
		name           string
		inputString    string
		lineNum        uint64
		initialCursor  cursorState
		expectedCursor cursorState
	}{
		{
			name:           "empty",
			inputString:    "",
			lineNum:        1,
			initialCursor:  cursorState{position: 0},
			expectedCursor: cursorState{position: 0},
		},
		{
			name:           "first line",
			inputString:    "abcd\nefgh\nijkl\n",
			lineNum:        0,
			initialCursor:  cursorState{position: 10},
			expectedCursor: cursorState{position: 0},
		},
		{
			name:           "last line",
			inputString:    "abcd\nefgh\nijkl\n",
			lineNum:        2,
			initialCursor:  cursorState{position: 0},
			expectedCursor: cursorState{position: 10},
		},
		{
			name:           "past last line, ending with newline",
			inputString:    "abcd\nefgh\nijkl\n",
			lineNum:        5,
			initialCursor:  cursorState{position: 0},
			expectedCursor: cursorState{position: 15},
		},
		{
			name:           "past last line, ending with character",
			inputString:    "abcd\nefgh\nijkl",
			lineNum:        5,
			initialCursor:  cursorState{position: 0},
			expectedCursor: cursorState{position: 10},
		},
		{
			name:           "middle line",
			inputString:    "abcd\nefgh\nijkl",
			lineNum:        1,
			initialCursor:  cursorState{position: 1},
			expectedCursor: cursorState{position: 5},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			textTree, err := text.NewTreeFromString(tc.inputString)
			require.NoError(t, err)
			state := BufferState{
				textTree: textTree,
				cursor:   tc.initialCursor,
			}
			loc := NewLineNumLocator(tc.lineNum)
			nextCursor := loc.Locate(&state)
			assert.Equal(t, tc.expectedCursor, nextCursor)
		})
	}
}

func TestLastLineLocator(t *testing.T) {
	testCases := []struct {
		name           string
		inputString    string
		initialCursor  cursorState
		expectedCursor cursorState
	}{
		{
			name:           "empty",
			inputString:    "",
			initialCursor:  cursorState{position: 0},
			expectedCursor: cursorState{position: 0},
		},
		{
			name:           "single newline",
			inputString:    "\n",
			initialCursor:  cursorState{position: 0},
			expectedCursor: cursorState{position: 1},
		},
		{
			name:           "multiple newlines",
			inputString:    "\n\n\n\n",
			initialCursor:  cursorState{position: 1},
			expectedCursor: cursorState{position: 4},
		},
		{
			name:           "from first line to last line, end with character",
			inputString:    "ab\ncd\nef",
			initialCursor:  cursorState{position: 1},
			expectedCursor: cursorState{position: 6},
		},
		{
			name:           "from first line to last line, end with newline",
			inputString:    "ab\ncd\nef\n",
			initialCursor:  cursorState{position: 1},
			expectedCursor: cursorState{position: 9},
		},
		{
			name:           "already on last line, move to start of line",
			inputString:    "ab\ncd\nef",
			initialCursor:  cursorState{position: 7},
			expectedCursor: cursorState{position: 6},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			textTree, err := text.NewTreeFromString(tc.inputString)
			require.NoError(t, err)
			state := BufferState{
				textTree: textTree,
				cursor:   tc.initialCursor,
			}
			loc := NewLastLineLocator()
			nextCursor := loc.Locate(&state)
			assert.Equal(t, tc.expectedCursor, nextCursor)
		})
	}
}

func TestNextWordStartLocator(t *testing.T) {
	testCases := []struct {
		name           string
		inputString    string
		syntaxLanguage syntax.Language
		initialCursor  cursorState
		expectedCursor cursorState
	}{
		{
			name:           "empty",
			inputString:    "",
			initialCursor:  cursorState{position: 0},
			expectedCursor: cursorState{position: 0},
		},
		{
			name:           "next word from current word, same line",
			inputString:    "abc   defg   hij",
			initialCursor:  cursorState{position: 1},
			expectedCursor: cursorState{position: 6},
		},
		{
			name:           "next word from whitespace, same line",
			inputString:    "abc   defg   hij",
			initialCursor:  cursorState{position: 4},
			expectedCursor: cursorState{position: 6},
		},
		{
			name:           "next word from different line",
			inputString:    "abc\n   123",
			initialCursor:  cursorState{position: 1},
			expectedCursor: cursorState{position: 7},
		},
		{
			name:           "next word to empty line",
			inputString:    "abc\n\n   123",
			initialCursor:  cursorState{position: 1},
			expectedCursor: cursorState{position: 4},
		},
		{
			name:           "empty line to next word",
			inputString:    "abc\n\n   123",
			initialCursor:  cursorState{position: 4},
			expectedCursor: cursorState{position: 8},
		},
		{
			name:           "multiple empty lines",
			inputString:    "\n\n\n\n",
			initialCursor:  cursorState{position: 1},
			expectedCursor: cursorState{position: 2},
		},
		{
			name:           "next syntax token",
			inputString:    "123+456",
			syntaxLanguage: syntax.LanguageGo,
			initialCursor:  cursorState{position: 1},
			expectedCursor: cursorState{position: 3},
		},
		{
			name:           "next syntax token skip empty",
			inputString:    "123    +      456",
			syntaxLanguage: syntax.LanguageGo,
			initialCursor:  cursorState{position: 1},
			expectedCursor: cursorState{position: 7},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			textTree, err := text.NewTreeFromString(tc.inputString)
			require.NoError(t, err)
			state := BufferState{
				textTree: textTree,
				cursor:   tc.initialCursor,
			}
			state.SetSyntax(tc.syntaxLanguage)
			loc := NewNextWordStartLocator()
			nextCursor := loc.Locate(&state)
			assert.Equal(t, tc.expectedCursor, nextCursor)
		})
	}
}

func TestPrevWordStartLocator(t *testing.T) {
	testCases := []struct {
		name           string
		inputString    string
		syntaxLanguage syntax.Language
		initialCursor  cursorState
		expectedCursor cursorState
	}{
		{
			name:           "empty",
			inputString:    "",
			initialCursor:  cursorState{position: 0},
			expectedCursor: cursorState{position: 0},
		},
		{
			name:           "prev word from current word, same line",
			inputString:    "abc   defg   hij",
			initialCursor:  cursorState{position: 6},
			expectedCursor: cursorState{position: 0},
		},
		{
			name:           "prev word from whitespace, same line",
			inputString:    "abc   defg   hij",
			initialCursor:  cursorState{position: 12},
			expectedCursor: cursorState{position: 6},
		},
		{
			name:           "prev word from different line",
			inputString:    "abc\n   123",
			initialCursor:  cursorState{position: 7},
			expectedCursor: cursorState{position: 0},
		},
		{
			name:           "prev word to empty line",
			inputString:    "abc\n\n   123",
			initialCursor:  cursorState{position: 8},
			expectedCursor: cursorState{position: 4},
		},
		{
			name:           "empty line to prev word",
			inputString:    "abc\n\n   123",
			initialCursor:  cursorState{position: 4},
			expectedCursor: cursorState{position: 0},
		},
		{
			name:           "multiple empty lines",
			inputString:    "\n\n\n\n",
			initialCursor:  cursorState{position: 2},
			expectedCursor: cursorState{position: 1},
		},
		{
			name:           "prev syntax token",
			inputString:    "123+456",
			syntaxLanguage: syntax.LanguageGo,
			initialCursor:  cursorState{position: 4},
			expectedCursor: cursorState{position: 3},
		},
		{
			name:           "prev syntax token skip empty",
			inputString:    "123    +      456",
			syntaxLanguage: syntax.LanguageGo,
			initialCursor:  cursorState{position: 14},
			expectedCursor: cursorState{position: 7},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			textTree, err := text.NewTreeFromString(tc.inputString)
			require.NoError(t, err)
			state := BufferState{
				textTree: textTree,
				cursor:   tc.initialCursor,
			}
			state.SetSyntax(tc.syntaxLanguage)
			loc := NewPrevWordStartLocator()
			nextCursor := loc.Locate(&state)
			assert.Equal(t, tc.expectedCursor, nextCursor)
		})
	}
}
