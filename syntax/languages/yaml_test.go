package languages

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/aretext/aretext/syntax/parser"
)

func TestYamlParseFunc(t *testing.T) {
	testCases := []struct {
		name     string
		text     string
		expected []TokenWithText
	}{
		{
			name: "single line comment",
			text: `# abc`,
			expected: []TokenWithText{
				{Text: `# abc`, Role: parser.TokenRoleComment},
			},
		},
		{
			name: "comment right before end of file",
			text: "#",
			expected: []TokenWithText{
				{
					Role: parser.TokenRoleComment,
					Text: "#",
				},
			},
		},
		{
			name: "key without quotes",
			text: "abc: xyz",
			expected: []TokenWithText{
				{Text: `abc:`, Role: yamlTokenRoleKey},
			},
		},
		{
			name: "key with single-quoted string",
			text: "'abc': xyz",
			expected: []TokenWithText{
				{Text: `'abc':`, Role: yamlTokenRoleKey},
			},
		},
		{
			name: "empty single-quoted string",
			text: `''`,
			expected: []TokenWithText{
				{Text: `''`, Role: parser.TokenRoleString},
			},
		},
		{
			name: "single-quoted string",
			text: `'abc'`,
			expected: []TokenWithText{
				{Text: `'abc'`, Role: parser.TokenRoleString},
			},
		},
		{
			name: "single-quoted string with escaped quote",
			text: `'ab''c'`,
			expected: []TokenWithText{
				{Text: `'ab''c'`, Role: parser.TokenRoleString},
			},
		},
		{
			name: "single-quoted string with newline",
			text: `'foo
bar
baz'`,
			expected: []TokenWithText{
				{Text: "'foo\nbar\nbaz'", Role: parser.TokenRoleString},
			},
		},
		{
			name: "double-quoted string with newline",
			text: `"foo
bar
baz"`,
			expected: []TokenWithText{
				{Text: "\"foo\nbar\nbaz\"", Role: parser.TokenRoleString},
			},
		},
		{
			name: "list item with unquoted scalar starting with hyphen",
			text: `- -s -w -X go test`,
			expected: []TokenWithText{
				{Text: `-`, Role: parser.TokenRoleOperator},
			},
		},
		{
			name: "doc with keys, number, and comment",
			text: `- foo: # this is a test
							bar: 123`,
			expected: []TokenWithText{
				{Text: `-`, Role: parser.TokenRoleOperator},
				{Text: `foo:`, Role: yamlTokenRoleKey},
				{Text: "# this is a test\n", Role: parser.TokenRoleComment},
				{Text: `bar:`, Role: yamlTokenRoleKey},
				{Text: `123`, Role: parser.TokenRoleNumber},
			},
		},
		{
			name: "unquoted string with number suffix",
			text: "foo: v0.1.2",
			expected: []TokenWithText{
				{Text: `foo:`, Role: yamlTokenRoleKey},
			},
		},
		{
			name: "unquoted scalar after key",
			text: "foo: 123 abc",
			expected: []TokenWithText{
				{Text: `foo:`, Role: yamlTokenRoleKey},
			},
		},
		{
			name: "unquoted scalar on next line after key",
			text: `
foo:
  123 abc "xyz" # test
`,
			expected: []TokenWithText{
				{Text: `foo:`, Role: yamlTokenRoleKey},
				{Text: "# test\n", Role: parser.TokenRoleComment},
			},
		},
		{
			name: "unquoted scalar containing colon",
			text: `image: mysql:5.7`,
			expected: []TokenWithText{
				{Text: `image:`, Role: yamlTokenRoleKey},
			},
		},
		{
			name: "block style indicator",
			text: `
foo: |
  test 123
  "abcxyz" # not a comment
  notkey: notval
  end
bar: 678`,
			expected: []TokenWithText{
				{Text: `foo:`, Role: yamlTokenRoleKey},
				{Text: `|`, Role: parser.TokenRoleOperator},
				{Text: "\n  test 123\n  \"abcxyz\" # not a comment\n  notkey: notval\n  end\n", Role: parser.TokenRoleString},
				{Text: `bar:`, Role: yamlTokenRoleKey},
				{Text: `678`, Role: parser.TokenRoleNumber},
			},
		},
		{
			name: "block style indicator with fold style plus",
			text: `
foo: |+
  bar`,
			expected: []TokenWithText{
				{Text: `foo:`, Role: yamlTokenRoleKey},
				{Text: `|`, Role: parser.TokenRoleOperator},
				{Text: `+`, Role: parser.TokenRoleOperator},
				{Text: "\n  bar", Role: parser.TokenRoleString},
			},
		},
		{
			name: "block style indicator with fold style minus",
			text: `
foo: |-
  bar`,
			expected: []TokenWithText{
				{Text: `foo:`, Role: yamlTokenRoleKey},
				{Text: `|`, Role: parser.TokenRoleOperator},
				{Text: `-`, Role: parser.TokenRoleOperator},
				{Text: "\n  bar", Role: parser.TokenRoleString},
			},
		},
		{
			name: "override operator",
			text: `<<: *def`,
			expected: []TokenWithText{
				{Text: `<<:`, Role: parser.TokenRoleOperator},
				{Text: `*def`, Role: yamlTokenRoleAliasOrAnchor},
			},
		},
		{
			name: "alias and anchor",
			text: `
foo: &ref
bar: *ref
`,
			expected: []TokenWithText{
				{Text: `foo:`, Role: yamlTokenRoleKey},
				{Text: `&ref`, Role: yamlTokenRoleAliasOrAnchor},
				{Text: `bar:`, Role: yamlTokenRoleKey},
				{Text: `*ref`, Role: yamlTokenRoleAliasOrAnchor},
			},
		},
		{
			name: "flow style map",
			text: `
foo: {key1: 123, key2: "value2", key3: abc 123}
bar: 789
`,
			expected: []TokenWithText{
				{Text: `foo:`, Role: yamlTokenRoleKey},
				{Text: `key1:`, Role: yamlTokenRoleKey},
				{Text: `123`, Role: parser.TokenRoleNumber},
				{Text: `key2:`, Role: yamlTokenRoleKey},
				{Text: `"value2"`, Role: parser.TokenRoleString},
				{Text: `key3:`, Role: yamlTokenRoleKey},
				{Text: `bar:`, Role: yamlTokenRoleKey},
				{Text: `789`, Role: parser.TokenRoleNumber},
			},
		},
		{
			name: "flow style array",
			text: `[abc 123, 456, "string value"]`,
			expected: []TokenWithText{
				{Text: `456`, Role: parser.TokenRoleNumber},
				{Text: `"string value"`, Role: parser.TokenRoleString},
			},
		},
		{
			name: "flow style with comments",
			text: `{
 foo: bar, # comment
 baz: bat
}`,
			expected: []TokenWithText{
				{Text: `foo:`, Role: yamlTokenRoleKey},
				{Text: "# comment\n", Role: parser.TokenRoleComment},
				{Text: `baz:`, Role: yamlTokenRoleKey},
			},
		},
		{
			name: "nested flow style",
			text: `{foo: [bar, 123, {baz: 456}]}`,
			expected: []TokenWithText{
				{Text: `foo:`, Role: yamlTokenRoleKey},
				{Text: `123`, Role: parser.TokenRoleNumber},
				{Text: `456`, Role: parser.TokenRoleNumber},
			},
		},
		{
			name: "nested flow style in block style",
			text: `
key1: {foo: [bar, 123, {baz: 456}]}
key2: 789
`,
			expected: []TokenWithText{
				{Text: `key1:`, Role: yamlTokenRoleKey},
				{Text: `foo:`, Role: yamlTokenRoleKey},
				{Text: `123`, Role: parser.TokenRoleNumber},
				{Text: `456`, Role: parser.TokenRoleNumber},
				{Text: `key2:`, Role: yamlTokenRoleKey},
				{Text: `789`, Role: parser.TokenRoleNumber},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tokens := ParseTokensWithText(YamlParseFunc(), tc.text)
			assert.Equal(t, tc.expected, tokens)
		})
	}
}
