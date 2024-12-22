package languages

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/aretext/aretext/editor/syntax/parser"
)

type markdownTest struct {
	name       string
	text       string
	expected   []TokenWithText
	skipReason string
}

func TestMarkdownParseFunc(t *testing.T) {
	testCases, err := loadCommonmarkTests()
	require.NoError(t, err)

	testCases = append(testCases, []markdownTest{
		{
			name: "fenced code block with underline",
			text: "```\n  ---\n```",
			expected: []TokenWithText{
				{
					Role: markdownCodeBlockRole,
					Text: "```\n  ---\n```",
				},
			},
		},
		{
			name: "fenced code block with trailing CRLF",
			text: "```\ntest\n```\r\nabcd",
			expected: []TokenWithText{
				{
					Role: markdownCodeBlockRole,
					Text: "```\ntest\n```\r\n",
				},
			},
		},
		{
			name: "fenced code block in emphasis",
			text: "*foo `code` bar*",
			expected: []TokenWithText{
				{
					Role: markdownEmphasisRole,
					Text: "*foo `code` bar*",
				},
			},
		},
		{
			name:     "paragraph followed by list with single hyphen and eof",
			text:     "foobar\n-",
			expected: []TokenWithText{},
		},
		{
			name:     "paragraph followed by list with single hyphen and newline",
			text:     "foobar\n-\n",
			expected: []TokenWithText{},
		},
		{
			name: "paragraph followed by list with single hyphen empty list",
			text: "foobar\n- \n",
			expected: []TokenWithText{
				{
					Role: markdownListBulletRole,
					Text: "-",
				},
			},
		},
	}...)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.skipReason != "" {
				t.Skip(tc.skipReason)
			}

			tokens := ParseTokensWithText(MarkdownParseFunc(), tc.text)
			assert.Equal(t, tc.expected, tokens)
		})
	}
}

func FuzzMarkdownParseFunc(f *testing.F) {
	testCases, err := loadCommonmarkTests()
	if err != nil {
		f.Fatalf("Could not load markdown test seeds: %s", err)
	}

	seeds := make([]string, 0, len(testCases))
	for _, tc := range testCases {
		seeds = append(seeds, tc.text)
	}

	FuzzParser(f, MarkdownParseFunc(), seeds)
}

func loadCommonmarkTests() ([]markdownTest, error) {
	type commonmarkToken struct {
		Role  string
		Start int
		End   int
		Text  string
	}

	type commonmarkTest struct {
		Name       string
		Markdown   string
		Tokens     []commonmarkToken
		SkipReason string
	}

	data, err := os.ReadFile("testdata/markdown/commonmark_0.3_tests.json")
	if err != nil {
		return nil, fmt.Errorf("os.ReadFile: %w", err)
	}

	var cmTests []commonmarkTest
	if err := json.Unmarshal(data, &cmTests); err != nil {
		return nil, fmt.Errorf("json.Unmarshal: %w", err)
	}

	mdTests := make([]markdownTest, 0, len(cmTests))
	for _, cmt := range cmTests {
		mdt := markdownTest{
			name:       cmt.Name,
			text:       cmt.Markdown,
			expected:   make([]TokenWithText, 0, len(cmt.Tokens)),
			skipReason: cmt.SkipReason,
		}

		for _, tok := range cmt.Tokens {
			var role parser.TokenRole
			switch tok.Role {
			case "CodeBlock":
				role = markdownCodeBlockRole
			case "CodeSpan":
				role = markdownCodeSpanRole
			case "Emphasis":
				role = markdownEmphasisRole
			case "StrongEmphasis":
				role = markdownStrongEmphasisRole
			case "Heading":
				role = markdownHeadingRole
			case "Link":
				role = markdownLinkRole
			case "ListNumber":
				role = markdownListNumberRole
			case "ListBullet":
				role = markdownListBulletRole
			case "ThematicBreak":
				role = markdownThematicBreakRole
			default:
				return nil, fmt.Errorf("Unrecognized role %s\n", tok.Role)
			}
			mdt.expected = append(mdt.expected, TokenWithText{
				Role: role,
				Text: tok.Text,
			})
		}

		mdTests = append(mdTests, mdt)
	}

	return mdTests, nil
}
