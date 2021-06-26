package syntax

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/aretext/aretext/syntax/parser"
)

func TestDevlogTokenizer(t *testing.T) {
	const toDoRole = parser.TokenRoleCustom1
	const inProgressRole = parser.TokenRoleCustom2
	const completedRole = parser.TokenRoleCustom3
	const blockedRole = parser.TokenRoleCustom4
	const codeBlockRole = parser.TokenRoleCustom5
	const tildeSeparatorRole = parser.TokenRoleCustom6

	testCases := []struct {
		name           string
		inputString    string
		expectedTokens []TokenWithText
	}{
		{
			name: "tasks in different states",
			inputString: strings.Join([]string{
				`^ in progress`,
				`+ completed`,
				`- blocked`,
				`* to do`,
			}, "\n"),
			expectedTokens: []TokenWithText{
				{Text: "^ in progress", Role: inProgressRole},
				{Text: "\n+ completed", Role: completedRole},
				{Text: "\n- blocked", Role: blockedRole},
				{Text: "\n*", Role: toDoRole},
				{Text: "to", Role: parser.TokenRoleWord},
				{Text: "do", Role: parser.TokenRoleWord},
			},
		},
		{
			name: "code block",
			inputString: strings.Join([]string{
				"before",
				"``` code`",
				"``block ```",
				"after",
			}, "\n"),
			expectedTokens: []TokenWithText{
				{Text: "before", Role: parser.TokenRoleWord},
				{Text: "``` code`\n``block ```", Role: codeBlockRole},
				{Text: "after", Role: parser.TokenRoleWord},
			},
		},
		{
			name: "tilde separator",
			inputString: strings.Join([]string{
				"before",
				"~~~",
				"after",
			}, "\n"),
			expectedTokens: []TokenWithText{
				{Text: "before", Role: parser.TokenRoleWord},
				{Text: "\n~~~", Role: tildeSeparatorRole},
				{Text: "after", Role: parser.TokenRoleWord},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tokens, err := ParseTokensWithText(LanguageDevlog, tc.inputString)
			require.NoError(t, err)
			assert.Equal(t, tc.expectedTokens, tokens)
		})
	}
}
