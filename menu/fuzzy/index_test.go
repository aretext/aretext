package fuzzy

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFuzzySearchIndex(t *testing.T) {
	testCases := []struct {
		name     string
		query    string
		records  []string
		expected []string
	}{
		{
			name:     "no records, empty query",
			query:    "",
			records:  nil,
			expected: []string{},
		},
		{
			name:     "no records, nonempty query",
			query:    "a",
			records:  nil,
			expected: []string{},
		},
		{
			name:     "some records, empty query",
			query:    "",
			records:  []string{"a", "b", "c"},
			expected: []string{},
		},
		{
			name:     "some records, query with only punctuation and whitespace",
			query:    ", _:",
			records:  []string{"a", "b", "c"},
			expected: []string{},
		},
		{
			name:  "some records, query matches both exact and partial",
			query: "foob",
			records: []string{
				"foo",
				"bar",
				"foobar",
				"barbaz",
				"barfoo",
				".foobar",
				"foo.bar",
			},
			expected: []string{
				"foobar",
				".foobar",
				"foo.bar",
				"foo",
			},
		},
		{
			name:  "some records, case-insensitive match",
			query: "FoO",
			records: []string{
				"fOo/first.txt",
				"Foo/second.txt",
				"foo/third.txt",
				"bar/first.txt",
				"bar/second.txt",
				"FoO/first.txt",
			},
			expected: []string{
				"FoO/first.txt",
				"Foo/second.txt",
				"foo/third.txt",
				"fOo/first.txt",
				"bar/first.txt",
			},
		},
		{
			name:     "records with shared prefix, select shorter",
			query:    "s",
			records:  []string{"save", "force save"},
			expected: []string{"save", "force save"},
		},
		{
			name:  "query path with multiple components",
			query: "foo/bar",
			records: []string{
				"foo/bar/test.txt",
				"foo/bar/test.go",
				"foo/baz/test.go",
				"doc.txt",
				"main.go",
			},
			expected: []string{
				"foo/bar/test.go",
				"foo/bar/test.txt",
				"foo/baz/test.go",
				"main.go",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			index := NewIndex(tc.records)
			recordIds := index.Search(tc.query)
			records := make([]string, len(recordIds))
			for i, recordId := range recordIds {
				records[i] = tc.records[recordId]
			}
			assert.Equal(t, tc.expected, records)
		})
	}
}
