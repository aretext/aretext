package fuzzy

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRankRecords(t *testing.T) {
	testCases := []struct {
		name     string
		query    string
		records  []string
		limit    int
		expected []string
	}{
		{
			name:  "rank multiple, no limit",
			query: "foobar",
			limit: 1000,
			records: []string{
				"barfoo",
				"xoo",
				"foo",
				"foobar",
				".foobar",
				"foo.bar",
				"xoobar",
			},
			expected: []string{
				"foobar",
				".foobar",
				"foo.bar",
				"xoobar",
				"foo",
				"barfoo",
				"xoo",
			},
		},
		{
			name:  "rank multiple with limit",
			query: "foobar",
			limit: 3,
			records: []string{
				"barfoo",
				"xoo",
				"foo",
				"foobar",
				".foobar",
				"foo.bar",
				"xoobar",
			},
			expected: []string{
				"foobar",
				".foobar",
				"foo.bar",
			},
		},
		{
			name:  "rank match at start of path component",
			query: "bar",
			limit: 1000,
			records: []string{
				"foo/foobar.go",
				"foo/bar.go",
				"foobar.go",
			},
			expected: []string{
				"foo/bar.go",
				"foobar.go",
				"foo/foobar.go",
			},
		},
		{
			name:  "rank case insensitive partial match",
			query: "foobar",
			limit: 1000,
			records: []string{
				"FooBar",
				"fooBar",
				"Foobar",
				"AooAar",
			},
			expected: []string{
				"Foobar",
				"fooBar",
				"FooBar",
				"AooAar",
			},
		},
		{
			name:  "rank penalize gaps",
			query: "set_tesx",
			limit: 1000,
			// These all have the same longest-common subsequence (LCS) length,
			// but set_test.go should rank highest because the query characters
			// align without gaps.
			records: []string{
				"menu/fuzzy/set_test.go",
				"state/edit_test.go",
				"state/quit_test.go",
				"state/task_test.go",
				"state/locator_test.go",
				"config/ruleset_test.go",
			},
			expected: []string{
				"menu/fuzzy/set_test.go",
				"config/ruleset_test.go",
				"state/edit_test.go",
				"state/quit_test.go",
				"state/task_test.go",
				"state/locator_test.go",
			},
		},
		{
			name:  "rank same score when normalized by length",
			query: "trie.go",
			limit: 100,
			records: []string{
				"text/tree.go",
				"text/tree_test.go",
				"menu/fuzzy/trie.go",
				"menu/fuzzy/trie_test.go",
			},
			// "text/tree.go" used to rank higher than "menu/fuzzy/trie.go"
			// when scores were normalized by length, but "trie.go" is a better
			// match because the query is an exact substring match.
			expected: []string{
				"menu/fuzzy/trie.go",
				"text/tree.go",
				"menu/fuzzy/trie_test.go",
				"text/tree_test.go",
			},
		},
		{
			name:  "rank start of string slightly higher than start of word",
			query: "fi",
			limit: 100,
			records: []string{
				"find and open",
				"fmt file",
			},
			expected: []string{
				"find and open",
				"fmt file",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var candidates []candidateRecord
			for recordId, record := range tc.records {
				candidates = append(candidates, candidateRecord{recordId, record})
			}
			var result []string
			for _, recordId := range rankRecords(candidates, tc.query, tc.limit) {
				result = append(result, tc.records[recordId])
			}
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestRankManyRecords(t *testing.T) {
	testCases := []struct {
		name       string
		numRecords int
	}{
		{
			name:       "almost one partition",
			numRecords: minRecordsPerPartition - 1,
		},
		{
			name:       "exactly one partition",
			numRecords: minRecordsPerPartition,
		},
		{
			name:       "barely two partitions",
			numRecords: minRecordsPerPartition + 1,
		},
		{
			name:       "many partitions, last partition has a few records",
			numRecords: minRecordsPerPartition*4 + 3,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			candidates := buildCandidates(tc.numRecords, 1)
			result := rankRecords(candidates, "query", tc.numRecords)
			assert.Equal(t, tc.numRecords, len(result))
		})
	}
}

func BenchmarkRankRecords(b *testing.B) {
	benchmarks := []struct {
		name       string
		numRecords int
		recordLen  int
	}{
		{name: "small, short records", numRecords: 128, recordLen: 16},
		{name: "small, long records", numRecords: 128, recordLen: 256},
		{name: "medium, short records", numRecords: 1024, recordLen: 16},
		{name: "medium, long records", numRecords: 1024, recordLen: 256},
		{name: "large, short records", numRecords: 65536, recordLen: 16},
		{name: "large, long records", numRecords: 65536, recordLen: 256},
	}
	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			candidates := buildCandidates(bm.numRecords, bm.recordLen)
			b.ResetTimer()
			for n := 0; n < b.N; n++ {
				rankRecords(candidates, "query", bm.numRecords)
			}
		})
	}
}

func buildCandidates(n int, recordLen int) []candidateRecord {
	recordRunes := make([]rune, recordLen)
	for i := 0; i < recordLen; i++ {
		recordRunes[i] = 'a'
	}
	record := string(recordRunes)

	candidates := make([]candidateRecord, n)
	for i := 0; i < n; i++ {
		candidates[i] = candidateRecord{recordId: i, record: record}
	}
	return candidates
}
