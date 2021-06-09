package syntax

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/aretext/aretext/syntax/parser"
)

func TestGitRebaseTokenizer(t *testing.T) {
	testCases := []struct {
		name           string
		inputString    string
		expectedTokens []TokenWithText
	}{
		{
			name:        "comment",
			inputString: "# Rebase 5c828d6..bb41094 onto 5c828d6 (1 command)",
			expectedTokens: []TokenWithText{
				{
					Text: "# Rebase 5c828d6..bb41094 onto 5c828d6 (1 command)",
					Role: parser.TokenRoleComment,
				},
			},
		},
		{
			name:        "keyword at start of line",
			inputString: "pick bc51064 Test commit",
			expectedTokens: []TokenWithText{
				{
					Text: "pick",
					Role: parser.TokenRoleKeyword,
				},
				{
					Text: "bc51064",
					Role: parser.TokenRoleWord,
				},
				{
					Text: "Test",
					Role: parser.TokenRoleWord,
				},
				{
					Text: "commit",
					Role: parser.TokenRoleWord,
				},
			},
		},
		{
			name:        "keyword past start of line",
			inputString: "edit reword pick",
			expectedTokens: []TokenWithText{
				{
					Text: "edit",
					Role: parser.TokenRoleKeyword,
				},
				{
					Text: "reword",
					Role: parser.TokenRoleWord,
				},
				{
					Text: "pick",
					Role: parser.TokenRoleWord,
				},
			},
		},
		{
			name:        "keyword prefix of another word",
			inputString: "pi",
			expectedTokens: []TokenWithText{
				{
					Text: "pi",
					Role: parser.TokenRoleWord,
				},
			},
		},
		{
			name:        "comment in commit message",
			inputString: "pick insert # in file",
			expectedTokens: []TokenWithText{
				{
					Text: "pick",
					Role: parser.TokenRoleKeyword,
				},
				{
					Text: "insert",
					Role: parser.TokenRoleWord,
				},
				{
					Text: "#",
					Role: parser.TokenRolePunctuation,
				},
				{
					Text: "in",
					Role: parser.TokenRoleWord,
				},
				{
					Text: "file",
					Role: parser.TokenRoleWord,
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tokens, err := ParseTokensWithText(LanguageGitRebase, tc.inputString)
			require.NoError(t, err)
			assert.Equal(t, tc.expectedTokens, tokens)
		})
	}
}
