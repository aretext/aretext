package exec

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGraphemeClusterWidth(t *testing.T) {
	testCases := []struct {
		name          string
		gc            []rune
		offset        uint64
		expectedWidth uint64
	}{
		{
			name:          "empty",
			gc:            []rune{},
			expectedWidth: 0,
		},
		{
			name:          "ascii printable",
			gc:            []rune{'a'},
			expectedWidth: 1,
		},
		{
			name:          "line feed",
			gc:            []rune{'\n'},
			expectedWidth: 0,
		},
		{
			name:          "carriage return and line feed",
			gc:            []rune{'\r', '\n'},
			expectedWidth: 0,
		},
		{
			name:          "tab at start of line",
			gc:            []rune{'\t'},
			expectedWidth: 4,
		},
		{
			name:          "tab at misaligned offset",
			gc:            []rune{'\t'},
			offset:        1,
			expectedWidth: 3,
		},
		{
			name:          "tab at aligned offset",
			gc:            []rune{'\t'},
			offset:        4,
			expectedWidth: 4,
		},
		{
			name:          "full width east-asian character",
			gc:            []rune{'界'},
			expectedWidth: 2,
		},
		{
			name:          "combining accent mark",
			gc:            []rune{'a', '\u0300'},
			expectedWidth: 1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			width := GraphemeClusterWidth(tc.gc, tc.offset)
			assert.Equal(t, tc.expectedWidth, width)
		})
	}
}
