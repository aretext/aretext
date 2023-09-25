package languages

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/aretext/aretext/syntax/parser"
)

func TestJsonParseFunc(t *testing.T) {
	testCases := []struct {
		name     string
		text     string
		expected []TokenWithText
	}{
		{
			name: "integer value",
			text: `123`,
			expected: []TokenWithText{
				{Text: `123`, Role: parser.TokenRoleNumber},
			},
		},
		{
			name: "negative integer value",
			text: `-1`,
			expected: []TokenWithText{
				{Text: `-1`, Role: parser.TokenRoleNumber},
			},
		},
		{
			name: "zero integer value",
			text: `0`,
			expected: []TokenWithText{
				{Text: `0`, Role: parser.TokenRoleNumber},
			},
		},
		{
			name: "exponentiated integer, capital e",
			text: `123E456`,
			expected: []TokenWithText{
				{Text: `123E456`, Role: parser.TokenRoleNumber},
			},
		},
		{
			name: "exponentiated integer, lowercase e",
			text: `123e456`,
			expected: []TokenWithText{
				{Text: `123e456`, Role: parser.TokenRoleNumber},
			},
		},
		{
			name: "exponentiated integer, negative exponent",
			text: "12E-3",
			expected: []TokenWithText{
				{Text: `12E-3`, Role: parser.TokenRoleNumber},
			},
		},
		{
			name: "float value zero",
			text: `0.0`,
			expected: []TokenWithText{
				{Text: `0.0`, Role: parser.TokenRoleNumber},
			},
		},
		{
			name: "float value greater than one",
			text: `123.456`,
			expected: []TokenWithText{
				{Text: `123.456`, Role: parser.TokenRoleNumber},
			},
		},
		{
			name: "float value less than one",
			text: `0.123`,
			expected: []TokenWithText{
				{Text: `0.123`, Role: parser.TokenRoleNumber},
			},
		},
		{
			name: "exponentiated float",
			text: `123.456E78`,
			expected: []TokenWithText{
				{Text: `123.456E78`, Role: parser.TokenRoleNumber},
			},
		},
		{
			name:     "number prefix",
			text:     "123abc",
			expected: []TokenWithText{},
		},
		{
			name:     "number suffix",
			text:     "abc123",
			expected: []TokenWithText{},
		},
		{
			name:     "number suffix with underscore",
			text:     "abc_123",
			expected: []TokenWithText{},
		},
		{
			name:     "number prefix starting with hyphen",
			text:     "-123abcd",
			expected: []TokenWithText{},
		},
		{
			name: "key with string value",
			text: `{"key": "abcd"}`,
			expected: []TokenWithText{
				{Text: `"key":`, Role: jsonTokenRoleKey},
				{Text: `"abcd"`, Role: parser.TokenRoleString},
			},
		},
		{
			name:     "incomplete key with escaped quote",
			text:     `"key\":`,
			expected: []TokenWithText{},
		},
		{
			name: "string with escaped quote",
			text: `"abc\"xyz"`,
			expected: []TokenWithText{
				{Text: `"abc\"xyz"`, Role: parser.TokenRoleString},
			},
		},
		{
			name: "string ending with escaped backslash",
			text: `"abc\\"`,
			expected: []TokenWithText{
				{Text: `"abc\\"`, Role: parser.TokenRoleString},
			},
		},
		{
			name: "incomplete string ending with escaped quote",
			text: `"abc\" 123`,
			expected: []TokenWithText{
				{Text: "123", Role: parser.TokenRoleNumber},
			},
		},
		{
			name:     "incomplete string ending with newline before quote",
			text:     "\"abc\n\"",
			expected: []TokenWithText{},
		},
		{
			name:     "string with line break",
			text:     "\"abc\nxyz\"",
			expected: []TokenWithText{},
		},
		{
			name: "string with escaped line break",
			text: `"abc\nxyz"`,
			expected: []TokenWithText{
				{Text: `"abc\nxyz"`, Role: parser.TokenRoleString},
			},
		},
		{
			name: "true value",
			text: `{"bool": true}`,
			expected: []TokenWithText{
				{Text: `"bool":`, Role: jsonTokenRoleKey},
				{Text: `true`, Role: parser.TokenRoleKeyword},
			},
		},
		{
			name: "false value",
			text: `{"bool": false}`,
			expected: []TokenWithText{
				{Text: `"bool":`, Role: jsonTokenRoleKey},
				{Text: `false`, Role: parser.TokenRoleKeyword},
			},
		},
		{
			name: "null value",
			text: `{"nullable": null}`,
			expected: []TokenWithText{
				{Text: `"nullable":`, Role: jsonTokenRoleKey},
				{Text: `null`, Role: parser.TokenRoleKeyword},
			},
		},
		{
			name:     "keyword prefix",
			text:     "nullable",
			expected: []TokenWithText{},
		},
		{
			name: "object with multiple keys",
			text: `{
				"k1": "v1",
				"k2": "v2"
			}`,
			expected: []TokenWithText{
				{Text: `"k1":`, Role: jsonTokenRoleKey},
				{Text: `"v1"`, Role: parser.TokenRoleString},
				{Text: `"k2":`, Role: jsonTokenRoleKey},
				{Text: `"v2"`, Role: parser.TokenRoleString},
			},
		},
		{
			name: "object with nested object",
			text: `{
				"nested": {
					"k1": 123,
					"k2": 456
				}
			}`,
			expected: []TokenWithText{
				{Text: `"nested":`, Role: jsonTokenRoleKey},
				{Text: `"k1":`, Role: jsonTokenRoleKey},
				{Text: `123`, Role: parser.TokenRoleNumber},
				{Text: `"k2":`, Role: jsonTokenRoleKey},
				{Text: `456`, Role: parser.TokenRoleNumber},
			},
		},
		{
			name: "spaces between key and colon",
			text: `{"key"      : 1}`,
			expected: []TokenWithText{
				{Text: `"key"      :`, Role: jsonTokenRoleKey},
				{Text: `1`, Role: parser.TokenRoleNumber},
			},
		},
		{
			name: "tabs between key and colon",
			text: "{\"key\"\t\t: 1}",
			expected: []TokenWithText{
				{Text: "\"key\"\t\t:", Role: jsonTokenRoleKey},
				{Text: "1", Role: parser.TokenRoleNumber},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tokens := ParseTokensWithText(JsonParseFunc(), tc.text)
			assert.Equal(t, tc.expected, tokens)
		})
	}
}

func BenchmarkJsonParser(b *testing.B) {
	BenchmarkParser(b, JsonParseFunc(), "testdata/json/test.json")
}
