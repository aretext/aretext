package display

import (
	"bufio"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type singleByteReader struct {
	s string
	i int
}

func newSingleByteReader(s string) io.Reader {
	return &singleByteReader{s, 0}
}

func (r *singleByteReader) Read(p []byte) (n int, err error) {
	n = copy(p, r.s[r.i:r.i+1])
	r.i++
	if r.i >= len(r.s) {
		err = io.EOF
	}
	return
}

func tokenize(t *testing.T, reader io.Reader) []string {
	scanner := bufio.NewScanner(reader)
	scanner.Split(splitUtf8Cells)
	tokens := make([]string, 0)
	for scanner.Scan() {
		tokens = append(tokens, scanner.Text())
	}
	require.NoError(t, scanner.Err())
	return tokens
}

func TestSplitUtf8Cells(t *testing.T) {
	testCases := []struct {
		name           string
		inputString    string
		expectedTokens []string
	}{
		{
			name:           "empty",
			inputString:    "",
			expectedTokens: []string{},
		},
		{
			name:           "single ASCII",
			inputString:    "a",
			expectedTokens: []string{"a"},
		},
		{
			name:           "two ASCII",
			inputString:    "ab",
			expectedTokens: []string{"a", "b"},
		},
		{
			name:           "combining acute accent",
			inputString:    "e\u0301",
			expectedTokens: []string{"e\u0301"},
		},
		{
			name:           "combining acute accent followed by ASCII",
			inputString:    "e\u0301x",
			expectedTokens: []string{"e\u0301", "x"},
		},
		{
			name:           "emoji with zero-width joiner",
			inputString:    "\U0001f468\u200d\U0001f469\u200d\U0001f466",
			expectedTokens: []string{"\U0001f468\u200d\U0001f469\u200d\U0001f466"},
		},
		{
			name:           "emoji with zero-width joiner, ASCII prefix",
			inputString:    "xyz\U0001f468\u200d\U0001f469\u200d\U0001f466",
			expectedTokens: []string{"x", "y", "z", "\U0001f468\u200d\U0001f469\u200d\U0001f466"},
		},
		{
			name:           "emoji with zero-width joiner, ASCII suffix",
			inputString:    "\U0001f468\u200d\U0001f469\u200d\U0001f466xyz",
			expectedTokens: []string{"\U0001f468\u200d\U0001f469\u200d\U0001f466", "x", "y", "z"},
		},
		{
			name:           "multiple zero-width joiners",
			inputString:    "\u200d\u200d\u200d",
			expectedTokens: []string{"\u200d", "\u200d", "\u200d"},
		},
		{
			name:           "newline with linefeed",
			inputString:    "ab\n\rc",
			expectedTokens: []string{"a", "b", "\n", "\r", "c"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			reader := strings.NewReader(tc.inputString)
			tokens := tokenize(t, reader)
			assert.Equal(t, tc.expectedTokens, tokens)
		})
	}
}

func TestSplitUtf8CellsSingleByteReader(t *testing.T) {
	s := "abcd\U0001f468\u200d\U0001f469\u200d\U0001f466xyz\n\re\u0301"
	reader := newSingleByteReader(s)
	tokens := tokenize(t, reader)
	expectedTokens := []string{
		"a", "b", "c", "d",
		"\U0001f468\u200d\U0001f469\u200d\U0001f466",
		"x", "y", "z",
		"\n", "\r",
		"e\u0301",
	}
	assert.Equal(t, expectedTokens, tokens)
}

func TestSplitUtf8CellsManyZeroWidth(t *testing.T) {
	var sb strings.Builder
	sb.WriteString("a")
	for i := 0; i < 65536; i++ {
		sb.WriteString("\u0301") // accent
	}
	s := sb.String()

	reader := newSingleByteReader(s)

	// If the SplitFunc were to wait indefinitely for a non-combining character, bufio.Scanner would panic here.
	// The test passes because the SplitFunc outputs a token when it has received a certain number of bytes,
	// even if all the characters are zero-width.
	tokens := tokenize(t, reader)
	assert.Equal(t, 65505, len(tokens))
	assert.Equal(t, s, strings.Join(tokens, ""))
}
