package locate

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/aretext/aretext/syntax"
	"github.com/aretext/aretext/text"
)

func TestNextWordStart(t *testing.T) {
	testCases := []struct {
		name             string
		inputString      string
		syntaxLanguage   syntax.Language
		includeEndOfFile bool
		pos              uint64
		expectedPos      uint64
	}{
		{
			name:        "empty",
			inputString: "",
			pos:         0,
			expectedPos: 0,
		},
		{
			name:        "next word from current word, same line",
			inputString: "abc   defg   hij",
			pos:         1,
			expectedPos: 6,
		},
		{
			name:        "next word from whitespace, same line",
			inputString: "abc   defg   hij",
			pos:         4,
			expectedPos: 6,
		},
		{
			name:        "next word from different line",
			inputString: "abc\n   123",
			pos:         1,
			expectedPos: 7,
		},
		{
			name:        "next word to empty line",
			inputString: "abc\n\n   123",
			pos:         1,
			expectedPos: 4,
		},
		{
			name:        "empty line to next word",
			inputString: "abc\n\n   123",
			pos:         4,
			expectedPos: 8,
		},
		{
			name:        "multiple empty lines",
			inputString: "\n\n\n\n",
			pos:         1,
			expectedPos: 2,
		},
		{
			name:             "last word, do not include eof",
			inputString:      "abc",
			includeEndOfFile: false,
			pos:              1,
			expectedPos:      2,
		},
		{
			name:             "last word, include eof",
			inputString:      "abc",
			includeEndOfFile: true,
			pos:              1,
			expectedPos:      3,
		},
		{
			name:           "next syntax token",
			inputString:    "123+456",
			syntaxLanguage: syntax.LanguageGo,
			pos:            1,
			expectedPos:    3,
		},
		{
			name:           "next syntax token skip empty",
			inputString:    "123    +      456",
			syntaxLanguage: syntax.LanguageGo,
			pos:            1,
			expectedPos:    7,
		},
		{
			name:           "syntax token starts with whitespace",
			inputString:    "//    foobar",
			syntaxLanguage: syntax.LanguageGo,
			pos:            0,
			expectedPos:    6,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			textTree, err := text.NewTreeFromString(tc.inputString)
			require.NoError(t, err)
			tokenTree, err := syntax.TokenizeString(tc.syntaxLanguage, tc.inputString)
			require.NoError(t, err)
			actualPos := NextWordStart(textTree, tokenTree, tc.pos, tc.includeEndOfFile)
			assert.Equal(t, tc.expectedPos, actualPos)
		})
	}
}

func TestNextWordStartInLine(t *testing.T) {
	testCases := []struct {
		name        string
		inputString string
		pos         uint64
		expectedPos uint64
	}{
		{
			name:        "empty",
			inputString: "",
			pos:         0,
			expectedPos: 0,
		},
		{
			name:        "start of word before another word",
			inputString: "abc  def",
			pos:         0,
			expectedPos: 5,
		},
		{
			name:        "middle of word before another word",
			inputString: "abc  def",
			pos:         1,
			expectedPos: 5,
		},
		{
			name:        "end of word before another word",
			inputString: "abc  def",
			pos:         2,
			expectedPos: 5,
		},
		{
			name:        "whitespace before word",
			inputString: "abc  def",
			pos:         3,
			expectedPos: 5,
		},
		{
			name:        "start of last word in document",
			inputString: "abc",
			pos:         0,
			expectedPos: 3,
		},
		{
			name:        "middle of last word in document",
			inputString: "abc",
			pos:         1,
			expectedPos: 3,
		},
		{
			name:        "end of last word in document",
			inputString: "abc",
			pos:         2,
			expectedPos: 3,
		},
		{
			name:        "last word in document single char",
			inputString: "a",
			pos:         0,
			expectedPos: 1,
		},
		{
			name:        "last word in line before next line",
			inputString: "abc\ndef",
			pos:         1,
			expectedPos: 3,
		},
		{
			name:        "last word in line with trailing whitespace before next line",
			inputString: "abc   \ndef",
			pos:         1,
			expectedPos: 6,
		},
		{
			name:        "single line, all whitespace, cursor at start",
			inputString: "   ",
			pos:         0,
			expectedPos: 3,
		},
		{
			name:        "single line, all whitespace, cursor in middle",
			inputString: "   ",
			pos:         1,
			expectedPos: 3,
		},
		{
			name:        "single line, all whitespace, cursor at end",
			inputString: "   ",
			pos:         2,
			expectedPos: 3,
		},
		{
			name:        "lines with all whitespace, first line",
			inputString: "   \n   ",
			pos:         1,
			expectedPos: 3,
		},
		{
			name:        "lines with all whitespace, last line",
			inputString: "   \n   ",
			pos:         5,
			expectedPos: 7,
		},
		{
			name:        "last word in line single char",
			inputString: "a\nbcd",
			pos:         0,
			expectedPos: 1,
		},
		{
			name:        "empty lines",
			inputString: "\n\n\n",
			pos:         1,
			expectedPos: 1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			textTree, err := text.NewTreeFromString(tc.inputString)
			require.NoError(t, err)
			actualPos := NextWordStartInLine(textTree, nil, tc.pos)
			assert.Equal(t, tc.expectedPos, actualPos)
		})
	}
}

func TestNextWordEnd(t *testing.T) {
	testCases := []struct {
		name           string
		inputString    string
		syntaxLanguage syntax.Language
		pos            uint64
		expectedPos    uint64
	}{
		{
			name:        "empty",
			inputString: "",
			pos:         0,
			expectedPos: 0,
		},
		{
			name:        "end of word from start of current word",
			inputString: "abc   defg   hij",
			pos:         6,
			expectedPos: 9,
		},
		{
			name:        "end of word from middle of current word",
			inputString: "abc   defg   hij",
			pos:         7,
			expectedPos: 9,
		},
		{
			name:        "next word from end of current word",
			inputString: "abc   defg   hij",
			pos:         2,
			expectedPos: 9,
		},
		{
			name:        "next word from whitespace",
			inputString: "abc   defg   hij",
			pos:         4,
			expectedPos: 9,
		},
		{
			name:        "next word past empty line",
			inputString: "abc\n\n   123   xyz",
			pos:         2,
			expectedPos: 10,
		},
		{
			name:        "empty line to next word",
			inputString: "abc\n\n   123  xyz",
			pos:         4,
			expectedPos: 10,
		},
		{
			name:           "next syntax token",
			inputString:    "123+456",
			syntaxLanguage: syntax.LanguageGo,
			pos:            2,
			expectedPos:    3,
		},
		{
			name:           "next syntax token skip empty",
			inputString:    "123    +      456",
			syntaxLanguage: syntax.LanguageGo,
			pos:            2,
			expectedPos:    7,
		},
		{
			name:           "next syntax token ends with whitespace",
			inputString:    `"    abcd    "`,
			syntaxLanguage: syntax.LanguageGo,
			pos:            8,
			expectedPos:    13,
		},
		{
			name:           "end of current syntax token",
			inputString:    "123+456",
			syntaxLanguage: syntax.LanguageGo,
			pos:            0,
			expectedPos:    2,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			textTree, err := text.NewTreeFromString(tc.inputString)
			require.NoError(t, err)
			tokenTree, err := syntax.TokenizeString(tc.syntaxLanguage, tc.inputString)
			require.NoError(t, err)
			actualPos := NextWordEnd(textTree, tokenTree, tc.pos)
			assert.Equal(t, tc.expectedPos, actualPos)
		})
	}
}

func TestPrevWordStart(t *testing.T) {
	testCases := []struct {
		name           string
		inputString    string
		syntaxLanguage syntax.Language
		pos            uint64
		expectedPos    uint64
	}{
		{
			name:        "empty",
			inputString: "",
			pos:         0,
			expectedPos: 0,
		},
		{
			name:        "prev word from current word, same line",
			inputString: "abc   defg   hij",
			pos:         6,
			expectedPos: 0,
		},
		{
			name:        "prev word from whitespace, same line",
			inputString: "abc   defg   hij",
			pos:         12,
			expectedPos: 6,
		},
		{
			name:        "prev word from different line",
			inputString: "abc\n   123",
			pos:         7,
			expectedPos: 0,
		},
		{
			name:        "prev word to empty line",
			inputString: "abc\n\n   123",
			pos:         8,
			expectedPos: 4,
		},
		{
			name:        "empty line to prev word",
			inputString: "abc\n\n   123",
			pos:         4,
			expectedPos: 0,
		},
		{
			name:        "multiple empty lines",
			inputString: "\n\n\n\n",
			pos:         2,
			expectedPos: 1,
		},
		{
			name:           "prev syntax token",
			inputString:    "123+456",
			syntaxLanguage: syntax.LanguageGo,
			pos:            4,
			expectedPos:    3,
		},
		{
			name:           "prev syntax token skip empty",
			inputString:    "123    +      456",
			syntaxLanguage: syntax.LanguageGo,
			pos:            14,
			expectedPos:    7,
		},
		{
			name:           "prev syntax token starts with whitespace",
			inputString:    "// abcd",
			syntaxLanguage: syntax.LanguageGo,
			pos:            3,
			expectedPos:    0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			textTree, err := text.NewTreeFromString(tc.inputString)
			require.NoError(t, err)
			tokenTree, err := syntax.TokenizeString(tc.syntaxLanguage, tc.inputString)
			require.NoError(t, err)
			actualPos := PrevWordStart(textTree, tokenTree, tc.pos)
			assert.Equal(t, tc.expectedPos, actualPos)
		})
	}
}

func TestCurrentWordStart(t *testing.T) {
	testCases := []struct {
		name           string
		inputString    string
		syntaxLanguage syntax.Language
		pos            uint64
		expectedPos    uint64
	}{
		{
			name:        "empty",
			inputString: "",
			pos:         0,
			expectedPos: 0,
		},
		{
			name:        "start of document",
			inputString: "abc   defg   hij",
			pos:         0,
			expectedPos: 0,
		},
		{
			name:        "start of word in middle of document",
			inputString: "abc   defg   hij",
			pos:         6,
			expectedPos: 6,
		},
		{
			name:        "middle of word to start of word",
			inputString: "abc   defg   hij",
			pos:         8,
			expectedPos: 6,
		},
		{
			name:        "end of word to start of word",
			inputString: "abc   defg   hij",
			pos:         9,
			expectedPos: 6,
		},
		{
			name:        "start of whitespace",
			inputString: "abc   defg   hij",
			pos:         3,
			expectedPos: 3,
		},
		{
			name:        "middle of whitespace",
			inputString: "abc   defg   hij",
			pos:         4,
			expectedPos: 3,
		},
		{
			name:        "end of whitespace",
			inputString: "abc   defg   hij",
			pos:         5,
			expectedPos: 3,
		},
		{
			name:        "word at start of line",
			inputString: "abc\nxyz",
			pos:         5,
			expectedPos: 4,
		},
		{
			name:        "whitespace at start of line",
			inputString: "abc\n    xyz",
			pos:         6,
			expectedPos: 4,
		},
		{
			name:        "empty line",
			inputString: "abc\n\n   123",
			pos:         4,
			expectedPos: 4,
		},
		{
			name:           "adjacent syntax tokens",
			inputString:    "123+456",
			syntaxLanguage: syntax.LanguageGo,
			pos:            5,
			expectedPos:    4,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			textTree, err := text.NewTreeFromString(tc.inputString)
			require.NoError(t, err)
			tokenTree, err := syntax.TokenizeString(tc.syntaxLanguage, tc.inputString)
			require.NoError(t, err)
			actualPos := CurrentWordStart(textTree, tokenTree, tc.pos)
			assert.Equal(t, tc.expectedPos, actualPos)
		})
	}
}

func TestCurrentWordEnd(t *testing.T) {
	testCases := []struct {
		name           string
		inputString    string
		syntaxLanguage syntax.Language
		pos            uint64
		expectedPos    uint64
	}{
		{
			name:        "empty",
			inputString: "",
			pos:         0,
			expectedPos: 0,
		},
		{
			name:        "end of document",
			inputString: "abc   defg   hijk",
			pos:         14,
			expectedPos: 17,
		},
		{
			name:        "start of word in middle of document",
			inputString: "abc   defg   hij",
			pos:         6,
			expectedPos: 10,
		},
		{
			name:        "middle of word to end of word",
			inputString: "abc   defg   hij",
			pos:         7,
			expectedPos: 10,
		},
		{
			name:        "end of word",
			inputString: "abc   defg   hij",
			pos:         9,
			expectedPos: 10,
		},
		{
			name:        "start of whitespace",
			inputString: "abc   defg   hij",
			pos:         3,
			expectedPos: 6,
		},
		{
			name:        "middle of whitespace",
			inputString: "abc   defg   hij",
			pos:         4,
			expectedPos: 6,
		},
		{
			name:        "end of whitespace",
			inputString: "abc   defg   hij",
			pos:         5,
			expectedPos: 6,
		},
		{
			name:        "word before end of line",
			inputString: "abc\nxyz",
			pos:         1,
			expectedPos: 3,
		},
		{
			name:        "whitespace at end of line",
			inputString: "abc     \nxyz",
			pos:         4,
			expectedPos: 8,
		},
		{
			name:        "empty line",
			inputString: "abc\n\n   123",
			pos:         4,
			expectedPos: 4,
		},
		{
			name:           "adjacent syntax tokens",
			inputString:    "123+456",
			syntaxLanguage: syntax.LanguageGo,
			pos:            1,
			expectedPos:    3,
		},
		{
			name:           "empty syntax token",
			inputString:    `{ "ab }`,
			syntaxLanguage: syntax.LanguageJson,
			pos:            2,
			expectedPos:    5,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			textTree, err := text.NewTreeFromString(tc.inputString)
			require.NoError(t, err)
			tokenTree, err := syntax.TokenizeString(tc.syntaxLanguage, tc.inputString)
			require.NoError(t, err)
			actualPos := CurrentWordEnd(textTree, tokenTree, tc.pos)
			assert.Equal(t, tc.expectedPos, actualPos)
		})
	}
}

func TestCurrentWordEndWithTrailingWhitespace(t *testing.T) {
	testCases := []struct {
		name           string
		inputString    string
		syntaxLanguage syntax.Language
		pos            uint64
		expectedPos    uint64
	}{
		{
			name:        "empty",
			inputString: "",
			pos:         0,
			expectedPos: 0,
		},
		{
			name:        "start of word at end of document",
			inputString: "abcd",
			pos:         0,
			expectedPos: 4,
		},
		{
			name:        "middle of word at end of document",
			inputString: "abcd",
			pos:         2,
			expectedPos: 4,
		},
		{
			name:        "end of word at end of document",
			inputString: "abcd",
			pos:         3,
			expectedPos: 4,
		},
		{
			name:        "on word before whitespace at end of document",
			inputString: "abc    ",
			pos:         2,
			expectedPos: 7,
		},
		{
			name:        "on whitespace at end of document",
			inputString: "abc    ",
			pos:         4,
			expectedPos: 7,
		},
		{
			name:        "on word with trailing whitespace before next word",
			inputString: "abc    def",
			pos:         2,
			expectedPos: 7,
		},
		{
			name:        "on word at end of line",
			inputString: "abc\ndef",
			pos:         1,
			expectedPos: 3,
		},
		{
			name:        "on word before whitespace at end of line",
			inputString: "abc   \ndef",
			pos:         1,
			expectedPos: 6,
		},
		{
			name:        "on word with trailing whitespace before word at end of line",
			inputString: "abc   def\nghi",
			pos:         1,
			expectedPos: 6,
		},
		{
			name:        "on whitespace at end of line",
			inputString: "abc   \nghi",
			pos:         4,
			expectedPos: 6,
		},
		{
			name:        "on empty line",
			inputString: "ab\n\ncd",
			pos:         2,
			expectedPos: 2,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			textTree, err := text.NewTreeFromString(tc.inputString)
			require.NoError(t, err)
			tokenTree, err := syntax.TokenizeString(tc.syntaxLanguage, tc.inputString)
			require.NoError(t, err)
			actualPos := CurrentWordEndWithTrailingWhitespace(textTree, tokenTree, tc.pos)
			assert.Equal(t, tc.expectedPos, actualPos)
		})
	}
}
