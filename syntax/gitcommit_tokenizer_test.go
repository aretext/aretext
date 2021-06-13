package syntax

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/aretext/aretext/syntax/parser"
)

func TestGitCommitTokenizer(t *testing.T) {
	inputString := `
This is a test commit
# Please enter the commit message for your changes. Lines starting
# with '#' will be ignored, and an empty message aborts the commit.
#
# On branch testbranch
# Changes to be committed:
#   new file:   test
#
# Untracked files:
#   sometest.go
#
`
	expectedTokens := []TokenWithText{
		{
			Text: "This",
			Role: parser.TokenRoleWord,
		},
		{
			Text: "is",
			Role: parser.TokenRoleWord,
		},
		{
			Text: "a",
			Role: parser.TokenRoleWord,
		},
		{
			Text: "test",
			Role: parser.TokenRoleWord,
		},
		{
			Text: "commit",
			Role: parser.TokenRoleWord,
		},
		{
			Text: "#",
			Role: parser.TokenRoleCommentDelimiter,
		},
		{
			Text: " Please enter the commit message for your changes. Lines starting",
			Role: parser.TokenRoleComment,
		},
		{
			Text: "#",
			Role: parser.TokenRoleCommentDelimiter,
		},
		{
			Text: " with '#' will be ignored, and an empty message aborts the commit.",
			Role: parser.TokenRoleComment,
		},
		{
			Text: "#",
			Role: parser.TokenRoleCommentDelimiter,
		},
		{
			Text: "#",
			Role: parser.TokenRoleCommentDelimiter,
		},
		{
			Text: " On branch testbranch",
			Role: parser.TokenRoleComment,
		},
		{
			Text: "#",
			Role: parser.TokenRoleCommentDelimiter,
		},
		{
			Text: " Changes to be committed:",
			Role: parser.TokenRoleComment,
		},
		{
			Text: "#",
			Role: parser.TokenRoleCommentDelimiter,
		},
		{
			Text: "   new file:   test",
			Role: parser.TokenRoleComment,
		},
		{
			Text: "#",
			Role: parser.TokenRoleCommentDelimiter,
		},
		{
			Text: "#",
			Role: parser.TokenRoleCommentDelimiter,
		},
		{
			Text: " Untracked files:",
			Role: parser.TokenRoleComment,
		},
		{
			Text: "#",
			Role: parser.TokenRoleCommentDelimiter,
		},
		{
			Text: "   sometest.go",
			Role: parser.TokenRoleComment,
		},
		{
			Text: "#",
			Role: parser.TokenRoleCommentDelimiter,
		},
	}

	tokens, err := ParseTokensWithText(LanguageGitCommit, inputString)
	require.NoError(t, err)
	assert.Equal(t, expectedTokens, tokens)
}