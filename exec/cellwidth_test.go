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
		{
			name:          "trademark symbol",
			gc:            []rune{'™'},
			expectedWidth: 1,
		},
		{
			name:          "left square double bracket",
			gc:            []rune{'⟦'},
			expectedWidth: 1,
		},
		{
			name:          "right square double bracket",
			gc:            []rune{'⟧'},
			expectedWidth: 1,
		},
		{
			name:          "left angle double bracket",
			gc:            []rune{'⟪'},
			expectedWidth: 1,
		},
		{
			name:          "right angle double bracket",
			gc:            []rune{'⟫'},
			expectedWidth: 1,
		},
		{
			name:          "thai",
			gc:            []rune{3588, 3657, 3635},
			expectedWidth: 2,
		},
		{
			name:          "combining character (angstrom)",
			gc:            []rune{'A', '\u030a'},
			expectedWidth: 1,
		},
		{
			name:          "emoticon (blowing a kiss)",
			gc:            []rune{'\U0001f618'},
			expectedWidth: 2,
		},
		{
			name:          "emoji (airplane)",
			gc:            []rune{'\u2708'},
			expectedWidth: 1,
		},
		{
			name:          "emoji (clover symbol)",
			gc:            []rune{'\u2318'},
			expectedWidth: 1,
		},
		{
			name:          "enclosed exclamation",
			gc:            []rune{'!', '\u20e3'},
			expectedWidth: 1,
		},
		{
			name:          "emoji zero-width joiner (female vampire)",
			gc:            []rune{'\U0001f9db', '\u200d', '\u2640'},
			expectedWidth: 2,
		},
		{
			name:          "emoji zero-width joiner (family, woman+girl+girl)",
			gc:            []rune{'\U0001f469', '\u200d', '\U0001f467', '\u200d', '\U0001f467'},
			expectedWidth: 2,
		},
		{
			name:          "region (usa)",
			gc:            []rune{'\U0001f1fa', '\U0001f1f8'},
			expectedWidth: 1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			width := GraphemeClusterWidth(tc.gc, tc.offset, 4)
			assert.Equal(t, tc.expectedWidth, width)
		})
	}
}
