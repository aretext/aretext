package syntax

import (
	"io/ioutil"
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
			name:        "number prefix",
			inputString: "123abc",
			expectedTokens: []TokenWithText{
				{Text: `123abc`, Role: parser.TokenRoleWord},
			},
		},
		{
			name:        "number suffix",
			inputString: "abc123",
			expectedTokens: []TokenWithText{
				{Text: `abc123`, Role: parser.TokenRoleWord},
			},
		},
		{
			name:        "number suffix with underscore",
			inputString: "abc_123",
			expectedTokens: []TokenWithText{
				{Text: `abc`, Role: parser.TokenRoleWord},
				{Text: `_`, Role: parser.TokenRolePunctuation},
				{Text: `123`, Role: parser.TokenRoleWord},
			},
		},
		{
			name:        "number prefix starting with hyphen",
			inputString: "-123abcd",
			expectedTokens: []TokenWithText{
				{Text: `-`, Role: parser.TokenRolePunctuation},
				{Text: `123abcd`, Role: parser.TokenRoleWord},
			},
		},
		{
			name:        "key with string value",
			inputString: `{"key": "abcd"}`,
			expectedTokens: []TokenWithText{
				{Text: `{`, Role: parser.TokenRolePunctuation},
				{Text: `"key":`, Role: parser.TokenRoleCustom1},
				{Text: `"`, Role: parser.TokenRoleStringQuote},
				{Text: `abcd`, Role: parser.TokenRoleString},
				{Text: `"`, Role: parser.TokenRoleStringQuote},
				{Text: `}`, Role: parser.TokenRolePunctuation},
			},
		},
		{
			name:        "incomplete key with escaped quote",
			inputString: `"key\":`,
			expectedTokens: []TokenWithText{
				{Text: "key", Role: parser.TokenRoleWord},
			},
		},
		{
			name:        "string with escaped quote",
			inputString: `"abc\"xyz"`,
			expectedTokens: []TokenWithText{
				{Text: `"`, Role: parser.TokenRoleStringQuote},
				{Text: `abc\"xyz`, Role: parser.TokenRoleString},
				{Text: `"`, Role: parser.TokenRoleStringQuote},
			},
		},
		{
			name:        "incomplete string ending with escaped quote",
			inputString: `"abc\" 123`,
			expectedTokens: []TokenWithText{
				{Text: "abc", Role: parser.TokenRoleWord},
				{Text: "123", Role: parser.TokenRoleNumber},
			},
		},
		{
			name:        "incomplete string ending with newline before quote",
			inputString: "\"abc\n\"",
			expectedTokens: []TokenWithText{
				{Text: "abc", Role: parser.TokenRoleWord},
			},
		},
		{
			name:        "string with line break",
			inputString: "\"abc\nxyz\"",
			expectedTokens: []TokenWithText{
				{Text: "abc", Role: parser.TokenRoleWord},
				{Text: "xyz", Role: parser.TokenRoleWord},
			},
		},
		{
			name:        "string with escaped line break",
			inputString: `"abc\nxyz"`,
			expectedTokens: []TokenWithText{
				{Text: `"`, Role: parser.TokenRoleStringQuote},
				{Text: `abc\nxyz`, Role: parser.TokenRoleString},
				{Text: `"`, Role: parser.TokenRoleStringQuote},
			},
		},
		{
			name:        "true value",
			inputString: `{"bool": true}`,
			expectedTokens: []TokenWithText{
				{Text: `{`, Role: parser.TokenRolePunctuation},
				{Text: `"bool":`, Role: parser.TokenRoleCustom1},
				{Text: `true`, Role: parser.TokenRoleKeyword},
				{Text: `}`, Role: parser.TokenRolePunctuation},
			},
		},
		{
			name:        "false value",
			inputString: `{"bool": false}`,
			expectedTokens: []TokenWithText{
				{Text: `{`, Role: parser.TokenRolePunctuation},
				{Text: `"bool":`, Role: parser.TokenRoleCustom1},
				{Text: `false`, Role: parser.TokenRoleKeyword},
				{Text: `}`, Role: parser.TokenRolePunctuation},
			},
		},
		{
			name:        "null value",
			inputString: `{"nullable": null}`,
			expectedTokens: []TokenWithText{
				{Text: `{`, Role: parser.TokenRolePunctuation},
				{Text: `"nullable":`, Role: parser.TokenRoleCustom1},
				{Text: `null`, Role: parser.TokenRoleKeyword},
				{Text: `}`, Role: parser.TokenRolePunctuation},
			},
		},
		{
			name:        "keyword prefix",
			inputString: "nullable",
			expectedTokens: []TokenWithText{
				{Text: `nullable`, Role: parser.TokenRoleWord},
			},
		},
		{
			name: "object with multiple keys",
			inputString: `{
				"k1": "v1",
				"k2": "v2"
			}`,
			expectedTokens: []TokenWithText{
				{Text: `{`, Role: parser.TokenRolePunctuation},
				{Text: `"k1":`, Role: parser.TokenRoleCustom1},
				{Text: `"`, Role: parser.TokenRoleStringQuote},
				{Text: `v1`, Role: parser.TokenRoleString},
				{Text: `"`, Role: parser.TokenRoleStringQuote},
				{Text: `,`, Role: parser.TokenRolePunctuation},
				{Text: `"k2":`, Role: parser.TokenRoleCustom1},
				{Text: `"`, Role: parser.TokenRoleStringQuote},
				{Text: `v2`, Role: parser.TokenRoleString},
				{Text: `"`, Role: parser.TokenRoleStringQuote},
				{Text: `}`, Role: parser.TokenRolePunctuation},
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
				{Text: `{`, Role: parser.TokenRolePunctuation},
				{Text: `"nested":`, Role: parser.TokenRoleCustom1},
				{Text: `{`, Role: parser.TokenRolePunctuation},
				{Text: `"k1":`, Role: parser.TokenRoleCustom1},
				{Text: `123`, Role: parser.TokenRoleNumber},
				{Text: `,`, Role: parser.TokenRolePunctuation},
				{Text: `"k2":`, Role: parser.TokenRoleCustom1},
				{Text: `456`, Role: parser.TokenRoleNumber},
				{Text: `}`, Role: parser.TokenRolePunctuation},
				{Text: `}`, Role: parser.TokenRolePunctuation},
			},
		},
		{
			name:        "spaces between key and colon",
			inputString: `{"key"      : 1}`,
			expectedTokens: []TokenWithText{
				{Text: `{`, Role: parser.TokenRolePunctuation},
				{Text: `"key"      :`, Role: parser.TokenRoleCustom1},
				{Text: `1`, Role: parser.TokenRoleNumber},
				{Text: `}`, Role: parser.TokenRolePunctuation},
			},
		},
		{
			name:        "tabs between key and colon",
			inputString: "{\"key\"\t\t: 1}",
			expectedTokens: []TokenWithText{
				{Text: `{`, Role: parser.TokenRolePunctuation},
				{Text: "\"key\"\t\t:", Role: parser.TokenRoleCustom1},
				{Text: "1", Role: parser.TokenRoleNumber},
				{Text: `}`, Role: parser.TokenRolePunctuation},
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
	data, err := ioutil.ReadFile("testdata/test.json")
	require.NoError(b, err)
	text := string(data)

	for i := 0; i < b.N; i++ {
		_, err := ParseTokensWithText(LanguageJson, text)
		require.NoError(b, err)
	}
}
