package segment

import (
	"io"
	"testing"

	"github.com/aretext/aretext/internal/pkg/text"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func gcWidthFunc(defaultWidth uint64) GraphemeClusterWidthFunc {
	return func(gc []rune, offsetInLine uint64) uint64 {
		if len(gc) == 0 {
			return 0
		}

		if gc[0] == '\n' {
			return 0
		}

		return defaultWidth
	}
}

func TestWrappedLineIter(t *testing.T) {
	testCases := []struct {
		name          string
		inputString   string
		maxLineWidth  uint64
		widthFunc     GraphemeClusterWidthFunc
		expectedLines []string
	}{
		{
			name:          "empty",
			inputString:   "",
			maxLineWidth:  10,
			widthFunc:     gcWidthFunc(1),
			expectedLines: []string{},
		},
		{
			name:          "single rune, less than max line width",
			inputString:   "a",
			maxLineWidth:  2,
			widthFunc:     gcWidthFunc(1),
			expectedLines: []string{"a"},
		},
		{
			name:          "single rune, equal to max line width",
			inputString:   "a",
			maxLineWidth:  1,
			widthFunc:     gcWidthFunc(1),
			expectedLines: []string{"a"},
		},
		{
			name:          "single rune, greater than max line width",
			inputString:   "a",
			maxLineWidth:  1,
			widthFunc:     gcWidthFunc(2),
			expectedLines: []string{"a"},
		},
		{
			name:          "multiple runes, less than max line width",
			inputString:   "abcd",
			maxLineWidth:  5,
			widthFunc:     gcWidthFunc(1),
			expectedLines: []string{"abcd"},
		},
		{
			name:          "multiple runes, equal to max line width",
			inputString:   "abcde",
			maxLineWidth:  5,
			widthFunc:     gcWidthFunc(1),
			expectedLines: []string{"abcde"},
		},
		{
			name:          "multiple runes, greater than max line width",
			inputString:   "abcdef",
			maxLineWidth:  5,
			widthFunc:     gcWidthFunc(1),
			expectedLines: []string{"abcde", "f"},
		},
		{
			name:          "multiple runes, each greater than max line width",
			inputString:   "abcdef",
			maxLineWidth:  1,
			widthFunc:     gcWidthFunc(2),
			expectedLines: []string{"a", "b", "c", "d", "e", "f"},
		},
		{
			name:          "single newline",
			inputString:   "\n",
			maxLineWidth:  5,
			widthFunc:     gcWidthFunc(1),
			expectedLines: []string{"\n"},
		},
		{
			name:          "multiple newlines",
			inputString:   "\n\n\n",
			maxLineWidth:  5,
			widthFunc:     gcWidthFunc(1),
			expectedLines: []string{"\n", "\n", "\n"},
		},
		{
			name:          "runes with newlines, no soft wrapping",
			inputString:   "abcd\nef\ngh\n",
			maxLineWidth:  5,
			widthFunc:     gcWidthFunc(1),
			expectedLines: []string{"abcd\n", "ef\n", "gh\n"},
		},
		{
			name:          "runes with newlines and soft wrapping",
			inputString:   "abcd\nefghijkl\nmnopqrstuvwxyz\n0123",
			maxLineWidth:  5,
			widthFunc:     gcWidthFunc(1),
			expectedLines: []string{"abcd\n", "efghi", "jkl\n", "mnopq", "rstuv", "wxyz\n", "0123"},
		},
		{
			name:          "runes with newlines and soft wrapping, each rune width greater than max line width",
			inputString:   "abcd\nefghijkl\nmnopqrstuvwxyz\n0123",
			maxLineWidth:  1,
			widthFunc:     gcWidthFunc(2),
			expectedLines: []string{"a", "b", "c", "d\n", "e", "f", "g", "h", "i", "j", "k", "l\n", "m", "n", "o", "p", "q", "r", "s", "t", "u", "v", "w", "x", "y", "z\n", "0", "1", "2", "3"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			wrapConfig := NewLineWrapConfig(tc.maxLineWidth, tc.widthFunc)
			reader := text.NewCloneableReaderFromString(tc.inputString)
			runeIter := text.NewCloneableForwardRuneIter(reader)
			wrappedLineIter := NewWrappedLineIter(runeIter, wrapConfig)
			lines := make([]string, 0)
			seg := NewSegment()
			for {
				err := wrappedLineIter.NextSegment(seg)
				if err == io.EOF {
					break
				}
				require.NoError(t, err)
				lines = append(lines, string(seg.Clone().Runes()))
			}
			assert.Equal(t, tc.expectedLines, lines)
		})
	}
}
