package text

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEscaper(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty",
			input:    "",
			expected: "<>",
		},
		{
			name:     "ascii",
			input:    "a",
			expected: "<U+0061>",
		},
		{
			name:     "multiple",
			input:    "abc",
			expected: "<U+0061,U+0062,U+0063>",
		},
		{
			name: "emoji",
			// "heart on fire"
			input:    "\u2764\ufe0f\u200d\u1f525",
			expected: "<U+2764,U+FE0F,U+200D,U+1F52,U+0035>",
		},
		{
			name: "supplementary character",
			// "grinning face", hex codepoint is five digits long
			input:    "\U0001f600",
			expected: "<U+1F600>",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			e := &Escaper{}
			actual := string(e.RunesToStr([]rune(tc.input)))
			assert.Equal(t, tc.expected, actual)
		})
	}
}
