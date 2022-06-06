package languages

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/aretext/aretext/syntax/parser"
)

func TestCParseFunc(t *testing.T) {
	testCases := []struct {
		name     string
		text     string
		expected []TokenWithText
	}{
		{
			name: "hello world",
			text: `
#include <stdio.h>

main()
{
    printf("hello, world\n");
}`,
			expected: []TokenWithText{
				{Text: "#include <stdio.h>\n", Role: cTokenRolePreprocessorDirective},
				{Text: "main", Role: parser.TokenRoleIdentifier},
				{Text: "printf", Role: parser.TokenRoleIdentifier},
				{Text: `"hello, world\n"`, Role: parser.TokenRoleString},
			},
		},
		{
			name: "conditionals and math",
			text: `if (x % 2 == 0) { return x + 10 } else { return x / 2 }`,
			expected: []TokenWithText{
				{Text: "if", Role: parser.TokenRoleKeyword},
				{Text: "x", Role: parser.TokenRoleIdentifier},
				{Text: "%", Role: parser.TokenRoleOperator},
				{Text: "2", Role: parser.TokenRoleNumber},
				{Text: "==", Role: parser.TokenRoleOperator},
				{Text: "0", Role: parser.TokenRoleNumber},
				{Text: "return", Role: parser.TokenRoleKeyword},
				{Text: "x", Role: parser.TokenRoleIdentifier},
				{Text: "+", Role: parser.TokenRoleOperator},
				{Text: "10", Role: parser.TokenRoleNumber},
				{Text: "else", Role: parser.TokenRoleKeyword},
				{Text: "return", Role: parser.TokenRoleKeyword},
				{Text: "x", Role: parser.TokenRoleIdentifier},
				{Text: "/", Role: parser.TokenRoleOperator},
				{Text: "2", Role: parser.TokenRoleNumber},
			},
		},
		{
			name: "header guard with comments",
			text: `
#ifndef _FILE_NAME_H_
#define _FILE_NAME_H_

/* code */

#endif // #ifndef _FILE_NAME_H_`,
			expected: []TokenWithText{
				{Text: "#ifndef _FILE_NAME_H_\n", Role: cTokenRolePreprocessorDirective},
				{Text: "#define _FILE_NAME_H_\n", Role: cTokenRolePreprocessorDirective},
				{Text: "/* code */", Role: parser.TokenRoleComment},
				{Text: "#endif // #ifndef _FILE_NAME_H_", Role: cTokenRolePreprocessorDirective},
			},
		},
		{
			name: "macro with continuation",
			text: `
#define  printboth(a, b)  \
   printf(#a " and " #b)
// comment`,
			expected: []TokenWithText{
				{Text: "#define  printboth(a, b)  \\\n   printf(#a \" and \" #b)\n", Role: cTokenRolePreprocessorDirective},
				{Text: `// comment`, Role: parser.TokenRoleComment},
			},
		},
		{
			name: "bit twiddling hacks",
			text: `r = x ^ ((x ^ y) & -(x < y)); // max(x, y)`,
			expected: []TokenWithText{
				{Text: "r", Role: parser.TokenRoleIdentifier},
				{Text: "=", Role: parser.TokenRoleOperator},
				{Text: "x", Role: parser.TokenRoleIdentifier},
				{Text: "^", Role: parser.TokenRoleOperator},
				{Text: "x", Role: parser.TokenRoleIdentifier},
				{Text: "^", Role: parser.TokenRoleOperator},
				{Text: "y", Role: parser.TokenRoleIdentifier},
				{Text: "&", Role: parser.TokenRoleOperator},
				{Text: "-", Role: parser.TokenRoleOperator},
				{Text: "x", Role: parser.TokenRoleIdentifier},
				{Text: "<", Role: parser.TokenRoleOperator},
				{Text: "y", Role: parser.TokenRoleIdentifier},
				{Text: "// max(x, y)", Role: parser.TokenRoleComment},
			},
		},
		{
			name: "integer and char declaration",
			text: `
int foo;
unsigned int bar = 42;
char quux = 'a';`,
			expected: []TokenWithText{
				{Text: "int", Role: parser.TokenRoleKeyword},
				{Text: "foo", Role: parser.TokenRoleIdentifier},
				{Text: "unsigned", Role: parser.TokenRoleKeyword},
				{Text: "int", Role: parser.TokenRoleKeyword},
				{Text: "bar", Role: parser.TokenRoleIdentifier},
				{Text: "=", Role: parser.TokenRoleOperator},
				{Text: "42", Role: parser.TokenRoleNumber},
				{Text: "char", Role: parser.TokenRoleKeyword},
				{Text: "quux", Role: parser.TokenRoleIdentifier},
				{Text: "=", Role: parser.TokenRoleOperator},
				{Text: "'a'", Role: parser.TokenRoleString},
			},
		},
		{
			name: "real number declaration",
			text: `
float foo;
double bar = 114.3943;`,
			expected: []TokenWithText{
				{Text: "float", Role: parser.TokenRoleKeyword},
				{Text: "foo", Role: parser.TokenRoleIdentifier},
				{Text: "double", Role: parser.TokenRoleKeyword},
				{Text: "bar", Role: parser.TokenRoleIdentifier},
				{Text: "=", Role: parser.TokenRoleOperator},
				{Text: "114.3943", Role: parser.TokenRoleNumber},
			},
		},
		{
			name: "hex number with suffix",
			text: "0xABCULL",
			expected: []TokenWithText{
				{Text: "0xABCULL", Role: parser.TokenRoleNumber},
			},
		},
		{
			name: "float with exponent",
			text: `
double x, y;

x = 5e2;
y = 5e-2;
`,
			expected: []TokenWithText{
				{Text: "double", Role: parser.TokenRoleKeyword},
				{Text: "x", Role: parser.TokenRoleIdentifier},
				{Text: "y", Role: parser.TokenRoleIdentifier},
				{Text: "x", Role: parser.TokenRoleIdentifier},
				{Text: "=", Role: parser.TokenRoleOperator},
				{Text: "5e2", Role: parser.TokenRoleNumber},
				{Text: "y", Role: parser.TokenRoleIdentifier},
				{Text: "=", Role: parser.TokenRoleOperator},
				{Text: "5e-2", Role: parser.TokenRoleNumber},
			},
		},
		{
			name: "unrecognized preprocessor directive",
			text: "#endifaaa",
			expected: []TokenWithText{
				{Text: "endifaaa", Role: parser.TokenRoleIdentifier},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tokens := ParseTokensWithText(CParseFunc(), tc.text)
			assert.Equal(t, tc.expected, tokens)
		})
	}
}

func BenchmarkCParser(b *testing.B) {
	ParserBenchmark(CParseFunc(), "testdata/c/hello.c")(b)
}
