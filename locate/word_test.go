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
		name                string
		inputString         string
		pos                 uint64
		count               uint64
		withPunct           bool
		stopAtEndOfLastLine bool
		expectedPos         uint64
	}{
		{
			name:        "empty",
			inputString: "",
			pos:         0,
			count:       1,
			expectedPos: 0,
		},
		{
			name:        "next word from current word, same line",
			inputString: "abc   defg   hij",
			pos:         1,
			count:       1,
			expectedPos: 6,
		},
		{
			name:        "next word from whitespace, same line",
			inputString: "abc   defg   hij",
			pos:         4,
			count:       1,
			expectedPos: 6,
		},
		{
			name:        "next word from different line",
			inputString: "abc\n   123",
			pos:         1,
			count:       1,
			expectedPos: 7,
		},
		{
			name:        "next word to empty line",
			inputString: "abc\n\n   123",
			pos:         1,
			count:       1,
			expectedPos: 4,
		},
		{
			name:        "line ending with punctuation",
			inputString: "abc.\n123",
			pos:         3,
			count:       1,
			expectedPos: 5,
		},
		{
			name:        "line ending with punctuation",
			inputString: "abc.\n123",
			pos:         3,
			count:       1,
			expectedPos: 5,
			withPunct:   true,
		},
		{
			name:        "punctuation before whitespace",
			inputString: "-   foo",
			pos:         0,
			count:       1,
			expectedPos: 4,
		},
		{
			name:        "punctuation before whitespace",
			inputString: "-   foo",
			pos:         0,
			count:       1,
			expectedPos: 4,
			withPunct:   true,
		},
		{
			name:        "empty line to next word",
			inputString: "abc\n\n   123",
			pos:         4,
			count:       1,
			expectedPos: 8,
		},
		{
			name:        "multiple empty lines",
			inputString: "\n\n\n\n",
			pos:         1,
			count:       1,
			expectedPos: 2,
		},
		{
			name:        "non-punctuation to punctuation",
			inputString: "abc/def/ghi",
			pos:         1,
			count:       1,
			expectedPos: 3,
		},
		{
			name:        "non-punctuation to punctuation",
			inputString: "abc/def/ghi",
			pos:         1,
			count:       1,
			expectedPos: 11,
			withPunct:   true,
		},
		{
			name:        "punctuation to non-punctuation",
			inputString: "abc/def/ghi",
			pos:         3,
			count:       1,
			expectedPos: 4,
		},
		{
			name:        "punctuation to non-punctuation",
			inputString: "abc/def/ghi",
			pos:         3,
			count:       1,
			expectedPos: 11,
			withPunct:   true,
		},
		{
			name:        "repeated punctuation",
			inputString: "abc////cde",
			pos:         3,
			count:       1,
			expectedPos: 7,
		},
		{
			name:        "repeated punctuation",
			inputString: "abc////cde",
			pos:         3,
			count:       1,
			expectedPos: 10,
			withPunct:   true,
		},
		{
			name:        "underscores treated as non-punctuation",
			inputString: "abc_def ghi",
			pos:         0,
			count:       1,
			expectedPos: 8,
		},
		{
			name:        "underscores treated as non-punctuation",
			inputString: "abc_def ghi",
			pos:         0,
			count:       1,
			expectedPos: 8,
			withPunct:   true,
		},
		{
			name:        "last word in document",
			inputString: "foo bar",
			pos:         5,
			count:       1,
			expectedPos: 7,
		},
		{
			name:        "count zero",
			inputString: "lorem ipsum dolor sit amet",
			pos:         2,
			count:       0,
			expectedPos: 2,
		},
		{
			name:        "count three",
			inputString: "lorem ipsum dolor sit amet",
			pos:         2,
			count:       3,
			expectedPos: 18,
		},
		{
			name:                "stop at end of last line",
			inputString:         "foo bar\nbaz bat\n",
			pos:                 4,
			count:               1,
			stopAtEndOfLastLine: true,
			expectedPos:         7,
		},
		{
			name:                "stop at end of last line, on empty line",
			inputString:         "foo\n\nbaz bat\n",
			pos:                 3,
			count:               1,
			stopAtEndOfLastLine: true,
			expectedPos:         3,
		},
		{
			name:                "stop at end of last line with count",
			inputString:         "foo bar\nbaz bat\nlorem",
			pos:                 0,
			count:               4,
			stopAtEndOfLastLine: true,
			expectedPos:         15,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			textTree, err := text.NewTreeFromString(tc.inputString)
			require.NoError(t, err)
			actualPos := NextWordStart(textTree, tc.pos, tc.count, tc.withPunct, tc.stopAtEndOfLastLine)
			assert.Equal(t, tc.expectedPos, actualPos)
		})
	}
}

func TestNextWordEnd(t *testing.T) {
	testCases := []struct {
		name        string
		inputString string
		pos         uint64
		count       uint64
		expectedPos uint64
		withPunct   bool
	}{
		{
			name:        "empty",
			inputString: "",
			pos:         0,
			count:       1,
			expectedPos: 0,
		},
		{
			name:        "end of word from start of current word",
			inputString: "abc   defg   hij",
			pos:         6,
			count:       1,
			expectedPos: 9,
		},
		{
			name:        "end of word from middle of current word",
			inputString: "abc   defg   hij",
			pos:         7,
			count:       1,
			expectedPos: 9,
		},
		{
			name:        "next word from end of current word",
			inputString: "abc   defg   hij",
			pos:         2,
			count:       1,
			expectedPos: 9,
		},
		{
			name:        "next word from whitespace",
			inputString: "abc   defg   hij",
			pos:         4,
			count:       1,
			expectedPos: 9,
		},
		{
			name:        "next word past empty line",
			inputString: "abc\n\n   123   xyz",
			pos:         2,
			count:       1,
			expectedPos: 10,
		},
		{
			name:        "empty line to next word",
			inputString: "abc\n\n   123  xyz",
			pos:         4,
			count:       1,
			expectedPos: 10,
		},
		{
			name:        "punctuation",
			inputString: "abc/def/ghi",
			pos:         1,
			count:       1,
			expectedPos: 2,
		},
		{
			name:        "punctuation",
			inputString: "abc/def/ghi",
			pos:         1,
			count:       1,
			expectedPos: 10,
			withPunct:   true,
		},
		{
			name:        "last word in document, third to last character",
			inputString: "foo bar",
			pos:         4,
			count:       1,
			expectedPos: 6,
		},
		{
			name:        "last word in document, second to last character",
			inputString: "foo bar",
			pos:         5,
			count:       1,
			expectedPos: 6,
		},
		{
			name:        "last word in document, last character",
			inputString: "foo bar",
			pos:         6,
			count:       1,
			expectedPos: 6,
		},
		{
			name:        "count zero",
			inputString: "lorem ipsum dolor sit amet",
			pos:         2,
			count:       0,
			expectedPos: 2,
		},
		{
			name:        "count three",
			inputString: "lorem ipsum dolor sit amet",
			pos:         2,
			count:       3,
			expectedPos: 16,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			textTree, err := text.NewTreeFromString(tc.inputString)
			require.NoError(t, err)
			actualPos := NextWordEnd(textTree, tc.pos, tc.count, tc.withPunct)
			assert.Equal(t, tc.expectedPos, actualPos)
		})
	}
}

func TestPrevWordStart(t *testing.T) {
	testCases := []struct {
		name        string
		inputString string
		pos         uint64
		count       uint64
		expectedPos uint64
		withPunct   bool
	}{
		{
			name:        "empty",
			inputString: "",
			pos:         0,
			count:       1,
			expectedPos: 0,
		},
		{
			name:        "prev word from current word, same line",
			inputString: "abc   defg   hij",
			pos:         6,
			count:       1,
			expectedPos: 0,
		},
		{
			name:        "prev word from whitespace, same line",
			inputString: "abc   defg   hij",
			pos:         12,
			count:       1,
			expectedPos: 6,
		},
		{
			name:        "prev word from different line",
			inputString: "abc\n   123",
			pos:         7,
			count:       1,
			expectedPos: 0,
		},
		{
			name:        "prev word to empty line",
			inputString: "abc\n\n   123",
			pos:         8,
			count:       1,
			expectedPos: 4,
		},
		{
			name:        "empty line to prev word",
			inputString: "abc\n\n   123",
			pos:         4,
			count:       1,
			expectedPos: 0,
		},
		{
			name:        "multiple empty lines",
			inputString: "\n\n\n\n",
			pos:         2,
			count:       1,
			expectedPos: 1,
		},
		{
			name:        "punctuation",
			inputString: "abc/def/ghi",
			pos:         5,
			count:       1,
			expectedPos: 4,
		},
		{
			name:        "line ending with punctuation",
			inputString: "abc.\n123",
			pos:         5,
			count:       1,
			expectedPos: 3,
		},
		{
			name:        "punctuation before whitespace",
			inputString: "-   foo",
			pos:         4,
			count:       1,
			expectedPos: 0,
		},
		{
			name:        "count zero",
			inputString: "lorem ipsum dolor sit amet",
			pos:         18,
			count:       0,
			expectedPos: 18,
		},
		{
			name:        "count three",
			inputString: "lorem ipsum dolor sit amet",
			pos:         25,
			count:       3,
			expectedPos: 12,
		},
		{
			name:        "count three, with punctuation",
			inputString: "lorem, ipsum, dolor, sit, amet",
			withPunct:   true,
			pos:         29,
			count:       3,
			expectedPos: 14,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			textTree, err := text.NewTreeFromString(tc.inputString)
			require.NoError(t, err)
			actualPos := PrevWordStart(textTree, tc.pos, tc.count, tc.withPunct)
			assert.Equal(t, tc.expectedPos, actualPos)
		})
	}
}

func TestWordObject(t *testing.T) {
	testCases := []struct {
		name             string
		inputString      string
		pos              uint64
		count            uint64
		expectedStartPos uint64
		expectedEndPos   uint64
	}{
		{
			name:             "empty",
			inputString:      "",
			pos:              0,
			count:            1,
			expectedStartPos: 0,
			expectedEndPos:   0,
		},
		{
			name:             "on start of leading whitespace before word",
			inputString:      "abc   def  ghi",
			pos:              3,
			count:            1,
			expectedStartPos: 3,
			expectedEndPos:   9,
		},
		{
			name:             "on middle of leading whitespace before word",
			inputString:      "abc   def  ghi",
			pos:              4,
			count:            1,
			expectedStartPos: 3,
			expectedEndPos:   9,
		},
		{
			name:             "on end of leading whitespace before word",
			inputString:      "abc   def  ghi",
			pos:              5,
			count:            1,
			expectedStartPos: 3,
			expectedEndPos:   9,
		},
		{
			name:             "on start of word with trailing whitespace",
			inputString:      "abc def    ghi",
			pos:              4,
			count:            1,
			expectedStartPos: 4,
			expectedEndPos:   11,
		},
		{
			name:             "on middle of word with trailing whitespace",
			inputString:      "abc def    ghi",
			pos:              5,
			count:            1,
			expectedStartPos: 4,
			expectedEndPos:   11,
		},
		{
			name:             "on end of word with trailing whitespace",
			inputString:      "abc def    ghi",
			pos:              6,
			count:            1,
			expectedStartPos: 4,
			expectedEndPos:   11,
		},
		{
			name:             "start of word after punctuation",
			inputString:      "abc/def/ghi",
			pos:              4,
			count:            1,
			expectedStartPos: 4,
			expectedEndPos:   7,
		},
		{
			name:             "middle of word after punctuation",
			inputString:      "abc/def/ghi",
			pos:              5,
			count:            1,
			expectedStartPos: 4,
			expectedEndPos:   7,
		},
		{
			name:             "end of word after punctuation",
			inputString:      "abc/def/ghi",
			pos:              6,
			count:            1,
			expectedStartPos: 4,
			expectedEndPos:   7,
		},
		{
			name:             "on punctuation surrounded by words",
			inputString:      "abc/def/ghi",
			pos:              3,
			count:            1,
			expectedStartPos: 3,
			expectedEndPos:   4,
		},
		{
			name:             "on punctuation surrounded by whitespace",
			inputString:      "a   /   b",
			pos:              4,
			count:            1,
			expectedStartPos: 4,
			expectedEndPos:   8,
		},
		{
			name:             "on multiple punctuation chars",
			inputString:      "abc///ghi",
			pos:              4,
			count:            1,
			expectedStartPos: 3,
			expectedEndPos:   6,
		},
		{
			name:             "on leading whitespace before punctuation",
			inputString:      "foo  {bar",
			pos:              3,
			count:            1,
			expectedStartPos: 3,
			expectedEndPos:   6,
		},
		{
			name:             "whitespace at start of line",
			inputString:      "abc\n    xyz",
			pos:              6,
			count:            1,
			expectedStartPos: 4,
			expectedEndPos:   11,
		},
		{

			name:             "empty line, indentation",
			inputString:      "abc\n\n   123",
			pos:              4,
			count:            1,
			expectedStartPos: 4,
			expectedEndPos:   11,
		},
		{

			name:             "empty line, no indentation",
			inputString:      "abc\n\n123",
			pos:              4,
			count:            1,
			expectedStartPos: 4,
			expectedEndPos:   8,
		},
		{
			name:             "start of word at end of document",
			inputString:      "abcd",
			pos:              0,
			count:            1,
			expectedStartPos: 0,
			expectedEndPos:   4,
		},
		{
			name:             "middle of word at end of document",
			inputString:      "abcd",
			pos:              2,
			count:            1,
			expectedStartPos: 0,
			expectedEndPos:   4,
		},
		{
			name:             "end of word at end of document",
			inputString:      "abcd",
			pos:              3,
			count:            1,
			expectedStartPos: 0,
			expectedEndPos:   4,
		},

		{
			name:             "on word before whitespace at end of document",
			inputString:      "abc    ",
			pos:              2,
			count:            1,
			expectedStartPos: 0,
			expectedEndPos:   7,
		},
		{
			name:             "on whitespace at end of document",
			inputString:      "abc    ",
			pos:              4,
			count:            1,
			expectedStartPos: 3,
			expectedEndPos:   7,
		},
		{
			name:             "count zero",
			inputString:      "lorem ipsum dolor sit amet",
			pos:              7,
			count:            0,
			expectedStartPos: 7,
			expectedEndPos:   7,
		},
		{
			name:             "count three with leading whitespace",
			inputString:      "lorem ipsum dolor sit amet",
			pos:              5,
			count:            3,
			expectedStartPos: 5,
			expectedEndPos:   21,
		},
		{
			name:             "count three with trailing whitespace",
			inputString:      "lorem ipsum dolor sit amet",
			pos:              7,
			count:            3,
			expectedStartPos: 6,
			expectedEndPos:   22,
		},
		{
			name:             "count three with punctuation",
			inputString:      "lorem.ipsum.dolor.sit.amet",
			pos:              5,
			count:            3,
			expectedStartPos: 5,
			expectedEndPos:   12,
		},
		{
			name:             "count three with punctuation with leading whitespace",
			inputString:      "lorem   ipsum.dolor.sit.amet",
			pos:              6,
			count:            3,
			expectedStartPos: 5,
			expectedEndPos:   19,
		},
		{
			name:             "count with multiple whitespace, leading whitespace",
			inputString:      "lorem     ipsum      dolor     sit      amet",
			pos:              7,
			count:            2,
			expectedStartPos: 5,
			expectedEndPos:   26,
		},
		{
			name:             "count with multiple whitespace, trailing whitespace",
			inputString:      "lorem     ipsum      dolor     sit      amet",
			pos:              2,
			count:            2,
			expectedStartPos: 0,
			expectedEndPos:   21,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			textTree, err := text.NewTreeFromString(tc.inputString)
			require.NoError(t, err)
			startPos, endPos := WordObject(textTree, tc.pos, tc.count)
			assert.Equal(t, tc.expectedStartPos, startPos)
			assert.Equal(t, tc.expectedEndPos, endPos)
		})
	}
}

func TestInnerWordObject(t *testing.T) {
	testCases := []struct {
		name             string
		inputString      string
		pos              uint64
		count            uint64
		expectedStartPos uint64
		expectedEndPos   uint64
	}{
		{
			name:             "empty",
			inputString:      "",
			pos:              0,
			count:            1,
			expectedStartPos: 0,
			expectedEndPos:   0,
		},
		{
			name:             "on start of leading whitespace before word",
			inputString:      "abc   def  ghi",
			pos:              3,
			count:            1,
			expectedStartPos: 3,
			expectedEndPos:   6,
		},
		{
			name:             "on middle of leading whitespace before word",
			inputString:      "abc   def  ghi",
			pos:              4,
			count:            1,
			expectedStartPos: 3,
			expectedEndPos:   6,
		},
		{
			name:             "on end of leading whitespace before word",
			inputString:      "abc   def  ghi",
			pos:              5,
			count:            1,
			expectedStartPos: 3,
			expectedEndPos:   6,
		},
		{
			name:             "on start of word with trailing whitespace",
			inputString:      "abc def    ghi",
			pos:              4,
			count:            1,
			expectedStartPos: 4,
			expectedEndPos:   7,
		},
		{
			name:             "on middle of word with trailing whitespace",
			inputString:      "abc def    ghi",
			pos:              5,
			count:            1,
			expectedStartPos: 4,
			expectedEndPos:   7,
		},
		{
			name:             "on end of word with trailing whitespace",
			inputString:      "abc def    ghi",
			pos:              6,
			count:            1,
			expectedStartPos: 4,
			expectedEndPos:   7,
		},
		{
			name:             "start of word after punctuation",
			inputString:      "abc/def/ghi",
			pos:              4,
			count:            1,
			expectedStartPos: 4,
			expectedEndPos:   7,
		},
		{
			name:             "middle of word after punctuation",
			inputString:      "abc/def/ghi",
			pos:              5,
			count:            1,
			expectedStartPos: 4,
			expectedEndPos:   7,
		},
		{
			name:             "end of word after punctuation",
			inputString:      "abc/def/ghi",
			pos:              6,
			count:            1,
			expectedStartPos: 4,
			expectedEndPos:   7,
		},
		{
			name:             "on punctuation surrounded by words",
			inputString:      "abc/def/ghi",
			pos:              3,
			count:            1,
			expectedStartPos: 3,
			expectedEndPos:   4,
		},
		{
			name:             "on punctuation surrounded by whitespace",
			inputString:      "a   /   b",
			pos:              4,
			count:            1,
			expectedStartPos: 4,
			expectedEndPos:   5,
		},
		{
			name:             "on multiple punctuation chars",
			inputString:      "abc///ghi",
			pos:              4,
			count:            1,
			expectedStartPos: 3,
			expectedEndPos:   6,
		},
		{
			name:             "on leading whitespace before punctuation",
			inputString:      "foo  {bar",
			pos:              3,
			count:            1,
			expectedStartPos: 3,
			expectedEndPos:   5,
		},
		{
			name:             "whitespace at start of line",
			inputString:      "abc\n    xyz",
			pos:              6,
			count:            1,
			expectedStartPos: 4,
			expectedEndPos:   8,
		},
		{

			name:             "empty line, indentation",
			inputString:      "abc\n\n   123",
			pos:              4,
			count:            1,
			expectedStartPos: 4,
			expectedEndPos:   4,
		},
		{

			name:             "empty line, no indentation",
			inputString:      "abc\n\n123",
			pos:              4,
			count:            1,
			expectedStartPos: 4,
			expectedEndPos:   4,
		},
		{
			name:             "word at end of line",
			inputString:      "foo bar\nbaz\nbat",
			pos:              5,
			count:            1,
			expectedStartPos: 4,
			expectedEndPos:   7,
		},
		{
			name:             "start of word at end of document",
			inputString:      "abcd",
			pos:              0,
			count:            1,
			expectedStartPos: 0,
			expectedEndPos:   4,
		},
		{
			name:             "middle of word at end of document",
			inputString:      "abcd",
			pos:              2,
			count:            1,
			expectedStartPos: 0,
			expectedEndPos:   4,
		},
		{
			name:             "end of word at end of document",
			inputString:      "abcd",
			pos:              3,
			count:            1,
			expectedStartPos: 0,
			expectedEndPos:   4,
		},

		{
			name:             "on word before whitespace at end of document",
			inputString:      "abc    ",
			pos:              2,
			count:            1,
			expectedStartPos: 0,
			expectedEndPos:   3,
		},
		{
			name:             "on whitespace at end of document",
			inputString:      "abc    ",
			pos:              4,
			count:            1,
			expectedStartPos: 3,
			expectedEndPos:   7,
		},
		{
			name:             "count zero",
			inputString:      "lorem ipsum dolor sit amet",
			pos:              7,
			count:            0,
			expectedStartPos: 7,
			expectedEndPos:   7,
		},
		{
			name:             "count three",
			inputString:      "lorem ipsum dolor sit amet",
			pos:              7,
			count:            3,
			expectedStartPos: 6,
			expectedEndPos:   17,
		},
		{
			name:             "count past end of line",
			inputString:      "lorem ipsum\ndolor\nsit amet",
			pos:              1,
			count:            5,
			expectedStartPos: 0,
			expectedEndPos:   21,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			textTree, err := text.NewTreeFromString(tc.inputString)
			require.NoError(t, err)
			startPos, endPos := InnerWordObject(textTree, tc.pos, tc.count)
			assert.Equal(t, tc.expectedStartPos, startPos)
			assert.Equal(t, tc.expectedEndPos, endPos)
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
