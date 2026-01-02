package languages

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/aretext/aretext/syntax/parser"
)

func TestSQLParseFunc(t *testing.T) {
	testCases := []struct {
		name     string
		text     string
		expected []TokenWithText
	}{
		{
			name: "single line comment",
			text: "-- comment",
			expected: []TokenWithText{
				{Text: "-- comment", Role: parser.TokenRoleComment},
			},
		},
		{
			name: "block comment",
			text: "/* abcd\n123 */",
			expected: []TokenWithText{
				{Text: "/* abcd\n123 */", Role: parser.TokenRoleComment},
			},
		},
		{
			name: "identifier followed by comment",
			text: "foobar  --comment",
			expected: []TokenWithText{
				{Text: `--comment`, Role: parser.TokenRoleComment},
			},
		},
		{
			name: "comment followed by identifier on next line",
			text: "foo\n-- comment\nbar",
			expected: []TokenWithText{
				{Text: "-- comment\n", Role: parser.TokenRoleComment},
			},
		},
		{
			name: "operators in equation",
			text: "x + y - z",
			expected: []TokenWithText{
				{Text: "+", Role: parser.TokenRoleOperator},
				{Text: "-", Role: parser.TokenRoleOperator},
			},
		},
		{
			name: "test multiline string",
			text: "'foo\nbar'",
			expected: []TokenWithText{
				{Text: "'foo\nbar'", Role: parser.TokenRoleString},
			},
		},
		{
			name: "decimal int literal, single digit",
			text: "7",
			expected: []TokenWithText{
				{Text: "7", Role: parser.TokenRoleNumber},
			},
		},
		{
			name: "decimal int literal, multiple digits",
			text: "2147483647",
			expected: []TokenWithText{
				{Text: "2147483647", Role: parser.TokenRoleNumber},
			},
		},
		{
			name: "hex literal",
			text: "0xdeadbeef",
			expected: []TokenWithText{
				{Text: "0xdeadbeef", Role: parser.TokenRoleNumber},
			},
		},
		{
			name: "float literal, point between digits",
			text: "3.14",
			expected: []TokenWithText{
				{Text: "3.14", Role: parser.TokenRoleNumber},
			},
		},
		{
			name: "float literal, point after digits",
			text: "10.",
			expected: []TokenWithText{
				{Text: "10.", Role: parser.TokenRoleNumber},
			},
		},
		{
			name: "float literal, point before digits",
			text: ".001",
			expected: []TokenWithText{
				{Text: ".001", Role: parser.TokenRoleNumber},
			},
		},
		{
			name: "float literal, exponent without point",
			text: "1e100",
			expected: []TokenWithText{
				{Text: "1e100", Role: parser.TokenRoleNumber},
			},
		},
		{
			name: "float literal, negative exponent",
			text: "3.14e-10",
			expected: []TokenWithText{
				{Text: "3.14e-10", Role: parser.TokenRoleNumber},
			},
		},
		{
			name: "float literal, zero digit and zero exponent",
			text: "0e0",
			expected: []TokenWithText{
				{Text: "0e0", Role: parser.TokenRoleNumber},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tokens := ParseTokensWithText(SQLParseFunc(), tc.text)
			assert.Equal(t, tc.expected, tokens)
		})
	}
}

func BenchmarkSQLParser(b *testing.B) {
	ParserBenchmark(b, SQLParseFunc(), "testdata/sql/example.sql")
}

func FuzzSQLParseFunc(f *testing.F) {
	seeds := LoadFuzzTestSeeds(f, "./testdata/sql/*")
	ParserFuzzTest(f, SQLParseFunc(), seeds)
}
