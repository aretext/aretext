package fuzzy

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEmptyTrie(t *testing.T) {
	trie := newTrie()
	recordIds := trie.topRecordIdsForPrefix("", nil, 100)
	assert.Equal(t, 0, recordIds.length())
}

func TestTopRecordIdsForPrefix(t *testing.T) {
	testCases := []struct {
		name     string
		prefix   string
		limit    int
		expected []int
	}{
		{
			name:     "empty prefix large limit",
			prefix:   "",
			limit:    100,
			expected: []int{1, 2, 3, 4, 5, 6, 7},
		},
		{
			name:     "empty prefix with limit",
			prefix:   "",
			limit:    3,
			expected: []int{1, 2, 4},
		},
		{
			name:     "prefix with single char",
			prefix:   "b",
			limit:    5,
			expected: []int{1, 2, 4, 5, 6},
		},
		{
			name:     "prefix with multiple chars",
			prefix:   "foo",
			limit:    5,
			expected: []int{1, 2, 3},
		},
		{
			name:     "exact match really long string",
			prefix:   "reallylongstringthatisreallylong",
			limit:    10,
			expected: []int{7},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			trie := newTrie()
			trie.insert("foo", 1)
			trie.insert("foo", 2)
			trie.insert("foobar", 3)
			trie.insert("baz", 5)
			trie.insert("baz", 4)
			trie.insert("bat", 6)
			trie.insert("reallylongstringthatisreallylong", 7)
			recordIds := trie.topRecordIdsForPrefix(tc.prefix, nil, tc.limit).toSlice()
			assert.Equal(t, tc.expected, recordIds)
		})
	}
}

func TestFilterTopRecordsByPrevRecords(t *testing.T) {
	trie := newTrie()
	trie.insert("foo", 1)
	trie.insert("foo", 2)
	trie.insert("foobar", 3)
	trie.insert("baz", 5)
	trie.insert("bat", 6)
	trie.insert("test", 7)
	trie.insert("fooreallylongstringthatisreallylong", 8)
	prevRecordIds := newRecordIdSet(0)
	prevRecordIds.add(3)
	prevRecordIds.add(5)
	prevRecordIds.add(6)
	recordIds := trie.topRecordIdsForPrefix("foo", prevRecordIds, 10).toSlice()
	expected := []int{3}
	assert.Equal(t, expected, recordIds)
}

func TestIncrementalQueryAddChars(t *testing.T) {
	trie := newTrie()
	trie.insert("foo", 1)
	trie.insert("foo", 2)
	trie.insert("foobar", 3)
	trie.insert("baz", 5)
	trie.insert("bat", 6)
	trie.insert("test", 7)

	query := "foobar"
	expected := [][]int{
		{1, 2, 3, 5, 6, 7}, // ""
		{1, 2, 3, 5, 6, 7}, // "f"
		{1, 2, 3, 5, 6, 7}, // "fo"
		{1, 2, 3},          // "foo"
		{1, 2, 3},          // "foob"
		{1, 2, 3},          // "fooba"
		{3},                // "foobar"
	}
	for i := 0; i <= len(query); i++ {
		recordIds := trie.topRecordIdsForPrefix(query[0:i], nil, 100).toSlice()
		assert.Equal(t, expected[i], recordIds)
	}
}

func TestIncrementalQueryDeleteChars(t *testing.T) {
	trie := newTrie()
	trie.insert("foo", 1)
	trie.insert("foo", 2)
	trie.insert("foobar", 3)
	trie.insert("baz", 5)
	trie.insert("bat", 6)
	trie.insert("test", 7)

	query := "foobar"
	expected := [][]int{
		{3},                // "foobar"
		{1, 2, 3},          // "fooba"
		{1, 2, 3},          // "foob"
		{1, 2, 3},          // "foo"
		{1, 2, 3, 5, 6, 7}, // "fo"
		{1, 2, 3, 5, 6, 7}, // "f"
		{1, 2, 3, 5, 6, 7}, // ""
	}
	for i := 0; i <= len(query); i++ {
		recordIds := trie.topRecordIdsForPrefix(query[0:len(query)-i], nil, 100).toSlice()
		assert.Equal(t, expected[i], recordIds)
	}
}
