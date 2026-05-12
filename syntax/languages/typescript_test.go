package languages

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/aretext/aretext/syntax/parser"
)

func TestTypescriptParseFunc(t *testing.T) {
	testCases := []struct {
		name     string
		text     string
		expected []TokenWithText
	}{
		{
			name: "line comment",
			text: "// comment",
			expected: []TokenWithText{
				{Text: "// comment", Role: parser.TokenRoleComment},
			},
		},
		{
			name: "empty line comment at end of file",
			text: "//",
			expected: []TokenWithText{
				{Text: "//", Role: parser.TokenRoleComment},
			},
		},
		{
			name: "block comment",
			text: "/* comment */",
			expected: []TokenWithText{
				{Text: "/* comment */", Role: parser.TokenRoleComment},
			},
		},
		{
			name: "multi-line block comment",
			text: "/* first\nsecond */",
			expected: []TokenWithText{
				{Text: "/* first\nsecond */", Role: parser.TokenRoleComment},
			},
		},
		{
			name: "jsdoc comment",
			text: "/** @param value */",
			expected: []TokenWithText{
				{Text: "/** @param value */", Role: parser.TokenRoleComment},
			},
		},
		{
			name: "single quote string",
			text: "'hello'",
			expected: []TokenWithText{
				{Text: "'hello'", Role: parser.TokenRoleString},
			},
		},
		{
			name: "double quote string",
			text: `"hello"`,
			expected: []TokenWithText{
				{Text: `"hello"`, Role: parser.TokenRoleString},
			},
		},
		{
			name: "string with escaped quote",
			text: `"say \"hello\""`,
			expected: []TokenWithText{
				{Text: `"say \"hello\""`, Role: parser.TokenRoleString},
			},
		},
		{
			name: "template string without interpolation",
			text: "`hello\nworld`",
			expected: []TokenWithText{
				{Text: "`hello\nworld`", Role: parser.TokenRoleString},
			},
		},
		{
			name: "unterminated single quote string stops at newline",
			text: "'hello\nconst x = 1",
			expected: []TokenWithText{
				{Text: "const", Role: parser.TokenRoleKeyword},
				{Text: "=", Role: parser.TokenRoleOperator},
				{Text: "1", Role: parser.TokenRoleNumber},
			},
		},
		{
			name: "decimal integer",
			text: "12345",
			expected: []TokenWithText{
				{Text: "12345", Role: parser.TokenRoleNumber},
			},
		},
		{
			name: "decimal integer with separators",
			text: "1_000_000",
			expected: []TokenWithText{
				{Text: "1_000_000", Role: parser.TokenRoleNumber},
			},
		},
		{
			name: "binary integer",
			text: "0b1010_0110",
			expected: []TokenWithText{
				{Text: "0b1010_0110", Role: parser.TokenRoleNumber},
			},
		},
		{
			name: "octal integer",
			text: "0o755",
			expected: []TokenWithText{
				{Text: "0o755", Role: parser.TokenRoleNumber},
			},
		},
		{
			name: "hex integer",
			text: "0xdead_beef",
			expected: []TokenWithText{
				{Text: "0xdead_beef", Role: parser.TokenRoleNumber},
			},
		},
		{
			name: "bigint literal",
			text: "123n",
			expected: []TokenWithText{
				{Text: "123n", Role: parser.TokenRoleNumber},
			},
		},
		{
			name: "float literal",
			text: "3.14159",
			expected: []TokenWithText{
				{Text: "3.14159", Role: parser.TokenRoleNumber},
			},
		},
		{
			name: "float with exponent",
			text: "6.022e23",
			expected: []TokenWithText{
				{Text: "6.022e23", Role: parser.TokenRoleNumber},
			},
		},
		{
			name: "float with negative exponent",
			text: "1.5e-10",
			expected: []TokenWithText{
				{Text: "1.5e-10", Role: parser.TokenRoleNumber},
			},
		},
		{
			name: "keywords in class declaration",
			text: "export abstract class Child extends Parent implements Interface {}",
			expected: []TokenWithText{
				{Text: "export", Role: parser.TokenRoleKeyword},
				{Text: "abstract", Role: parser.TokenRoleKeyword},
				{Text: "class", Role: parser.TokenRoleKeyword},
				{Text: "extends", Role: parser.TokenRoleKeyword},
				{Text: "implements", Role: parser.TokenRoleKeyword},
			},
		},
		{
			name: "keywords in interface declaration",
			text: "interface KeyValuePair<K, V> extends Array<K | V> { readonly key: K }",
			expected: []TokenWithText{
				{Text: "interface", Role: parser.TokenRoleKeyword},
				{Text: "<", Role: parser.TokenRoleOperator},
				{Text: ">", Role: parser.TokenRoleOperator},
				{Text: "extends", Role: parser.TokenRoleKeyword},
				{Text: "<", Role: parser.TokenRoleOperator},
				{Text: "|", Role: parser.TokenRoleOperator},
				{Text: ">", Role: parser.TokenRoleOperator},
				{Text: "readonly", Role: parser.TokenRoleKeyword},
				{Text: ":", Role: parser.TokenRoleOperator},
			},
		},
		{
			name: "variable declarations and primitive type keywords",
			text: `const answer: number = 42; let label: string = "ok"; var done: boolean = false;`,
			expected: []TokenWithText{
				{Text: "const", Role: parser.TokenRoleKeyword},
				{Text: ":", Role: parser.TokenRoleOperator},
				{Text: "number", Role: parser.TokenRoleKeyword},
				{Text: "=", Role: parser.TokenRoleOperator},
				{Text: "42", Role: parser.TokenRoleNumber},
				{Text: "let", Role: parser.TokenRoleKeyword},
				{Text: ":", Role: parser.TokenRoleOperator},
				{Text: "string", Role: parser.TokenRoleKeyword},
				{Text: "=", Role: parser.TokenRoleOperator},
				{Text: `"ok"`, Role: parser.TokenRoleString},
				{Text: "var", Role: parser.TokenRoleKeyword},
				{Text: ":", Role: parser.TokenRoleOperator},
				{Text: "boolean", Role: parser.TokenRoleKeyword},
				{Text: "=", Role: parser.TokenRoleOperator},
				{Text: "false", Role: parser.TokenRoleKeyword},
			},
		},
		{
			name: "function with return type",
			text: "async function fetchValue(id: string): Promise<number> { return await get(id); }",
			expected: []TokenWithText{
				{Text: "async", Role: parser.TokenRoleKeyword},
				{Text: "function", Role: parser.TokenRoleKeyword},
				{Text: ":", Role: parser.TokenRoleOperator},
				{Text: "string", Role: parser.TokenRoleKeyword},
				{Text: ":", Role: parser.TokenRoleOperator},
				{Text: "<", Role: parser.TokenRoleOperator},
				{Text: "number", Role: parser.TokenRoleKeyword},
				{Text: ">", Role: parser.TokenRoleOperator},
				{Text: "return", Role: parser.TokenRoleKeyword},
				{Text: "await", Role: parser.TokenRoleKeyword},
			},
		},
		{
			name: "arrow function operators",
			text: "const inc = (value: number): number => value + 1",
			expected: []TokenWithText{
				{Text: "const", Role: parser.TokenRoleKeyword},
				{Text: "=", Role: parser.TokenRoleOperator},
				{Text: ":", Role: parser.TokenRoleOperator},
				{Text: "number", Role: parser.TokenRoleKeyword},
				{Text: ":", Role: parser.TokenRoleOperator},
				{Text: "number", Role: parser.TokenRoleKeyword},
				{Text: "=>", Role: parser.TokenRoleOperator},
				{Text: "+", Role: parser.TokenRoleOperator},
				{Text: "1", Role: parser.TokenRoleNumber},
			},
		},
		{
			name: "compound operators",
			text: "a += b ?? c?.d !== e && f || g ** h",
			expected: []TokenWithText{
				{Text: "+=", Role: parser.TokenRoleOperator},
				{Text: "??", Role: parser.TokenRoleOperator},
				{Text: "?.", Role: parser.TokenRoleOperator},
				{Text: "!==", Role: parser.TokenRoleOperator},
				{Text: "&&", Role: parser.TokenRoleOperator},
				{Text: "||", Role: parser.TokenRoleOperator},
				{Text: "**", Role: parser.TokenRoleOperator},
			},
		},
		{
			name: "type aliases and unions",
			text: "type Result<T> = T | Error | null | undefined",
			expected: []TokenWithText{
				{Text: "type", Role: parser.TokenRoleKeyword},
				{Text: "<", Role: parser.TokenRoleOperator},
				{Text: ">", Role: parser.TokenRoleOperator},
				{Text: "=", Role: parser.TokenRoleOperator},
				{Text: "|", Role: parser.TokenRoleOperator},
				{Text: "|", Role: parser.TokenRoleOperator},
				{Text: "null", Role: parser.TokenRoleKeyword},
				{Text: "|", Role: parser.TokenRoleOperator},
				{Text: "undefined", Role: parser.TokenRoleKeyword},
			},
		},
		{
			name: "modifier operators in mapped types",
			text: "type Mutable<T> = { -readonly [P in keyof T]+?: T[P] }",
			expected: []TokenWithText{
				{Text: "type", Role: parser.TokenRoleKeyword},
				{Text: "<", Role: parser.TokenRoleOperator},
				{Text: ">", Role: parser.TokenRoleOperator},
				{Text: "=", Role: parser.TokenRoleOperator},
				{Text: "-", Role: parser.TokenRoleOperator},
				{Text: "readonly", Role: parser.TokenRoleKeyword},
				{Text: "in", Role: parser.TokenRoleKeyword},
				{Text: "keyof", Role: parser.TokenRoleKeyword},
				{Text: "+?", Role: parser.TokenRoleOperator},
				{Text: ":", Role: parser.TokenRoleOperator},
			},
		},
		{
			name: "comments strings and operators together",
			text: `const url = "https://example.com"; // keep scheme slash in string`,
			expected: []TokenWithText{
				{Text: "const", Role: parser.TokenRoleKeyword},
				{Text: "=", Role: parser.TokenRoleOperator},
				{Text: `"https://example.com"`, Role: parser.TokenRoleString},
				{Text: "// keep scheme slash in string", Role: parser.TokenRoleComment},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tokens := ParseTokensWithText(TypescriptParseFunc(), tc.text)
			assert.Equal(t, tc.expected, tokens)
		})
	}
}

func FuzzTypescriptParseFunc(f *testing.F) {
	seeds := LoadFuzzTestSeeds(f, "./testdata/typescript/*")
	ParserFuzzTest(f, TypescriptParseFunc(), seeds)
}
