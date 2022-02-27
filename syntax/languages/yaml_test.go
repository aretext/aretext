package languages

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/aretext/aretext/syntax/parser"
)

func TestYamlParseFunc(t *testing.T) {
	const tokenRoleKey = parser.TokenRoleCustom1
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
				{Text: `abc:`, Role: tokenRoleKey},
			},
		},
		{
			name: "key with single-quoted string",
			text: "'abc': xyz",
			expected: []TokenWithText{
				{Text: `'abc':`, Role: tokenRoleKey},
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
			name: "doc with keys, number, and comment",
			text: `- foo: # this is a test
							bar: 123`,
			expected: []TokenWithText{
				{Text: `foo:`, Role: tokenRoleKey},
				{Text: "# this is a test\n", Role: parser.TokenRoleComment},
				{Text: `bar:`, Role: tokenRoleKey},
				{Text: `123`, Role: parser.TokenRoleNumber},
			},
		},
		{
			name: "unquoted string with number suffix",
			text: "foo: v0.1.2",
			expected: []TokenWithText{
				{Text: `foo:`, Role: tokenRoleKey},
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
