package fuzzy

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

func recordIdSetToSlice(r *recordIdSet) []int {
	var recordIds []int
	r.forEach(func(id int) { recordIds = append(recordIds, id) })
	sort.Ints(recordIds)
	return recordIds
}

func TestEmptyTrie(t *testing.T) {
	trie := newTrie()
	recordIds := trie.recordIdsForPrefix("", nil)
	assert.Equal(t, 0, len(recordIdSetToSlice(recordIds)))
}

func TestRecordIdsForPrefix(t *testing.T) {
	testCases := []struct {
		name     string
		prefix   string
		expected []int
	}{
		{
			name:     "empty prefix",
			prefix:   "",
			expected: []int{1, 2, 3, 4, 5, 6, 7},
		},
		{
			name:     "prefix with single char",
			prefix:   "b",
			expected: []int{1, 2, 3, 4, 5, 6, 7},
		},
		{
			name:     "prefix with multiple chars",
			prefix:   "foo",
			expected: []int{1, 2, 3},
		},
		{
			name:     "exact match really long string",
			prefix:   "reallylongstringthatisreallylong",
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
			recordIds := recordIdSetToSlice(trie.recordIdsForPrefix(tc.prefix, nil))
			assert.Equal(t, tc.expected, recordIds)
		})
	}
}

func TestFilterRecordsByPrevRecords(t *testing.T) {
	trie := newTrie()
	trie.insert("foo", 1)
	trie.insert("foo", 2)
	trie.insert("foobar", 3)
	trie.insert("baz", 5)
	trie.insert("bat", 6)
	trie.insert("test", 7)
	trie.insert("fooreallylongstringthatisreallylong", 8)
	prevRecordIds := newRecordIdSet()
	prevRecordIds.add(3)
	prevRecordIds.add(5)
	prevRecordIds.add(6)
	recordIds := recordIdSetToSlice(trie.recordIdsForPrefix("foo", prevRecordIds))
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
		recordIds := recordIdSetToSlice(trie.recordIdsForPrefix(query[0:i], nil))
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
		recordIds := recordIdSetToSlice(trie.recordIdsForPrefix(query[0:len(query)-i], nil))
		assert.Equal(t, expected[i], recordIds)
	}
}
