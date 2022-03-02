package text

import (
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildPrefixTable(t *testing.T) {
	testCases := []struct {
		name     string
		pattern  string
		expected []int
	}{
		{
			name:     "empty",
			pattern:  "",
			expected: []int{},
		},
		{
			name:     "single char",
			pattern:  "a",
			expected: []int{0},
		},
		{
			name:     "all unique",
			pattern:  "abcdef",
			expected: []int{0, 0, 0, 0, 0, 0},
		},
		{
			name:     "prefixes",
			pattern:  "ababababca",
			expected: []int{0, 0, 1, 2, 3, 4, 5, 6, 0, 1},
		},
		{
			name:     "more prefixes",
			pattern:  "ababbabbabbababbabb",
			expected: []int{0, 0, 1, 2, 0, 1, 2, 0, 1, 2, 0, 1, 2, 3, 4, 5, 6, 7, 8},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			pt := buildPrefixTable(tc.pattern)
			assert.Equal(t, tc.expected, pt)
		})
	}
}

var searchTestCases = []struct {
	name         string
	q            string
	s            string
	expectFound  bool
	expectOffset uint64
}{
	{
		name:        "empty string, empty query",
		q:           "",
		s:           "",
		expectFound: false,
	},
	{
		name:        "empty string, non-empty query",
		q:           "abc",
		s:           "",
		expectFound: false,
	},
	{
		name:        "non-empty string, empty query",
		q:           "",
		s:           "abc",
		expectFound: false,
	},
	{
		name:        "find single char in short string, not found",
		q:           "x",
		s:           "abcd",
		expectFound: false,
	},
	{
		name:         "find single char at beginning of short string",
		q:            "x",
		s:            "xabcd",
		expectFound:  true,
		expectOffset: 0,
	},
	{
		name:         "find single char in middle of short string",
		q:            "a",
		s:            "xyzabc",
		expectFound:  true,
		expectOffset: 3,
	},
	{
		name:         "find single char at beginning of short string",
		q:            "x",
		s:            "abcdx",
		expectFound:  true,
		expectOffset: 4,
	},
	{
		name:         "exact match short string",
		q:            "abcd1234",
		s:            "abcd1234",
		expectFound:  true,
		expectOffset: 0,
	},
	{
		name:         "repeating prefix",
		q:            "ababababa",
		s:            "xxxxxxxxabcababcababababayyyyyyy",
		expectFound:  true,
		expectOffset: 16,
	},
	{
		name:         "long string",
		q:            "abcabba",
		s:            Repeat('x', 512) + "abcabba" + Repeat('y', 1024),
		expectFound:  true,
		expectOffset: 512,
	},
	{
		name:         "multi-byte unicode",
		q:            "丅丆",
		s:            "丂丄丅丆丏 ¢ह€한",
		expectFound:  true,
		expectOffset: 2,
	},
}

func TestSearchNextInReader(t *testing.T) {
	for _, tc := range searchTestCases {
		t.Run(tc.name, func(t *testing.T) {
			ok, offset, err := NewSearcher(tc.q).NextInReader(strings.NewReader(tc.s))
			assert.Equal(t, tc.expectFound, ok)
			assert.Equal(t, tc.expectOffset, offset)
			assert.NoError(t, err)
		})
	}
}

func TestSearchNextInReaderWithSingleByteReader(t *testing.T) {
	for _, tc := range searchTestCases {
		t.Run(tc.name, func(t *testing.T) {
			r := NewSingleByteReader(tc.s)
			ok, offset, err := NewSearcher(tc.q).NextInReader(r)
			assert.Equal(t, tc.expectFound, ok)
			assert.Equal(t, tc.expectOffset, offset)
			assert.NoError(t, err)
		})
	}
}

func TestSearchNextInReaderWithLimit(t *testing.T) {
	testCases := []struct {
		name         string
		s            string
		q            string
		limit        uint64
		expectFound  bool
		expectOffset uint64
	}{
		{
			name:        "ascii, limit zero",
			s:           "abcd123",
			q:           "cd",
			limit:       0,
			expectFound: false,
		},
		{
			name:         "ascii, find before limit",
			s:            "abcd123",
			q:            "cd",
			limit:        4,
			expectFound:  true,
			expectOffset: 2,
		},
		{
			name:        "ascii, find on limit",
			s:           "abcd123",
			q:           "cd",
			limit:       3,
			expectFound: false,
		},
		{
			name:        "ascii, find after limit",
			s:           "abcd123",
			q:           "cd",
			limit:       1,
			expectFound: false,
		},
		{
			name:         "multi-byte unicode, find before limit",
			q:            "丅丆",
			s:            "丂丄丅丆丏 ¢ह€한",
			limit:        4,
			expectFound:  true,
			expectOffset: 2,
		},
		{
			name:        "multi-byte unicode, find on limit",
			q:           "丅丆",
			s:           "丂丄丅丆丏 ¢ह€한",
			limit:       3,
			expectFound: false,
		},
		{
			name:        "multi-byte unicode, find after limit",
			q:           "丅丆",
			s:           "丂丄丅丆丏 ¢ह€한",
			limit:       2,
			expectFound: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			r := strings.NewReader(tc.s)
			searcher := NewSearcher(tc.q).Limit(tc.limit)
			ok, offset, err := searcher.NextInReader(r)
			assert.Equal(t, tc.expectFound, ok)
			assert.Equal(t, tc.expectOffset, offset)
			assert.NoError(t, err)
		})
	}
}

func TestSearchLastInReader(t *testing.T) {
	testCases := []struct {
		name         string
		q            string
		s            string
		expectFound  bool
		expectOffset uint64
	}{
		{
			name:        "empty string, empty query",
			q:           "",
			s:           "",
			expectFound: false,
		},
		{
			name:        "empty string, non-empty query",
			q:           "abc",
			s:           "",
			expectFound: false,
		},
		{
			name:        "non-empty string, empty query",
			q:           "",
			s:           "abc",
			expectFound: false,
		},
		{
			name:        "no matches",
			q:           "xyz",
			s:           "abcdefghijklmnop",
			expectFound: false,
		},
		{
			name:         "single match",
			q:            "xy",
			s:            "abcdxyz",
			expectFound:  true,
			expectOffset: 4,
		},
		{
			name:         "multiple matches",
			q:            "xy",
			s:            "abcdxyzxyz123",
			expectFound:  true,
			expectOffset: 7,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			r := strings.NewReader(tc.s)
			searcher := NewSearcher(tc.q)
			ok, offset, err := searcher.LastInReader(r)
			assert.Equal(t, tc.expectFound, ok)
			assert.Equal(t, tc.expectOffset, offset)
			assert.NoError(t, err)
		})
	}
}

// queryAtEndReader outputs n space characters followed by a query string.
type queryAtEndReader struct {
	n int
	q string
	i int
}

func (r *queryAtEndReader) Read(buf []byte) (int, error) {
	if r.i >= r.n+len(r.q) {
		return 0, io.EOF
	}

	var j int
	for r.i < r.n && j < len(buf) {
		buf[j] = ' '
		r.i++
		j++
	}

	if r.i < r.n+len(r.q) && j < len(buf) {
		buf[j] = r.q[r.i-r.n]
		r.i++
		j++
	}

	return j, nil
}

func BenchmarkFindAtEnd(b *testing.B) {
	for i := 0; i < b.N; i++ {
		r := &queryAtEndReader{
			n: 100000,
			q: "abcdxyz1234",
		}
		ok, _, err := NewSearcher(r.q).NextInReader(r)
		assert.True(b, ok)
		assert.NoError(b, err)
	}
}
