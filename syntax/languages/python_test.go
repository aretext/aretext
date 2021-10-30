package languages

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/aretext/aretext/syntax/parser"
)

func TestPythonParseFunc(t *testing.T) {
	testCases := []struct {
		name     string
		text     string
		expected []TokenWithText
	}{
		{
			name: "comment at start of document",
			text: "# comment",
			expected: []TokenWithText{
				{Text: `# comment`, Role: parser.TokenRoleComment},
			},
		},
		{
			name: "single comment hash",
			text: "#",
			expected: []TokenWithText{
				{Text: `#`, Role: parser.TokenRoleComment},
			},
		},
		{
			name: "identifier followed by comment",
			text: "foobar #comment",
			expected: []TokenWithText{
				{Text: "foobar", Role: parser.TokenRoleIdentifier},
				{Text: `#comment`, Role: parser.TokenRoleComment},
			},
		},
		{
			name: "comment followed by identifier on next line",
			text: "foo\n# comment\nbar",
			expected: []TokenWithText{
				{Text: "foo", Role: parser.TokenRoleIdentifier},
				{Text: "# comment\n", Role: parser.TokenRoleComment},
				{Text: "bar", Role: parser.TokenRoleIdentifier},
			},
		},
		{
			name: "encoding declaration",
			text: "# -*- coding: <encoding-name> -*-",
			expected: []TokenWithText{
				{Text: "# -*- coding: <encoding-name> -*-", Role: parser.TokenRoleComment},
			},
		},
		{
			name: "keywords in for-range loop",
			text: "for i in range(len(l)):",
			expected: []TokenWithText{
				{Text: "for", Role: parser.TokenRoleKeyword},
				{Text: "i", Role: parser.TokenRoleIdentifier},
				{Text: "in", Role: parser.TokenRoleKeyword},
				{Text: "range", Role: parser.TokenRoleIdentifier},
				{Text: "len", Role: parser.TokenRoleIdentifier},
				{Text: "l", Role: parser.TokenRoleIdentifier},
			},
		},
		{
			name: "operators in equation",
			text: "x + y - z",
			expected: []TokenWithText{
				{Text: "x", Role: parser.TokenRoleIdentifier},
				{Text: "+", Role: parser.TokenRoleOperator},
				{Text: "y", Role: parser.TokenRoleIdentifier},
				{Text: "-", Role: parser.TokenRoleOperator},
				{Text: "z", Role: parser.TokenRoleIdentifier},
			},
		},
		{
			name: "comparison operators",
			text: "a <= b and c > d or e != f and g == h",
			expected: []TokenWithText{
				{Text: "a", Role: parser.TokenRoleIdentifier},
				{Text: "<=", Role: parser.TokenRoleOperator},
				{Text: "b", Role: parser.TokenRoleIdentifier},
				{Text: "and", Role: parser.TokenRoleKeyword},
				{Text: "c", Role: parser.TokenRoleIdentifier},
				{Text: ">", Role: parser.TokenRoleOperator},
				{Text: "d", Role: parser.TokenRoleIdentifier},
				{Text: "or", Role: parser.TokenRoleKeyword},
				{Text: "e", Role: parser.TokenRoleIdentifier},
				{Text: "!=", Role: parser.TokenRoleOperator},
				{Text: "f", Role: parser.TokenRoleIdentifier},
				{Text: "and", Role: parser.TokenRoleKeyword},
				{Text: "g", Role: parser.TokenRoleIdentifier},
				{Text: "==", Role: parser.TokenRoleOperator},
				{Text: "h", Role: parser.TokenRoleIdentifier},
			},
		},
		{
			name: "test short string, single quote",
			text: `'foo\nbar'`,
			expected: []TokenWithText{
				{Text: `'foo\nbar'`, Role: parser.TokenRoleString},
			},
		},
		{
			name: "test short string, double quote",
			text: `"foo\nbar"`,
			expected: []TokenWithText{
				{Text: `"foo\nbar"`, Role: parser.TokenRoleString},
			},
		},
		{
			name: "test long string, single quote",
			text: "'''foo\n'\nbar\n'''",
			expected: []TokenWithText{
				{Text: "'''foo\n'\nbar\n'''", Role: parser.TokenRoleString},
			},
		},
		{
			name: "test long string, double quote",
			text: "\"\"\"foo\n\n\"\nbar\"\"\"",
			expected: []TokenWithText{
				{Text: "\"\"\"foo\n\n\"\nbar\"\"\"", Role: parser.TokenRoleString},
			},
		},
		{
			name: "string with byte prefix",
			text: `b"foobar"`,
			expected: []TokenWithText{
				{Text: `b"foobar"`, Role: parser.TokenRoleString},
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
			name: "decimal int literal, multiple digits with separators",
			text: "100_000_000_000",
			expected: []TokenWithText{
				{Text: "100_000_000_000", Role: parser.TokenRoleNumber},
			},
		},
		{
			name: "binary literal",
			text: "0b_1110_0101",
			expected: []TokenWithText{
				{Text: "0b_1110_0101", Role: parser.TokenRoleNumber},
			},
		},
		{
			name: "octal literal",
			text: "0o377",
			expected: []TokenWithText{
				{Text: "0o377", Role: parser.TokenRoleNumber},
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
		{
			name: "float literal, separators",
			text: "3.14_15_93",
			expected: []TokenWithText{
				{Text: "3.14_15_93", Role: parser.TokenRoleNumber},
			},
		},
		{
			name: "imaginary number, point float decimal in middle",
			text: "3.14j",
			expected: []TokenWithText{
				{Text: "3.14j", Role: parser.TokenRoleNumber},
			},
		},
		{
			name: "imaginary number, point float decimal at end",
			text: "10.j",
			expected: []TokenWithText{
				{Text: "10.j", Role: parser.TokenRoleNumber},
			},
		},
		{
			name: "imaginary number, int literal",
			text: "10j",
			expected: []TokenWithText{
				{Text: "10j", Role: parser.TokenRoleNumber},
			},
		},
		{
			name: "imaginary number, exponent ",
			text: "3.14e-10j",
			expected: []TokenWithText{
				{Text: "3.14e-10j", Role: parser.TokenRoleNumber},
			},
		},
		{
			name: "imaginary number, separators ",
			text: "3.14_15_93j",
			expected: []TokenWithText{
				{Text: "3.14_15_93j", Role: parser.TokenRoleNumber},
			},
		},
		{
			name: "full program",
			text: `import sys

def main():
    if len(sys.argv) < 2:
        print("Usage: {} NAME".format(sys.argv[0]))
        sys.exit(1)
    name = sys.argv[1]
    print("Hello {}!".format(name))

if __name__ == "__main__":
    main()
`,
			expected: []TokenWithText{
				{Role: parser.TokenRoleKeyword, Text: "import"},
				{Role: parser.TokenRoleIdentifier, Text: "sys"},
				{Role: parser.TokenRoleKeyword, Text: "def"},
				{Role: parser.TokenRoleIdentifier, Text: "main"},
				{Role: parser.TokenRoleKeyword, Text: "if"},
				{Role: parser.TokenRoleIdentifier, Text: "len"},
				{Role: parser.TokenRoleIdentifier, Text: "sys"},
				{Role: parser.TokenRoleIdentifier, Text: "argv"},
				{Role: parser.TokenRoleOperator, Text: "<"},
				{Role: parser.TokenRoleNumber, Text: "2"},
				{Role: parser.TokenRoleIdentifier, Text: "print"},
				{Role: parser.TokenRoleString, Text: "\"Usage: {} NAME\""},
				{Role: parser.TokenRoleIdentifier, Text: "format"},
				{Role: parser.TokenRoleIdentifier, Text: "sys"},
				{Role: parser.TokenRoleIdentifier, Text: "argv"},
				{Role: parser.TokenRoleNumber, Text: "0"},
				{Role: parser.TokenRoleIdentifier, Text: "sys"},
				{Role: parser.TokenRoleIdentifier, Text: "exit"},
				{Role: parser.TokenRoleNumber, Text: "1"},
				{Role: parser.TokenRoleIdentifier, Text: "name"},
				{Role: parser.TokenRoleIdentifier, Text: "sys"},
				{Role: parser.TokenRoleIdentifier, Text: "argv"},
				{Role: parser.TokenRoleNumber, Text: "1"},
				{Role: parser.TokenRoleIdentifier, Text: "print"},
				{Role: parser.TokenRoleString, Text: "\"Hello {}!\""},
				{Role: parser.TokenRoleIdentifier, Text: "format"},
				{Role: parser.TokenRoleIdentifier, Text: "name"},
				{Role: parser.TokenRoleKeyword, Text: "if"},
				{Role: parser.TokenRoleIdentifier, Text: "__name__"},
				{Role: parser.TokenRoleOperator, Text: "=="},
				{Role: parser.TokenRoleString, Text: "\"__main__\""},
				{Role: parser.TokenRoleIdentifier, Text: "main"},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tokens := ParseTokensWithText(PythonParseFunc(), tc.text)
			assert.Equal(t, tc.expected, tokens)
		})
	}
}
