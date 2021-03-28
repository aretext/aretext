package segment

import (
	"io"
	"strconv"
	"testing"

	"github.com/aretext/aretext/text"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//go:generate go run gen_test_cases.go --dataPath data/GraphemeBreakTest.txt --outputPath grapheme_cluster_test_cases.go

func graphemeClusterIterFromString(s string) Iter {
	reader := text.NewCloneableReaderFromString(s)
	runeIter := text.NewCloneableForwardRuneIter(reader)
	return NewGraphemeClusterIter(runeIter)
}

func reverseGraphemeClusterIterFromString(s string) Iter {
	reader := text.NewCloneableReaderFromString(text.Reverse(s))
	runeIter := text.NewCloneableBackwardRuneIter(reader)
	return NewReverseGraphemeClusterIter(runeIter)
}

func TestGraphemeClusterIterEmptyString(t *testing.T) {
	iter := graphemeClusterIterFromString("")
	seg := Empty()
	err := iter.NextSegment(seg)
	assert.Equal(t, io.EOF, err)
}

func TestGraphemeClusterIterPastEOF(t *testing.T) {
	iter := graphemeClusterIterFromString("abc")

	seg := Empty()
	err := iter.NextSegment(seg)
	require.NoError(t, err)
	assert.Equal(t, []rune{'a'}, seg.Runes())

	err = iter.NextSegment(seg)
	require.NoError(t, err)
	assert.Equal(t, []rune{'b'}, seg.Runes())

	err = iter.NextSegment(seg)
	require.NoError(t, err)
	assert.Equal(t, []rune{'c'}, seg.Runes())

	err = iter.NextSegment(seg)
	assert.Equal(t, io.EOF, err)
}

func collectSegments(t *testing.T, iter Iter) ([][]rune, error) {
	segments := make([][]rune, 0)
	seg := Empty()
	for {
		err := iter.NextSegment(seg)
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}
		segments = append(segments, seg.Clone().Runes())
	}
	return segments, nil
}

func TestGraphemeClusterIterUnicodeTestCases(t *testing.T) {
	for i, tc := range graphemeBreakTestCases() {
		t.Run(strconv.FormatInt(int64(i), 10), func(t *testing.T) {
			iter := graphemeClusterIterFromString(tc.inputString)
			segments, err := collectSegments(t, iter)
			require.NoError(t, err, tc.description)
			assert.Equal(t, tc.segments, segments, tc.description)
		})
	}
}

func reverseSegments(segments [][]rune) [][]rune {
	if len(segments) == 0 {
		return segments
	}

	reversed := make([][]rune, len(segments))
	for i := 0; i < len(segments); i++ {
		reversed[i] = segments[len(segments)-i-1]
	}

	return reversed
}

func TestReverseGraphemeClusterIterUnicodeTestCases(t *testing.T) {
	for i, tc := range graphemeBreakTestCases() {
		t.Run(strconv.FormatInt(int64(i), 10), func(t *testing.T) {
			iter := reverseGraphemeClusterIterFromString(tc.inputString)
			segments, err := collectSegments(t, iter)
			require.NoError(t, err, tc.description)
			assert.Equal(t, tc.segments, reverseSegments(segments), tc.description)
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

func BenchmarkGraphemeClusterIter(b *testing.B) {
	runeIter := &infiniteRuneIter{c: 'a'}
	iter := NewGraphemeClusterIter(runeIter)
	seg := Empty()
	for i := 0; i < b.N; i++ {
		if err := iter.NextSegment(seg); err != nil {
			b.Fatalf("error occurred: %v\n", err)
		}
	}
}
