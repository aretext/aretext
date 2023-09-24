package languages

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/aretext/aretext/syntax/parser"
)

func TestRustParseFunc(t *testing.T) {
	testCases := []struct {
		name     string
		text     string
		expected []TokenWithText
	}{
		{
			name: "empty line comment",
			text: "//",
			expected: []TokenWithText{
				{Text: "//", Role: parser.TokenRoleComment},
			},
		},
		{
			name: "line comment",
			text: "// foo bar",
			expected: []TokenWithText{
				{Text: "// foo bar", Role: parser.TokenRoleComment},
			},
		},
		{
			name: "line doc comment",
			text: "//! foo bar",
			expected: []TokenWithText{
				{Text: "//! foo bar", Role: parser.TokenRoleComment},
			},
		},
		{
			name: "line doc comment with two bangs",
			text: "//!! foo bar",
			expected: []TokenWithText{
				{Text: "//!! foo bar", Role: parser.TokenRoleComment},
			},
		},
		{
			name: "empty block comment",
			text: "/**/",
			expected: []TokenWithText{
				{Text: "/**/", Role: parser.TokenRoleComment},
			},
		},
		{
			name: "block comment with single asterisk",
			text: "/***/",
			expected: []TokenWithText{
				{Text: "/***/", Role: parser.TokenRoleComment},
			},
		},
		{
			name: "multi-line block comment",
			text: "/** this\nis a\ncomment*/",
			expected: []TokenWithText{
				{Text: "/** this\nis a\ncomment*/", Role: parser.TokenRoleComment},
			},
		},
		{
			name: "block doc comment",
			text: "/*! foo \n bar */",
			expected: []TokenWithText{
				{Text: "/*! foo \n bar */", Role: parser.TokenRoleComment},
			},
		},
		{
			name: "character",
			text: "'H'",
			expected: []TokenWithText{
				{Text: "'H'", Role: parser.TokenRoleString},
			},
		},
		{
			name: "quote string",
			text: `"Hello"`,
			expected: []TokenWithText{
				{Text: `"Hello"`, Role: parser.TokenRoleString},
			},
		},
		{
			name: "multi-line quote string",
			text: `
const MSG: &'static str = "multi
line
string";`,
			expected: []TokenWithText{
				{Text: "const", Role: parser.TokenRoleKeyword},
				{Text: ":", Role: parser.TokenRoleOperator},
				{Text: "&", Role: parser.TokenRoleOperator},
				{Text: "'static", Role: parser.TokenRoleCustom1},
				{Text: "=", Role: parser.TokenRoleOperator},
				{Text: "\"multi\nline\nstring\"", Role: parser.TokenRoleString},
			},
		},
		{
			name: "raw string",
			text: `r#"Hello"#`,
			expected: []TokenWithText{
				{Text: `r#"Hello"#`, Role: parser.TokenRoleString},
			},
		},
		{
			name: "raw string with multiple hash delimiters",
			text: `r##"foo #"# bar"##`,
			expected: []TokenWithText{
				{Text: `r##"foo #"# bar"##`, Role: parser.TokenRoleString},
			},
		},
		{
			name: "byte",
			text: `b'H'`,
			expected: []TokenWithText{
				{Text: `b'H'`, Role: parser.TokenRoleString},
			},
		},
		{
			name: "byte string",
			text: `b"Hello"`,
			expected: []TokenWithText{
				{Text: `b"Hello"`, Role: parser.TokenRoleString},
			},
		},
		{
			name: "raw byte string",
			text: `br#"hello"#`,
			expected: []TokenWithText{
				{Text: `br#"hello"#`, Role: parser.TokenRoleString},
			},
		},
		{
			name: "integer literal",
			text: "123",
			expected: []TokenWithText{
				{Text: "123", Role: parser.TokenRoleNumber},
			},
		},
		{
			name: "integer literal, i32 suffix",
			text: "123i32",
			expected: []TokenWithText{
				{Text: "123i32", Role: parser.TokenRoleNumber},
			},
		},
		{
			name: "integer literal, u32 suffix",
			text: "123u32",
			expected: []TokenWithText{
				{Text: "123u32", Role: parser.TokenRoleNumber},
			},
		},
		{
			name: "integer literal, u32 suffix with separator",
			text: "123_u32",
			expected: []TokenWithText{
				{Text: "123_u32", Role: parser.TokenRoleNumber},
			},
		},
		{
			name: "hex literal",
			text: "0xff",
			expected: []TokenWithText{
				{Text: "0xff", Role: parser.TokenRoleNumber},
			},
		},
		{
			name: "hex literal, u8 suffix with separator",
			text: "0xff_u8",
			expected: []TokenWithText{
				{Text: "0xff_u8", Role: parser.TokenRoleNumber},
			},
		},
		{
			name: "octal literal",
			text: "0o70",
			expected: []TokenWithText{
				{Text: "0o70", Role: parser.TokenRoleNumber},
			},
		},
		{
			name: "octal literal, i16 suffix with separator",
			text: "0o70_i16",
			expected: []TokenWithText{
				{Text: "0o70_i16", Role: parser.TokenRoleNumber},
			},
		},
		{
			name: "binary literal with separators",
			text: "0b1111_1111_1001_0000",
			expected: []TokenWithText{
				{Text: "0b1111_1111_1001_0000", Role: parser.TokenRoleNumber},
			},
		},
		{
			name: "binary literal with separators and i164 suffix",
			text: "0b1111_1111_1001_0000i64",
			expected: []TokenWithText{
				{Text: "0b1111_1111_1001_0000i64", Role: parser.TokenRoleNumber},
			},
		},
		{
			name: "binary literal with repeated separators",
			text: "0b________1",
			expected: []TokenWithText{
				{Text: "0b________1", Role: parser.TokenRoleNumber},
			},
		},
		{
			name: "zero suffix usize",
			text: "0usize",
			expected: []TokenWithText{
				{Text: "0usize", Role: parser.TokenRoleNumber},
			},
		},
		{
			name: "zero with invalid suffix",
			text: "0invalidsuffix",
			expected: []TokenWithText{
				{Text: "0", Role: parser.TokenRoleNumber},
			},
		},
		{
			name: "integer with invalid digits",
			text: "123AFB43",
			expected: []TokenWithText{
				{Text: "123", Role: parser.TokenRoleNumber},
			},
		},
		{
			name: "binary with invalid digits",
			text: "0b0102",
			expected: []TokenWithText{
				{Text: "0b010", Role: parser.TokenRoleNumber},
				{Text: "2", Role: parser.TokenRoleNumber},
			},
		},
		{
			name: "octal with invalid digits",
			text: "0o0581",
			expected: []TokenWithText{
				{Text: "0o05", Role: parser.TokenRoleNumber},
				{Text: "81", Role: parser.TokenRoleNumber},
			},
		},
		{
			name: "binary literal has too few digits",
			text: "0b_",
			expected: []TokenWithText{
				{Text: "0", Role: parser.TokenRoleNumber},
			},
		},
		{
			name: "hex literal has too few digits",
			text: "0x_",
			expected: []TokenWithText{
				{Text: "0", Role: parser.TokenRoleNumber},
			},
		},
		{
			name: "octal literal has too few digits",
			text: "0b_",
			expected: []TokenWithText{
				{Text: "0", Role: parser.TokenRoleNumber},
			},
		},
		{
			name: "float literal, decimal point and suffix",
			text: "123.0f64",
			expected: []TokenWithText{
				{Text: "123.0f64", Role: parser.TokenRoleNumber},
			},
		},
		{
			name: "float literal, zero then decimal point and suffix",
			text: "0.1f32",
			expected: []TokenWithText{
				{Text: "0.1f32", Role: parser.TokenRoleNumber},
			},
		},
		{
			name: "float literal, exponent",
			text: "12E+99_f64",
			expected: []TokenWithText{
				{Text: "12E+99_f64", Role: parser.TokenRoleNumber},
			},
		},
		{
			name: "float literal, decimal with suffix",
			text: "5f32",
			expected: []TokenWithText{
				{Text: "5f32", Role: parser.TokenRoleNumber},
			},
		},
		{
			name: "float literal, decimal point at end",
			text: "2.",
			expected: []TokenWithText{
				{Text: "2.", Role: parser.TokenRoleNumber},
			},
		},
		{
			name: "keywords and identifiers",
			text: "let x = y",
			expected: []TokenWithText{
				{Text: "let", Role: parser.TokenRoleKeyword},
				{Text: "=", Role: parser.TokenRoleOperator},
			},
		},
		{
			name:     "identifier ascii",
			text:     "foo",
			expected: []TokenWithText{},
		},
		{
			name:     "identifier with underscore prefix",
			text:     "_foo",
			expected: []TokenWithText{},
		},
		{
			name:     "raw identifier",
			text:     "r#true",
			expected: []TokenWithText{},
		},
		{
			name:     "identifier with non-ascii unicode",
			text:     "Москва",
			expected: []TokenWithText{},
		},
		{
			name:     "identifier with chinese unicode",
			text:     "東京",
			expected: []TokenWithText{},
		},
		{
			name: "lifetime label",
			text: "'static",
			expected: []TokenWithText{
				{Text: "'static", Role: rustTokenRoleLifetime},
			},
		},
		{
			name: "lifetime underscore",
			text: "'_",
			expected: []TokenWithText{
				{Text: "'_", Role: rustTokenRoleLifetime},
			},
		},
		{
			name: "raw identifier cannot use crate",
			text: "r#crate",
			expected: []TokenWithText{
				{Text: "#", Role: parser.TokenRoleOperator},
				{Text: "crate", Role: parser.TokenRoleKeyword},
			},
		},
		{
			name: "raw identifier cannot use super",
			text: "r#super",
			expected: []TokenWithText{
				{Text: "#", Role: parser.TokenRoleOperator},
				{Text: "super", Role: parser.TokenRoleKeyword},
			},
		},
		{
			name: "raw identifier cannot use self",
			text: "r#self",
			expected: []TokenWithText{
				{Text: "#", Role: parser.TokenRoleOperator},
				{Text: "self", Role: parser.TokenRoleKeyword},
			},
		},
		{
			name: "raw identifier cannot use Self",
			text: "r#Self",
			expected: []TokenWithText{
				{Text: "#", Role: parser.TokenRoleOperator},
				{Text: "Self", Role: parser.TokenRoleKeyword},
			},
		},
		{
			name:     "raw identifier with excluded term prefix",
			text:     "r#superbloom",
			expected: []TokenWithText{},
		},
		{
			name: "lifetimes in function def",
			text: `fn print_refs<'a, 'b>(x: &'a i32, y: &'b i32)`,
			expected: []TokenWithText{
				{Text: "fn", Role: parser.TokenRoleKeyword},
				{Text: "<", Role: parser.TokenRoleOperator},
				{Text: "'a", Role: rustTokenRoleLifetime},
				{Text: "'b", Role: rustTokenRoleLifetime},
				{Text: ">", Role: parser.TokenRoleOperator},
				{Text: ":", Role: parser.TokenRoleOperator},
				{Text: "&", Role: parser.TokenRoleOperator},
				{Text: "'a", Role: rustTokenRoleLifetime},
				{Text: ":", Role: parser.TokenRoleOperator},
				{Text: "&", Role: parser.TokenRoleOperator},
				{Text: "'b", Role: rustTokenRoleLifetime},
			},
		},
		{
			name: "bitwise operator assignment",
			text: "x &= y",
			expected: []TokenWithText{
				{Text: "&=", Role: parser.TokenRoleOperator},
			},
		},
		{
			name: "fizzbuzz",
			text: `
fn main() {
    // A counter variable
    let mut n = 1;

    // Loop while n is less than 101
    while n < 101 {
        if n % 15 == 0 {
            println!("fizzbuzz");
        } else if n % 3 == 0 {
            println!("fizz");
        } else if n % 5 == 0 {
            println!("buzz");
        } else {
            println!("{}", n);
        }

        // Increment counter
        n += 1;
    }
}`,
			expected: []TokenWithText{
				{Role: parser.TokenRoleKeyword, Text: "fn"},
				{Role: parser.TokenRoleComment, Text: "// A counter variable\n"},
				{Role: parser.TokenRoleKeyword, Text: "let"},
				{Role: parser.TokenRoleKeyword, Text: "mut"},
				{Role: parser.TokenRoleOperator, Text: "="},
				{Role: parser.TokenRoleNumber, Text: "1"},
				{Role: parser.TokenRoleComment, Text: "// Loop while n is less than 101\n"},
				{Role: parser.TokenRoleKeyword, Text: "while"},
				{Role: parser.TokenRoleOperator, Text: "<"},
				{Role: parser.TokenRoleNumber, Text: "101"},
				{Role: parser.TokenRoleKeyword, Text: "if"},
				{Role: parser.TokenRoleOperator, Text: "%"},
				{Role: parser.TokenRoleNumber, Text: "15"},
				{Role: parser.TokenRoleOperator, Text: "=="},
				{Role: parser.TokenRoleNumber, Text: "0"},
				{Role: parser.TokenRoleOperator, Text: "!"},
				{Role: parser.TokenRoleString, Text: "\"fizzbuzz\""},
				{Role: parser.TokenRoleKeyword, Text: "else"},
				{Role: parser.TokenRoleKeyword, Text: "if"},
				{Role: parser.TokenRoleOperator, Text: "%"},
				{Role: parser.TokenRoleNumber, Text: "3"},
				{Role: parser.TokenRoleOperator, Text: "=="},
				{Role: parser.TokenRoleNumber, Text: "0"},
				{Role: parser.TokenRoleOperator, Text: "!"},
				{Role: parser.TokenRoleString, Text: "\"fizz\""},
				{Role: parser.TokenRoleKeyword, Text: "else"},
				{Role: parser.TokenRoleKeyword, Text: "if"},
				{Role: parser.TokenRoleOperator, Text: "%"},
				{Role: parser.TokenRoleNumber, Text: "5"},
				{Role: parser.TokenRoleOperator, Text: "=="},
				{Role: parser.TokenRoleNumber, Text: "0"},
				{Role: parser.TokenRoleOperator, Text: "!"},
				{Role: parser.TokenRoleString, Text: "\"buzz\""},
				{Role: parser.TokenRoleKeyword, Text: "else"},
				{Role: parser.TokenRoleOperator, Text: "!"},
				{Role: parser.TokenRoleString, Text: "\"{}\""},
				{Role: parser.TokenRoleComment, Text: "// Increment counter\n"},
				{Role: parser.TokenRoleOperator, Text: "+="},
				{Role: parser.TokenRoleNumber, Text: "1"},
			},
		},
		{
			name: "hex literal followed by other tokens",
			text: `
const VAR: u16 = 0x101;
pub enum TEST{}
`,
			expected: []TokenWithText{
				{Role: parser.TokenRoleKeyword, Text: "const"},
				{Role: parser.TokenRoleOperator, Text: ":"},
				{Role: parser.TokenRoleOperator, Text: "="},
				{Role: parser.TokenRoleNumber, Text: "0x101"},
				{Role: parser.TokenRoleKeyword, Text: "pub"},
				{Role: parser.TokenRoleKeyword, Text: "enum"},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tokens := ParseTokensWithText(RustParseFunc(), tc.text)
			assert.Equal(t, tc.expected, tokens)
		})
	}
}

func BenchmarkRustParser(b *testing.B) {
	BenchmarkParser(b, RustParseFunc(), "testdata/rust/hello.rs")
}
