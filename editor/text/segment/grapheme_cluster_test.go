package segment

import (
	"io"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/aretext/aretext/editor/text"
)

//go:generate go run gen_test_cases.go --prefix graphemeBreak --dataPath data/GraphemeBreakTest.txt --outputPath grapheme_cluster_test_cases.go

func graphemeClusterIterFromString(s string) GraphemeClusterIter {
	tree, err := text.NewTreeFromString(s)
	if err != nil {
		panic(err)
	}
	reader := tree.ReaderAtPosition(0)
	return NewGraphemeClusterIter(reader)
}

func reverseGraphemeClusterIterFromString(s string) ReverseGraphemeClusterIter {
	tree, err := text.NewTreeFromString(s)
	if err != nil {
		panic(err)
	}
	reader := tree.ReverseReaderAtPosition(tree.NumChars())
	return NewReverseGraphemeClusterIter(reader)
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

func TestGraphemeClusterIterUnicodeTestCases(t *testing.T) {
	for i, tc := range graphemeBreakTestCases() {
		t.Run(strconv.FormatInt(int64(i), 10), func(t *testing.T) {
			iter := graphemeClusterIterFromString(tc.inputString)
			segments := make([][]rune, 0)
			seg := Empty()
			for {
				err := iter.NextSegment(seg)
				if err == io.EOF {
					break
				} else {
					require.NoError(t, err, tc.description)
				}
				segRunes := make([]rune, seg.NumRunes())
				copy(segRunes, seg.Runes())
				segments = append(segments, segRunes)
			}
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
			segments := make([][]rune, 0)
			seg := Empty()
			for {
				err := iter.NextSegment(seg)
				if err == io.EOF {
					break
				} else {
					require.NoError(t, err, tc.description)
				}
				segRunes := make([]rune, seg.NumRunes())
				copy(segRunes, seg.Runes())
				segments = append(segments, segRunes)
			}
			assert.Equal(t, tc.segments, reverseSegments(segments), tc.description)
		})
	}
}

func BenchmarkGraphemeClusterIter(b *testing.B) {
	tree, err := text.NewTreeFromString("abcdefghijkl")
	if err != nil {
		b.Fatalf("error occurred: %v\n", err)
	}

	reader := tree.ReaderAtPosition(0)
	seg := Empty()
	for i := 0; i < b.N; i++ {
		iter := NewGraphemeClusterIter(reader)
		if err := iter.NextSegment(seg); err != nil {
			b.Fatalf("error occurred: %v\n", err)
		}
	}
}
