package breaks

import (
	"io"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wedaly/aretext/internal/pkg/text"
)

//go:generate go run gen_test_cases.go --dataPath data/GraphemeBreakTest.txt --outputPath grapheme_clusters_test_cases.go

func graphemeClusterBreakIterFromString(s string) BreakIter {
	reader := text.NewCloneableReaderFromString(s)
	runeIter := text.NewCloneableForwardRuneIter(reader)
	return NewGraphemeClusterBreakIter(runeIter)
}

func reverseGraphemeClusterBreakIterFromString(s string) BreakIter {
	reader := text.NewCloneableReaderFromString(text.Reverse(s))
	runeIter := text.NewCloneableBackwardRuneIter(reader)
	return NewReverseGraphemeClusterBreakIter(runeIter)
}

func TestGraphemeClusterBreakIterEmptyString(t *testing.T) {
	iter := graphemeClusterBreakIterFromString("")
	bp, err := iter.NextBreak()
	require.NoError(t, err)
	assert.Equal(t, uint64(0), bp)
}

func TestGraphemeClusterBreakIterPastEOF(t *testing.T) {
	iter := graphemeClusterBreakIterFromString("abc")
	bp, _ := iter.NextBreak()
	assert.Equal(t, uint64(0), bp)
	bp, _ = iter.NextBreak()
	assert.Equal(t, uint64(1), bp)
	bp, _ = iter.NextBreak()
	assert.Equal(t, uint64(2), bp)
	bp, _ = iter.NextBreak()
	assert.Equal(t, uint64(3), bp)
	_, err := iter.NextBreak()
	assert.Equal(t, io.EOF, err)
}

func TestGraphemeClusterBreakIterUnicodeTestCases(t *testing.T) {
	for i, tc := range graphemeBreakTestCases() {
		t.Run(strconv.FormatInt(int64(i), 10), func(t *testing.T) {
			iter := graphemeClusterBreakIterFromString(tc.inputString)
			breakPoints := make([]uint64, 0)
			for {
				bp, err := iter.NextBreak()
				if err == io.EOF {
					break
				} else if err != nil {
					require.NoError(t, err, tc.description)
				}

				breakPoints = append(breakPoints, bp)
			}
			assert.Equal(t, tc.breakPoints, breakPoints, tc.description)
		})
	}
}

func reverseBreakpoints(breakpoints []uint64) []uint64 {
	if len(breakpoints) == 0 {
		return breakpoints
	}

	reversed := make([]uint64, len(breakpoints))
	numRunes := breakpoints[len(breakpoints)-1]
	for i := 0; i < len(breakpoints); i++ {
		reversed[i] = numRunes - breakpoints[len(breakpoints)-i-1]
	}
	return reversed
}

func TestReverseGraphemeClusterBreakIterUnicodeTestCases(t *testing.T) {
	for i, tc := range graphemeBreakTestCases() {
		t.Run(strconv.FormatInt(int64(i), 10), func(t *testing.T) {
			iter := reverseGraphemeClusterBreakIterFromString(tc.inputString)
			breakPoints := make([]uint64, 0)
			for {
				bp, err := iter.NextBreak()
				if err == io.EOF {
					break
				} else if err != nil {
					require.NoError(t, err, tc.description)
				}

				breakPoints = append(breakPoints, bp)
			}
			expectedBreakPoints := reverseBreakpoints(tc.breakPoints)
			assert.Equal(t, expectedBreakPoints, breakPoints, tc.description)
		})
	}
}

type infiniteRuneIter struct {
	c rune
}

func (r *infiniteRuneIter) NextRune() (rune, error) {
	return r.c, nil
}

func (r *infiniteRuneIter) Clone() text.CloneableRuneIter {
	return &infiniteRuneIter{c: r.c}
}

func BenchmarkGraphemeClusterBreakIter(b *testing.B) {
	runeIter := &infiniteRuneIter{c: 'a'}
	iter := NewGraphemeClusterBreakIter(runeIter)
	for i := 0; i < b.N; i++ {
		if _, err := iter.NextBreak(); err != nil {
			b.Fatalf("error occurred: %v\n", err)
		}
	}
}
