package locate

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/aretext/aretext/syntax"
	"github.com/aretext/aretext/syntax/parser"
	"github.com/aretext/aretext/text"
)

func TestMatchingCodeBlockDelimiter(t *testing.T) {
	testCases := []struct {
		name           string
		inputString    string
		pos            uint64
		syntaxLanguage syntax.Language
		expectMatch    bool
		expectPos      uint64
	}{
		{
			name:        "empty",
			inputString: "",
			pos:         0,
			expectMatch: false,
		},
		{
			name:        "not on paren or brace",
			inputString: "{ abcd }",
			pos:         3,
			expectMatch: false,
		},
		{
			name:        "on brace with no match",
			inputString: "foo { bar baz",
			pos:         4,
			expectMatch: false,
		},
		{
			name:        "match paren",
			inputString: "foo ( bar ) baz",
			pos:         4,
			expectMatch: true,
			expectPos:   10,
		},
		{
			name:        "match brace",
			inputString: "foo { bar } baz",
			pos:         4,
			expectMatch: true,
			expectPos:   10,
		},
		{
			name:        "match bracket",
			inputString: "foo [ bar ] baz",
			pos:         4,
			expectMatch: true,
			expectPos:   10,
		},
		{
			name:        "match with nesting",
			inputString: "Lorem (ipsum (dolor (sit (amet) consectetur) adipiscing) elit) sed",
			pos:         13,
			expectMatch: true,
			expectPos:   55,
		},
		{
			name: "match ignore nesting in Go strings and comments",
			inputString: `func test() bool {
	if x == 2 {
		// ignore }
		y := "ignore }"
		return true
	}
}`,
			syntaxLanguage: syntax.LanguageGo,
			pos:            30,
			expectMatch:    true,
			expectPos:      79,
		},
		{
			name: "match within Go string",
			inputString: `func() bool {
	x := "{lorem {ipsum}}"
}`,
			syntaxLanguage: syntax.LanguageGo,
			pos:            21,
			expectMatch:    true,
			expectPos:      35,
		},
		{
			name: "match within Go comment",
			inputString: `func() bool {
/*
	{ foo { bar } baz }
*/
	return true
}`,
			syntaxLanguage: syntax.LanguageGo,
			pos:            18,
			expectMatch:    true,
			expectPos:      36,
		},
		{
			name: "do not match from different Go comments",
			inputString: `func() bool {
// open { in comment
// close } in a different comment
}`,
			syntaxLanguage: syntax.LanguageGo,
			pos:            22,
			expectMatch:    false,
		},
		{
			name:           "do not match from inside Go string to outside Go string",
			inputString:    `"foo { " }`,
			syntaxLanguage: syntax.LanguageGo,
			pos:            5,
			expectMatch:    false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			textTree, syntaxParser := textTreeAndSyntaxParser(t, tc.inputString, tc.syntaxLanguage)
			actualPos, ok := MatchingCodeBlockDelimiter(textTree, syntaxParser, tc.pos)
			assert.Equal(t, tc.expectMatch, ok)
			if ok {
				assert.Equal(t, tc.expectPos, actualPos)

				// Verify that we get back the original position from the matched position.
				originalPos, ok := MatchingCodeBlockDelimiter(textTree, syntaxParser, actualPos)
				assert.True(t, ok)
				assert.Equal(t, tc.pos, originalPos)
			}
		})
	}
}

func TestNextUnmatchedCloseBrace(t *testing.T) {
	testCases := []struct {
		name           string
		inputString    string
		pos            uint64
		syntaxLanguage syntax.Language
		expectMatch    bool
		expectPos      uint64
	}{
		{
			name:        "empty",
			inputString: "",
			pos:         0,
			expectMatch: false,
		},
		{
			name:        "no match",
			inputString: "abcd 1234",
			pos:         2,
			expectMatch: false,
		},
		{
			name:        "on open brace",
			inputString: "{ a { b { c { d } } } }",
			pos:         4,
			expectMatch: true,
			expectPos:   20,
		},
		{
			name:        "after open brace",
			inputString: "{ a { b { c { d } } } }",
			pos:         5,
			expectMatch: true,
			expectPos:   20,
		},
		{
			name: "ignore brace in Go comment",
			inputString: `{
abc
	{
	// }
	}
}`,
			syntaxLanguage: syntax.LanguageGo,
			pos:            2,
			expectMatch:    true,
			expectPos:      18,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			textTree, syntaxParser := textTreeAndSyntaxParser(t, tc.inputString, tc.syntaxLanguage)
			actualPos, ok := NextUnmatchedCloseBrace(textTree, syntaxParser, tc.pos)
			assert.Equal(t, tc.expectMatch, ok)
			if ok {
				assert.Equal(t, tc.expectPos, actualPos)
			}
		})
	}
}

func TestPrevUnmatchedOpenBrace(t *testing.T) {
	testCases := []struct {
		name           string
		inputString    string
		pos            uint64
		syntaxLanguage syntax.Language
		expectMatch    bool
		expectPos      uint64
	}{
		{
			name:        "empty",
			inputString: "",
			pos:         0,
			expectMatch: false,
		},
		{
			name:        "no match",
			inputString: "abcd 1234",
			pos:         6,
			expectMatch: false,
		},
		{
			name:        "on close brace",
			inputString: "{ a { b { c { d } } } }",
			pos:         20,
			expectMatch: true,
			expectPos:   4,
		},
		{
			name:        "after close brace",
			inputString: "{ a { b { c { d } } } }",
			pos:         19,
			expectMatch: true,
			expectPos:   4,
		},
		{
			name: "ignore brace in Go comment",
			inputString: `{
	{
	// {
	}
abc
}`,
			syntaxLanguage: syntax.LanguageGo,
			pos:            15,
			expectMatch:    true,
			expectPos:      0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			textTree, syntaxParser := textTreeAndSyntaxParser(t, tc.inputString, tc.syntaxLanguage)
			actualPos, ok := PrevUnmatchedOpenBrace(textTree, syntaxParser, tc.pos)
			assert.Equal(t, tc.expectMatch, ok)
			if ok {
				assert.Equal(t, tc.expectPos, actualPos)
			}
		})
	}
}

func textTreeAndSyntaxParser(t *testing.T, s string, syntaxLanguage syntax.Language) (*text.Tree, *parser.P) {
	textTree, err := text.NewTreeFromString(s)
	require.NoError(t, err)

	syntaxParser := syntax.ParserForLanguage(syntaxLanguage)
	if syntaxParser != nil {
		syntaxParser.ParseAll(textTree)
	}

	return textTree, syntaxParser
}
