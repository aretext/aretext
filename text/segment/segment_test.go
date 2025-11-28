package segment

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSegmentHasNewline(t *testing.T) {
	testCases := []struct {
		name     string
		str      string
		expected bool
	}{
		{name: "empty", str: "", expected: false},
		{name: "LF", str: "\n", expected: true},
		{name: "CRLF", str: "\r\n", expected: true},
		{name: "ascii", str: "a", expected: false},
		{name: "single line", str: "abcdefg", expected: false},
		{name: "two lines", str: "abc\ndefg", expected: true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			seg := Empty()
			for _, r := range tc.str {
				seg.Append(r)
			}
			assert.Equal(t, tc.expected, seg.HasNewline())
		})
	}
}

func TestSegmentIsWhitespace(t *testing.T) {
	testCases := []struct {
		name     string
		str      string
		expected bool
	}{
		{name: "empty", str: "", expected: false},
		{name: "space", str: " ", expected: true},
		{name: "tab", str: "\t", expected: true},
		{name: "ascii", str: "a", expected: false},
		{name: "multiple spaces", str: "    ", expected: true},
		{name: "ascii and spaces", str: "a b c", expected: false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			seg := Empty()
			for _, r := range tc.str {
				seg.Append(r)
			}
			assert.Equal(t, tc.expected, seg.IsWhitespace())
		})
	}
}

func TestIterRunes(t *testing.T) {
	testCases := []struct{
		name string
		str string
	}{
		{name: "empty", str: ""},
		{name: "one", str: "a"},
		{name: "several", str: "abcd"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			seg := Empty()
			for _, r := range tc.str {
				seg.Append(r)
			}

			retrievedRunes := make([]rune, 0, seg.NumRunes())
			for r := range seg.IterRunes() {
				retrievedRunes = append(retrievedRunes, r)
			}
			assert.Equal(t, tc.str, string(retrievedRunes))
		})
	}
}

