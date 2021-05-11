package syntax

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/aretext/aretext/syntax/parser"
)

func TestGolangTokenizer(t *testing.T) {
	testCases := []struct {
		name           string
		inputString    string
		expectedTokens []TokenWithText
	}{
		{
			name:        "line comment",
			inputString: "// comment",
			expectedTokens: []TokenWithText{
				{Text: `//`, Role: parser.TokenRoleCommentDelimiter},
				{Text: ` comment`, Role: parser.TokenRoleComment},
			},
		},
		{
			name:        "general comment",
			inputString: "/* abcd\n123 */",
			expectedTokens: []TokenWithText{
				{Text: "/*", Role: parser.TokenRoleCommentDelimiter},
				{Text: " abcd\n123 ", Role: parser.TokenRoleComment},
				{Text: "*/", Role: parser.TokenRoleCommentDelimiter},
			},
		},
		{
			name:        "variable declaration",
			inputString: `var foo []int`,
			expectedTokens: []TokenWithText{
				{Text: "var", Role: parser.TokenRoleKeyword},
				{Text: "foo", Role: parser.TokenRoleIdentifier},
				{Text: "[", Role: parser.TokenRolePunctuation},
				{Text: "]", Role: parser.TokenRolePunctuation},
				{Text: "int", Role: parser.TokenRoleKeyword},
			},
		},
		{
			name:        "operators",
			inputString: "a + b / c",
			expectedTokens: []TokenWithText{
				{Text: "a", Role: parser.TokenRoleIdentifier},
				{Text: "+", Role: parser.TokenRoleOperator},
				{Text: "b", Role: parser.TokenRoleIdentifier},
				{Text: "/", Role: parser.TokenRoleOperator},
				{Text: "c", Role: parser.TokenRoleIdentifier},
			},
		},
		{
			name:        "raw string",
			inputString: "`abcd\n123`",
			expectedTokens: []TokenWithText{
				{Text: "`", Role: parser.TokenRoleStringQuote},
				{Text: "abcd\n123", Role: parser.TokenRoleString},
				{Text: "`", Role: parser.TokenRoleStringQuote},
			},
		},
		{
			name:        "interpreted string",
			inputString: `"abcd"`,
			expectedTokens: []TokenWithText{
				{Text: `"`, Role: parser.TokenRoleStringQuote},
				{Text: `abcd`, Role: parser.TokenRoleString},
				{Text: `"`, Role: parser.TokenRoleStringQuote},
			},
		},
		{
			name:        "interpreted string empty",
			inputString: `""`,
			expectedTokens: []TokenWithText{
				{Text: `"`, Role: parser.TokenRoleStringQuote},
				{Text: `"`, Role: parser.TokenRoleStringQuote},
			},
		},
		{
			name:        "interpreted string with escaped quote",
			inputString: `"ab\"cd"`,
			expectedTokens: []TokenWithText{
				{Text: `"`, Role: parser.TokenRoleStringQuote},
				{Text: `ab\"cd`, Role: parser.TokenRoleString},
				{Text: `"`, Role: parser.TokenRoleStringQuote},
			},
		},
		{
			name:        "incomplete interpreted string ending with escaped quote",
			inputString: `"abc\" 123`,
			expectedTokens: []TokenWithText{
				{Text: `abc`, Role: parser.TokenRoleIdentifier},
				{Text: `123`, Role: parser.TokenRoleNumber},
			},
		},
		{
			name:        "incomplete interpreted string with newline before quote",
			inputString: "\"abc\n\"",
			expectedTokens: []TokenWithText{
				{Text: `abc`, Role: parser.TokenRoleIdentifier},
			},
		},
		{
			name:        "rune",
			inputString: `'a'`,
			expectedTokens: []TokenWithText{
				{Text: `'`, Role: parser.TokenRoleStringQuote},
				{Text: `a`, Role: parser.TokenRoleString},
				{Text: `'`, Role: parser.TokenRoleStringQuote},
			},
		},
		{
			name:        "identifier with underscore prefix",
			inputString: "_x9",
			expectedTokens: []TokenWithText{
				{Text: `_x9`, Role: parser.TokenRoleIdentifier},
			},
		},
		{
			name:        "identifier with mixed case",
			inputString: "ThisVariableIsExported",
			expectedTokens: []TokenWithText{
				{Text: `ThisVariableIsExported`, Role: parser.TokenRoleIdentifier},
			},
		},
		{
			name:        "identifier with non-ascii unicode",
			inputString: "αβ",
			expectedTokens: []TokenWithText{
				{Text: "αβ", Role: parser.TokenRoleIdentifier},
			},
		},
		{
			name:        "integer with underscore",
			inputString: "4_2",
			expectedTokens: []TokenWithText{
				{Text: `4_2`, Role: parser.TokenRoleNumber},
			},
		},
		{
			name:        "octal without o separator",
			inputString: "0600",
			expectedTokens: []TokenWithText{
				{Text: `0600`, Role: parser.TokenRoleNumber},
			},
		},
		{
			name:        "integer with leading zero and underscore",
			inputString: "0_600",
			expectedTokens: []TokenWithText{
				{Text: `0_600`, Role: parser.TokenRoleNumber},
			},
		},
		{
			name:        "octal with lowercase o",
			inputString: "0o600",
			expectedTokens: []TokenWithText{
				{Text: `0o600`, Role: parser.TokenRoleNumber},
			},
		},
		{
			name:        "octal with uppercase O",
			inputString: "0O600",
			expectedTokens: []TokenWithText{
				{Text: `0O600`, Role: parser.TokenRoleNumber},
			},
		},
		{
			name:        "hex with lowercase x",
			inputString: "0xBadFace",
			expectedTokens: []TokenWithText{
				{Text: `0xBadFace`, Role: parser.TokenRoleNumber},
			},
		},
		{
			name:        "hex with lowercase x and underscore",
			inputString: "0xBad_Face",
			expectedTokens: []TokenWithText{
				{Text: `0xBad_Face`, Role: parser.TokenRoleNumber},
			},
		},
		{
			name:        "hex with leading underscore",
			inputString: "0x_67_7a_2f_cc_40_c6",
			expectedTokens: []TokenWithText{
				{Text: `0x_67_7a_2f_cc_40_c6`, Role: parser.TokenRoleNumber},
			},
		},
		{
			name:        "long number no underscores",
			inputString: "170141183460469231731687303715884105727",
			expectedTokens: []TokenWithText{
				{Text: `170141183460469231731687303715884105727`, Role: parser.TokenRoleNumber},
			},
		},
		{
			name:        "long number with underscores",
			inputString: "170_141183_460469_231731_687303_715884_105727",
			expectedTokens: []TokenWithText{
				{Text: `170_141183_460469_231731_687303_715884_105727`, Role: parser.TokenRoleNumber},
			},
		},
		{
			name:        "identifier with leading underscore and digits",
			inputString: "_42",
			expectedTokens: []TokenWithText{
				{Text: `_42`, Role: parser.TokenRoleIdentifier},
			},
		},
		{
			name:        "invalid number with digits and trailing underscore",
			inputString: "42_",
			expectedTokens: []TokenWithText{
				{Text: `42`, Role: parser.TokenRoleNumber},
				{Text: `_`, Role: parser.TokenRoleIdentifier},
			},
		},
		{
			name:        "invalid number with multiple underscores",
			inputString: "4__2",
			expectedTokens: []TokenWithText{
				{Text: `4`, Role: parser.TokenRoleNumber},
				{Text: `__2`, Role: parser.TokenRoleIdentifier},
			},
		},
		{
			name:        "invalid hex with leading underscore",
			inputString: "0_xBadFace",
			expectedTokens: []TokenWithText{
				{Text: `0`, Role: parser.TokenRoleNumber},
				{Text: `_xBadFace`, Role: parser.TokenRoleIdentifier},
			},
		},
		{
			name:        "floating point zero with decimal",
			inputString: "0.",
			expectedTokens: []TokenWithText{
				{Text: `0.`, Role: parser.TokenRoleNumber},
			},
		},
		{
			name:        "floating point with decimal",
			inputString: "72.40",
			expectedTokens: []TokenWithText{
				{Text: `72.40`, Role: parser.TokenRoleNumber},
			},
		},
		{
			name:        "floating point with decimal and leading zero",
			inputString: "072.40",
			expectedTokens: []TokenWithText{
				{Text: `072.40`, Role: parser.TokenRoleNumber},
			},
		},
		{
			name:        "floating point with single leading decimal",
			inputString: "2.71828",
			expectedTokens: []TokenWithText{
				{Text: `2.71828`, Role: parser.TokenRoleNumber},
			},
		},
		{
			name:        "floating point with exponent zero",
			inputString: "1.e+0",
			expectedTokens: []TokenWithText{
				{Text: `1.e+0`, Role: parser.TokenRoleNumber},
			},
		},
		{
			name:        "floating point with negative exponent",
			inputString: "6.67428e-11",
			expectedTokens: []TokenWithText{
				{Text: `6.67428e-11`, Role: parser.TokenRoleNumber},
			},
		},
		{
			name:        "exponent with uppercase E",
			inputString: "1E6",
			expectedTokens: []TokenWithText{
				{Text: `1E6`, Role: parser.TokenRoleNumber},
			},
		},
		{
			name:        "floating point decimal no leading zero",
			inputString: ".25",
			expectedTokens: []TokenWithText{
				{Text: `.25`, Role: parser.TokenRoleNumber},
			},
		},
		{
			name:        "floating point decimal no leading zero with exponent",
			inputString: ".12345E+5",
			expectedTokens: []TokenWithText{
				{Text: `.12345E+5`, Role: parser.TokenRoleNumber},
			},
		},
		{
			name:        "floating point decimal with underscore",
			inputString: "1_5.",
			expectedTokens: []TokenWithText{
				{Text: `1_5.`, Role: parser.TokenRoleNumber},
			},
		},
		{
			name:        "floating point exponent with underscore",
			inputString: "0.15e+0_2",
			expectedTokens: []TokenWithText{
				{Text: `0.15e+0_2`, Role: parser.TokenRoleNumber},
			},
		},
		{
			name:        "floating point hex with negative exponent",
			inputString: "0x1p-2",
			expectedTokens: []TokenWithText{
				{Text: `0x1p-2`, Role: parser.TokenRoleNumber},
			},
		},
		{
			name:        "floating point hex with positive exponent",
			inputString: "0x2.p10",
			expectedTokens: []TokenWithText{
				{Text: `0x2.p10`, Role: parser.TokenRoleNumber},
			},
		},
		{
			name:        "floating point hex with plus zero exponent",
			inputString: "0x1.Fp+0",
			expectedTokens: []TokenWithText{
				{Text: `0x1.Fp+0`, Role: parser.TokenRoleNumber},
			},
		},
		{
			name:        "floating point hex with minus zero exponent",
			inputString: "0X.8p-0",
			expectedTokens: []TokenWithText{
				{Text: `0X.8p-0`, Role: parser.TokenRoleNumber},
			},
		},
		{
			name:        "floating point hex with underscores",
			inputString: "0X_1FFFP-16",
			expectedTokens: []TokenWithText{
				{Text: `0X_1FFFP-16`, Role: parser.TokenRoleNumber},
			},
		},
		{
			name:        "floating point hex minus integer",
			inputString: "0x15e-2",
			expectedTokens: []TokenWithText{
				{Text: `0x15e`, Role: parser.TokenRoleNumber},
				{Text: `-`, Role: parser.TokenRoleOperator},
				{Text: `2`, Role: parser.TokenRoleNumber},
			},
		},
		{
			name:        "floating point hex invalid, mantissa has no digits",
			inputString: "0x.p1",
			expectedTokens: []TokenWithText{
				{Text: `0`, Role: parser.TokenRoleNumber},
				{Text: `x`, Role: parser.TokenRoleIdentifier},
				{Text: `.`, Role: parser.TokenRolePunctuation},
				{Text: `p1`, Role: parser.TokenRoleIdentifier},
			},
		},
		{
			name:        "floating point hex invalid, p exponent requires hexadecimal mantissa",
			inputString: "1p-2",
			expectedTokens: []TokenWithText{
				{Text: `1`, Role: parser.TokenRoleNumber},
				{Text: `p`, Role: parser.TokenRoleIdentifier},
				{Text: `-`, Role: parser.TokenRoleOperator},
				{Text: `2`, Role: parser.TokenRoleNumber},
			},
		},
		{
			name:        "floating point hex invalid, hexadecimal mantissa requires p exponent",
			inputString: "0x1.5e-2",
			expectedTokens: []TokenWithText{
				{Text: `0x1`, Role: parser.TokenRoleNumber},
				{Text: `.5e-2`, Role: parser.TokenRoleNumber},
			},
		},
		{
			name:        "floating point hex invalid, _ must separate successive digits before decimal point",
			inputString: "1_.5",
			expectedTokens: []TokenWithText{
				{Text: `1`, Role: parser.TokenRoleNumber},
				{Text: `_`, Role: parser.TokenRoleIdentifier},
				{Text: `.5`, Role: parser.TokenRoleNumber},
			},
		},
		{
			name:        "floating point hex invalid, _ must separate successive digits after decimal point",
			inputString: "1._5",
			expectedTokens: []TokenWithText{
				{Text: `1.`, Role: parser.TokenRoleNumber},
				{Text: `_5`, Role: parser.TokenRoleIdentifier},
			},
		},
		{
			name:        "floating point hex invalid, _ must separate successive digits before exponent",
			inputString: "1.5_e1",
			expectedTokens: []TokenWithText{
				{Text: `1.5`, Role: parser.TokenRoleNumber},
				{Text: `_e1`, Role: parser.TokenRoleIdentifier},
			},
		},
		{
			name:        "floating point hex invalid, _ must separate successive digits after exponent",
			inputString: "1.5e_1",
			expectedTokens: []TokenWithText{
				{Text: `1.5`, Role: parser.TokenRoleNumber},
				{Text: `e_1`, Role: parser.TokenRoleIdentifier},
			},
		},
		{
			name:        "floating point hex invalid, _ must separate successive digits at end",
			inputString: "1.5e1_",
			expectedTokens: []TokenWithText{
				{Text: `1.5e1`, Role: parser.TokenRoleNumber},
				{Text: `_`, Role: parser.TokenRoleIdentifier},
			},
		},
		{
			name:        "imaginary zero",
			inputString: "0i",
			expectedTokens: []TokenWithText{
				{Text: `0i`, Role: parser.TokenRoleNumber},
			},
		},
		{
			name:        "imaginary decimal with leading zero",
			inputString: "0123i",
			expectedTokens: []TokenWithText{
				{Text: `0123i`, Role: parser.TokenRoleNumber},
			},
		},
		{
			name:        "imaginary octal",
			inputString: "0o123i",
			expectedTokens: []TokenWithText{
				{Text: `0o123i`, Role: parser.TokenRoleNumber},
			},
		},
		{
			name:        "imaginary hex",
			inputString: "0xabci",
			expectedTokens: []TokenWithText{
				{Text: `0xabci`, Role: parser.TokenRoleNumber},
			},
		},
		{
			name:        "imaginary floating point zero",
			inputString: "0.i",
			expectedTokens: []TokenWithText{
				{Text: `0.i`, Role: parser.TokenRoleNumber},
			},
		},
		{
			name:        "imaginary floating point with decimal",
			inputString: "2.71828i",
			expectedTokens: []TokenWithText{
				{Text: `2.71828i`, Role: parser.TokenRoleNumber},
			},
		},
		{
			name:        "imaginary floating point zero exponent",
			inputString: "1.e+0i",
			expectedTokens: []TokenWithText{
				{Text: `1.e+0i`, Role: parser.TokenRoleNumber},
			},
		},
		{
			name:        "imaginary floating point negative exponent",
			inputString: "6.67428e-11i",
			expectedTokens: []TokenWithText{
				{Text: `6.67428e-11i`, Role: parser.TokenRoleNumber},
			},
		},
		{
			name:        "imaginary floating point uppercase exponent",
			inputString: "1E6i",
			expectedTokens: []TokenWithText{
				{Text: `1E6i`, Role: parser.TokenRoleNumber},
			},
		},
		{
			name:        "imaginary floating point no leading zero",
			inputString: ".25i",
			expectedTokens: []TokenWithText{
				{Text: `.25i`, Role: parser.TokenRoleNumber},
			},
		},
		{
			name:        "imaginary floating point no leading zero with exponent",
			inputString: ".12345E+5i",
			expectedTokens: []TokenWithText{
				{Text: `.12345E+5i`, Role: parser.TokenRoleNumber},
			},
		},
		{
			name:        "imaginary floating point hex with negative exponent",
			inputString: "0x1p-2i",
			expectedTokens: []TokenWithText{
				{Text: `0x1p-2i`, Role: parser.TokenRoleNumber},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tokens, err := ParseTokensWithText(LanguageGo, tc.inputString)
			require.NoError(t, err)
			assert.Equal(t, tc.expectedTokens, tokens)
		})
	}
}
