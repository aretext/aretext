package segment

import (
	"io"
	"sort"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/aretext/aretext/text"
)

//go:generate go run gen_test_cases.go --prefix lineBreak --dataPath data/LineBreakTest.txt --outputPath line_break_test_cases.go

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

func TestLineBreaker(t *testing.T) {
	// The test cases assume the tailoring of numbers from Example 7 or Section 8.2,
	// which we haven't implemented, so skip those.
	// See https://www.unicode.org/reports/tr14/#Testing
	skipTests := []int{
		1132, 1134, 1136, 1138, 2844, 2846, 4396, 4398, 4444, 4446,
		4568, 4570, 4616, 4618, 5080, 5082, 5334, 6120, 6122, 6124,
		6126, 7448, 7457, 7462, 7547, 7548, 7549, 7550, 7551, 7552,
		7554, 7555, 7556, 7557, 7558,
	}

	for i, tc := range lineBreakTestCases() {
		t.Run(strconv.FormatInt(int64(i), 10), func(t *testing.T) {
			if x := sort.SearchInts(skipTests, i); x < len(skipTests) && skipTests[x] == i {
				t.Skip()
				return
			}

			var lb LineBreaker
			var segments [][]rune
			var seg []rune
			for _, r := range tc.inputString {
				decision := lb.ProcessRune(r)

				if len(seg) > 0 && (decision == AllowLineBreakBefore || decision == RequireLineBreakBefore) {
					segments = append(segments, seg)
					seg = nil
				}

				seg = append(seg, r)

				if len(seg) > 0 && (decision == RequireLineBreakAfter) {
					segments = append(segments, seg)
					seg = nil
				}
			}
			if len(seg) > 0 {
				segments = append(segments, seg)
			}
			assert.Equal(t, tc.segments, segments, tc.description)
		})
	}
}

func TestLineBreakerForceBreaks(t *testing.T) {
	testCases := []struct {
		name              string
		inputString       string
		expectedDecisions []LineBreakDecision
	}{
		{
			name:        "line feed",
			inputString: "abc\ndef",
			expectedDecisions: []LineBreakDecision{
				AllowLineBreakBefore,
				NoLineBreak,
				NoLineBreak,
				RequireLineBreakAfter,
				AllowLineBreakBefore,
				NoLineBreak,
				NoLineBreak,
			},
		},
		{
			name:        "carriage return line feed",
			inputString: "abc\r\ndef",
			expectedDecisions: []LineBreakDecision{
				AllowLineBreakBefore,
				NoLineBreak,
				NoLineBreak,
				NoLineBreak,
				RequireLineBreakAfter,
				AllowLineBreakBefore,
				NoLineBreak,
				NoLineBreak,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var lb LineBreaker
			for i, r := range tc.inputString {
				decision := lb.ProcessRune(r)
				assert.Equal(t, tc.expectedDecisions[i], decision, "Wrong line break decision at pos %d", i)
			}
		})
	}
}

func TestWrappedLineIter(t *testing.T) {
	testCases := []struct {
		name            string
		inputString     string
		maxLineWidth    uint64
		allowCharBreaks bool
		widthFunc       GraphemeClusterWidthFunc
		expectedLines   []string
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
		{
			name:          "line break at word boundaries",
			inputString:   "Lorem ipsum dolor sit amet",
			maxLineWidth:  13,
			widthFunc:     gcWidthFunc(1),
			expectedLines: []string{"Lorem ipsum ", "dolor sit ", "amet"},
		},
		{
			name:            "line break at character boundaries",
			inputString:     "Lorem ipsum dolor sit amet",
			maxLineWidth:    13,
			allowCharBreaks: true,
			widthFunc:       gcWidthFunc(1),
			expectedLines:   []string{"Lorem ipsum d", "olor sit amet"},
		},
		{
			name:          "hard line break at CR",
			inputString:   "Lorem\ripsum dolor sit amet",
			maxLineWidth:  13,
			widthFunc:     gcWidthFunc(1),
			expectedLines: []string{"Lorem\r", "ipsum dolor ", "sit amet"},
		},
		{
			name:          "hard line break at CR LF",
			inputString:   "Lorem\r\nipsum dolor sit amet",
			maxLineWidth:  13,
			widthFunc:     gcWidthFunc(1),
			expectedLines: []string{"Lorem\r\n", "ipsum dolor ", "sit amet"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			wrapConfig := NewLineWrapConfig(tc.maxLineWidth, tc.allowCharBreaks, tc.widthFunc)
			tree, err := text.NewTreeFromString(tc.inputString)
			require.NoError(t, err)
			wrappedLineIter := NewWrappedLineIter(wrapConfig, tree, 0)
			lines := make([]string, 0)
			seg := Empty()
			for {
				err := wrappedLineIter.NextSegment(seg)
				if err == io.EOF {
					break
				}
				require.NoError(t, err)
				lines = append(lines, string(seg.Runes()))
			}
			assert.Equal(t, tc.expectedLines, lines)
		})
	}
}
