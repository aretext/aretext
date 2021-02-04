package parser

import (
	"testing"
	"unicode"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseRegexp(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected Regexp
	}{
		{
			name:     "single character",
			input:    "a",
			expected: regexpChar{char: 'a'},
		},
		{
			name:     "newline",
			input:    "\n",
			expected: regexpChar{char: '\n'},
		},
		{
			name:  "non-ascii unicode",
			input: "\u2603",
			expected: regexpConcat{
				left: regexpConcat{
					left:  regexpChar{char: 0xe2},
					right: regexpChar{char: 0x98},
				},
				right: regexpChar{char: 0x83},
			},
		},
		{
			name:     "escaped quote",
			input:    "\"",
			expected: regexpChar{char: '"'},
		},
		{
			name:     "escaped star",
			input:    `\*`,
			expected: regexpChar{char: '*'},
		},
		{
			name:     "escaped plus",
			input:    `\+`,
			expected: regexpChar{char: '+'},
		},
		{
			name:     "escaped left paren",
			input:    `\(`,
			expected: regexpChar{char: '('},
		},
		{
			name:     "escaped right paren",
			input:    `\)`,
			expected: regexpChar{char: ')'},
		},
		{
			name:     "escaped backslash",
			input:    `\\`,
			expected: regexpChar{char: '\\'},
		},
		{
			name:     "escaped pipe",
			input:    `\|`,
			expected: regexpChar{char: '|'},
		},
		{
			name:     "escape question mark",
			input:    `\?`,
			expected: regexpChar{char: '?'},
		},
		{
			name:     "escape left bracket",
			input:    `\[`,
			expected: regexpChar{char: '['},
		},
		{
			name:     "escape right bracket",
			input:    `\]`,
			expected: regexpChar{char: ']'},
		},
		{
			name:     "escape dot",
			input:    `\.`,
			expected: regexpChar{char: '.'},
		},
		{
			name:     "escape carat",
			input:    `\^`,
			expected: regexpChar{char: '^'},
		},
		{
			name:     "escape dollar sign",
			input:    `\$`,
			expected: regexpChar{char: '$'},
		},
		{
			name:     "escape line feed",
			input:    `\n`,
			expected: regexpChar{char: '\n'},
		},
		{
			name:     "escape tab",
			input:    `\t`,
			expected: regexpChar{char: '\t'},
		},
		{
			name:     "escape form feed",
			input:    `\f`,
			expected: regexpChar{char: '\f'},
		},
		{
			name:     "escape carriage return",
			input:    `\r`,
			expected: regexpChar{char: '\r'},
		},
		{
			name:  "concatenate two characters",
			input: "ab",
			expected: regexpConcat{
				left:  regexpChar{char: 'a'},
				right: regexpChar{char: 'b'},
			},
		},
		{
			name:  "concatenate three characters",
			input: "abc",
			expected: regexpConcat{
				left: regexpConcat{
					left:  regexpChar{char: 'a'},
					right: regexpChar{char: 'b'},
				},
				right: regexpChar{char: 'c'},
			},
		},
		{
			name:  "union two characters",
			input: "a|b",
			expected: regexpUnion{
				left:  regexpChar{char: 'a'},
				right: regexpChar{char: 'b'},
			},
		},
		{
			name:  "union three characters",
			input: "a|b|c",
			expected: regexpUnion{
				left: regexpChar{char: 'a'},
				right: regexpUnion{
					left:  regexpChar{char: 'b'},
					right: regexpChar{char: 'c'},
				},
			},
		},
		{
			name:  "union of two concatenations",
			input: "ab|cd",
			expected: regexpUnion{
				left: regexpConcat{
					left:  regexpChar{'a'},
					right: regexpChar{'b'},
				},
				right: regexpConcat{
					left:  regexpChar{'c'},
					right: regexpChar{'d'},
				},
			},
		},
		{
			name:  "union of three concatenations",
			input: "ab|cd|ef",
			expected: regexpUnion{
				left: regexpConcat{
					left:  regexpChar{'a'},
					right: regexpChar{'b'},
				},
				right: regexpUnion{
					left: regexpConcat{
						left:  regexpChar{'c'},
						right: regexpChar{'d'},
					},
					right: regexpConcat{
						left:  regexpChar{'e'},
						right: regexpChar{'f'},
					},
				},
			},
		},
		{
			name:  "star single character",
			input: "a*",
			expected: regexpStar{
				child: regexpChar{char: 'a'},
			},
		},
		{
			name:  "star after concatenation",
			input: "ab*",
			expected: regexpConcat{
				left: regexpChar{char: 'a'},
				right: regexpStar{
					child: regexpChar{char: 'b'},
				},
			},
		},
		{
			name:  "star after union",
			input: "a|b*",
			expected: regexpUnion{
				left: regexpChar{char: 'a'},
				right: regexpStar{
					child: regexpChar{char: 'b'},
				},
			},
		},
		{
			name:  "star after paren expression",
			input: "(a)*",
			expected: regexpStar{
				child: regexpParenExpr{
					child: regexpChar{char: 'a'},
				},
			},
		},
		{
			name:  "star after paren expression with concat",
			input: "(ab)*",
			expected: regexpStar{
				child: regexpParenExpr{
					child: regexpConcat{
						left:  regexpChar{char: 'a'},
						right: regexpChar{char: 'b'},
					},
				},
			},
		},
		{
			name:  "plus single char",
			input: "a+",
			expected: regexpConcat{
				left: regexpChar{char: 'a'},
				right: regexpStar{
					child: regexpChar{char: 'a'},
				},
			},
		},
		{
			name:  "plus after concatenation",
			input: "ab+",
			expected: regexpConcat{
				left: regexpChar{char: 'a'},
				right: regexpConcat{
					left: regexpChar{char: 'b'},
					right: regexpStar{
						child: regexpChar{char: 'b'},
					},
				},
			},
		},
		{
			name:  "plus after paren expression",
			input: "(ab)+",
			expected: regexpConcat{
				left: regexpParenExpr{
					child: regexpConcat{
						left:  regexpChar{char: 'a'},
						right: regexpChar{char: 'b'},
					},
				},
				right: regexpStar{
					child: regexpParenExpr{
						child: regexpConcat{
							left:  regexpChar{char: 'a'},
							right: regexpChar{char: 'b'},
						},
					},
				},
			},
		},
		{
			name:  "question mark after single char",
			input: "a?",
			expected: regexpUnion{
				left:  regexpEmpty{},
				right: regexpChar{char: 'a'},
			},
		},
		{
			name:  "question mark after concatenation",
			input: "ab?",
			expected: regexpConcat{
				left: regexpChar{char: 'a'},
				right: regexpUnion{
					left:  regexpEmpty{},
					right: regexpChar{char: 'b'},
				},
			},
		},
		{
			name:  "question mark after paren",
			input: "(a)?",
			expected: regexpUnion{
				left: regexpEmpty{},
				right: regexpParenExpr{
					child: regexpChar{char: 'a'},
				},
			},
		},
		{
			name:     "empty paren expression",
			input:    "()",
			expected: regexpParenExpr{child: regexpEmpty{}},
		},
		{
			name:  "single paren expression",
			input: "(a)",
			expected: regexpParenExpr{
				child: regexpChar{char: 'a'},
			},
		},
		{
			name:  "concatenated paren expressions",
			input: "(a)(b)",
			expected: regexpConcat{
				left: regexpParenExpr{
					child: regexpChar{char: 'a'},
				},
				right: regexpParenExpr{
					child: regexpChar{char: 'b'},
				},
			},
		},
		{
			name:  "paren expr union",
			input: "(a|b|c)",
			expected: regexpParenExpr{
				child: regexpUnion{
					left: regexpChar{char: 'a'},
					right: regexpUnion{
						left:  regexpChar{char: 'b'},
						right: regexpChar{char: 'c'},
					},
				},
			},
		},
		{
			name:  "nested paren expression",
			input: "((a))",
			expected: regexpParenExpr{
				child: regexpParenExpr{
					child: regexpChar{char: 'a'},
				},
			},
		},
		{
			name:  "nested paren expression with concatenation",
			input: "((a)b(c))",
			expected: regexpParenExpr{
				child: regexpConcat{
					left: regexpConcat{
						left: regexpParenExpr{
							child: regexpChar{char: 'a'},
						},
						right: regexpChar{char: 'b'},
					},
					right: regexpParenExpr{
						child: regexpChar{char: 'c'},
					},
				},
			},
		},
		{
			name:  "character class with single character",
			input: "[a]",
			expected: regexpCharClass{
				chars: []byte{'a'},
			},
		},
		{
			name:  "character class with multiple characters",
			input: "[ab]",
			expected: regexpCharClass{
				chars: []byte{'a', 'b'},
			},
		},
		{
			name:  "negated character class with single character",
			input: "[^a]",
			expected: regexpCharClass{
				negated: true,
				chars:   []byte{'a'},
			},
		},
		{
			name:  "character class with escape carat",
			input: `[\^]`,
			expected: regexpCharClass{
				chars: []byte{'^'},
			},
		},
		{
			name:  "character class with escape right bracket",
			input: `[\]]`,
			expected: regexpCharClass{
				chars: []byte{']'},
			},
		},
		{
			name:  "character class with escape backslash",
			input: `[\\]`,
			expected: regexpCharClass{
				chars: []byte{'\\'},
			},
		},
		{
			name:  "negated character class with multiple characters",
			input: "[^ab]",
			expected: regexpCharClass{
				negated: true,
				chars:   []byte{'a', 'b'},
			},
		},
		{
			name:  "character class range",
			input: "[a-c]",
			expected: regexpCharClass{
				chars: []byte{'a', 'b', 'c'},
			},
		},
		{
			name:  "multiple character class ranges",
			input: "[a-cD-F]",
			expected: regexpCharClass{
				chars: []byte{'a', 'b', 'c', 'D', 'E', 'F'},
			},
		},
		{
			name:  "negated character class",
			input: "[^a-c]",
			expected: regexpCharClass{
				negated: true,
				chars:   []byte{'a', 'b', 'c'},
			},
		},
		{
			name:  "character class with ranges and individual chars",
			input: "[Ab-d]",
			expected: regexpCharClass{
				chars: []byte{'A', 'b', 'c', 'd'},
			},
		},
		{
			name:     "dot character class",
			input:    ".",
			expected: regexpCharClass{negated: true},
		},
		{
			name:  "dot character class in concat",
			input: "a.b",
			expected: regexpConcat{
				left: regexpConcat{
					left:  regexpChar{'a'},
					right: regexpCharClass{negated: true},
				},
				right: regexpChar{'b'},
			},
		},
		{
			name:     "unicode letter char class",
			input:    `\p{L}`,
			expected: regexpUnicodeCategory{rangeTable: unicode.Letter},
		},
		{
			name:     "unicode number, decimal digit",
			input:    `\p{Nd}`,
			expected: regexpUnicodeCategory{rangeTable: unicode.Nd},
		},
		{
			name:  "unicode letter char class with prefix and suffix",
			input: `x\p{L}y`,
			expected: regexpConcat{
				left: regexpConcat{
					left:  regexpChar{'x'},
					right: regexpUnicodeCategory{rangeTable: unicode.Letter},
				},
				right: regexpChar{'y'},
			},
		},
		{
			name:     "start of text",
			input:    "^",
			expected: regexpStartOfText{},
		},
		{
			name:  "start of text in concat",
			input: "^a",
			expected: regexpConcat{
				left:  regexpStartOfText{},
				right: regexpChar{'a'},
			},
		},
		{
			name:     "end of text",
			input:    "$",
			expected: regexpEndOfText{},
		},
		{
			name:  "end of text in concat",
			input: "a$",
			expected: regexpConcat{
				left:  regexpChar{'a'},
				right: regexpEndOfText{},
			},
		},
		{
			name:  "complex expression",
			input: "(a|b|cd)*abb",
			expected: regexpConcat{
				left: regexpConcat{
					left: regexpConcat{
						left: regexpStar{
							child: regexpParenExpr{
								child: regexpUnion{
									left: regexpChar{char: 'a'},
									right: regexpUnion{
										left: regexpChar{char: 'b'},
										right: regexpConcat{
											left:  regexpChar{char: 'c'},
											right: regexpChar{char: 'd'},
										},
									},
								},
							},
						},
						right: regexpChar{char: 'a'},
					},
					right: regexpChar{char: 'b'},
				},
				right: regexpChar{char: 'b'},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			r, err := ParseRegexp(tc.input)
			require.NoError(t, err)
			assert.Equal(t, tc.expected, r)
		})
	}
}

func TestParseRegexpErrors(t *testing.T) {
	testCases := []struct {
		name          string
		input         string
		expectedError string
	}{
		{
			name:          "empty",
			input:         "",
			expectedError: "Unexpected end of regular expression",
		},
		{
			name:          "only left paren",
			input:         "(",
			expectedError: "Unexpected end of regular expression",
		},
		{
			name:          "only right paren",
			input:         ")",
			expectedError: "Unexpected closing paren",
		},
		{
			name:          "missing closing paren",
			input:         "(abc",
			expectedError: "Expected closing paren",
		},
		{
			name:          "missing opening paren",
			input:         "abc)",
			expectedError: "Unexpected closing paren",
		},
		{
			name:          "only star",
			input:         "*",
			expectedError: "Expected characters before star",
		},
		{
			name:          "star at start of string",
			input:         "*abcd",
			expectedError: "Expected characters before star",
		},
		{
			name:          "only union",
			input:         "|",
			expectedError: "Unexpected end of regular expression",
		},
		{
			name:          "union at start of string",
			input:         "|abcd",
			expectedError: "Expected characters before union",
		},
		{
			name:          "only plus",
			input:         "+",
			expectedError: "Expected characters before plus",
		},
		{
			name:          "plus at start of string",
			input:         "+abcd",
			expectedError: "Expected characters before plus",
		},
		{
			name:          "only question mark",
			input:         "?",
			expectedError: "Expected characters before question mark",
		},
		{
			name:          "question mark at start of string",
			input:         "?abcd",
			expectedError: "Expected characters before question mark",
		},
		{
			name:          "invalid escape sequence",
			input:         `abc\`,
			expectedError: "Invalid escape sequence",
		},
		{
			name:          "missing closing bracket in character class",
			input:         `[abc`,
			expectedError: "Expected closing bracket",
		},
		{
			name:          "invalid escape sequence in character class",
			input:         `[\`,
			expectedError: "Invalid escape sequence",
		},
		{
			name:          "missing start of character class range",
			input:         "[-B]",
			expectedError: "Unexpected '-' in character class",
		},
		{
			name:          "missing end of character class range",
			input:         "[A-]",
			expectedError: "Unexpected ']' in character class",
		},
		{
			name:          "only hyphen in character class",
			input:         "[-]",
			expectedError: "Unexpected '-' in character class",
		},
		{
			name:          "multiple hyphens in character class",
			input:         "[---]",
			expectedError: "Unexpected '-' in character class",
		},
		{
			name:          "unsupported unicode category",
			input:         `\pX`,
			expectedError: "Unsupported unicode category",
		},
		{
			name:          "unicode char class missing name",
			input:         `\p`,
			expectedError: "Unsupported unicode category",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ParseRegexp(tc.input)
			assert.EqualError(t, err, tc.expectedError)
		})
	}
}
