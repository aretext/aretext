package languages

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/aretext/aretext/syntax/parser"
)

func TestBashParseFunc(t *testing.T) {
	testCases := []struct {
		name     string
		text     string
		expected []TokenWithText
	}{
		{
			name: "comment",
			text: "# this is a comment",
			expected: []TokenWithText{
				{
					Role: parser.TokenRoleComment,
					Text: "# this is a comment",
				},
			},
		},
		{
			name: "if condition",
			text: `
if $var; then
echo "hello";
fi`,
			expected: []TokenWithText{
				{Role: parser.TokenRoleKeyword, Text: `if`},
				{Role: bashTokenRoleVariable, Text: `$var`},
				{Role: parser.TokenRoleKeyword, Text: `then`},
				{Role: parser.TokenRoleString, Text: `"hello"`},
				{Role: parser.TokenRoleKeyword, Text: `fi`},
			},
		},
		{
			name: "while loop",
			text: `
while $var; do
echo "hello";
done`,
			expected: []TokenWithText{
				{Role: parser.TokenRoleKeyword, Text: `while`},
				{Role: bashTokenRoleVariable, Text: `$var`},
				{Role: parser.TokenRoleKeyword, Text: `do`},
				{Role: parser.TokenRoleString, Text: `"hello"`},
				{Role: parser.TokenRoleKeyword, Text: `done`},
			},
		},
		{
			name: "case statement",
			text: `
case $var in
	foo) echo "hello"
	bar) echo "goodbye"
	*) echo "ok"
esac`,
			expected: []TokenWithText{
				{Role: parser.TokenRoleKeyword, Text: `case`},
				{Role: bashTokenRoleVariable, Text: `$var`},
				{Role: parser.TokenRoleKeyword, Text: `in`},
				{Role: parser.TokenRoleString, Text: `"hello"`},
				{Role: parser.TokenRoleString, Text: `"goodbye"`},
				{Role: parser.TokenRoleString, Text: `"ok"`},
				{Role: parser.TokenRoleKeyword, Text: `esac`},
			},
		},
		{
			name: "variable followed by hyphen",
			text: "$FOO-bar",
			expected: []TokenWithText{
				{Role: bashTokenRoleVariable, Text: `$FOO`},
			},
		},
		{
			name: "variable brace expansion",
			text: `${PATH:-}`,
			expected: []TokenWithText{
				{Role: bashTokenRoleVariable, Text: `${PATH:-}`},
			},
		},
		{
			name: "variable brace expansion with quoted close brace",
			text: `${FOO:-"close with }"}`,
			expected: []TokenWithText{
				{Role: bashTokenRoleVariable, Text: `${FOO:-"close with }"}`},
			},
		},
		{
			name: "variable positional argument",
			text: `[ $# -ne 2 ] && { echo "Usage: $0 VERSION NAME"; exit 1; }`,
			expected: []TokenWithText{
				{Role: bashTokenRoleVariable, Text: `$#`},
				{Role: parser.TokenRoleOperator, Text: `&&`},
				{Role: parser.TokenRoleString, Text: `"Usage: $0 VERSION NAME"`},
			},
		},
		{
			name: "subshell",
			text: `echo $(pwd)`,
			expected: []TokenWithText{
				{Role: parser.TokenRoleOperator, Text: `$`},
			},
		},
		{
			name: "backquote expansion",
			text: "rm `find . -name '*.go'`",
			expected: []TokenWithText{
				{Role: bashTokenRoleBackquoteExpansion, Text: "`find . -name '*.go'`"},
			},
		},
		{
			name: "file redirect",
			text: "go test > out.txt",
			expected: []TokenWithText{
				{Role: parser.TokenRoleOperator, Text: ">"},
			},
		},
		{
			name: "file redirect with ampersand",
			text: "go test &> out.txt",
			expected: []TokenWithText{
				{Role: parser.TokenRoleOperator, Text: "&>"},
			},
		},
		{
			name: "pipe",
			text: `echo "foo" | wl-copy`,
			expected: []TokenWithText{
				{Role: parser.TokenRoleString, Text: `"foo"`},
				{Role: parser.TokenRoleOperator, Text: `|`},
			},
		},
		{
			name: "regex match",
			text: `[[ $line =~ [[:space:]]*(a)?b ]]`,
			expected: []TokenWithText{
				{Role: bashTokenRoleVariable, Text: `$line`},
				{Role: parser.TokenRoleOperator, Text: `=~`},
			},
		},
		{
			name: "not condition",
			text: `if ! grep $foo; then echo "not found"; fi`,
			expected: []TokenWithText{
				{Role: parser.TokenRoleKeyword, Text: `if`},
				{Role: parser.TokenRoleOperator, Text: `!`},
				{Role: bashTokenRoleVariable, Text: `$foo`},
				{Role: parser.TokenRoleKeyword, Text: `then`},
				{Role: parser.TokenRoleString, Text: `"not found"`},
				{Role: parser.TokenRoleKeyword, Text: `fi`},
			},
		},
		{
			name: "double quote escaped quote",
			text: `"abcd \" xyz"`,
			expected: []TokenWithText{
				{Role: parser.TokenRoleString, Text: `"abcd \" xyz"`},
			},
		},
		{
			name: "double quote string multi-line",
			text: `FOO="
a
b
c"`,
			expected: []TokenWithText{
				{Role: parser.TokenRoleOperator, Text: `=`},
				{Role: parser.TokenRoleString, Text: "\"\na\nb\nc\""},
			},
		},
		{
			name: "double quote variable expansion",
			text: `"var=$VAR"`,
			expected: []TokenWithText{
				{Role: parser.TokenRoleString, Text: `"var=$VAR"`},
			},
		},
		{
			name: "escaped dollar sign before variable expansion",
			text: `\$${PATH}`,
			expected: []TokenWithText{
				{Role: bashTokenRoleVariable, Text: "${PATH}"},
			},
		},
		{
			name: "single quote string",
			text: `'abc defgh'`,
			expected: []TokenWithText{
				{Role: parser.TokenRoleString, Text: `'abc defgh'`},
			},
		},
		{
			name: "double quote string with subshell expansion",
			text: `echo "echo $(echo "\"foo\"")"`,
			expected: []TokenWithText{
				{Role: parser.TokenRoleString, Text: `"echo $(echo "\"foo\"")"`},
			},
		},
		{
			name: "double quote string with variable expansion",
			text: `echo "${FOO:-"foo"}"`,
			expected: []TokenWithText{
				{Role: parser.TokenRoleString, Text: `"${FOO:-"foo"}"`},
			},
		},
		{
			name: "double quote string with backquote expansion",
			text: "echo \"`echo \"hello\"`\"",
			expected: []TokenWithText{
				{Role: parser.TokenRoleString, Text: "\"`echo \"hello\"`\""},
			},
		},
		{
			name: "double quote string with escaped $ then variable expansion",
			text: `echo "\$${PATH}"`,
			expected: []TokenWithText{
				{Role: parser.TokenRoleString, Text: `"\$${PATH}"`},
			},
		},
		{
			name:     "unterminated double quote",
			text:     `echo "`,
			expected: []TokenWithText{},
		},
		{
			name:     "unterminated single quote",
			text:     `echo '`,
			expected: []TokenWithText{},
		},
		{
			name:     "unterminated backquote",
			text:     "echo `",
			expected: []TokenWithText{},
		},
		{
			name: "heredoc",
			text: `
cat << EOF
this is
some heredoc
text
EOF
`,
			expected: []TokenWithText{
				{Role: parser.TokenRoleOperator, Text: `<<`},
				{Role: parser.TokenRoleString, Text: `EOF
this is
some heredoc
text
EOF`},
			},
		},
		{
			name: "heredoc indented end word",
			text: `
cat << EOF
this is
some heredoc
text
	  EOF
EOF
`,
			expected: []TokenWithText{
				{Role: parser.TokenRoleOperator, Text: `<<`},
				{Role: parser.TokenRoleString, Text: `EOF
this is
some heredoc
text
	  EOF
EOF`},
			},
		},
		{
			name: "heredoc dash then word without whitespace",
			text: `
cat <<<-FOO
heredoc text
FOO
`,
			expected: []TokenWithText{
				{Role: parser.TokenRoleOperator, Text: `<<<-`},
				{Role: parser.TokenRoleString, Text: `FOO
heredoc text
FOO`},
			},
		},
		{
			name: "heredoc contains end word prefix",
			text: `
cat << EOF
EOFANDTHENSOME
EOF AND THEN SOME
EOF
`,
			expected: []TokenWithText{
				{Role: parser.TokenRoleOperator, Text: `<<`},
				{Role: parser.TokenRoleString, Text: `EOF
EOFANDTHENSOME
EOF AND THEN SOME
EOF`},
			},
		},
		{
			name: "heredoc contains partial end word",
			text: `
cat << ENDWORD
END
ENDWORD
`,
			expected: []TokenWithText{
				{Role: parser.TokenRoleOperator, Text: `<<`},
				{Role: parser.TokenRoleString, Text: `ENDWORD
END
ENDWORD`},
			},
		},
		{
			name: "heredoc no word",
			text: `
cat <<
echo "hello"
`,
			expected: []TokenWithText{
				{Role: parser.TokenRoleOperator, Text: `<<`},
				{Role: parser.TokenRoleString, Text: `"hello"`},
			},
		},
		{
			name: "heredoc EOF before word",
			text: `cat <<`,
			expected: []TokenWithText{
				{Role: parser.TokenRoleOperator, Text: `<<`},
			},
		},
		{
			name: "heredoc single-quoted word",
			text: `
cat << 'EOF'
heredoc text
EOF`,
			expected: []TokenWithText{
				{Role: parser.TokenRoleOperator, Text: `<<`},
				{Role: parser.TokenRoleString, Text: `'EOF'
heredoc text
EOF`},
			},
		},
		{
			name: "heredoc double-quoted word",
			text: `
cat << "EOF"
heredoc text
EOF`,
			expected: []TokenWithText{
				{Role: parser.TokenRoleOperator, Text: `<<`},
				{Role: parser.TokenRoleString, Text: `"EOF"
heredoc text
EOF`},
			},
		},
		{
			name: "heredoc empty quoted word",
			text: `
cat << ""
heredoc text

echo 'hello'`,
			expected: []TokenWithText{
				{Role: parser.TokenRoleOperator, Text: `<<`},
				{Role: parser.TokenRoleString, Text: `""
heredoc text
`},
				{Role: parser.TokenRoleString, Text: `'hello'`},
			},
		},
		{
			name: "heredoc one single quote",
			text: `
cat << '
foo`,
			expected: []TokenWithText{
				{Role: parser.TokenRoleOperator, Text: `<<`},
			},
		},
		{
			name: "heredoc one double quote",
			text: `
cat << "
foo`,
			expected: []TokenWithText{
				{Role: parser.TokenRoleOperator, Text: `<<`},
			},
		},
		{
			name: "heredoc backslash quoted word",
			text: `
cat << \EOF
heredoc text
EOF`,
			expected: []TokenWithText{
				{Role: parser.TokenRoleOperator, Text: `<<`},
				{Role: parser.TokenRoleString, Text: `\EOF
heredoc text
EOF`},
			},
		},
		{
			name: "heredoc backslash empty word",
			text: `
cat << \
heredoc text
EOF`,
			expected: []TokenWithText{
				{Role: parser.TokenRoleOperator, Text: `<<`},
				{Role: parser.TokenRoleString, Text: `\
heredoc text
`},
			},
		},
		{
			name: "function name with dash",
			text: `foo-for-bar() { echo "foo bar" }`,
			expected: []TokenWithText{
				{Role: parser.TokenRoleString, Text: `"foo bar"`},
			},
		},
		{
			name: "variable assignment with home expansion",
			text: `path=~/foo/bar`,
			expected: []TokenWithText{
				{Role: parser.TokenRoleOperator, Text: `=`},
			},
		},
		{
			name: "append operator",
			text: `GLOB+="foo"`,
			expected: []TokenWithText{
				{Role: parser.TokenRoleOperator, Text: `+=`},
				{Role: parser.TokenRoleString, Text: `"foo"`},
			},
		},
		{
			name: "conditional with regex start of line",
			text: `[[ $line =~ ^"initial string" ]]`,
			expected: []TokenWithText{
				{Role: bashTokenRoleVariable, Text: `$line`},
				{Role: parser.TokenRoleOperator, Text: `=~`},
				{Role: parser.TokenRoleOperator, Text: `^`},
				{Role: parser.TokenRoleString, Text: `"initial string"`},
			},
		},
		{
			name: "conditional with exact string match",
			text: `[[ $line == "test" ]]`,
			expected: []TokenWithText{
				{Role: bashTokenRoleVariable, Text: `$line`},
				{Role: parser.TokenRoleOperator, Text: `==`},
				{Role: parser.TokenRoleString, Text: `"test"`},
			},
		},
		{
			name: "conditional with lexicographic order comparison",
			text: `[[ $line > "test" ]]`,
			expected: []TokenWithText{
				{Role: bashTokenRoleVariable, Text: `$line`},
				{Role: parser.TokenRoleOperator, Text: `>`},
				{Role: parser.TokenRoleString, Text: `"test"`},
			},
		},
		{
			name: "if statement with conditional",
			text: `if [[ $line == "test"]]; then x=~/foo/bar; fi`,
			expected: []TokenWithText{
				{Role: parser.TokenRoleKeyword, Text: `if`},
				{Role: bashTokenRoleVariable, Text: `$line`},
				{Role: parser.TokenRoleOperator, Text: `==`},
				{Role: parser.TokenRoleString, Text: `"test"`},
				{Role: parser.TokenRoleKeyword, Text: `then`},
				{Role: parser.TokenRoleOperator, Text: `=`},
				{Role: parser.TokenRoleKeyword, Text: `fi`},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tokens := ParseTokensWithText(BashParseFunc(), tc.text)
			assert.Equal(t, tc.expected, tokens)
		})
	}
}

func FuzzBashParseFunc(f *testing.F) {
	seeds := LoadFuzzTestSeeds(f, "./testdata/bash/*")
	ParserFuzzTest(f, BashParseFunc(), seeds)
}
