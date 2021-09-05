package languages

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/aretext/aretext/syntax2/parser"
)

func TestDevlogParseFunc(t *testing.T) {
	const toDoRole = parser.TokenRoleCustom1
	const inProgressRole = parser.TokenRoleCustom2
	const completedRole = parser.TokenRoleCustom3
	const blockedRole = parser.TokenRoleCustom4
	const codeBlockRole = parser.TokenRoleCustom5
	const tildeSeparatorRole = parser.TokenRoleCustom6

	testCases := []struct {
		name     string
		text     string
		expected []TokenWithText
	}{
		{
			name: "tasks in different states",
			text: strings.Join([]string{
				`^ in progress`,
				`+ completed`,
				`- blocked`,
				`* to do`,
			}, "\n"),
			expected: []TokenWithText{
				{Text: "^ in progress\n", Role: inProgressRole},
				{Text: "+ completed\n", Role: completedRole},
				{Text: "- blocked\n", Role: blockedRole},
				{Text: "*", Role: toDoRole},
			},
		},
		{
			name: "code block",
			text: strings.Join([]string{
				"before",
				"``` code`",
				"``block ```",
				"after",
			}, "\n"),
			expected: []TokenWithText{
				{Text: "``` code`\n``block ```", Role: codeBlockRole},
			},
		},
		{
			name: "tilde separator",
			text: strings.Join([]string{
				"before",
				"~~~",
				"after",
			}, "\n"),
			expected: []TokenWithText{
				{Text: "~~~", Role: tildeSeparatorRole},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tokens := ParseTokensWithText(DevlogParseFunc(), tc.text)
			assert.Equal(t, tc.expected, tokens)
		})
	}
}
