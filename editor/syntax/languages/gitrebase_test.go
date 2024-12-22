package languages

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/aretext/aretext/editor/syntax/parser"
)

func TestGitRebaseParseFunc(t *testing.T) {
	testCases := []struct {
		name     string
		text     string
		expected []TokenWithText
	}{
		{
			name: "comment",
			text: "# Rebase 5c828d6..bb41094 onto 5c828d6 (1 command)",
			expected: []TokenWithText{
				{
					Text: "# Rebase 5c828d6..bb41094 onto 5c828d6 (1 command)",
					Role: parser.TokenRoleComment,
				},
			},
		},
		{
			name: "keyword at start of line",
			text: "pick bc51064 Test commit",
			expected: []TokenWithText{
				{
					Text: "pick",
					Role: parser.TokenRoleKeyword,
				},
			},
		},
		{
			name: "keyword past start of line",
			text: "edit reword pick",
			expected: []TokenWithText{
				{
					Text: "edit",
					Role: parser.TokenRoleKeyword,
				},
			},
		},
		{
			name:     "keyword prefix of another word",
			text:     "pi",
			expected: []TokenWithText{},
		},
		{
			name: "keyword after newline",
			text: "\nreword test",
			expected: []TokenWithText{
				{
					Text: "reword",
					Role: parser.TokenRoleKeyword,
				},
			},
		},
		{
			name: "comment in commit message",
			text: "pick insert # in file",
			expected: []TokenWithText{
				{
					Text: "pick",
					Role: parser.TokenRoleKeyword,
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tokens := ParseTokensWithText(GitRebaseParseFunc(), tc.text)
			assert.Equal(t, tc.expected, tokens)
		})
	}
}
