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
			name:        "match angle brackets",
			inputString: "foo < bar > baz",
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

func TestNextUnmatchedCloseDelimiter(t *testing.T) {
	testCases := []struct {
		name           string
		delimiterPair  DelimiterPair
		inputString    string
		pos            uint64
		syntaxLanguage syntax.Language
		expectMatch    bool
		expectPos      uint64
	}{
		{
			name:          "empty",
			delimiterPair: BracePair,
			inputString:   "",
			pos:           0,
			expectMatch:   false,
		},
		{
			name:          "no match braces",
			delimiterPair: BracePair,
			inputString:   "abcd 1234",
			pos:           2,
			expectMatch:   false,
		},
		{
			name:          "on open brace",
			delimiterPair: BracePair,
			inputString:   "{ a { b { c { d } } } }",
			pos:           4,
			expectMatch:   true,
			expectPos:     20,
		},
		{
			name:          "after open brace",
			delimiterPair: BracePair,
			inputString:   "{ a { b { c { d } } } }",
			pos:           5,
			expectMatch:   true,
			expectPos:     20,
		},
		{
			name:          "ignore brace in Go comment",
			delimiterPair: BracePair,
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
		{
			name:          "no match parens",
			delimiterPair: ParenPair,
			inputString:   "abcd 1234",
			pos:           2,
			expectMatch:   false,
		},
		{
			name:          "on open paren",
			delimiterPair: ParenPair,
			inputString:   "( a ( b ( c ( d ) ) ) )",
			pos:           4,
			expectMatch:   true,
			expectPos:     20,
		},
		{
			name:          "after open paren",
			delimiterPair: ParenPair,
			inputString:   "( a ( b ( c ( d ) ) ) )",
			pos:           5,
			expectMatch:   true,
			expectPos:     20,
		},
		{
			name:          "ignore paren in Go comment",
			delimiterPair: ParenPair,
			inputString: `(
abc
	(
	// )
	)
)`,
			syntaxLanguage: syntax.LanguageGo,
			pos:            2,
			expectMatch:    true,
			expectPos:      18,
		},
		{
			name:           "unmatched paren in Go string",
			delimiterPair:  ParenPair,
			inputString:    `(x == "(y + z)")`,
			syntaxLanguage: syntax.LanguageGo,
			pos:            9,
			expectMatch:    true,
			expectPos:      13,
		},
		{
			name:           "matched paren in Go string",
			delimiterPair:  ParenPair,
			inputString:    `(x == "y + (z)")`,
			syntaxLanguage: syntax.LanguageGo,
			pos:            9,
			expectMatch:    true,
			expectPos:      15,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			textTree, syntaxParser := textTreeAndSyntaxParser(t, tc.inputString, tc.syntaxLanguage)
			actualPos, ok := NextUnmatchedCloseDelimiter(tc.delimiterPair, textTree, syntaxParser, tc.pos)
			assert.Equal(t, tc.expectMatch, ok)
			if ok {
				assert.Equal(t, tc.expectPos, actualPos)
			}
		})
	}
}

func TestPrevUnmatchedOpenDelimiter(t *testing.T) {
	testCases := []struct {
		name           string
		delimiterPair  DelimiterPair
		inputString    string
		pos            uint64
		syntaxLanguage syntax.Language
		expectMatch    bool
		expectPos      uint64
	}{
		{
			name:          "empty",
			delimiterPair: BracePair,
			inputString:   "",
			pos:           0,
			expectMatch:   false,
		},
		{
			name:          "no match braces",
			delimiterPair: BracePair,
			inputString:   "abcd 1234",
			pos:           6,
			expectMatch:   false,
		},
		{
			name:          "on close brace",
			delimiterPair: BracePair,
			inputString:   "{ a { b { c { d } } } }",
			pos:           20,
			expectMatch:   true,
			expectPos:     4,
		},
		{
			name:          "after close brace",
			delimiterPair: BracePair,
			inputString:   "{ a { b { c { d } } } }",
			pos:           19,
			expectMatch:   true,
			expectPos:     4,
		},
		{
			name:          "ignore brace in Go comment",
			delimiterPair: BracePair,
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
		{
			name:          "no match parens",
			delimiterPair: ParenPair,
			inputString:   "abcd 1234",
			pos:           6,
			expectMatch:   false,
		},
		{
			name:          "on close paren",
			delimiterPair: ParenPair,
			inputString:   "( a ( b ( c ( d ) ) ) )",
			pos:           20,
			expectMatch:   true,
			expectPos:     4,
		},
		{
			name:          "after close paren",
			delimiterPair: ParenPair,
			inputString:   "( a ( b ( c ( d ) ) ) )",
			pos:           19,
			expectMatch:   true,
			expectPos:     4,
		},
		{
			name:          "ignore paren in Go comment",
			delimiterPair: ParenPair,
			inputString: `(
	(
	// (
	)
abc
)`,
			syntaxLanguage: syntax.LanguageGo,
			pos:            15,
			expectMatch:    true,
			expectPos:      0,
		},
		{
			name:           "unmatched paren in Go string",
			delimiterPair:  ParenPair,
			inputString:    `(x == "(y + z)")`,
			syntaxLanguage: syntax.LanguageGo,
			pos:            9,
			expectMatch:    true,
			expectPos:      7,
		},
		{
			name:           "matched paren in Go string",
			delimiterPair:  ParenPair,
			inputString:    `(x == "(y) + z")`,
			syntaxLanguage: syntax.LanguageGo,
			pos:            10,
			expectMatch:    true,
			expectPos:      0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			textTree, syntaxParser := textTreeAndSyntaxParser(t, tc.inputString, tc.syntaxLanguage)
			actualPos, ok := PrevUnmatchedOpenDelimiter(tc.delimiterPair, textTree, syntaxParser, tc.pos)
			assert.Equal(t, tc.expectMatch, ok)
			if ok {
				assert.Equal(t, tc.expectPos, actualPos)
			}
		})
	}
}

func TestDelimitedBlock(t *testing.T) {
	testCases := []struct {
		name              string
		inputString       string
		pos               uint64
		syntaxLanguage    syntax.Language
		delimiterPair     DelimiterPair
		includeDelimiters bool
		expectStartPos    uint64
		expectEndPos      uint64
	}{
		{
			name:           "empty",
			inputString:    "",
			pos:            0,
			delimiterPair:  ParenPair,
			expectStartPos: 0,
			expectEndPos:   0,
		},
		{
			name:           "start of unmatched start paren",
			inputString:    "(abc",
			pos:            0,
			delimiterPair:  ParenPair,
			expectStartPos: 0,
			expectEndPos:   0,
		},
		{
			name:           "after unmatched start paren",
			inputString:    "(abc",
			pos:            2,
			delimiterPair:  ParenPair,
			expectStartPos: 2,
			expectEndPos:   2,
		},
		{
			name:           "end of unmatched end paren",
			inputString:    "a)bc",
			pos:            1,
			delimiterPair:  ParenPair,
			expectStartPos: 1,
			expectEndPos:   1,
		},
		{
			name:           "before unmatched end paren",
			inputString:    "a)bc",
			pos:            0,
			delimiterPair:  ParenPair,
			expectStartPos: 0,
			expectEndPos:   0,
		},
		{
			name:           "start of matched paren, no content",
			inputString:    "()",
			pos:            0,
			delimiterPair:  ParenPair,
			expectStartPos: 1,
			expectEndPos:   1,
		},
		{
			name:           "end of matched paren, no content",
			inputString:    "()",
			pos:            1,
			delimiterPair:  ParenPair,
			expectStartPos: 1,
			expectEndPos:   1,
		},
		{
			name:           "start of matched paren, content",
			inputString:    "(abc)",
			pos:            0,
			delimiterPair:  ParenPair,
			expectStartPos: 1,
			expectEndPos:   4,
		},
		{
			name:           "after start paren, content",
			inputString:    "(abc)",
			pos:            1,
			delimiterPair:  ParenPair,
			expectStartPos: 1,
			expectEndPos:   4,
		},
		{
			name:           "before end paren, content",
			inputString:    "(abc)",
			pos:            3,
			delimiterPair:  ParenPair,
			expectStartPos: 1,
			expectEndPos:   4,
		},
		{
			name:           "end of matched paren, content",
			inputString:    "(abc)",
			pos:            4,
			delimiterPair:  ParenPair,
			expectStartPos: 1,
			expectEndPos:   4,
		},
		{
			name:           "before nested paren",
			inputString:    "(a(b)c)",
			pos:            1,
			delimiterPair:  ParenPair,
			expectStartPos: 1,
			expectEndPos:   6,
		},
		{
			name:           "on nested paren",
			inputString:    "(a(b)c)",
			pos:            2,
			delimiterPair:  ParenPair,
			expectStartPos: 3,
			expectEndPos:   4,
		},
		{
			name:              "include parens",
			inputString:       "x (abc) y",
			pos:               4,
			delimiterPair:     ParenPair,
			includeDelimiters: true,
			expectStartPos:    2,
			expectEndPos:      7,
		},
		{
			name:           "parens within Go string",
			inputString:    `(x == "(y + z)")`,
			syntaxLanguage: syntax.LanguageGo,
			pos:            10,
			delimiterPair:  ParenPair,
			expectStartPos: 8,
			expectEndPos:   13,
		},
		{
			name:           "unmatched parens within Go string",
			inputString:    `(x == "(y + z")`,
			syntaxLanguage: syntax.LanguageGo,
			pos:            10,
			delimiterPair:  ParenPair,
			expectStartPos: 1,
			expectEndPos:   14,
		},
		{
			name: "end of nested braces",
			inputString: `
{
	foo: 123,
	bar: {
		baz: {
			bat: 456,
		},
	},
}`,
			pos:            53,
			delimiterPair:  BracePair,
			expectStartPos: 2,
			expectEndPos:   53,
		},
		{
			name:           "end of nested braces, within Go string",
			inputString:    `{{foo: "{a{b}c}", bar: "}"}}`,
			syntaxLanguage: syntax.LanguageGo,
			pos:            13,
			delimiterPair:  BracePair,
			expectStartPos: 9,
			expectEndPos:   14,
		},
		{
			name:           "end of nested braces, outside Go string",
			inputString:    `{{foo: "{a{b}c}", bar: "{"}}`,
			syntaxLanguage: syntax.LanguageGo,
			pos:            26,
			delimiterPair:  BracePair,
			expectStartPos: 2,
			expectEndPos:   26,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			textTree, syntaxParser := textTreeAndSyntaxParser(t, tc.inputString, tc.syntaxLanguage)
			actualStartPos, actualEndPos := DelimitedBlock(tc.delimiterPair, textTree, syntaxParser, tc.includeDelimiters, tc.pos)
			assert.Equal(t, tc.expectStartPos, actualStartPos)
			assert.Equal(t, tc.expectEndPos, actualEndPos)
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
