package languages

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/aretext/aretext/syntax/parser"
)

func TestGolangParseFunc(t *testing.T) {
	testCases := []struct {
		name     string
		text     string
		expected []TokenWithText
	}{
		{
			name: "line comment",
			text: "// comment",
			expected: []TokenWithText{
				{Text: `// comment`, Role: parser.TokenRoleComment},
			},
		},
		{
			name: "empty comment at end of file",
			text: "//",
			expected: []TokenWithText{
				{Text: "//", Role: parser.TokenRoleComment},
			},
		},
		{
			name: "general comment",
			text: "/* abcd\n123 */",
			expected: []TokenWithText{
				{Text: "/* abcd\n123 */", Role: parser.TokenRoleComment},
			},
		},
		{
			name: "variable declaration",
			text: `var foo []int`,
			expected: []TokenWithText{
				{Text: "var", Role: parser.TokenRoleKeyword},
				{Text: "foo", Role: parser.TokenRoleIdentifier},
				{Text: "int", Role: parser.TokenRoleKeyword},
			},
		},
		{
			name: "operators",
			text: "a + b / c",
			expected: []TokenWithText{
				{Text: "a", Role: parser.TokenRoleIdentifier},
				{Text: "+", Role: parser.TokenRoleOperator},
				{Text: "b", Role: parser.TokenRoleIdentifier},
				{Text: "/", Role: parser.TokenRoleOperator},
				{Text: "c", Role: parser.TokenRoleIdentifier},
			},
		},
		{
			name: "raw string",
			text: "`abcd\n123`",
			expected: []TokenWithText{
				{Text: "`abcd\n123`", Role: parser.TokenRoleString},
			},
		},
		{
			name: "interpreted string",
			text: `"abcd"`,
			expected: []TokenWithText{
				{Text: `"abcd"`, Role: parser.TokenRoleString},
			},
		},
		{
			name: "interpreted string empty",
			text: `""`,
			expected: []TokenWithText{
				{Text: `""`, Role: parser.TokenRoleString},
			},
		},
		{
			name: "interpreted string with escaped quote",
			text: `"ab\"cd"`,
			expected: []TokenWithText{
				{Text: `"ab\"cd"`, Role: parser.TokenRoleString},
			},
		},
		{
			name: "interpreted string ending with escaped backslash",
			text: `"abc\\"`,
			expected: []TokenWithText{
				{Text: `"abc\\"`, Role: parser.TokenRoleString},
			},
		},
		{
			name: "incomplete interpreted string ending with escaped quote",
			text: `"abc\" 123`,
			expected: []TokenWithText{
				{Text: `abc`, Role: parser.TokenRoleIdentifier},
				{Text: `123`, Role: parser.TokenRoleNumber},
			},
		},
		{
			name: "incomplete interpreted string with newline before quote",
			text: "\"abc\n\"",
			expected: []TokenWithText{
				{Text: `abc`, Role: parser.TokenRoleIdentifier},
			},
		},
		{
			name: "rune",
			text: `'a'`,
			expected: []TokenWithText{
				{Text: "'a'", Role: parser.TokenRoleString},
			},
		},
		{
			name: "rune with escaped newline",
			text: `'\n'`,
			expected: []TokenWithText{
				{Text: `'\n'`, Role: parser.TokenRoleString},
			},
		},
		{
			name: "rune with escaped quote",
			text: `'\''`,
			expected: []TokenWithText{
				{Text: `'\''`, Role: parser.TokenRoleString},
			},
		},
		{
			name: "rune with escaped backslash",
			text: `'\\'`,
			expected: []TokenWithText{
				{Text: `'\\'`, Role: parser.TokenRoleString},
			},
		},
		{
			name:     "incomplete rune with newline before quote",
			text:     "'\n'",
			expected: []TokenWithText{},
		},
		{
			name: "identifier with underscore prefix",
			text: "_x9",
			expected: []TokenWithText{
				{Text: `_x9`, Role: parser.TokenRoleIdentifier},
			},
		},
		{
			name: "identifier with mixed case",
			text: "ThisVariableIsExported",
			expected: []TokenWithText{
				{Text: `ThisVariableIsExported`, Role: parser.TokenRoleIdentifier},
			},
		},
		{
			name: "identifier with non-ascii unicode",
			text: "αβ",
			expected: []TokenWithText{
				{Text: "αβ", Role: parser.TokenRoleIdentifier},
			},
		},
		{
			name: "integer with underscore",
			text: "4_2",
			expected: []TokenWithText{
				{Text: `4_2`, Role: parser.TokenRoleNumber},
			},
		},
		{
			name: "octal without o separator",
			text: "0600",
			expected: []TokenWithText{
				{Text: `0600`, Role: parser.TokenRoleNumber},
			},
		},
		{
			name: "integer with leading zero and underscore",
			text: "0_600",
			expected: []TokenWithText{
				{Text: `0_600`, Role: parser.TokenRoleNumber},
			},
		},
		{
			name: "octal with lowercase o",
			text: "0o600",
			expected: []TokenWithText{
				{Text: `0o600`, Role: parser.TokenRoleNumber},
			},
		},
		{
			name: "octal with uppercase O",
			text: "0O600",
			expected: []TokenWithText{
				{Text: `0O600`, Role: parser.TokenRoleNumber},
			},
		},
		{
			name: "octal denoted by leading zero",
			text: "0123",
			expected: []TokenWithText{
				{Text: "0123", Role: parser.TokenRoleNumber},
			},
		},
		{
			name: "hex with lowercase x",
			text: "0xBadFace",
			expected: []TokenWithText{
				{Text: `0xBadFace`, Role: parser.TokenRoleNumber},
			},
		},
		{
			name: "hex with lowercase x and underscore",
			text: "0xBad_Face",
			expected: []TokenWithText{
				{Text: `0xBad_Face`, Role: parser.TokenRoleNumber},
			},
		},
		{
			name: "hex with leading underscore",
			text: "0x_67_7a_2f_cc_40_c6",
			expected: []TokenWithText{
				{Text: `0x_67_7a_2f_cc_40_c6`, Role: parser.TokenRoleNumber},
			},
		},
		{
			name: "long number no underscores",
			text: "170141183460469231731687303715884105727",
			expected: []TokenWithText{
				{Text: `170141183460469231731687303715884105727`, Role: parser.TokenRoleNumber},
			},
		},
		{
			name: "long number with underscores",
			text: "170_141183_460469_231731_687303_715884_105727",
			expected: []TokenWithText{
				{Text: `170_141183_460469_231731_687303_715884_105727`, Role: parser.TokenRoleNumber},
			},
		},
		{
			name: "identifier with leading underscore and digits",
			text: "_42",
			expected: []TokenWithText{
				{Text: `_42`, Role: parser.TokenRoleIdentifier},
			},
		},
		{
			name: "invalid number with digits and trailing underscore",
			text: "42_",
			expected: []TokenWithText{
				{Text: `42`, Role: parser.TokenRoleNumber},
				{Text: `_`, Role: parser.TokenRoleIdentifier},
			},
		},
		{
			name: "invalid number with multiple underscores",
			text: "4__2",
			expected: []TokenWithText{
				{Text: `4`, Role: parser.TokenRoleNumber},
				{Text: `__2`, Role: parser.TokenRoleIdentifier},
			},
		},
		{
			name: "invalid hex with leading underscore",
			text: "0_xBadFace",
			expected: []TokenWithText{
				{Text: `0`, Role: parser.TokenRoleNumber},
				{Text: `_xBadFace`, Role: parser.TokenRoleIdentifier},
			},
		},
		{
			name: "floating point zero with decimal",
			text: "0.",
			expected: []TokenWithText{
				{Text: `0.`, Role: parser.TokenRoleNumber},
			},
		},
		{
			name: "floating point with decimal",
			text: "72.40",
			expected: []TokenWithText{
				{Text: `72.40`, Role: parser.TokenRoleNumber},
			},
		},
		{
			name: "floating point with decimal and leading zero",
			text: "072.40",
			expected: []TokenWithText{
				{Text: `072.40`, Role: parser.TokenRoleNumber},
			},
		},
		{
			name: "floating point with single leading decimal",
			text: "2.71828",
			expected: []TokenWithText{
				{Text: `2.71828`, Role: parser.TokenRoleNumber},
			},
		},
		{
			name: "floating point with exponent zero",
			text: "1.e+0",
			expected: []TokenWithText{
				{Text: `1.e+0`, Role: parser.TokenRoleNumber},
			},
		},
		{
			name: "floating point with negative exponent",
			text: "6.67428e-11",
			expected: []TokenWithText{
				{Text: `6.67428e-11`, Role: parser.TokenRoleNumber},
			},
		},
		{
			name: "exponent with uppercase E",
			text: "1E6",
			expected: []TokenWithText{
				{Text: `1E6`, Role: parser.TokenRoleNumber},
			},
		},
		{
			name: "floating point decimal no leading zero",
			text: ".25",
			expected: []TokenWithText{
				{Text: `.25`, Role: parser.TokenRoleNumber},
			},
		},
		{
			name: "floating point decimal no leading zero with exponent",
			text: ".12345E+5",
			expected: []TokenWithText{
				{Text: `.12345E+5`, Role: parser.TokenRoleNumber},
			},
		},
		{
			name: "floating point decimal with underscore",
			text: "1_5.",
			expected: []TokenWithText{
				{Text: `1_5.`, Role: parser.TokenRoleNumber},
			},
		},
		{
			name: "floating point exponent with underscore",
			text: "0.15e+0_2",
			expected: []TokenWithText{
				{Text: `0.15e+0_2`, Role: parser.TokenRoleNumber},
			},
		},
		{
			name: "floating point hex with negative exponent",
			text: "0x1p-2",
			expected: []TokenWithText{
				{Text: `0x1p-2`, Role: parser.TokenRoleNumber},
			},
		},
		{
			name: "floating point hex with positive exponent",
			text: "0x2.p10",
			expected: []TokenWithText{
				{Text: `0x2.p10`, Role: parser.TokenRoleNumber},
			},
		},
		{
			name: "floating point hex with plus zero exponent",
			text: "0x1.Fp+0",
			expected: []TokenWithText{
				{Text: `0x1.Fp+0`, Role: parser.TokenRoleNumber},
			},
		},
		{
			name: "floating point hex with minus zero exponent",
			text: "0X.8p-0",
			expected: []TokenWithText{
				{Text: `0X.8p-0`, Role: parser.TokenRoleNumber},
			},
		},
		{
			name: "floating point hex with underscores",
			text: "0X_1FFFP-16",
			expected: []TokenWithText{
				{Text: `0X_1FFFP-16`, Role: parser.TokenRoleNumber},
			},
		},
		{
			name: "floating point hex minus integer",
			text: "0x15e-2",
			expected: []TokenWithText{
				{Text: `0x15e`, Role: parser.TokenRoleNumber},
				{Text: `-`, Role: parser.TokenRoleOperator},
				{Text: `2`, Role: parser.TokenRoleNumber},
			},
		},
		{
			name: "floating point hex invalid, mantissa has no digits",
			text: "0x.p1",
			expected: []TokenWithText{
				{Text: `0`, Role: parser.TokenRoleNumber},
				{Text: `x`, Role: parser.TokenRoleIdentifier},
				{Text: `p1`, Role: parser.TokenRoleIdentifier},
			},
		},
		{
			name: "floating point hex invalid, p exponent requires hexadecimal mantissa",
			text: "1p-2",
			expected: []TokenWithText{
				{Text: `1`, Role: parser.TokenRoleNumber},
				{Text: `p`, Role: parser.TokenRoleIdentifier},
				{Text: `-`, Role: parser.TokenRoleOperator},
				{Text: `2`, Role: parser.TokenRoleNumber},
			},
		},
		{
			name: "floating point hex invalid, hexadecimal mantissa requires p exponent",
			text: "0x1.5e-2",
			expected: []TokenWithText{
				{Text: `0x1`, Role: parser.TokenRoleNumber},
				{Text: `.5e-2`, Role: parser.TokenRoleNumber},
			},
		},
		{
			name: "floating point hex invalid, _ must separate successive digits before decimal point",
			text: "1_.5",
			expected: []TokenWithText{
				{Text: `1`, Role: parser.TokenRoleNumber},
				{Text: `_`, Role: parser.TokenRoleIdentifier},
				{Text: `.5`, Role: parser.TokenRoleNumber},
			},
		},
		{
			name: "floating point hex invalid, _ must separate successive digits after decimal point",
			text: "1._5",
			expected: []TokenWithText{
				{Text: `1.`, Role: parser.TokenRoleNumber},
				{Text: `_5`, Role: parser.TokenRoleIdentifier},
			},
		},
		{
			name: "floating point hex invalid, _ must separate successive digits before exponent",
			text: "1.5_e1",
			expected: []TokenWithText{
				{Text: `1.5`, Role: parser.TokenRoleNumber},
				{Text: `_e1`, Role: parser.TokenRoleIdentifier},
			},
		},
		{
			name: "floating point hex invalid, _ must separate successive digits after exponent",
			text: "1.5e_1",
			expected: []TokenWithText{
				{Text: `1.5`, Role: parser.TokenRoleNumber},
				{Text: `e_1`, Role: parser.TokenRoleIdentifier},
			},
		},
		{
			name: "floating point hex invalid, _ must separate successive digits at end",
			text: "1.5e1_",
			expected: []TokenWithText{
				{Text: `1.5e1`, Role: parser.TokenRoleNumber},
				{Text: `_`, Role: parser.TokenRoleIdentifier},
			},
		},
		{
			name: "imaginary zero",
			text: "0i",
			expected: []TokenWithText{
				{Text: `0i`, Role: parser.TokenRoleNumber},
			},
		},
		{
			name: "imaginary decimal with leading zero",
			text: "0123i",
			expected: []TokenWithText{
				{Text: `0123i`, Role: parser.TokenRoleNumber},
			},
		},
		{
			name: "imaginary octal",
			text: "0o123i",
			expected: []TokenWithText{
				{Text: `0o123i`, Role: parser.TokenRoleNumber},
			},
		},
		{
			name: "imaginary hex",
			text: "0xabci",
			expected: []TokenWithText{
				{Text: `0xabci`, Role: parser.TokenRoleNumber},
			},
		},
		{
			name: "imaginary floating point zero",
			text: "0.i",
			expected: []TokenWithText{
				{Text: `0.i`, Role: parser.TokenRoleNumber},
			},
		},
		{
			name: "imaginary floating point with decimal",
			text: "2.71828i",
			expected: []TokenWithText{
				{Text: `2.71828i`, Role: parser.TokenRoleNumber},
			},
		},
		{
			name: "imaginary floating point zero exponent",
			text: "1.e+0i",
			expected: []TokenWithText{
				{Text: `1.e+0i`, Role: parser.TokenRoleNumber},
			},
		},
		{
			name: "imaginary floating point negative exponent",
			text: "6.67428e-11i",
			expected: []TokenWithText{
				{Text: `6.67428e-11i`, Role: parser.TokenRoleNumber},
			},
		},
		{
			name: "imaginary floating point uppercase exponent",
			text: "1E6i",
			expected: []TokenWithText{
				{Text: `1E6i`, Role: parser.TokenRoleNumber},
			},
		},
		{
			name: "imaginary floating point no leading zero",
			text: ".25i",
			expected: []TokenWithText{
				{Text: `.25i`, Role: parser.TokenRoleNumber},
			},
		},
		{
			name: "imaginary floating point no leading zero with exponent",
			text: ".12345E+5i",
			expected: []TokenWithText{
				{Text: `.12345E+5i`, Role: parser.TokenRoleNumber},
			},
		},
		{
			name: "imaginary floating point hex with negative exponent",
			text: "0x1p-2i",
			expected: []TokenWithText{
				{Text: `0x1p-2i`, Role: parser.TokenRoleNumber},
			},
		},
		{
			name: "const declaration",
			text: `const foo = "test"`,
			expected: []TokenWithText{
				{Text: "const", Role: parser.TokenRoleKeyword},
				{Text: "foo", Role: parser.TokenRoleIdentifier},
				{Text: "=", Role: parser.TokenRoleOperator},
				{Text: `"test"`, Role: parser.TokenRoleString},
			},
		},
		{
			name: "interface with underlying type",
			text: `
interface {
	~int
	String() string
}`,
			expected: []TokenWithText{
				{Text: "interface", Role: parser.TokenRoleKeyword},
				{Text: "~", Role: parser.TokenRoleOperator},
				{Text: "int", Role: parser.TokenRoleKeyword},
				{Text: "String", Role: parser.TokenRoleIdentifier},
				{Text: "string", Role: parser.TokenRoleKeyword},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tokens := ParseTokensWithText(GolangParseFunc(), tc.text)
			assert.Equal(t, tc.expected, tokens)
		})
	}
}
