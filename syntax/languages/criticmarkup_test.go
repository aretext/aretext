package languages

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/aretext/aretext/syntax/parser"
)

func TestCriticMarkupParseFunc(t *testing.T) {
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
			name: "addition",
			text: "Lorem ipsum dolor{++ sit++} amet",
			expected: []TokenWithText{
				{Text: `{++ sit++}`, Role: criticMarkupAddRole},
			},
		},
		{
			name: "deletion",
			text: "Lorem{-- ipsum--} dolor sit amet",
			expected: []TokenWithText{
				{Text: `{-- ipsum--}`, Role: criticMarkupDelRole},
			},
		},
		{
			name: "deletion with hyphen U+2010",
			text: "Lorem{‐‐ ipsum‐‐} dolor sit amet",
			expected: []TokenWithText{
				{Text: `{‐‐ ipsum‐‐}`, Role: criticMarkupDelRole},
			},
		},
		{
			name: "comment",
			text: "Hello {>>This is a comment<<} world",
			expected: []TokenWithText{
				{Text: `{>>This is a comment<<}`, Role: criticMarkupCommentRole},
			},
		},
		{
			name: "highlight",
			text: "Hello {==world!==}{>> classic intro <<}",
			expected: []TokenWithText{
				{Text: `{==world!==}`, Role: criticMarkupHighlightRole},
				{Text: `{>> classic intro <<}`, Role: criticMarkupCommentRole},
			},
		},
		{
			name: "embed in markdown title",
			text: "# foo {++ bar ++} baz",
			expected: []TokenWithText{
				{Text: "# foo ", Role: markdownHeadingRole},
				{Text: "{++ bar ++}", Role: criticMarkupAddRole},
				{Text: " baz", Role: markdownHeadingRole},
			},
		},
		{
			name: "delete between paragraphs",
			text: "abcd{--\n\ne--}fghi",
			expected: []TokenWithText{
				{Text: "{--\n\ne--}", Role: criticMarkupDelRole},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tokens := ParseTokensWithText(CriticMarkupParseFunc(), tc.text)
			assert.Equal(t, tc.expected, tokens)
		})
	}
}

func TestCriticMarkupConsolidateTokens(t *testing.T) {
	testCases := []struct {
		name               string
		markdownTokens     []parser.ComputedToken
		criticMarkupTokens []parser.ComputedToken
		expected           []parser.ComputedToken
	}{
		{
			name:               "both empty",
			markdownTokens:     nil,
			criticMarkupTokens: nil,
			expected:           nil,
		},
		{
			name: "markdown without criticmarkup",
			markdownTokens: []parser.ComputedToken{
				{Offset: 0, Length: 2},
			},
			criticMarkupTokens: nil,
			expected: []parser.ComputedToken{
				{Offset: 0, Length: 2},
			},
		},
		{
			name:           "criticmarkup without markdown",
			markdownTokens: nil,
			criticMarkupTokens: []parser.ComputedToken{
				{Offset: 0, Length: 2},
			},
			expected: []parser.ComputedToken{
				{Offset: 0, Length: 2},
			},
		},
		{
			name: "non-overlapping, markdown before criticmarkup",
			markdownTokens: []parser.ComputedToken{
				{Offset: 0, Length: 2},
			},
			criticMarkupTokens: []parser.ComputedToken{
				{Offset: 3, Length: 4},
			},
			expected: []parser.ComputedToken{
				{Offset: 0, Length: 2},
				{Offset: 3, Length: 4},
			},
		},
		{
			name: "non-overlapping, criticmarkup before markdown",
			markdownTokens: []parser.ComputedToken{
				{Offset: 3, Length: 4},
			},
			criticMarkupTokens: []parser.ComputedToken{
				{Offset: 0, Length: 2},
			},
			expected: []parser.ComputedToken{
				{Offset: 0, Length: 2},
				{Offset: 3, Length: 4},
			},
		},
		{
			name: "overlap, truncate",
			markdownTokens: []parser.ComputedToken{
				{Offset: 0, Length: 2},
				{Offset: 4, Length: 2},
			},
			criticMarkupTokens: []parser.ComputedToken{
				{Offset: 1, Length: 4},
			},
			expected: []parser.ComputedToken{
				{Offset: 0, Length: 1},
				{Offset: 1, Length: 4},
				{Offset: 5, Length: 1},
			},
		},
		{
			name: "overlap, replace single token aligned",
			markdownTokens: []parser.ComputedToken{
				{Offset: 1, Length: 2, Role: markdownHeadingRole},
			},
			criticMarkupTokens: []parser.ComputedToken{
				{Offset: 1, Length: 2, Role: criticMarkupAddRole},
			},
			expected: []parser.ComputedToken{
				{Offset: 1, Length: 2, Role: criticMarkupAddRole},
			},
		},
		{
			name: "overlap, replace single token misaligned",
			markdownTokens: []parser.ComputedToken{
				{Offset: 1, Length: 2},
			},
			criticMarkupTokens: []parser.ComputedToken{
				{Offset: 0, Length: 4},
			},
			expected: []parser.ComputedToken{
				{Offset: 0, Length: 4},
			},
		},
		{
			name: "overlap, replace multiple tokens",
			markdownTokens: []parser.ComputedToken{
				{Offset: 1, Length: 1},
				{Offset: 2, Length: 1},
				{Offset: 3, Length: 1},
				{Offset: 4, Length: 1},
				{Offset: 5, Length: 1},
			},
			criticMarkupTokens: []parser.ComputedToken{
				{Offset: 2, Length: 3},
			},
			expected: []parser.ComputedToken{
				{Offset: 1, Length: 1},
				{Offset: 2, Length: 3},
				{Offset: 5, Length: 1},
			},
		},
		{
			name: "overlap, split",
			markdownTokens: []parser.ComputedToken{
				{Offset: 0, Length: 4},
			},
			criticMarkupTokens: []parser.ComputedToken{
				{Offset: 1, Length: 1},
			},
			expected: []parser.ComputedToken{
				{Offset: 0, Length: 1},
				{Offset: 1, Length: 1},
				{Offset: 2, Length: 2},
			},
		},
		{
			name: "multiple criticmarkup tokens",
			markdownTokens: []parser.ComputedToken{
				{Offset: 0, Length: 2},
				{Offset: 3, Length: 5},
			},
			criticMarkupTokens: []parser.ComputedToken{
				{Offset: 1, Length: 1},
				{Offset: 2, Length: 2},
			},
			expected: []parser.ComputedToken{
				{Offset: 0, Length: 1},
				{Offset: 1, Length: 1},
				{Offset: 2, Length: 2},
				{Offset: 4, Length: 4},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := criticMarkupConsolidateTokens(tc.markdownTokens, tc.criticMarkupTokens)
			assert.Equal(t, tc.expected, actual)
		})
	}
}
