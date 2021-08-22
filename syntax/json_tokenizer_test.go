package syntax

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/aretext/aretext/syntax/parser"
)

func TestJsonTokenizer(t *testing.T) {
	testCases := []struct {
		name           string
		inputString    string
		expectedTokens []TokenWithText
	}{
		{
			name:        "integer value",
			inputString: `123`,
			expectedTokens: []TokenWithText{
				{Text: `123`, Role: parser.TokenRoleNumber},
			},
		},
		{
			name:        "negative integer value",
			inputString: `-1`,
			expectedTokens: []TokenWithText{
				{Text: `-1`, Role: parser.TokenRoleNumber},
			},
		},
		{
			name:        "zero integer value",
			inputString: `0`,
			expectedTokens: []TokenWithText{
				{Text: `0`, Role: parser.TokenRoleNumber},
			},
		},
		{
			name:        "exponentiated integer, capital e",
			inputString: `123E456`,
			expectedTokens: []TokenWithText{
				{Text: `123E456`, Role: parser.TokenRoleNumber},
			},
		},
		{
			name:        "exponentiated integer, lowercase e",
			inputString: `123e456`,
			expectedTokens: []TokenWithText{
				{Text: `123e456`, Role: parser.TokenRoleNumber},
			},
		},
		{
			name:        "exponentiated integer, negative exponent",
			inputString: "12E-3",
			expectedTokens: []TokenWithText{
				{Text: `12E-3`, Role: parser.TokenRoleNumber},
			},
		},
		{
			name:        "float value zero",
			inputString: `0.0`,
			expectedTokens: []TokenWithText{
				{Text: `0.0`, Role: parser.TokenRoleNumber},
			},
		},
		{
			name:        "float value greater than one",
			inputString: `123.456`,
			expectedTokens: []TokenWithText{
				{Text: `123.456`, Role: parser.TokenRoleNumber},
			},
		},
		{
			name:        "float value less than one",
			inputString: `0.123`,
			expectedTokens: []TokenWithText{
				{Text: `0.123`, Role: parser.TokenRoleNumber},
			},
		},
		{
			name:        "exponentiated float",
			inputString: `123.456E78`,
			expectedTokens: []TokenWithText{
				{Text: `123.456E78`, Role: parser.TokenRoleNumber},
			},
		},
		{
			name:           "number prefix",
			inputString:    "123abc",
			expectedTokens: []TokenWithText{},
		},
		{
			name:           "number suffix",
			inputString:    "abc123",
			expectedTokens: []TokenWithText{},
		},
		{
			name:           "number suffix with underscore",
			inputString:    "abc_123",
			expectedTokens: []TokenWithText{},
		},
		{
			name:           "number prefix starting with hyphen",
			inputString:    "-123abcd",
			expectedTokens: []TokenWithText{},
		},
		{
			name:        "key with string value",
			inputString: `{"key": "abcd"}`,
			expectedTokens: []TokenWithText{
				{Text: `"key":`, Role: parser.TokenRoleCustom1},
				{Text: `"abcd"`, Role: parser.TokenRoleString},
			},
		},
		{
			name:           "incomplete key with escaped quote",
			inputString:    `"key\":`,
			expectedTokens: []TokenWithText{},
		},
		{
			name:        "string with escaped quote",
			inputString: `"abc\"xyz"`,
			expectedTokens: []TokenWithText{
				{Text: `"abc\"xyz"`, Role: parser.TokenRoleString},
			},
		},
		{
			name:        "string ending with escaped backslash",
			inputString: `"abc\\"`,
			expectedTokens: []TokenWithText{
				{Text: `"abc\\"`, Role: parser.TokenRoleString},
			},
		},
		{
			name:        "incomplete string ending with escaped quote",
			inputString: `"abc\" 123`,
			expectedTokens: []TokenWithText{
				{Text: "123", Role: parser.TokenRoleNumber},
			},
		},
		{
			name:           "incomplete string ending with newline before quote",
			inputString:    "\"abc\n\"",
			expectedTokens: []TokenWithText{},
		},
		{
			name:           "string with line break",
			inputString:    "\"abc\nxyz\"",
			expectedTokens: []TokenWithText{},
		},
		{
			name:        "string with escaped line break",
			inputString: `"abc\nxyz"`,
			expectedTokens: []TokenWithText{
				{Text: `"abc\nxyz"`, Role: parser.TokenRoleString},
			},
		},
		{
			name:        "true value",
			inputString: `{"bool": true}`,
			expectedTokens: []TokenWithText{
				{Text: `"bool":`, Role: parser.TokenRoleCustom1},
				{Text: `true`, Role: parser.TokenRoleKeyword},
			},
		},
		{
			name:        "false value",
			inputString: `{"bool": false}`,
			expectedTokens: []TokenWithText{
				{Text: `"bool":`, Role: parser.TokenRoleCustom1},
				{Text: `false`, Role: parser.TokenRoleKeyword},
			},
		},
		{
			name:        "null value",
			inputString: `{"nullable": null}`,
			expectedTokens: []TokenWithText{
				{Text: `"nullable":`, Role: parser.TokenRoleCustom1},
				{Text: `null`, Role: parser.TokenRoleKeyword},
			},
		},
		{
			name:           "keyword prefix",
			inputString:    "nullable",
			expectedTokens: []TokenWithText{},
		},
		{
			name: "object with multiple keys",
			inputString: `{
				"k1": "v1",
				"k2": "v2"
			}`,
			expectedTokens: []TokenWithText{
				{Text: `"k1":`, Role: parser.TokenRoleCustom1},
				{Text: `"v1"`, Role: parser.TokenRoleString},
				{Text: `"k2":`, Role: parser.TokenRoleCustom1},
				{Text: `"v2"`, Role: parser.TokenRoleString},
			},
		},
		{
			name: "object with nested object",
			inputString: `{
				"nested": {
					"k1": 123,
					"k2": 456
				}
			}`,
			expectedTokens: []TokenWithText{
				{Text: `"nested":`, Role: parser.TokenRoleCustom1},
				{Text: `"k1":`, Role: parser.TokenRoleCustom1},
				{Text: `123`, Role: parser.TokenRoleNumber},
				{Text: `"k2":`, Role: parser.TokenRoleCustom1},
				{Text: `456`, Role: parser.TokenRoleNumber},
			},
		},
		{
			name:        "spaces between key and colon",
			inputString: `{"key"      : 1}`,
			expectedTokens: []TokenWithText{
				{Text: `"key"      :`, Role: parser.TokenRoleCustom1},
				{Text: `1`, Role: parser.TokenRoleNumber},
			},
		},
		{
			name:        "tabs between key and colon",
			inputString: "{\"key\"\t\t: 1}",
			expectedTokens: []TokenWithText{
				{Text: "\"key\"\t\t:", Role: parser.TokenRoleCustom1},
				{Text: "1", Role: parser.TokenRoleNumber},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tokens, err := ParseTokensWithText(LanguageJson, tc.inputString)
			require.NoError(t, err)
			assert.Equal(t, tc.expectedTokens, tokens)
		})
	}
}

func BenchmarkJsonTokenizer(b *testing.B) {
	data, err := os.ReadFile("testdata/test.json")
	require.NoError(b, err)
	text := string(data)

	for i := 0; i < b.N; i++ {
		_, err := ParseTokensWithText(LanguageJson, text)
		require.NoError(b, err)
	}
}
