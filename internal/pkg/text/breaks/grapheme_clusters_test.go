package breaks

import (
	"io"
	"strconv"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wedaly/aretext/internal/pkg/text"
)

//go:generate go run gen_test_cases.go --dataPath data/GraphemeBreakTest.txt --outputPath grapheme_clusters_test_cases.go

func TestGraphemeClusterBreakIterEmptyString(t *testing.T) {
	reader := text.NewCloneableReaderFromString("")
	finder := NewGraphemeClusterBreakIter(reader)
	bp, err := finder.NextBreak()
	require.NoError(t, err)
	assert.Equal(t, uint64(0), bp)
}

func TestGraphemeClusterBreakIterPastEOF(t *testing.T) {
	reader := text.NewCloneableReaderFromString("abc")
	finder := NewGraphemeClusterBreakIter(reader)
	bp, _ := finder.NextBreak()
	assert.Equal(t, uint64(0), bp)
	bp, _ = finder.NextBreak()
	assert.Equal(t, uint64(1), bp)
	bp, _ = finder.NextBreak()
	assert.Equal(t, uint64(2), bp)
	bp, _ = finder.NextBreak()
	assert.Equal(t, uint64(3), bp)
	_, err := finder.NextBreak()
	assert.Equal(t, io.EOF, err)
}

func TestGraphemeClusterBreakIterUnicodeTestCases(t *testing.T) {
	for i, tc := range graphemeBreakTestCases() {
		t.Run(strconv.FormatInt(int64(i), 10), func(t *testing.T) {
			reader := text.NewCloneableReaderFromString(tc.inputString)
			finder := NewGraphemeClusterBreakIter(reader)
			breakPoints := make([]uint64, 0)
			for {
				bp, err := finder.NextBreak()
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

type infiniteCharReader struct {
	c rune
}

func (r *infiniteCharReader) Read(p []byte) (n int, err error) {
	w := utf8.RuneLen(r.c)
	for n+w < len(p) {
		n += utf8.EncodeRune(p, r.c)
	}
	return n, nil
}

func (r *infiniteCharReader) Clone() text.CloneableReader {
	return &infiniteCharReader{c: r.c}
}

func BenchmarkGraphemeClusterBreakIter(b *testing.B) {
	reader := &infiniteCharReader{c: 'a'}
	finder := NewGraphemeClusterBreakIter(reader)
	for i := 0; i < b.N; i++ {
		if _, err := finder.NextBreak(); err != nil {
			b.Fatalf("error occurred: %v\n", err)
		}
	}
}
