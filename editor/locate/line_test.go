package locate

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/aretext/aretext/editor/text"
)

func TestClosestCharOnLine(t *testing.T) {
	testCases := []struct {
		name        string
		inputString string
		pos         uint64
		expectedPos uint64
	}{
		{
			name:        "empty string, cursor at origin",
			inputString: "",
			pos:         0,
			expectedPos: 0,
		},
		{
			name:        "empty string, cursor past origin",
			inputString: "",
			pos:         1,
			expectedPos: 0,
		},
		{
			name:        "cursor already on line at beginning of file",
			inputString: "abcd",
			pos:         0,
			expectedPos: 0,
		},
		{
			name:        "cursor already on line at in middle of line",
			inputString: "abcd",
			pos:         2,
			expectedPos: 2,
		},
		{
			name:        "cursor already on line at beginning of line",
			inputString: "abcd\nefg",
			pos:         5,
			expectedPos: 5,
		},
		{
			name:        "cursor already on line at end of line",
			inputString: "abcd\nefg",
			pos:         3,
			expectedPos: 3,
		},
		{
			name:        "cursor past end of file by single char",
			inputString: "abcd",
			pos:         4,
			expectedPos: 3,
		},
		{
			name:        "cursor past end of file by multiple chars",
			inputString: "abcd",
			pos:         10,
			expectedPos: 3,
		},
		{
			name:        "cursor on newline",
			inputString: "abcd\nefgh",
			pos:         4,
			expectedPos: 3,
		},
		{
			name:        "cursor on newline preceded by newline",
			inputString: "abcd\n\nefgh",
			pos:         5,
			expectedPos: 5,
		},
		{
			name:        "cursor at newline in file with only newline",
			inputString: "\n",
			pos:         0,
			expectedPos: 0,
		},
		{
			name:        "cursor at newline in file with multiple newlines",
			inputString: "\n\n\n",
			pos:         2,
			expectedPos: 2,
		},
		{
			name:        "cursor at newline with carriage return, on line feed",
			inputString: "abcd\r\nefgh",
			pos:         5,
			expectedPos: 3,
		},
		{
			name:        "cursor at newline with carriage return, on carriage return",
			inputString: "abcd\r\nefgh",
			pos:         4,
			expectedPos: 3,
		},
		{
			name:        "cursor on newline ending with multi-char grapheme cluster",
			inputString: "abcde\u0301\nfgh",
			pos:         6,
			expectedPos: 4,
		},
		{
			name:        "cursor on newline with carriage return ending with multi-char grapheme cluster",
			inputString: "abcde\u0301\r\nfgh",
			pos:         7,
			expectedPos: 4,
		},
		{
			name:        "cursor past end of text ending with multi-char grapheme cluster",
			inputString: "abcde\u0301",
			pos:         6,
			expectedPos: 4,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			textTree, err := text.NewTreeFromString(tc.inputString)
			require.NoError(t, err)
			actualPos := ClosestCharOnLine(textTree, tc.pos)
			assert.Equal(t, tc.expectedPos, actualPos)
		})
	}
}

func TestStartOfLineAbove(t *testing.T) {
	testCases := []struct {
		name        string
		inputString string
		count       uint64
		pos         uint64
		expectedPos uint64
	}{
		{
			name:        "empty",
			inputString: "",
			count:       1,
			pos:         0,
			expectedPos: 0,
		},
		{
			name:        "start of line above",
			inputString: "abcd\nefgh\nijkl\nmnop",
			count:       2,
			pos:         17,
			expectedPos: 5,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			textTree, err := text.NewTreeFromString(tc.inputString)
			require.NoError(t, err)
			actualPos := StartOfLineAbove(textTree, tc.count, tc.pos)
			assert.Equal(t, tc.expectedPos, actualPos)
		})
	}
}

func TestStartOfLineBelow(t *testing.T) {
	testCases := []struct {
		name        string
		inputString string
		count       uint64
		pos         uint64
		expectedPos uint64
	}{
		{
			name:        "empty",
			inputString: "",
			count:       1,
			pos:         0,
			expectedPos: 0,
		},
		{
			name:        "start of line below",
			inputString: "abcd\nefgh\nijkl\nmnop",
			count:       2,
			pos:         3,
			expectedPos: 10,
		},
		{
			name:        "ends with newline",
			inputString: "abcd\nefgh\nijkl\n",
			count:       5,
			pos:         1,
			expectedPos: 15,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			textTree, err := text.NewTreeFromString(tc.inputString)
			require.NoError(t, err)
			actualPos := StartOfLineBelow(textTree, tc.count, tc.pos)
			assert.Equal(t, tc.expectedPos, actualPos)
		})
	}
}

func TestNextLineBoundary(t *testing.T) {
	testCases := []struct {
		name                   string
		inputString            string
		includeEndOfLineOrFile bool
		pos                    uint64
		expectedPos            uint64
	}{
		{
			name:        "empty",
			inputString: "",
			pos:         0,
			expectedPos: 0,
		},
		{
			name:        "read to line break",
			inputString: "abcd\nefgh",
			pos:         2,
			expectedPos: 3,
		},
		{
			name:        "read from last line",
			inputString: "abcd\nefgh",
			pos:         6,
			expectedPos: 8,
		},
		{
			name:                   "read include end of line",
			inputString:            "abcd\nefgh",
			pos:                    2,
			includeEndOfLineOrFile: true,
			expectedPos:            4,
		},
		{
			name:        "read include end of file",
			inputString: "abcd\nefgh",
			pos:         6,
			expectedPos: 8,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			textTree, err := text.NewTreeFromString(tc.inputString)
			require.NoError(t, err)
			actualPos := NextLineBoundary(textTree, tc.includeEndOfLineOrFile, tc.pos)
			assert.Equal(t, tc.expectedPos, actualPos)
		})
	}
}

func TestPrevLineBoundary(t *testing.T) {
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
			name:        "read from first line",
			inputString: "abcd\nefgh",
			pos:         2,
			expectedPos: 0,
		},
		{
			name:        "read to line break",
			inputString: "abcd\nefgh",
			pos:         8,
			expectedPos: 5,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			textTree, err := text.NewTreeFromString(tc.inputString)
			require.NoError(t, err)
			actualPos := PrevLineBoundary(textTree, tc.pos)
			assert.Equal(t, tc.expectedPos, actualPos)
		})
	}
}

func TestStartOfLineNum(t *testing.T) {
	testCases := []struct {
		name        string
		inputString string
		lineNum     uint64
		pos         uint64
		expectedPos uint64
	}{
		{
			name:        "empty",
			inputString: "",
			lineNum:     1,
			pos:         0,
			expectedPos: 0,
		},
		{
			name:        "first line",
			inputString: "abcd\nefgh\nijkl\n",
			lineNum:     0,
			pos:         10,
			expectedPos: 0,
		},
		{
			name:        "last line",
			inputString: "abcd\nefgh\nijkl\n",
			lineNum:     2,
			pos:         0,
			expectedPos: 10,
		},
		{
			name:        "past last line, ending with newline",
			inputString: "abcd\nefgh\nijkl\n",
			lineNum:     5,
			pos:         0,
			expectedPos: 15,
		},
		{
			name:        "past last line, ending with character",
			inputString: "abcd\nefgh\nijkl",
			lineNum:     5,
			pos:         0,
			expectedPos: 10,
		},
		{
			name:        "middle line",
			inputString: "abcd\nefgh\nijkl",
			lineNum:     1,
			pos:         1,
			expectedPos: 5,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			textTree, err := text.NewTreeFromString(tc.inputString)
			require.NoError(t, err)
			actualPos := StartOfLineNum(textTree, tc.lineNum)
			assert.Equal(t, tc.expectedPos, actualPos)
		})
	}
}

func TestStartOfLastLine(t *testing.T) {
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
			name:        "single newline",
			inputString: "\n",
			pos:         0,
			expectedPos: 1,
		},
		{
			name:        "multiple newlines",
			inputString: "\n\n\n\n",
			pos:         1,
			expectedPos: 4,
		},
		{
			name:        "from first line to last line, end with character",
			inputString: "ab\ncd\nef",
			pos:         1,
			expectedPos: 6,
		},
		{
			name:        "from first line to last line, end with newline",
			inputString: "ab\ncd\nef\n",
			pos:         1,
			expectedPos: 9,
		},
		{
			name:        "already on last line, move to start of line",
			inputString: "ab\ncd\nef",
			pos:         7,
			expectedPos: 6,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			textTree, err := text.NewTreeFromString(tc.inputString)
			require.NoError(t, err)
			actualPos := StartOfLastLine(textTree)
			assert.Equal(t, tc.expectedPos, actualPos)
		})
	}
}

func TestPosToLineNumAndCol(t *testing.T) {
	testCases := []struct {
		name            string
		inputString     string
		pos             uint64
		expectedLineNum uint64
		expectedCol     uint64
	}{
		{
			name:            "empty string",
			inputString:     "",
			pos:             0,
			expectedLineNum: 0,
			expectedCol:     0,
		},
		{
			name:            "empty string, position past EOF",
			inputString:     "",
			pos:             5,
			expectedLineNum: 0,
			expectedCol:     0,
		},
		{
			name:            "single line, pos at start of line",
			inputString:     "foobar",
			pos:             0,
			expectedLineNum: 0,
			expectedCol:     0,
		},
		{
			name:            "single line, pos in middle of line",
			inputString:     "foobar",
			pos:             3,
			expectedLineNum: 0,
			expectedCol:     3,
		},
		{
			name:            "single line, pos past end of line",
			inputString:     "foobar",
			pos:             24,
			expectedLineNum: 0,
			expectedCol:     6,
		},
		{
			name:            "single line, multi-char grapheme cluster",
			inputString:     "abc\u0301xyz",
			pos:             6,
			expectedLineNum: 0,
			expectedCol:     5,
		},
		{
			name:            "multiple lines, start of second line",
			inputString:     "abc\ndef\nghi",
			pos:             4,
			expectedLineNum: 1,
			expectedCol:     0,
		},
		{
			name:            "multiple lines, middle of second line",
			inputString:     "abc\ndef\nghi",
			pos:             5,
			expectedLineNum: 1,
			expectedCol:     1,
		},
		{
			name:            "multiple lines, end of second line",
			inputString:     "abc\ndef\nghi",
			pos:             6,
			expectedLineNum: 1,
			expectedCol:     2,
		},
		{
			name:            "multiple lines, past end of second line",
			inputString:     "abc\ndef\nghi",
			pos:             7,
			expectedLineNum: 1,
			expectedCol:     3,
		},
		{
			name:            "on blank line",
			inputString:     "abc\n\n\ndef",
			pos:             4,
			expectedLineNum: 1,
			expectedCol:     0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			textTree, err := text.NewTreeFromString(tc.inputString)
			require.NoError(t, err)
			lineNum, col := PosToLineNumAndCol(textTree, tc.pos)
			assert.Equal(t, tc.expectedLineNum, lineNum)
			assert.Equal(t, tc.expectedCol, col)
		})
	}
}

func TestLineNumAndColToPos(t *testing.T) {
	testCases := []struct {
		name        string
		inputString string
		lineNum     uint64
		col         uint64
		expectedPos uint64
	}{
		{
			name:        "empty string",
			inputString: "",
			lineNum:     0,
			col:         0,
			expectedPos: 0,
		},
		{
			name:        "empty string, line past EOF",
			inputString: "",
			lineNum:     4,
			col:         3,
			expectedPos: 0,
		},
		{
			name:        "single line, at start of line",
			inputString: "abcd",
			lineNum:     0,
			col:         0,
			expectedPos: 0,
		},
		{
			name:        "single line, in middle of line",
			inputString: "abcd",
			lineNum:     0,
			col:         2,
			expectedPos: 2,
		},
		{
			name:        "single line, at end of line",
			inputString: "abcd",
			lineNum:     0,
			col:         3,
			expectedPos: 3,
		},
		{
			name:        "single line, past end of line",
			inputString: "abcd",
			lineNum:     0,
			col:         5,
			expectedPos: 3,
		},
		{
			name:        "single line, multi-char grapheme cluster",
			inputString: "abc\u0301xyz",
			lineNum:     0,
			col:         5,
			expectedPos: 6,
		},
		{
			name:        "multiple lines, at start of second line",
			inputString: "abc\ndef\nghi",
			lineNum:     1,
			col:         0,
			expectedPos: 4,
		},
		{
			name:        "multiple lines, at middle of second line",
			inputString: "abc\ndef\nghi",
			lineNum:     1,
			col:         1,
			expectedPos: 5,
		},
		{
			name:        "multiple lines, at end of second line",
			inputString: "abc\ndef\nghi",
			lineNum:     1,
			col:         2,
			expectedPos: 6,
		},
		{
			name:        "multiple lines, past end of second line",
			inputString: "abc\ndef\nghi",
			lineNum:     1,
			col:         3,
			expectedPos: 6,
		},
		{
			name:        "on blank line",
			inputString: "abc\n\n\ndef",
			lineNum:     1,
			col:         0,
			expectedPos: 4,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			textTree, err := text.NewTreeFromString(tc.inputString)
			require.NoError(t, err)
			pos := LineNumAndColToPos(textTree, tc.lineNum, tc.col)
			assert.Equal(t, tc.expectedPos, pos)
		})
	}
}
