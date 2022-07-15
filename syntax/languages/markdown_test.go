package languages

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/aretext/aretext/syntax/parser"
	"github.com/aretext/aretext/text"
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

	testCases = append(testCases, markdownTest{
		name: "fenced code block with underline",
		text: "```\n  ---\n```",
		expected: []TokenWithText{
			{
				Role: markdownCodeBlockRole,
				Text: "```\n  ---\n```",
			},
		},
	})

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

	for _, tc := range testCases {
		f.Add(tc.text)
	}

	parseFunc := MarkdownParseFunc()
	f.Fuzz(func(t *testing.T, data string) {
		p := parser.New(parseFunc)
		tree, err := text.NewTreeFromString(data)
		if errors.Is(err, text.InvalidUtf8Error) {
			t.Skip()
		}
		require.NoError(t, err)
		p.ParseAll(tree)
		p.TokensIntersectingRange(0, math.MaxUint64)
	})
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
		return nil, errors.Wrapf(err, "os.ReadFile")
	}

	var cmTests []commonmarkTest
	if err := json.Unmarshal(data, &cmTests); err != nil {
		return nil, errors.Wrapf(err, "json.Unmarshal")
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
