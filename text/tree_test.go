package text

import (
	"io"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func lines(numLines int, charsPerLine int) string {
	lines := make([]string, 0, numLines)
	currentChar := byte(65)

	for i := 0; i < numLines; i++ {
		l := Repeat(rune(currentChar), charsPerLine)
		lines = append(lines, l)
		currentChar++
		if currentChar > 90 { // letter Z
			currentChar = 65 // letter A
		}
	}

	return strings.Join(lines, "\n")
}

func TestEmptyTree(t *testing.T) {
	tree := NewTree()
	assert.Equal(t, "", tree.String())
}

func TestTreeBulkLoadAndReadAll(t *testing.T) {
	testCases := []struct {
		name string
		text string
	}{
		{"empty", ""},
		{"single ASCII char", "a"},
		{"multiple ASCII chars", "abcdefg"},
		{"very long ASCII chars", Repeat('a', 300000)},
		{"single 2-byte char", "£"},
		{"multiple 2-byte chars", "£ôƊ"},
		{"very long 2-byte chars", Repeat('£', 300000)},
		{"single 3-byte char", "፴"},
		{"multiple 3-byte chars:", "፴ऴஅ"},
		{"very long 3-byte char", Repeat('፴', 3000000)},
		{"single 4-byte char", "\U0010AAAA"},
		{"multiple 4-byte chars", "\U0010AAAA\U0010BBBB\U0010CCCC"},
		{"very long 4-byte chars", Repeat('\U0010AAAA', 300000)},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tree, err := NewTreeFromString(tc.text)
			require.NoError(t, err)
			text := tree.String()
			assert.Equal(t, tc.text, text, "original str had len %d, output string had len %d", len(tc.text), len(text))
		})
	}
}

func TestReaderStartLocation(t *testing.T) {
	testCases := []struct {
		name  string
		runes []rune
	}{
		{
			name:  "short, ASCII",
			runes: []rune{'a', 'b', 'c', 'd'},
		},
		{
			name:  "short, mixed width characters",
			runes: []rune{'a', '£', 'b', '፴', 'c', 'd', '\U0010AAAA', 'e', 'ऴ'},
		},
		{
			name:  "medium, ASCII",
			runes: []rune(Repeat('a', 4096)),
		},
		{
			name:  "short, 2-byte chars",
			runes: []rune(Repeat('£', 10)),
		},
		{
			name:  "medium, 2-byte chars",
			runes: []rune(Repeat('£', 4096)),
		},
		{
			name:  "short, 3-byte chars",
			runes: []rune(Repeat('፴', 5)),
		},
		{
			name:  "medium, 3-byte chars",
			runes: []rune(Repeat('፴', 4096)),
		},
		{
			name:  "short, 4-byte chars",
			runes: []rune(Repeat('\U0010AAAA', 5)),
		},
		{
			name:  "medium, 4-byte chars",
			runes: []rune(Repeat('\U0010AAAA', 4096)),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tree, err := NewTreeFromString(string(tc.runes))
			require.NoError(t, err)

			// Check a reader starting from each character position to the end
			for i := 0; i < len(tc.runes); i++ {
				reader := tree.ReaderAtPosition(uint64(i), ReadDirectionForward)
				retrieved, err := io.ReadAll(reader)
				require.NoError(t, err)
				require.Equal(t, string(tc.runes[i:]), string(retrieved), "invalid substring starting from character at position %d (expect len = %d, actual len = %d)", i, len(string(tc.runes[i:])), len(string(retrieved)))
			}
		})
	}
}

func TestReaderPastLastCharacter(t *testing.T) {
	testCases := []struct {
		name string
		text string
		pos  uint64
	}{
		{
			name: "empty, position zero",
			text: "",
			pos:  0,
		},
		{
			name: "empty, position one",
			text: "",
			pos:  1,
		},
		{
			name: "single char, position one",
			text: "a",
			pos:  1,
		},
		{
			name: "single char, position two",
			text: "a",
			pos:  2,
		},
		{
			name: "full leaf, position at end of leaf",
			text: Repeat('a', maxBytesPerLeaf),
			pos:  maxBytesPerLeaf,
		},
		{
			name: "full leaf, position one after end of leaf",
			text: Repeat('b', maxBytesPerLeaf),
			pos:  maxBytesPerLeaf + 1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tree, err := NewTreeFromString(tc.text)
			require.NoError(t, err)
			reader := tree.ReaderAtPosition(tc.pos, ReadDirectionForward)
			retrieved, err := io.ReadAll(reader)
			require.NoError(t, err)
			assert.Equal(t, "", string(retrieved))
		})
	}
}

func TestLineStartPosition(t *testing.T) {
	testCases := []struct {
		name string
		text string
	}{
		{
			name: "empty",
			text: "",
		},
		{
			name: "single newline",
			text: "\n",
		},
		{
			name: "two newlines",
			text: "\n\n",
		},
		{
			name: "single line, same leaf",
			text: lines(1, 12),
		},
		{
			name: "single line, multiple leaves",
			text: lines(1, 4096),
		},
		{
			name: "two lines, same leaf",
			text: lines(2, 4),
		},
		{
			name: "two lines, multiple leaves",
			text: lines(2, 4096),
		},
		{
			name: "many lines, single character per line",
			text: lines(4096, 1),
		},
		{
			name: "many lines, many characters per line",
			text: lines(4096, 1024),
		},
		{
			name: "many lines, newline on previous leaf",
			text: lines(1024, maxBytesPerLeaf-1),
		},
		{
			name: "many lines, newline on next leaf",
			text: lines(1024, maxBytesPerLeaf),
		},
	}

	linePositionsFromTree := func(tree *Tree, numLines int) []uint64 {
		linePositions := make([]uint64, 0, numLines)
		for i := 0; i < numLines; i++ {
			linePositions = append(linePositions, tree.LineStartPosition(uint64(i)))
		}
		return linePositions
	}

	linePositionsFromString := func(s string) []uint64 {
		var pos uint64
		linePositions := make([]uint64, 0)
		for _, line := range strings.Split(s, "\n") {
			linePositions = append(linePositions, pos)
			pos += uint64(utf8.RuneCountInString(line)) + 1
		}
		return linePositions
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			expectLinePositions := linePositionsFromString(tc.text)
			tree, err := NewTreeFromString(tc.text)
			require.NoError(t, err)
			actualLinePositions := linePositionsFromTree(tree, len(expectLinePositions))
			assert.Equal(t, expectLinePositions, actualLinePositions)
		})
	}
}

func TestLineNumForPosition(t *testing.T) {
	testCases := []struct {
		name          string
		text          string
		position      uint64
		expectLineNum uint64
	}{
		{
			name:          "empty",
			text:          "",
			position:      0,
			expectLineNum: 0,
		},
		{
			name:          "single line, start of line",
			text:          "abcd",
			position:      0,
			expectLineNum: 0,
		},
		{
			name:          "single line, middle of line",
			text:          "abcd",
			position:      2,
			expectLineNum: 0,
		},
		{
			name:          "single line, end of line",
			text:          "abcd",
			position:      3,
			expectLineNum: 0,
		},
		{
			name:          "single line, past end of line",
			text:          "abcd",
			position:      4,
			expectLineNum: 0,
		},
		{
			name:          "single line ending in newline, middle of line",
			text:          "abcd\n",
			position:      2,
			expectLineNum: 0,
		},
		{
			name:          "multiple lines, on first line",
			text:          "abcd\nefgh",
			position:      2,
			expectLineNum: 0,
		},
		{
			name:          "multiple lines, on newline",
			text:          "abcd\nefgh",
			position:      5,
			expectLineNum: 1,
		},
		{
			name:          "multiple lines, on start of second line",
			text:          "abcd\nefgh",
			position:      6,
			expectLineNum: 1,
		},
		{
			name:          "multiple lines, on end of second line",
			text:          "abcd\nefgh",
			position:      9,
			expectLineNum: 1,
		},
		{
			name:          "multiple newlines",
			text:          "\n\n\n\n\n",
			position:      2,
			expectLineNum: 2,
		},
		{
			name:          "many lines",
			text:          lines(4096, 1024),
			position:      1025 * 100,
			expectLineNum: 100,
		},
		{
			name:          "many lines, newline on previous leaf",
			text:          lines(1024, maxBytesPerLeaf-1),
			position:      maxBytesPerLeaf * 100,
			expectLineNum: 100,
		},
		{
			name:          "many lines, past end of file",
			text:          lines(4096, 1024),
			position:      1025 * 4096,
			expectLineNum: 4095,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tree, err := NewTreeFromString(tc.text)
			require.NoError(t, err)
			lineNum := tree.LineNumForPosition(tc.position)
			assert.Equal(t, tc.expectLineNum, lineNum)
		})
	}
}

func TestReaderPastLastLine(t *testing.T) {
	testCases := []struct {
		name    string
		text    string
		lineNum uint64
	}{
		{
			name:    "empty, line zero",
			text:    "",
			lineNum: 0,
		},
		{
			name:    "empty, line one",
			text:    "",
			lineNum: 1,
		},
		{
			name:    "single line, line one",
			text:    "abcdefgh",
			lineNum: 1,
		},
		{
			name:    "single line, line two",
			text:    "abcdefgh",
			lineNum: 2,
		},
		{
			name:    "multiple lines, one past last line",
			text:    "abc\ndefg\nhijk",
			lineNum: 3,
		},
		{
			name:    "multiple lines, two past last line",
			text:    "abc\ndefg\nhijk",
			lineNum: 4,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tree, err := NewTreeFromString(tc.text)
			require.NoError(t, err)
			lineStartPos := tree.LineStartPosition(tc.lineNum)
			expectLineStartPos := uint64(utf8.RuneCountInString(tc.text))
			assert.Equal(t, expectLineStartPos, lineStartPos)
		})
	}
}

func TestReadBackwards(t *testing.T) {
	testCases := []struct {
		name        string
		inputString string
		position    uint64
		expect      string
	}{
		{
			name:        "empty",
			position:    0,
			inputString: "",
			expect:      "",
		},
		{
			name:        "single ASCII character",
			position:    1,
			inputString: "a",
			expect:      "a",
		},
		{
			name:        "multiple ASCII characters",
			position:    2,
			inputString: "abcd",
			expect:      "ba",
		},
		{
			name:        "multiple non-ASCII characters",
			position:    3,
			inputString: "a£፴cd",
			expect:      Reverse("a£፴"),
		},
		{
			name:        "long string with non-ASCII characters",
			inputString: Repeat('፴', 4096),
			position:    2048,
			expect:      Reverse(Repeat('፴', 2048)),
		},
		{
			name:        "all characters from end",
			inputString: "abcdefgh",
			position:    8,
			expect:      Reverse("abcdefgh"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tree, err := NewTreeFromString(tc.inputString)
			require.NoError(t, err)

			reader := tree.ReaderAtPosition(tc.position, ReadDirectionBackward)
			retrieved, err := io.ReadAll(reader)
			require.NoError(t, err)
			require.Equal(t, tc.expect, string(retrieved))
		})
	}
}

func TestReaderSeekBackward(t *testing.T) {
	testCases := []struct {
		name         string
		inputString  string
		readPosition uint64
		seekOffset   uint64
		expect       string
	}{
		{
			name:         "empty, seek offset zero",
			inputString:  "",
			readPosition: 0,
			seekOffset:   0,
			expect:       "",
		},
		{
			name:         "empty, seek offset one",
			inputString:  "",
			readPosition: 0,
			seekOffset:   1,
			expect:       "",
		},
		{
			name:         "single character, seek to start of string",
			inputString:  "a",
			readPosition: 1,
			seekOffset:   1,
			expect:       "a",
		},
		{
			name:         "single character, seek past start of string",
			inputString:  "a",
			readPosition: 1,
			seekOffset:   10,
			expect:       "a",
		},
		{
			name:         "multiple characters, seek a few characters back",
			inputString:  "abcd1234",
			readPosition: 8,
			seekOffset:   3,
			expect:       "234",
		},
		{
			name:         "very long string, short seek",
			inputString:  Repeat('a', 300000),
			readPosition: 300000,
			seekOffset:   100,
			expect:       Repeat('a', 100),
		},
		{
			name:         "very long string, long seek",
			inputString:  Repeat('a', 300000),
			readPosition: 300000,
			seekOffset:   10000,
			expect:       Repeat('a', 10000),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tree, err := NewTreeFromString(tc.inputString)
			require.NoError(t, err)

			reader := tree.ReaderAtPosition(tc.readPosition, ReadDirectionForward)
			err = reader.SeekBackward(tc.seekOffset)
			require.NoError(t, err)

			retrieved, err := io.ReadAll(reader)
			require.NoError(t, err)
			require.Equal(t, tc.expect, string(retrieved))
		})
	}
}

func TestReadToEndThenSeekBackward(t *testing.T) {
	s := Repeat('a', 1000)
	tree, err := NewTreeFromString(s)
	require.NoError(t, err)

	reader := tree.ReaderAtPosition(0, ReadDirectionForward)
	_, err = io.ReadAll(reader)
	require.NoError(t, err)

	err = reader.SeekBackward(100)
	require.NoError(t, err)

	retrieved, err := io.ReadAll(reader)
	require.NoError(t, err)

	expect := Repeat('a', 100)
	require.Equal(t, expect, string(retrieved))
}

func TestInsertAtPosition(t *testing.T) {
	testCases := []struct {
		name        string
		initialText string
		insertPos   uint64
		insertChar  rune
		expectText  string
	}{
		{
			name:        "empty, insert ASCII",
			initialText: "",
			insertPos:   0,
			insertChar:  'a',
			expectText:  "a",
		},
		{
			name:        "empty, insert 2-byte char",
			initialText: "",
			insertPos:   0,
			insertChar:  '£',
			expectText:  "£",
		},
		{
			name:        "empty, insert 3-byte char",
			initialText: "",
			insertPos:   0,
			insertChar:  'ऴ',
			expectText:  "ऴ",
		},
		{
			name:        "empty, insert 4-byte char",
			initialText: "",
			insertPos:   0,
			insertChar:  '\U0010AAAA',
			expectText:  "\U0010AAAA",
		},
		{
			name:        "insert ASCII at beginning",
			initialText: "abcdefgh",
			insertPos:   0,
			insertChar:  'x',
			expectText:  "xabcdefgh",
		},
		{
			name:        "insert 2-byte char at beginning",
			initialText: "abcƊe",
			insertPos:   0,
			insertChar:  'ô',
			expectText:  "ôabcƊe",
		},
		{
			name:        "insert 3-byte char at beginning",
			initialText: "ab፴cƊe",
			insertPos:   0,
			insertChar:  'ऴ',
			expectText:  "ऴab፴cƊe",
		},
		{
			name:        "insert 4-byte char at beginning",
			initialText: "ab፴cƊe",
			insertPos:   0,
			insertChar:  '\U0010AAAA',
			expectText:  "\U0010AAAAab፴cƊe",
		},
		{
			name:        "insert ASCII just before end",
			initialText: "abc",
			insertPos:   2,
			insertChar:  'x',
			expectText:  "abxc",
		},
		{
			name:        "insert 2-byte char just before end",
			initialText: "abcƊe",
			insertPos:   4,
			insertChar:  'ô',
			expectText:  "abcƊôe",
		},
		{
			name:        "insert 3-byte char just before end",
			initialText: "ab፴cƊe",
			insertPos:   5,
			insertChar:  'ऴ',
			expectText:  "ab፴cƊऴe",
		},
		{
			name:        "insert 4-byte char just before end",
			initialText: "ab፴cƊe",
			insertPos:   5,
			insertChar:  '\U0010AAAA',
			expectText:  "ab፴cƊ\U0010AAAAe",
		},
		{
			name:        "insert ASCII at end",
			initialText: "abc",
			insertPos:   3,
			insertChar:  'x',
			expectText:  "abcx",
		},
		{
			name:        "insert 2-byte char at end",
			initialText: "abcƊe",
			insertPos:   5,
			insertChar:  'ô',
			expectText:  "abcƊeô",
		},
		{
			name:        "insert 3-byte char just before end",
			initialText: "ab፴cƊe",
			insertPos:   6,
			insertChar:  'ऴ',
			expectText:  "ab፴cƊeऴ",
		},
		{
			name:        "insert 4-byte char just before end",
			initialText: "ab፴cƊe",
			insertPos:   6,
			insertChar:  '\U0010AAAA',
			expectText:  "ab፴cƊe\U0010AAAA",
		},
		{
			name:        "insert ASCII past end",
			initialText: "abc",
			insertPos:   1000,
			insertChar:  'x',
			expectText:  "abcx",
		},
		{
			name:        "insert 2-byte char at end",
			initialText: "abcƊe",
			insertPos:   1000,
			insertChar:  'ô',
			expectText:  "abcƊeô",
		},
		{
			name:        "insert 3-byte char at end",
			initialText: "ab፴cƊe",
			insertPos:   1000,
			insertChar:  'ऴ',
			expectText:  "ab፴cƊeऴ",
		},
		{
			name:        "insert 4-byte char at end",
			initialText: "ab፴cƊe",
			insertPos:   1000,
			insertChar:  '\U0010AAAA',
			expectText:  "ab፴cƊe\U0010AAAA",
		},
		{
			name:        "insert ASCII in middle",
			initialText: "abcdefgh",
			insertPos:   3,
			insertChar:  'x',
			expectText:  "abcxdefgh",
		},
		{
			name:        "insert 2-byte char in middle",
			initialText: "abcƊe",
			insertPos:   3,
			insertChar:  'ô',
			expectText:  "abcôƊe",
		},
		{
			name:        "insert 3-byte char in middle",
			initialText: "ab፴cƊe",
			insertPos:   3,
			insertChar:  'ऴ',
			expectText:  "ab፴ऴcƊe",
		},
		{
			name:        "insert 4-byte char in middle",
			initialText: "ab፴cƊe",
			insertPos:   3,
			insertChar:  '\U0010AAAA',
			expectText:  "ab፴\U0010AAAAcƊe",
		},
		{
			name:        "insert ASCII before long string",
			initialText: Repeat('a', 4096),
			insertPos:   0,
			insertChar:  'x',
			expectText:  "x" + Repeat('a', 4096),
		},
		{
			name:        "insert 2-byte char before long string",
			initialText: Repeat('£', 4096),
			insertPos:   0,
			insertChar:  'ô',
			expectText:  "ô" + Repeat('£', 4096),
		},
		{
			name:        "insert 3-byte char before long string",
			initialText: Repeat('፴', 4096),
			insertPos:   0,
			insertChar:  'ऴ',
			expectText:  "ऴ" + Repeat('፴', 4096),
		},
		{
			name:        "insert 4-byte char before long string",
			initialText: Repeat('\U0010AAAA', 4096),
			insertPos:   0,
			insertChar:  '\U0010BBBB',
			expectText:  "\U0010BBBB" + Repeat('\U0010AAAA', 4096),
		},
		{
			name:        "insert ASCII in middle of long string",
			initialText: Repeat('a', 4096),
			insertPos:   2000,
			insertChar:  'x',
			expectText:  Repeat('a', 2000) + "x" + Repeat('a', 2096),
		},
		{
			name:        "insert 2-byte char in middle of  long string",
			initialText: Repeat('£', 4096),
			insertPos:   2000,
			insertChar:  'ô',
			expectText:  Repeat('£', 2000) + "ô" + Repeat('£', 2096),
		},
		{
			name:        "insert 3-byte char in middle of  long string",
			initialText: Repeat('፴', 4096),
			insertPos:   2000,
			insertChar:  'ऴ',
			expectText:  Repeat('፴', 2000) + "ऴ" + Repeat('፴', 2096),
		},
		{
			name:        "insert 4-byte char in middle of  long string",
			initialText: Repeat('\U0010AAAA', 4096),
			insertPos:   2000,
			insertChar:  '\U0010BBBB',
			expectText:  Repeat('\U0010AAAA', 2000) + "\U0010BBBB" + Repeat('\U0010AAAA', 2096),
		},
		{
			name:        "insert ASCII at end of long string",
			initialText: Repeat('a', 4096),
			insertPos:   4096,
			insertChar:  'x',
			expectText:  Repeat('a', 4096) + "x",
		},
		{
			name:        "insert 2-byte char at end of  long string",
			initialText: Repeat('£', 4096),
			insertPos:   4096,
			insertChar:  'ô',
			expectText:  Repeat('£', 4096) + "ô",
		},
		{
			name:        "insert 3-byte char at end of  long string",
			initialText: Repeat('፴', 4096),
			insertPos:   4096,
			insertChar:  'ऴ',
			expectText:  Repeat('፴', 4096) + "ऴ",
		},
		{
			name:        "insert 4-byte char at end of  long string",
			initialText: Repeat('\U0010AAAA', 4096),
			insertPos:   4096,
			insertChar:  '\U0010BBBB',
			expectText:  Repeat('\U0010AAAA', 4096) + "\U0010BBBB",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tree, err := NewTreeFromString(tc.initialText)
			require.NoError(t, err)
			err = tree.InsertAtPosition(tc.insertPos, tc.insertChar)
			require.NoError(t, err)
			assert.Equal(t, tc.expectText, tree.String())
		})
	}
}

func TestInsertManySequential(t *testing.T) {
	testCases := []struct {
		name string
		text string
	}{
		{
			name: "several ASCII chars",
			text: "abcd",
		},
		{
			name: "several 2-byte chars",
			text: "£ôƊ",
		},
		{
			name: "several 3-byte chars",
			text: "፴ऴஅ",
		},
		{
			name: "several 4-byte chars",
			text: "\U0010AAAA\U0010BBBB\U0010CCCC",
		},
		{
			name: "many ASCII chars",
			text: Repeat('a', 4096),
		},
		{
			name: "many 2-byte chars",
			text: Repeat('Ɗ', 4096),
		},
		{
			name: "many 3-byte chars",
			text: Repeat('፴', 4096),
		},
		{
			name: "many 4-byte chars",
			text: Repeat('\U0010AAAA', 4096),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tree := NewTree()
			i := uint64(0)
			for _, c := range tc.text {
				err := tree.InsertAtPosition(i, c)
				require.NoError(t, err)
				i++
			}
			actualText := tree.String()
			assert.Equal(t, tc.text, actualText, "input text len %v, output text len %v", len(tc.text), len(actualText))
		})
	}
}

func TestInsertNewline(t *testing.T) {
	testCases := []struct {
		name            string
		initialText     string
		insertPos       uint64
		retrieveLineNum uint64
		expectLine      string
	}{
		{
			name:            "empty string",
			initialText:     "",
			insertPos:       0,
			retrieveLineNum: 1,
			expectLine:      "",
		},
		{
			name:            "middle of string",
			initialText:     "abcdefgh",
			insertPos:       3,
			retrieveLineNum: 1,
			expectLine:      "defgh",
		},
		{
			name:            "after existing newline",
			initialText:     "ab\nhijkl",
			insertPos:       5,
			retrieveLineNum: 2,
			expectLine:      "jkl",
		},
		{
			name:            "very long string",
			initialText:     Repeat('a', 4096),
			insertPos:       4095,
			retrieveLineNum: 1,
			expectLine:      "a",
		},
		{
			name:            "very long string with existing newlines",
			initialText:     lines(4096, 10),
			insertPos:       1000,
			retrieveLineNum: 4096,
			expectLine:      "NNNNNNNNNN",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tree, err := NewTreeFromString(tc.initialText)
			require.NoError(t, err)

			err = tree.InsertAtPosition(tc.insertPos, '\n')
			require.NoError(t, err)

			lineStartPos := tree.LineStartPosition(tc.retrieveLineNum)
			reader := tree.ReaderAtPosition(lineStartPos, ReadDirectionForward)
			text, err := io.ReadAll(reader)
			require.NoError(t, err)
			assert.Equal(t, tc.expectLine, string(text))
		})
	}
}

func TestInsertInvalidUtf8(t *testing.T) {
	tree := NewTree()
	err := tree.InsertAtPosition(0, rune(-1))
	assert.Error(t, err)
	assert.Equal(t, "", tree.String())
}

func TestDeleteAtPosition(t *testing.T) {
	testCases := []struct {
		name              string
		inputText         string
		deletePos         uint64
		expectDidDelete   bool
		expectDeletedRune rune
		expectText        string
	}{
		{
			name:            "empty",
			inputText:       "",
			deletePos:       0,
			expectDidDelete: false,
			expectText:      "",
		},
		{
			name:              "single character",
			inputText:         "A",
			deletePos:         0,
			expectDidDelete:   true,
			expectDeletedRune: 'A',
			expectText:        "",
		},
		{
			name:            "single character, delete past end",
			inputText:       "A",
			deletePos:       1,
			expectDidDelete: false,
			expectText:      "A",
		},
		{
			name:              "two characters, delete first",
			inputText:         "AB",
			deletePos:         0,
			expectDidDelete:   true,
			expectDeletedRune: 'A',
			expectText:        "B",
		},
		{
			name:              "two characters, delete second",
			inputText:         "AB",
			deletePos:         1,
			expectDidDelete:   true,
			expectDeletedRune: 'B',
			expectText:        "A",
		},
		{
			name:              "multi-byte character, delete before",
			inputText:         "a£b",
			deletePos:         0,
			expectDidDelete:   true,
			expectDeletedRune: 'a',
			expectText:        "£b",
		},
		{
			name:              "multi-byte character, delete on",
			inputText:         "a£b",
			deletePos:         1,
			expectDidDelete:   true,
			expectDeletedRune: '£',
			expectText:        "ab",
		},
		{
			name:              "multi-byte character, delete after",
			inputText:         "a£b",
			deletePos:         2,
			expectDidDelete:   true,
			expectDeletedRune: 'b',
			expectText:        "a£",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tree, err := NewTreeFromString(tc.inputText)
			require.NoError(t, err)
			didDelete, r := tree.DeleteAtPosition(tc.deletePos)
			assert.Equal(t, tc.expectDidDelete, didDelete)
			assert.Equal(t, tc.expectDeletedRune, r)
			assert.Equal(t, tc.expectText, tree.String())
		})
	}
}

func TestDeleteAllCharsInLongStringFromBeginning(t *testing.T) {
	testCases := []struct {
		name string
		text string
	}{
		{
			name: "ASCII",
			text: Repeat('a', 4096),
		},
		{
			name: "2-byte chars",
			text: Repeat('£', 4096),
		},
		{
			name: "3-byte chars",
			text: Repeat('፴', 4096),
		},
		{
			name: "4-byte chars",
			text: Repeat('\U0010AAAA', 4096),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tree, err := NewTreeFromString(tc.text)
			require.NoError(t, err)
			for i := 0; i < len(tc.text); i++ {
				tree.DeleteAtPosition(0)
			}
			assert.Equal(t, "", tree.String())
		})
	}
}

func TestDeleteAllCharsInLongStringFromEnd(t *testing.T) {
	testCases := []struct {
		name string
		text string
	}{
		{
			name: "ASCII",
			text: Repeat('a', 4096),
		},
		{
			name: "2-byte chars",
			text: Repeat('£', 4096),
		},
		{
			name: "3-byte chars",
			text: Repeat('፴', 4096),
		},
		{
			name: "4-byte chars",
			text: Repeat('\U0010AAAA', 4096),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tree, err := NewTreeFromString(tc.text)
			require.NoError(t, err)
			for i := len(tc.text) - 1; i >= 0; i-- {
				tree.DeleteAtPosition(0)
			}
			assert.Equal(t, "", tree.String())
		})
	}
}

func TestDeleteNewline(t *testing.T) {
	tree, err := NewTreeFromString(lines(4096, 100))
	require.NoError(t, err)

	lineStart := tree.LineStartPosition(4094)
	expectLineStart := uint64(413494)
	assert.Equal(t, expectLineStart, lineStart)

	tree.DeleteAtPosition(100) // delete first newline
	lineStart = tree.LineStartPosition(4094)
	expectLineStart += 100
	assert.Equal(t, expectLineStart, lineStart)
}

func TestNodeSplit(t *testing.T) {
	s := Repeat('x', 1339)
	tree, err := NewTreeFromString(s)
	require.NoError(t, err)

	assert.Equal(t, uint64(1339), tree.NumChars())
	assert.Equal(t, 1339, len(tree.String()))
	require.NoError(t, tree.InsertAtPosition(0, 'a'))
	require.NoError(t, tree.InsertAtPosition(1, 'b'))
	assert.Equal(t, uint64(1341), tree.NumChars())
	assert.Equal(t, 1341, len(tree.String()))
}

func BenchmarkLoad(b *testing.B) {
	benchmarks := []struct {
		name     string
		numBytes int
	}{
		{name: "small", numBytes: 16},
		{name: "medium", numBytes: 4096},
		{name: "large", numBytes: 1048576},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			text := Repeat('a', bm.numBytes)
			for n := 0; n < b.N; n++ {
				_, err := NewTreeFromString(text)
				if err != nil {
					b.Fatalf("err = %v", err)
				}
			}
		})
	}
}

func BenchmarkRead(b *testing.B) {
	benchmarks := []struct {
		name     string
		numBytes int
	}{
		{name: "small", numBytes: 16},
		{name: "medium", numBytes: 4096},
		{name: "large", numBytes: 1048576},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			text := Repeat('a', bm.numBytes)
			tree, err := NewTreeFromString(text)
			if err != nil {
				b.Fatalf("err = %v", err)
			}

			for n := 0; n < b.N; n++ {
				reader := tree.ReaderAtPosition(0, ReadDirectionForward)
				_, err := io.ReadAll(reader)
				if err != nil {
					b.Fatalf("err = %v", err)
				}
			}
		})
	}
}

func BenchmarkInsert(b *testing.B) {
	benchmarks := []struct {
		name           string
		numBytesInTree int
	}{
		{name: "empty", numBytesInTree: 0},
		{name: "small", numBytesInTree: 16},
		{name: "medium", numBytesInTree: 4096},
		{name: "large", numBytesInTree: 1048576},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			text := Repeat('a', bm.numBytesInTree)
			tree, err := NewTreeFromString(text)
			if err != nil {
				b.Fatalf("err = %v", err)
			}

			insertPos := uint64(bm.numBytesInTree / 2)

			for n := 0; n < b.N; n++ {
				// This is a little inaccurate because we're modifying the same tree on each iteration.
				err = tree.InsertAtPosition(insertPos, 'x')
				if err != nil {
					b.Fatalf("err = %v", err)
				}
			}
		})
	}
}
