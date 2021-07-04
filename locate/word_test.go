package locate

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/aretext/aretext/text"
	"github.com/aretext/aretext/text/segment"
)

func TestNextWordStart(t *testing.T) {
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
			name:        "next word from current word, same line",
			inputString: "abc   defg   hij",
			pos:         1,
			expectedPos: 6,
		},
		{
			name:        "next word from whitespace, same line",
			inputString: "abc   defg   hij",
			pos:         4,
			expectedPos: 6,
		},
		{
			name:        "next word from different line",
			inputString: "abc\n   123",
			pos:         1,
			expectedPos: 7,
		},
		{
			name:        "next word to empty line",
			inputString: "abc\n\n   123",
			pos:         1,
			expectedPos: 4,
		},
		{
			name:        "empty line to next word",
			inputString: "abc\n\n   123",
			pos:         4,
			expectedPos: 8,
		},
		{
			name:        "multiple empty lines",
			inputString: "\n\n\n\n",
			pos:         1,
			expectedPos: 2,
		},
		{
			name:        "non-punctuation to punctuation",
			inputString: "abc/def/ghi",
			pos:         1,
			expectedPos: 3,
		},
		{
			name:        "punctuation to non-punctuation",
			inputString: "abc/def/ghi",
			pos:         3,
			expectedPos: 4,
		},
		{
			name:        "repeated punctuation",
			inputString: "abc////cde",
			pos:         3,
			expectedPos: 7,
		},
		{
			name:        "underscores treated as non-punctuation",
			inputString: "abc_def ghi",
			pos:         0,
			expectedPos: 8,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			textTree, err := text.NewTreeFromString(tc.inputString)
			require.NoError(t, err)
			actualPos := NextWordStart(textTree, tc.pos)
			assert.Equal(t, tc.expectedPos, actualPos)
		})
	}
}

func TestNextWordStartInLine(t *testing.T) {
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
			name:        "start of word before another word",
			inputString: "abc  def",
			pos:         0,
			expectedPos: 5,
		},
		{
			name:        "middle of word before another word",
			inputString: "abc  def",
			pos:         1,
			expectedPos: 5,
		},
		{
			name:        "end of word before another word",
			inputString: "abc  def",
			pos:         2,
			expectedPos: 5,
		},
		{
			name:        "whitespace before word",
			inputString: "abc  def",
			pos:         3,
			expectedPos: 5,
		},
		{
			name:        "start of last word in document",
			inputString: "abc",
			pos:         0,
			expectedPos: 3,
		},
		{
			name:        "middle of last word in document",
			inputString: "abc",
			pos:         1,
			expectedPos: 3,
		},
		{
			name:        "end of last word in document",
			inputString: "abc",
			pos:         2,
			expectedPos: 3,
		},
		{
			name:        "last word in document single char",
			inputString: "a",
			pos:         0,
			expectedPos: 1,
		},
		{
			name:        "last word in line before next line",
			inputString: "abc\ndef",
			pos:         1,
			expectedPos: 3,
		},
		{
			name:        "last word in line with trailing whitespace before next line",
			inputString: "abc   \ndef",
			pos:         1,
			expectedPos: 6,
		},
		{
			name:        "single line, all whitespace, cursor at start",
			inputString: "   ",
			pos:         0,
			expectedPos: 3,
		},
		{
			name:        "single line, all whitespace, cursor in middle",
			inputString: "   ",
			pos:         1,
			expectedPos: 3,
		},
		{
			name:        "single line, all whitespace, cursor at end",
			inputString: "   ",
			pos:         2,
			expectedPos: 3,
		},
		{
			name:        "lines with all whitespace, first line",
			inputString: "   \n   ",
			pos:         1,
			expectedPos: 3,
		},
		{
			name:        "lines with all whitespace, last line",
			inputString: "   \n   ",
			pos:         5,
			expectedPos: 7,
		},
		{
			name:        "last word in line single char",
			inputString: "a\nbcd",
			pos:         0,
			expectedPos: 1,
		},
		{
			name:        "empty lines",
			inputString: "\n\n\n",
			pos:         1,
			expectedPos: 1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			textTree, err := text.NewTreeFromString(tc.inputString)
			require.NoError(t, err)
			actualPos := NextWordStartInLine(textTree, tc.pos)
			assert.Equal(t, tc.expectedPos, actualPos)
		})
	}
}

func TestNextWordEnd(t *testing.T) {
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
			name:        "end of word from start of current word",
			inputString: "abc   defg   hij",
			pos:         6,
			expectedPos: 9,
		},
		{
			name:        "end of word from middle of current word",
			inputString: "abc   defg   hij",
			pos:         7,
			expectedPos: 9,
		},
		{
			name:        "next word from end of current word",
			inputString: "abc   defg   hij",
			pos:         2,
			expectedPos: 9,
		},
		{
			name:        "next word from whitespace",
			inputString: "abc   defg   hij",
			pos:         4,
			expectedPos: 9,
		},
		{
			name:        "next word past empty line",
			inputString: "abc\n\n   123   xyz",
			pos:         2,
			expectedPos: 10,
		},
		{
			name:        "empty line to next word",
			inputString: "abc\n\n   123  xyz",
			pos:         4,
			expectedPos: 10,
		},
		{
			name:        "punctuation",
			inputString: "abc/def/ghi",
			pos:         1,
			expectedPos: 2,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			textTree, err := text.NewTreeFromString(tc.inputString)
			require.NoError(t, err)
			actualPos := NextWordEnd(textTree, tc.pos)
			assert.Equal(t, tc.expectedPos, actualPos)
		})
	}
}

func TestPrevWordStart(t *testing.T) {
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
			name:        "prev word from current word, same line",
			inputString: "abc   defg   hij",
			pos:         6,
			expectedPos: 0,
		},
		{
			name:        "prev word from whitespace, same line",
			inputString: "abc   defg   hij",
			pos:         12,
			expectedPos: 6,
		},
		{
			name:        "prev word from different line",
			inputString: "abc\n   123",
			pos:         7,
			expectedPos: 0,
		},
		{
			name:        "prev word to empty line",
			inputString: "abc\n\n   123",
			pos:         8,
			expectedPos: 4,
		},
		{
			name:        "empty line to prev word",
			inputString: "abc\n\n   123",
			pos:         4,
			expectedPos: 0,
		},
		{
			name:        "multiple empty lines",
			inputString: "\n\n\n\n",
			pos:         2,
			expectedPos: 1,
		},
		{
			name:        "punctuation",
			inputString: "abc/def/ghi",
			pos:         5,
			expectedPos: 4,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			textTree, err := text.NewTreeFromString(tc.inputString)
			require.NoError(t, err)
			actualPos := PrevWordStart(textTree, tc.pos)
			assert.Equal(t, tc.expectedPos, actualPos)
		})
	}
}

func TestCurrentWordStart(t *testing.T) {
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
			name:        "start of document",
			inputString: "abc   defg   hij",
			pos:         0,
			expectedPos: 0,
		},
		{
			name:        "start of word in middle of document",
			inputString: "abc   defg   hij",
			pos:         6,
			expectedPos: 6,
		},
		{
			name:        "middle of word to start of word",
			inputString: "abc   defg   hij",
			pos:         8,
			expectedPos: 6,
		},
		{
			name:        "end of word to start of word",
			inputString: "abc   defg   hij",
			pos:         9,
			expectedPos: 6,
		},
		{
			name:        "start of whitespace",
			inputString: "abc   defg   hij",
			pos:         3,
			expectedPos: 3,
		},
		{
			name:        "middle of whitespace",
			inputString: "abc   defg   hij",
			pos:         4,
			expectedPos: 3,
		},
		{
			name:        "end of whitespace",
			inputString: "abc   defg   hij",
			pos:         5,
			expectedPos: 3,
		},
		{
			name:        "word at start of line",
			inputString: "abc\nxyz",
			pos:         5,
			expectedPos: 4,
		},
		{
			name:        "whitespace at start of line",
			inputString: "abc\n    xyz",
			pos:         6,
			expectedPos: 4,
		},
		{
			name:        "empty line",
			inputString: "abc\n\n   123",
			pos:         4,
			expectedPos: 4,
		},
		{
			name:        "from non-punctuation, stop at punctuation",
			inputString: "abc/def/ghi",
			pos:         5,
			expectedPos: 4,
		},
		{
			name:        "on single punctuation char",
			inputString: "abc/ghi",
			pos:         3,
			expectedPos: 3,
		},
		{
			name:        "on multiple punctuation chars",
			inputString: "abc///ghi",
			pos:         4,
			expectedPos: 3,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			textTree, err := text.NewTreeFromString(tc.inputString)
			require.NoError(t, err)
			actualPos := CurrentWordStart(textTree, tc.pos)
			assert.Equal(t, tc.expectedPos, actualPos)
		})
	}
}

func TestCurrentWordEnd(t *testing.T) {
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
			name:        "end of document",
			inputString: "abc   defg   hijk",
			pos:         14,
			expectedPos: 17,
		},
		{
			name:        "start of word in middle of document",
			inputString: "abc   defg   hij",
			pos:         6,
			expectedPos: 10,
		},
		{
			name:        "middle of word to end of word",
			inputString: "abc   defg   hij",
			pos:         7,
			expectedPos: 10,
		},
		{
			name:        "end of word",
			inputString: "abc   defg   hij",
			pos:         9,
			expectedPos: 10,
		},
		{
			name:        "start of whitespace",
			inputString: "abc   defg   hij",
			pos:         3,
			expectedPos: 6,
		},
		{
			name:        "middle of whitespace",
			inputString: "abc   defg   hij",
			pos:         4,
			expectedPos: 6,
		},
		{
			name:        "end of whitespace",
			inputString: "abc   defg   hij",
			pos:         5,
			expectedPos: 6,
		},
		{
			name:        "word before end of line",
			inputString: "abc\nxyz",
			pos:         1,
			expectedPos: 3,
		},
		{
			name:        "whitespace at end of line",
			inputString: "abc     \nxyz",
			pos:         4,
			expectedPos: 8,
		},
		{
			name:        "empty line",
			inputString: "abc\n\n   123",
			pos:         4,
			expectedPos: 4,
		},
		{
			name:        "punctuation",
			inputString: "abc/def/ghi",
			pos:         5,
			expectedPos: 7,
		},
		{
			name:        "from non-punctuation, stop at punctuation",
			inputString: "abc/def/ghi",
			pos:         5,
			expectedPos: 7,
		},
		{
			name:        "on single punctuation char",
			inputString: "abc/ghi",
			pos:         3,
			expectedPos: 4,
		},
		{
			name:        "on multiple punctuation chars",
			inputString: "abc///ghi",
			pos:         4,
			expectedPos: 6,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			textTree, err := text.NewTreeFromString(tc.inputString)
			require.NoError(t, err)
			actualPos := CurrentWordEnd(textTree, tc.pos)
			assert.Equal(t, tc.expectedPos, actualPos)
		})
	}
}

func TestCurrentWordEndWithTrailingWhitespace(t *testing.T) {
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
			name:        "start of word at end of document",
			inputString: "abcd",
			pos:         0,
			expectedPos: 4,
		},
		{
			name:        "middle of word at end of document",
			inputString: "abcd",
			pos:         2,
			expectedPos: 4,
		},
		{
			name:        "end of word at end of document",
			inputString: "abcd",
			pos:         3,
			expectedPos: 4,
		},
		{
			name:        "on word before whitespace at end of document",
			inputString: "abc    ",
			pos:         2,
			expectedPos: 7,
		},
		{
			name:        "on whitespace at end of document",
			inputString: "abc    ",
			pos:         4,
			expectedPos: 7,
		},
		{
			name:        "on word with trailing whitespace before next word",
			inputString: "abc    def",
			pos:         2,
			expectedPos: 7,
		},
		{
			name:        "on word at end of line",
			inputString: "abc\ndef",
			pos:         1,
			expectedPos: 3,
		},
		{
			name:        "on word before whitespace at end of line",
			inputString: "abc   \ndef",
			pos:         1,
			expectedPos: 6,
		},
		{
			name:        "on word with trailing whitespace before word at end of line",
			inputString: "abc   def\nghi",
			pos:         1,
			expectedPos: 6,
		},
		{
			name:        "on whitespace at end of line",
			inputString: "abc   \nghi",
			pos:         4,
			expectedPos: 6,
		},
		{
			name:        "on empty line",
			inputString: "ab\n\ncd",
			pos:         2,
			expectedPos: 2,
		},
		{
			name:        "on punctuation followed by non-whitespace",
			inputString: "ab,cd",
			pos:         2,
			expectedPos: 3,
		},
		{
			name:        "on punctuation followed by whitespace",
			inputString: "ab,  cd",
			pos:         2,
			expectedPos: 5,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			textTree, err := text.NewTreeFromString(tc.inputString)
			require.NoError(t, err)
			actualPos := CurrentWordEndWithTrailingWhitespace(textTree, tc.pos)
			assert.Equal(t, tc.expectedPos, actualPos)
		})
	}
}

func TestIsPunct(t *testing.T) {
	testCases := []struct {
		r           rune
		expectPunct bool
	}{
		{r: '\x00', expectPunct: false},
		{r: '\x01', expectPunct: false},
		{r: '\x02', expectPunct: false},
		{r: '\x03', expectPunct: false},
		{r: '\x04', expectPunct: false},
		{r: '\x05', expectPunct: false},
		{r: '\x06', expectPunct: false},
		{r: '\a', expectPunct: false},
		{r: '\b', expectPunct: false},
		{r: '\t', expectPunct: false},
		{r: '\n', expectPunct: false},
		{r: '\v', expectPunct: false},
		{r: '\f', expectPunct: false},
		{r: '\r', expectPunct: false},
		{r: '\x0e', expectPunct: false},
		{r: '\x0f', expectPunct: false},
		{r: '\x10', expectPunct: false},
		{r: '\x11', expectPunct: false},
		{r: '\x12', expectPunct: false},
		{r: '\x13', expectPunct: false},
		{r: '\x14', expectPunct: false},
		{r: '\x15', expectPunct: false},
		{r: '\x16', expectPunct: false},
		{r: '\x17', expectPunct: false},
		{r: '\x18', expectPunct: false},
		{r: '\x19', expectPunct: false},
		{r: '\x1a', expectPunct: false},
		{r: '\x1b', expectPunct: false},
		{r: '\x1c', expectPunct: false},
		{r: '\x1d', expectPunct: false},
		{r: '\x1e', expectPunct: false},
		{r: '\x1f', expectPunct: false},
		{r: ' ', expectPunct: false},
		{r: '!', expectPunct: true},
		{r: '"', expectPunct: true},
		{r: '#', expectPunct: true},
		{r: '$', expectPunct: true},
		{r: '%', expectPunct: true},
		{r: '&', expectPunct: true},
		{r: '\'', expectPunct: true},
		{r: '(', expectPunct: true},
		{r: ')', expectPunct: true},
		{r: '*', expectPunct: true},
		{r: '+', expectPunct: true},
		{r: ',', expectPunct: true},
		{r: '-', expectPunct: true},
		{r: '.', expectPunct: true},
		{r: '/', expectPunct: true},
		{r: '0', expectPunct: false},
		{r: '1', expectPunct: false},
		{r: '2', expectPunct: false},
		{r: '3', expectPunct: false},
		{r: '4', expectPunct: false},
		{r: '5', expectPunct: false},
		{r: '6', expectPunct: false},
		{r: '7', expectPunct: false},
		{r: '8', expectPunct: false},
		{r: '9', expectPunct: false},
		{r: ':', expectPunct: true},
		{r: ';', expectPunct: true},
		{r: '<', expectPunct: true},
		{r: '=', expectPunct: true},
		{r: '>', expectPunct: true},
		{r: '?', expectPunct: true},
		{r: '@', expectPunct: true},
		{r: 'A', expectPunct: false},
		{r: 'B', expectPunct: false},
		{r: 'C', expectPunct: false},
		{r: 'D', expectPunct: false},
		{r: 'E', expectPunct: false},
		{r: 'F', expectPunct: false},
		{r: 'G', expectPunct: false},
		{r: 'H', expectPunct: false},
		{r: 'I', expectPunct: false},
		{r: 'J', expectPunct: false},
		{r: 'K', expectPunct: false},
		{r: 'L', expectPunct: false},
		{r: 'M', expectPunct: false},
		{r: 'N', expectPunct: false},
		{r: 'O', expectPunct: false},
		{r: 'P', expectPunct: false},
		{r: 'Q', expectPunct: false},
		{r: 'R', expectPunct: false},
		{r: 'S', expectPunct: false},
		{r: 'T', expectPunct: false},
		{r: 'U', expectPunct: false},
		{r: 'V', expectPunct: false},
		{r: 'W', expectPunct: false},
		{r: 'X', expectPunct: false},
		{r: 'Y', expectPunct: false},
		{r: 'Z', expectPunct: false},
		{r: '[', expectPunct: true},
		{r: '\\', expectPunct: true},
		{r: ']', expectPunct: true},
		{r: '^', expectPunct: true},
		{r: '_', expectPunct: false},
		{r: '`', expectPunct: true},
		{r: 'a', expectPunct: false},
		{r: 'b', expectPunct: false},
		{r: 'c', expectPunct: false},
		{r: 'd', expectPunct: false},
		{r: 'e', expectPunct: false},
		{r: 'f', expectPunct: false},
		{r: 'g', expectPunct: false},
		{r: 'h', expectPunct: false},
		{r: 'i', expectPunct: false},
		{r: 'j', expectPunct: false},
		{r: 'k', expectPunct: false},
		{r: 'l', expectPunct: false},
		{r: 'm', expectPunct: false},
		{r: 'n', expectPunct: false},
		{r: 'o', expectPunct: false},
		{r: 'p', expectPunct: false},
		{r: 'q', expectPunct: false},
		{r: 'r', expectPunct: false},
		{r: 's', expectPunct: false},
		{r: 't', expectPunct: false},
		{r: 'u', expectPunct: false},
		{r: 'v', expectPunct: false},
		{r: 'w', expectPunct: false},
		{r: 'x', expectPunct: false},
		{r: 'y', expectPunct: false},
		{r: 'z', expectPunct: false},
		{r: '{', expectPunct: true},
		{r: '|', expectPunct: true},
		{r: '}', expectPunct: true},
		{r: '~', expectPunct: true},
		{r: '\u007f', expectPunct: false},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%q", tc.r), func(t *testing.T) {
			seg := segment.Empty()
			seg.Extend([]rune{tc.r})
			assert.Equal(t, tc.expectPunct, isPunct(seg))
		})
	}
}
