package locate

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/aretext/aretext/editor/text"
)

func TestNextCharInLine(t *testing.T) {
	testCases := []struct {
		name                   string
		inputString            string
		pos                    uint64
		count                  uint64
		includeEndOfLineOrFile bool
		expectedPos            uint64
	}{
		{
			name:        "empty string",
			inputString: "",
			pos:         0,
			count:       1,
			expectedPos: 0,
		},
		{
			name:        "first char, ASCII string",
			inputString: "abcd",
			pos:         0,
			count:       1,
			expectedPos: 1,
		},
		{
			name:        "second char, ASCII string",
			inputString: "abcd",
			pos:         1,
			count:       1,
			expectedPos: 2,
		},
		{
			name:        "last char, ASCII string",
			inputString: "abcd",
			pos:         3,
			count:       1,
			expectedPos: 3,
		},
		{
			name:        "multi-char grapheme cluster",
			inputString: "e\u0301xyz",
			pos:         0,
			count:       1,
			expectedPos: 2,
		},
		{
			name:        "up to end of line",
			inputString: "ab\ncd",
			pos:         1,
			count:       1,
			expectedPos: 1,
		},
		{
			name:        "at end of line",
			inputString: "ab\ncd",
			pos:         2,
			count:       1,
			expectedPos: 2,
		},
		{
			name:        "end of line with carriage return",
			inputString: "ab\r\ncd",
			pos:         1,
			count:       1,
			expectedPos: 1,
		},
		{
			name:        "move cursor multiple chars within line",
			inputString: "abcdefgh",
			pos:         2,
			count:       3,
			expectedPos: 5,
		},
		{
			name:        "move cursor multiple chars to end of line",
			inputString: "abcd\nefgh",
			pos:         1,
			count:       2,
			expectedPos: 3,
		},
		{
			name:        "move cursor multiple chars past end of line",
			inputString: "abcd\nefgh",
			pos:         1,
			count:       5,
			expectedPos: 3,
		},
		{
			name:        "move cursor multiple chars past end of string",
			inputString: "abcd",
			pos:         0,
			count:       100,
			expectedPos: 3,
		},
		{
			name:                   "include end of line",
			inputString:            "abcd\nefgh",
			pos:                    2,
			count:                  5,
			includeEndOfLineOrFile: true,
			expectedPos:            4,
		},
		{
			name:                   "include end of file",
			inputString:            "abcd",
			pos:                    2,
			count:                  5,
			includeEndOfLineOrFile: true,
			expectedPos:            4,
		},
		{
			name:                   "first character when including end of line or file",
			inputString:            "abcdefgh",
			pos:                    0,
			count:                  1,
			includeEndOfLineOrFile: true,
			expectedPos:            1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			textTree, err := text.NewTreeFromString(tc.inputString)
			require.NoError(t, err)
			actualPos := NextCharInLine(textTree, tc.count, tc.includeEndOfLineOrFile, tc.pos)
			assert.Equal(t, tc.expectedPos, actualPos)
		})
	}
}

func TestPrevCharInLine(t *testing.T) {
	testCases := []struct {
		name                   string
		inputString            string
		pos                    uint64
		count                  uint64
		includeEndOfLineOrFile bool
		expectedPos            uint64
	}{
		{
			name:        "empty",
			inputString: "",
			pos:         0,
			count:       1,
			expectedPos: 0,
		},
		{
			name:        "last char, ASCII string",
			inputString: "abcd",
			pos:         3,
			count:       1,
			expectedPos: 2,
		},
		{
			name:        "second-to-last char, ASCII string",
			inputString: "abcd",
			pos:         2,
			count:       1,
			expectedPos: 1,
		},
		{
			name:        "first char, ASCII string",
			inputString: "abcd",
			pos:         0,
			count:       1,
			expectedPos: 0,
		},
		{
			name:        "first char in line",
			inputString: "ab\ncd",
			pos:         3,
			count:       1,
			expectedPos: 3,
		},
		{
			name:        "multi-char grapheme cluster",
			inputString: "abe\u0301xyz",
			pos:         4,
			count:       1,
			expectedPos: 2,
		},
		{
			name:        "move cursor multiple chars within line",
			inputString: "abcdefgh",
			pos:         4,
			count:       3,
			expectedPos: 1,
		},
		{
			name:        "move cursor multiple chars to beginning of line",
			inputString: "ab\ncdefgh",
			pos:         5,
			count:       2,
			expectedPos: 3,
		},
		{
			name:        "move cursor multiple chars past beginning of line",
			inputString: "ab\ncdefgh",
			pos:         5,
			count:       4,
			expectedPos: 3,
		},
		{
			name:                   "include end of previous line",
			inputString:            "abcd\nefgh",
			pos:                    5,
			count:                  1,
			includeEndOfLineOrFile: true,
			expectedPos:            4,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			textTree, err := text.NewTreeFromString(tc.inputString)
			require.NoError(t, err)
			actualPos := PrevCharInLine(textTree, tc.count, tc.includeEndOfLineOrFile, tc.pos)
			assert.Equal(t, tc.expectedPos, actualPos)
		})
	}
}

func TestPrevChar(t *testing.T) {
	testCases := []struct {
		name        string
		inputString string
		pos         uint64
		count       uint64
		expectedPos uint64
	}{
		{
			name:        "empty string",
			inputString: "",
			pos:         0,
			count:       1,
			expectedPos: 0,
		},
		{
			name:        "back single char, same line",
			inputString: "abc\ndef",
			pos:         5,
			count:       1,
			expectedPos: 4,
		},
		{
			name:        "back single char, prev line",
			inputString: "abc\ndef",
			pos:         3,
			count:       1,
			expectedPos: 2,
		},
		{
			name:        "back multi-char grapheme cluster",
			inputString: "e\u0301xyz",
			pos:         2,
			count:       1,
			expectedPos: 0,
		},
		{
			name:        "back multiple chars, within document",
			inputString: "abc\ndef",
			pos:         5,
			count:       3,
			expectedPos: 2,
		},
		{
			name:        "back multiple chars, outside document",
			inputString: "abc\ndef",
			pos:         5,
			count:       100,
			expectedPos: 0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			textTree, err := text.NewTreeFromString(tc.inputString)
			require.NoError(t, err)
			actualPos := PrevChar(textTree, tc.count, tc.pos)
			assert.Equal(t, tc.expectedPos, actualPos)
		})
	}
}

func TestNextMatchingCharInLine(t *testing.T) {
	testCases := []struct {
		name        string
		inputString string
		char        rune
		count       uint64
		includeChar bool
		pos         uint64
		expectFound bool
		expectedPos uint64
	}{
		{
			name:        "empty string",
			inputString: "",
			char:        'x',
			count:       1,
			pos:         0,
			expectFound: false,
			expectedPos: 0,
		},
		{
			name:        "not found on first line",
			inputString: "abcxyz",
			char:        'm',
			count:       1,
			pos:         1,
			expectFound: false,
			expectedPos: 0,
		},
		{
			name:        "count zero finds nothing",
			inputString: "abcxyz",
			char:        'x',
			count:       0,
			pos:         1,
			expectFound: false,
			expectedPos: 0,
		},
		{
			name:        "found on first line, include",
			inputString: "abcxyz",
			char:        'x',
			count:       1,
			includeChar: true,
			pos:         1,
			expectFound: true,
			expectedPos: 3,
		},
		{
			name:        "found on first line, exclude",
			inputString: "abcxyz",
			char:        'x',
			count:       1,
			includeChar: false,
			pos:         1,
			expectFound: true,
			expectedPos: 2,
		},
		{
			name:        "found on first line, count > 0",
			inputString: "abcxyzxyz",
			char:        'x',
			count:       2,
			includeChar: true,
			pos:         1,
			expectFound: true,
			expectedPos: 6,
		},
		{
			name:        "next match on subsequent line",
			inputString: "abc\nxyz",
			char:        'x',
			count:       1,
			includeChar: true,
			pos:         1,
			expectFound: false,
			expectedPos: 0,
		},
		{
			name:        "match at end of current line",
			inputString: "abc\nabx\nyz",
			char:        'x',
			count:       1,
			includeChar: true,
			pos:         4,
			expectFound: true,
			expectedPos: 6,
		},
		{
			name:        "no match character same as under cursor",
			inputString: "ab",
			char:        'a',
			count:       1,
			includeChar: false,
			pos:         0,
			expectFound: false,
			expectedPos: 0,
		},
		{
			name:        "match character same as under cursor",
			inputString: "xaaaaaaaxbbbb",
			char:        'x',
			count:       1,
			includeChar: false,
			pos:         0,
			expectFound: true,
			expectedPos: 7,
		},
		{
			name:        "match next character same as character under cursor",
			inputString: "aab",
			char:        'a',
			count:       1,
			includeChar: false,
			pos:         0,
			expectFound: true,
			expectedPos: 0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			textTree, err := text.NewTreeFromString(tc.inputString)
			require.NoError(t, err)
			found, actualPos := NextMatchingCharInLine(textTree, tc.char, tc.count, tc.includeChar, tc.pos)
			assert.Equal(t, tc.expectFound, found)
			assert.Equal(t, tc.expectedPos, actualPos)
		})
	}
}

func TestPrevMatchingCharInLine(t *testing.T) {
	testCases := []struct {
		name        string
		inputString string
		char        rune
		count       uint64
		includeChar bool
		pos         uint64
		expectFound bool
		expectedPos uint64
	}{
		{
			name:        "empty string",
			inputString: "",
			char:        'x',
			count:       1,
			pos:         0,
			expectFound: false,
			expectedPos: 0,
		},
		{
			name:        "not found on first line",
			inputString: "abcxyz",
			char:        'm',
			count:       1,
			pos:         5,
			expectFound: false,
			expectedPos: 0,
		},
		{
			name:        "count zero finds nothing",
			inputString: "abcxyz",
			char:        'x',
			count:       0,
			pos:         5,
			expectFound: false,
			expectedPos: 0,
		},
		{
			name:        "found on first line, include",
			inputString: "abcxyz",
			char:        'x',
			count:       1,
			includeChar: true,
			pos:         5,
			expectFound: true,
			expectedPos: 3,
		},
		{
			name:        "found on first line, exclude",
			inputString: "abcxyz",
			char:        'x',
			count:       1,
			includeChar: false,
			pos:         5,
			expectFound: true,
			expectedPos: 4,
		},
		{
			name:        "found on first line, count > 0",
			inputString: "abcxyzxyz",
			char:        'x',
			count:       2,
			includeChar: true,
			pos:         8,
			expectFound: true,
			expectedPos: 3,
		},
		{
			name:        "next match on previous line",
			inputString: "abcx\nyz",
			char:        'x',
			count:       1,
			includeChar: true,
			pos:         6,
			expectFound: false,
			expectedPos: 0,
		},
		{
			name:        "match at start of current line",
			inputString: "abc\nxab\nyz",
			char:        'x',
			count:       1,
			includeChar: true,
			pos:         6,
			expectFound: true,
			expectedPos: 4,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			textTree, err := text.NewTreeFromString(tc.inputString)
			require.NoError(t, err)
			found, actualPos := PrevMatchingCharInLine(textTree, tc.char, tc.count, tc.includeChar, tc.pos)
			assert.Equal(t, tc.expectFound, found)
			assert.Equal(t, tc.expectedPos, actualPos)
		})
	}
}

func TestPrevAutoIndent(t *testing.T) {
	testCases := []struct {
		name              string
		inputString       string
		autoIndentEnabled bool
		pos               uint64
		expectedPos       uint64
	}{
		{
			name:        "empty string",
			inputString: "",
			pos:         0,
			expectedPos: 0,
		},
		{
			name:        "multiple tabs, autoindent disabled",
			inputString: "\t\t",
			pos:         2,
			expectedPos: 2,
		},
		{
			name:              "single space, autoindent enabled",
			inputString:       " ",
			autoIndentEnabled: true,
			pos:               1,
			expectedPos:       0,
		},
		{
			name:              "multiple spaces, autoindent enabled",
			inputString:       "        ",
			autoIndentEnabled: true,
			pos:               8,
			expectedPos:       4,
		},
		{
			name:              "multiple tabs, autoindent enabled",
			inputString:       "\t\t",
			autoIndentEnabled: true,
			pos:               2,
			expectedPos:       1,
		},
		{
			name:              "mixed tabs and spaces, autoindent enabled",
			inputString:       " \t",
			autoIndentEnabled: true,
			pos:               2,
			expectedPos:       0,
		},
		{
			name:              "no tabs or spaces, autoindent enabled",
			inputString:       "ab",
			autoIndentEnabled: true,
			pos:               2,
			expectedPos:       2,
		},
		{
			name:              "start of line, autoindent enabled",
			inputString:       "ab\ncd",
			autoIndentEnabled: true,
			pos:               2,
			expectedPos:       2,
		},
		{
			name:              "end of document, autoindent enabled",
			inputString:       "ab\n\n",
			autoIndentEnabled: true,
			pos:               3,
			expectedPos:       3,
		},
		{
			name:              "spaces within line aligned, autoindent enabled",
			inputString:       "abcd    ef",
			autoIndentEnabled: true,
			pos:               8,
			expectedPos:       4,
		},
		{
			name:              "spaces within line misaligned, autoindent enabled",
			inputString:       "ab    cd",
			autoIndentEnabled: true,
			pos:               6,
			expectedPos:       4,
		},
		{
			name:              "tabs within line, autoindent enabled",
			inputString:       "ab\t\tcd",
			autoIndentEnabled: true,
			pos:               4,
			expectedPos:       3,
		},
		{
			name:              "spaces within line but not before cursor, autoindent enabled",
			inputString:       "ab    cdef",
			autoIndentEnabled: true,
			pos:               7,
			expectedPos:       7,
		},
		{
			name:              "spaces at end of line less than tab size, autoindent enabled",
			inputString:       "abcdef  ",
			autoIndentEnabled: true,
			pos:               8,
			expectedPos:       6,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			textTree, err := text.NewTreeFromString(tc.inputString)
			require.NoError(t, err)
			actualPos := PrevAutoIndent(textTree, tc.autoIndentEnabled, 4, tc.pos)
			assert.Equal(t, tc.expectedPos, actualPos)
		})
	}
}

func TestNextNonWhitespaceOrNewline(t *testing.T) {
	testCases := []struct {
		name        string
		inputString string
		pos         uint64
		expectedPos uint64
	}{
		{
			name:        "empty",
			inputString: "",
			pos:         0,
			expectedPos: 0,
		},
		{
			name:        "no movement",
			inputString: "   abcd   ",
			pos:         4,
			expectedPos: 4,
		},
		{
			name:        "movement",
			inputString: "   abcd   ",
			pos:         1,
			expectedPos: 3,
		},
		{
			name:        "stop before newline on empty line",
			inputString: "abcd\n\n\nefgh",
			pos:         5,
			expectedPos: 5,
		},
		{
			name:        "stop before newline at end of line",
			inputString: "abcd\nefghi",
			pos:         3,
			expectedPos: 3,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			textTree, err := text.NewTreeFromString(tc.inputString)
			require.NoError(t, err)
			actualPos := NextNonWhitespaceOrNewline(textTree, tc.pos)
			assert.Equal(t, tc.expectedPos, actualPos)
		})
	}
}

func TestNextNewline(t *testing.T) {
	testCases := []struct {
		name        string
		inputString string
		pos         uint64
		expectedOk  bool
		expectedPos uint64
		expectedLen uint64
	}{
		{
			name:        "empty",
			inputString: "",
			pos:         0,
			expectedOk:  false,
		},
		{
			name:        "last line",
			inputString: "abcd",
			pos:         2,
			expectedOk:  false,
		},
		{
			name:        "before LF",
			inputString: "abc\ndef",
			pos:         1,
			expectedOk:  true,
			expectedPos: 3,
			expectedLen: 1,
		},
		{
			name:        "on LF",
			inputString: "abc\ndef",
			pos:         3,
			expectedOk:  true,
			expectedPos: 3,
			expectedLen: 1,
		},
		{
			name:        "before CR LF",
			inputString: "abc\r\ndef",
			pos:         1,
			expectedOk:  true,
			expectedPos: 3,
			expectedLen: 2,
		},
		{
			name:        "on CR LF",
			inputString: "abc\r\ndef",
			pos:         3,
			expectedOk:  true,
			expectedPos: 3,
			expectedLen: 2,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			textTree, err := text.NewTreeFromString(tc.inputString)
			require.NoError(t, err)
			actualPos, actualLen, actualOk := NextNewline(textTree, tc.pos)
			assert.Equal(t, tc.expectedOk, actualOk)
			assert.Equal(t, tc.expectedPos, actualPos)
			assert.Equal(t, tc.expectedLen, actualLen)
		})
	}
}

func TestNumGraphemeClustersInRange(t *testing.T) {
	testCases := []struct {
		name          string
		inputString   string
		startPos      uint64
		endPos        uint64
		expectedCount uint64
	}{
		{
			name:          "empty text",
			inputString:   "",
			startPos:      0,
			endPos:        0,
			expectedCount: 0,
		},
		{
			name:          "empty range",
			inputString:   "abcdefgh",
			startPos:      1,
			endPos:        1,
			expectedCount: 0,
		},
		{
			name:          "single-rune grapheme clusters",
			inputString:   "abcdefgh",
			startPos:      1,
			endPos:        4,
			expectedCount: 3,
		},
		{
			name:          "multi-rune grapheme clusters",
			inputString:   "ᄀ̈각각̈͏",
			startPos:      0,
			endPos:        6,
			expectedCount: 3,
		},
		{
			name:          "past end of file",
			inputString:   "abcdefgh",
			startPos:      3,
			endPos:        100,
			expectedCount: 5,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			textTree, err := text.NewTreeFromString(tc.inputString)
			require.NoError(t, err)
			actualCount := NumGraphemeClustersInRange(textTree, tc.startPos, tc.endPos)
			assert.Equal(t, tc.expectedCount, actualCount)
		})
	}
}
