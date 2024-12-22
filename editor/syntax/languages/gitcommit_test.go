package languages

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/aretext/aretext/editor/syntax/parser"
)

func TestGitCommitParseFunc(t *testing.T) {
	testCases := []struct {
		name     string
		text     string
		expected []TokenWithText
	}{
		{
			name:     "empty",
			text:     "",
			expected: []TokenWithText{},
		},
		{
			name:     "single line, not a comment",
			text:     "this is a commit",
			expected: []TokenWithText{},
		},
		{
			name: "single line, comment",
			text: "# this is a comment",
			expected: []TokenWithText{
				{
					Role: parser.TokenRoleComment,
					Text: "# this is a comment",
				},
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
			name: "multiple lines, mix of comments and non-comments",
			text: "this is a commit\n# this is a comment\n#\n# end",
			expected: []TokenWithText{
				{
					Role: parser.TokenRoleComment,
					Text: "# this is a comment\n",
				},
				{
					Role: parser.TokenRoleComment,
					Text: "#\n",
				},
				{
					Role: parser.TokenRoleComment,
					Text: "# end",
				},
			},
		},
		{
			name:     "non-comment with hash symbol",
			text:     "Commit msg with a '#'",
			expected: []TokenWithText{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tokens := ParseTokensWithText(GitCommitParseFunc(), tc.text)
			assert.Equal(t, tc.expected, tokens)
		})
	}
}
