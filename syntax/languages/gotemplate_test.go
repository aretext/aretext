package languages

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/aretext/aretext/syntax/parser"
)

func TestGoTemplateParseFunc(t *testing.T) {
	testCases := []struct {
		name     string
		text     string
		expected []TokenWithText
	}{
		{
			name:     "text with no actions",
			text:     "abcd",
			expected: []TokenWithText{},
		},
		{
			name: "action interpolate variable",
			text: "abc {{ .Test }} xyz",
			expected: []TokenWithText{
				{Text: `{{`, Role: parser.TokenRoleOperator},
				{Text: `}}`, Role: parser.TokenRoleOperator},
			},
		},
		{
			name: "action trim whitespace",
			text: "{{- .Test -}}",
			expected: []TokenWithText{
				{Text: `{{-`, Role: parser.TokenRoleOperator},
				{Text: `-}}`, Role: parser.TokenRoleOperator},
			},
		},
		{
			name: "comment",
			text: "{{ /* abc */ }}",
			expected: []TokenWithText{
				{Text: `{{`, Role: parser.TokenRoleOperator},
				{Text: `/* abc */`, Role: parser.TokenRoleComment},
				{Text: `}}`, Role: parser.TokenRoleOperator},
			},
		},
		{
			name: "string",
			text: `{{ "abc" }}`,
			expected: []TokenWithText{
				{Text: `{{`, Role: parser.TokenRoleOperator},
				{Text: `"abc"`, Role: parser.TokenRoleString},
				{Text: `}}`, Role: parser.TokenRoleOperator},
			},
		},
		{
			name: "if else",
			text: "{{ if .Value }} abc {{ else }} xyz {{ end }}",
			expected: []TokenWithText{
				{Text: `{{`, Role: parser.TokenRoleOperator},
				{Text: `if`, Role: parser.TokenRoleKeyword},
				{Text: `}}`, Role: parser.TokenRoleOperator},
				{Text: `{{`, Role: parser.TokenRoleOperator},
				{Text: `else`, Role: parser.TokenRoleKeyword},
				{Text: `}}`, Role: parser.TokenRoleOperator},
				{Text: `{{`, Role: parser.TokenRoleOperator},
				{Text: `end`, Role: parser.TokenRoleKeyword},
				{Text: `}}`, Role: parser.TokenRoleOperator},
			},
		},
		{
			name: "pipe",
			text: `{{"abcd" | printf "%q"}}`,
			expected: []TokenWithText{
				{Text: `{{`, Role: parser.TokenRoleOperator},
				{Text: `"abcd"`, Role: parser.TokenRoleString},
				{Text: `|`, Role: parser.TokenRoleOperator},
				{Text: `"%q"`, Role: parser.TokenRoleString},
				{Text: `}}`, Role: parser.TokenRoleOperator},
			},
		},
		{
			name: "variable assignment",
			text: `{{with $x := "abc"}}{{$x}}{{end}}`,
			expected: []TokenWithText{
				{Text: `{{`, Role: parser.TokenRoleOperator},
				{Text: `with`, Role: parser.TokenRoleKeyword},
				{Text: `$`, Role: parser.TokenRoleOperator},
				{Text: `:=`, Role: parser.TokenRoleOperator},
				{Text: `"abc"`, Role: parser.TokenRoleString},
				{Text: `}}`, Role: parser.TokenRoleOperator},
				{Text: `{{`, Role: parser.TokenRoleOperator},
				{Text: `$`, Role: parser.TokenRoleOperator},
				{Text: `}}`, Role: parser.TokenRoleOperator},
				{Text: `{{`, Role: parser.TokenRoleOperator},
				{Text: `end`, Role: parser.TokenRoleKeyword},
				{Text: `}}`, Role: parser.TokenRoleOperator},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tokens := ParseTokensWithText(GoTemplateParseFunc(), tc.text)
			assert.Equal(t, tc.expected, tokens)
		})
	}
}
