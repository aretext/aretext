package locate

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/aretext/aretext/editor/syntax"
)

func TestStringObject(t *testing.T) {
	testCases := []struct {
		name           string
		inputString    string
		pos            uint64
		syntaxLanguage syntax.Language
		quoteRune      rune
		includeQuotes  bool
		expectStartPos uint64
		expectEndPos   uint64
	}{
		{
			name:           "empty",
			inputString:    "",
			pos:            0,
			expectStartPos: 0,
			expectEndPos:   0,
		},
		{
			name:           "on start quote",
			inputString:    `x "abcd" y`,
			pos:            2,
			quoteRune:      '"',
			includeQuotes:  true,
			expectStartPos: 2,
			expectEndPos:   8,
		},
		{
			name:           "include quotes",
			inputString:    `"abcd"`,
			pos:            3,
			quoteRune:      '"',
			includeQuotes:  true,
			expectStartPos: 0,
			expectEndPos:   6,
		},
		{
			name:           "exclude quotes",
			inputString:    `"abcd"`,
			pos:            3,
			quoteRune:      '"',
			includeQuotes:  false,
			expectStartPos: 1,
			expectEndPos:   5,
		},
		{
			name:           "include quotes, empty string",
			inputString:    `""`,
			pos:            1,
			quoteRune:      '"',
			includeQuotes:  true,
			expectStartPos: 0,
			expectEndPos:   2,
		},
		{
			name:           "exclude quotes, empty string",
			inputString:    `""`,
			pos:            1,
			quoteRune:      '"',
			includeQuotes:  false,
			expectStartPos: 1,
			expectEndPos:   1,
		},
		{
			name:           "string syntax token, include quotes",
			inputString:    `"ab\"cd"`,
			pos:            2,
			quoteRune:      '"',
			includeQuotes:  true,
			expectStartPos: 0,
			expectEndPos:   8,
			syntaxLanguage: syntax.LanguageGo,
		},
		{
			name:           "string syntax token, exclude quotes",
			inputString:    `"ab\"cd"`,
			pos:            2,
			quoteRune:      '"',
			includeQuotes:  false,
			expectStartPos: 1,
			expectEndPos:   7,
			syntaxLanguage: syntax.LanguageGo,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			textTree, syntaxParser := textTreeAndSyntaxParser(t, tc.inputString, tc.syntaxLanguage)
			actualStartPos, actualEndPos := StringObject(tc.quoteRune, textTree, syntaxParser, tc.includeQuotes, tc.pos)
			assert.Equal(t, tc.expectStartPos, actualStartPos)
			assert.Equal(t, tc.expectEndPos, actualEndPos)
		})
	}
}
