package text

import (
	"strings"
	"testing"
	"testing/iotest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAlign(t *testing.T) {
	testCases := []struct {
		name      string
		leftText  string
		rightText string
		expected  []LineMatch
	}{
		{
			name:      "both empty",
			leftText:  "",
			rightText: "",
			expected:  []LineMatch{},
		},
		{
			name:      "left empty, right non-empty",
			leftText:  "",
			rightText: "a",
			expected:  nil,
		},
		{
			name:      "left non-empty, right empty",
			leftText:  "a",
			rightText: "",
			expected:  nil,
		},
		{
			name:      "single line, exact match",
			leftText:  "abcd",
			rightText: "abcd",
			expected: []LineMatch{
				{LeftLineNum: 0, RightLineNum: 0},
			},
		},
		{
			name:      "single line, mismatch",
			leftText:  "abc",
			rightText: "xyz",
			expected:  nil,
		},
		{
			name:      "three lines, all match",
			leftText:  "abc\ndef\nghi",
			rightText: "abc\ndef\nghi",
			expected: []LineMatch{
				{LeftLineNum: 0, RightLineNum: 0},
				{LeftLineNum: 1, RightLineNum: 1},
				{LeftLineNum: 2, RightLineNum: 2},
			},
		},
		{
			name:      "three lines, second two match",
			leftText:  "abc\ndef\nghi",
			rightText: "xyz\ndef\nghi",
			expected: []LineMatch{
				{LeftLineNum: 1, RightLineNum: 1},
				{LeftLineNum: 2, RightLineNum: 2},
			},
		},
		{
			name:      "three lines, first two match",
			leftText:  "abc\ndef\nghi",
			rightText: "abc\ndef\nxyz",
			expected: []LineMatch{
				{LeftLineNum: 0, RightLineNum: 0},
				{LeftLineNum: 1, RightLineNum: 1},
			},
		},
		{
			name:      "three lines, first and last match",
			leftText:  "abc\ndef\nghi",
			rightText: "abc\nxyz\nghi",
			expected: []LineMatch{
				{LeftLineNum: 0, RightLineNum: 0},
				{LeftLineNum: 2, RightLineNum: 2},
			},
		},
		{
			name:      "three lines, only middle matches",
			leftText:  "abc\ndef\nghi",
			rightText: "lmn\ndef\nxyz",
			expected: []LineMatch{
				{LeftLineNum: 1, RightLineNum: 1},
			},
		},
		{
			name:      "duplicate lines, all match",
			leftText:  "ab\nab\nab",
			rightText: "ab\nab\nab",
			expected: []LineMatch{
				{LeftLineNum: 0, RightLineNum: 0},
				{LeftLineNum: 1, RightLineNum: 1},
				{LeftLineNum: 2, RightLineNum: 2},
			},
		},
		{
			name:      "duplicate lines, some match",
			leftText:  "ab\nab\ncd\ncd\n",
			rightText: "xy\nab\nxy\ncd\n",
			expected:  nil, // no unique lines to align
		},
		{
			name:      "all blank lines match",
			leftText:  "\n\n\n",
			rightText: "\n\n\n",
			expected: []LineMatch{
				{LeftLineNum: 0, RightLineNum: 0},
				{LeftLineNum: 1, RightLineNum: 1},
				{LeftLineNum: 2, RightLineNum: 2},
			},
		},
		{
			name:      "some blank lines",
			leftText:  "a\n\nb\n\nc",
			rightText: "\n\nb\n\nd\nc",
			expected: []LineMatch{
				{LeftLineNum: 1, RightLineNum: 1},
				{LeftLineNum: 2, RightLineNum: 2},
				{LeftLineNum: 3, RightLineNum: 3},
				{LeftLineNum: 4, RightLineNum: 5},
			},
		},
		{
			name:      "add lines in middle",
			leftText:  "a\nb\nx",
			rightText: "a\nb\nc\nd\ne\nx",
			expected: []LineMatch{
				{LeftLineNum: 0, RightLineNum: 0},
				{LeftLineNum: 1, RightLineNum: 1},
				{LeftLineNum: 2, RightLineNum: 5},
			},
		},
		{
			name:      "delete lines in middle",
			leftText:  "a\nb\nc\nd",
			rightText: "a\nd",
			expected: []LineMatch{
				{LeftLineNum: 0, RightLineNum: 0},
				{LeftLineNum: 3, RightLineNum: 1},
			},
		},
		{
			name:      "add lines before",
			leftText:  "a\nb\nc",
			rightText: "x\ny\nz\na\nb\nc",
			expected: []LineMatch{
				{LeftLineNum: 0, RightLineNum: 3},
				{LeftLineNum: 1, RightLineNum: 4},
				{LeftLineNum: 2, RightLineNum: 5},
			},
		},
		{
			name:      "delete lines before",
			leftText:  "x\ny\nz\na\nb\nc",
			rightText: "a\nb\nc",
			expected: []LineMatch{
				{LeftLineNum: 3, RightLineNum: 0},
				{LeftLineNum: 4, RightLineNum: 1},
				{LeftLineNum: 5, RightLineNum: 2},
			},
		},
		{
			name:      "add lines after",
			leftText:  "a\nb\nc\n",
			rightText: "a\nb\nc\nx\ny\nz",
			expected: []LineMatch{
				{LeftLineNum: 0, RightLineNum: 0},
				{LeftLineNum: 1, RightLineNum: 1},
				{LeftLineNum: 2, RightLineNum: 2},
			},
		},
		{
			name:      "delete lines after",
			leftText:  "a\nb\nc\nx\ny\nz",
			rightText: "a\nb\nc\n",
			expected: []LineMatch{
				{LeftLineNum: 0, RightLineNum: 0},
				{LeftLineNum: 1, RightLineNum: 1},
				{LeftLineNum: 2, RightLineNum: 2},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			matches, err := Align(strings.NewReader(tc.leftText), strings.NewReader(tc.rightText))
			require.NoError(t, err)
			assert.Equal(t, tc.expected, matches)
		})
	}
}

func TestAlignOneByteReader(t *testing.T) {
	leftReader := iotest.OneByteReader(strings.NewReader("abc\ndef\nghi"))
	rightReader := iotest.OneByteReader(strings.NewReader("abc\nxyz\nghi"))
	diff, err := Align(leftReader, rightReader)
	require.NoError(t, err)
	expected :=
		[]LineMatch{
			{LeftLineNum: 0, RightLineNum: 0},
			{LeftLineNum: 2, RightLineNum: 2},
		}
	assert.Equal(t, expected, diff)
}
