package rules

import (
	"strings"

	"github.com/aretext/aretext/syntax/parser"
)

var GolangRules []parser.TokenizerRule

// Based on the "Go Programming Language Specification"
// https://golang.org/ref/spec
func init() {
	decimalDigits := `[0-9](_?[0-9])*`
	binaryDigits := `[01](_?[01])*`
	octalDigits := `[0-7](_?[0-7])*`
	hexDigits := `[0-9A-Fa-f](_?[0-9A-Fa-f])*`

	decimalLiteral := `0|([1-9](_?(` + decimalDigits + `))?)`
	binaryLiteral := `0[bB]_?` + binaryDigits
	octalLiteral := `0[oO]?_?` + octalDigits
	hexLiteral := `0[xX]_?` + hexDigits
	intLiteral := strings.Join([]string{
		`(` + decimalLiteral + `)`,
		`(` + binaryLiteral + `)`,
		`(` + octalLiteral + `)`,
		`(` + hexLiteral + `)`,
	}, "|")

	decimalExponent := `[eE](\+|-)?` + decimalDigits
	decimalFloatLiteral := strings.Join([]string{
		decimalDigits + `\.(` + decimalDigits + `)?` + `(` + decimalExponent + `)?`,
		decimalDigits + decimalExponent,
		`\.` + decimalDigits + `(` + decimalExponent + `)?`,
	}, "|")
	hexExponent := `[pP](\+|-)?` + decimalDigits
	hexMantissa := strings.Join([]string{
		`_?` + hexDigits + `\.(` + hexDigits + `)?`,
		`_?` + hexDigits,
		`\.` + hexDigits,
	}, "|")
	hexFloatLiteral := `0[xX]` + `(` + hexMantissa + `)` + hexExponent
	floatLiteral := strings.Join([]string{
		`(` + decimalFloatLiteral + `)`,
		`(` + hexFloatLiteral + `)`,
	}, "|")

	imaginaryLiteral := `(` + decimalDigits + `|` + intLiteral + `|` + floatLiteral + `)i`

	GolangRules = []parser.TokenizerRule{
		// Line Comment
		{
			Regexp:    `//[^\n]*`,
			TokenRole: parser.TokenRoleComment,
		},

		// General comment
		// https://blog.ostermiller.org/finding-comments-in-source-code-using-regular-expressions/
		{
			Regexp:    `/\*([^*]|(\*+[^*/]))*\*+/`,
			TokenRole: parser.TokenRoleComment,
		},

		// Keywords
		{
			Regexp: strings.Join([]string{
				"break",
				"default",
				"func",
				"interface",
				"select",
				"case",
				"defer",
				"go",
				"map",
				"struct",
				"chan",
				"else",
				"goto",
				"package",
				"switch",
				"const",
				"fallthrough",
				"if",
				"range",
				"type",
				"continue",
				"for",
				"import",
				"return",
				"var",
			}, "|"),
			TokenRole: parser.TokenRoleKeyword,
		},

		// Predeclared identifiers
		{
			Regexp: strings.Join([]string{
				"bool",
				"byte",
				"complex64",
				"complex128",
				"error",
				"float32",
				"float64",
				"int",
				"int8",
				"int16",
				"int32",
				"int64",
				"rune",
				"string",
				"uint",
				"uint8",
				"uint16",
				"uint32",
				"uint64",
				"uintptr",
				"true",
				"false",
				"iota",
				"nil",
				"append",
				"cap",
				"close",
				"complex",
				"copy",
				"delete",
				"imag",
				"len",
				"make",
				"new",
				"panic",
				"print",
				"println",
				"real",
				"recover",
			}, "|"),
			TokenRole: parser.TokenRoleKeyword,
		},

		// Identifiers
		{
			Regexp:    `(\p{L}|_)((\p{L}|_)|\p{Nd})*`,
			TokenRole: parser.TokenRoleIdentifier,
		},

		// Operators
		{
			Regexp: strings.Join([]string{
				`\+`, `&`, `\+=`, `&=`, `&&`, `==`, `!=`,
				`-`, `\|`, `-=`, `\|=`, `\|\|`, `<`, `<=`,
				`\*`, `\^`, `\*=`, `\^=`, `<-`, `>`, `>=`,
				`/`, `<<`, `/=`, `<<=`, `\+\+`, `=`, `:=`,
				`%`, `>>`, `%=`, `>>=`, `--`, `!`,
				`&\^`, `&\^=`,
			}, "|"),
			TokenRole: parser.TokenRoleOperator,
		},

		// Punctuation
		{
			Regexp: strings.Join([]string{
				`\(`, `\)`, `\[`, `\]`, `\{`, `\}`,
				`,`, `;`, `\.\.\.`, `\.`, `:`,
			}, "|"),
			TokenRole: parser.TokenRolePunctuation,
		},

		// Integer literal
		{
			Regexp:    intLiteral,
			TokenRole: parser.TokenRoleNumber,
		},

		// Floating point literal
		{
			Regexp:    floatLiteral,
			TokenRole: parser.TokenRoleNumber,
		},

		// Imaginary literal
		{
			Regexp:    imaginaryLiteral,
			TokenRole: parser.TokenRoleNumber,
		},

		// Rune literal
		{
			Regexp:    "'[^']*'",
			TokenRole: parser.TokenRoleString,
		},

		// Raw string literal
		{
			Regexp:    "`[^`]*`",
			TokenRole: parser.TokenRoleString,
		},

		// Interpreted string literal
		{
			Regexp:    `"([^\"\n]|\\")*"`,
			TokenRole: parser.TokenRoleString,
		},
	}
}
