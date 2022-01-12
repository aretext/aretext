package fuzzy

import (
	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

// Maximum number of records returned for a given search query.
const maxSearchResults = 100

// Index provides efficient fuzzy search over a fixed set of records.
// It is designed to return results at interactive speeds (tens of ms)
// for hundreds of thousands of records.
type Index struct {
	records     []string
	keywordTrie *trie
}

// NewIndex constructs a new index for the set of records.
// Records are identified by indices in the records slice.
// The ranking algorithm takes time proportional to the length of each record,
// so the caller should truncate long record strings.
func NewIndex(records []string) *Index {
	keywordTrie := newTrie()
	for recordId, record := range records {
		iterKeywords(record, func(keyword string) {
			// The trie is responsible for deduplicating record IDs for the keyword.
			keywordTrie.insert(keyword, recordId)
		})
	}
	return &Index{records, keywordTrie}
}

// Search returns record IDs that match the query, ordered from most- to least-relevant.
// The ranking algorithm takes time proportional to the length of the query,
// so the caller should truncate long queries.
func (idx *Index) Search(query string) []int {
	recordIds := idx.findTopMatchingRecords(query)
	return idx.rankRecords(query, recordIds)
}

func (idx *Index) findTopMatchingRecords(query string) *recordIdSet {
	var recordIds *recordIdSet
	iterKeywords(query, func(queryKeyword string) {
		// For the first keyword, recordIds is nil, so the trie will search all records.
		// On subsequent queries, recordIds will be non-nil, so the trie will search
		// only records within that set.
		// The final set will contain only records that have ALL the query keywords.
		recordIds = idx.keywordTrie.recordIdsForPrefix(queryKeyword, recordIds)
	})
	if recordIds == nil {
		// The query didn't have any keywords, so it doesn't match any records.
		recordIds = newRecordIdSet()
	}
	return recordIds
}

func (idx *Index) rankRecords(query string, recordIds *recordIdSet) []int {
	var candidates []candidateRecord
	recordIds.forEach(func(recordId int) {
		// We are NOT lowercasing the strings here
		// so that case-sensitive matches will be ranked higher
		// than case-insensitive matches.
		candidates = append(candidates, candidateRecord{
			recordId: recordId,
			record:   norm.NFC.String(idx.records[recordId]),
		})
	})
	return rankRecords(candidates, norm.NFC.String(query), maxSearchResults)
}

// iterKeywords iterates through the keywords in a string.
// Keywords are substrings separated by spaces or punctuation.
func iterKeywords(s string, f func(string)) {
	var sb strings.Builder
	for _, r := range strings.ToLower(norm.NFC.String(s)) {
		if unicode.IsSpace(r) || unicode.IsPunct(r) {
			if sb.Len() > 0 {
				f(sb.String())
				sb.Reset()
			}
		} else {
			sb.WriteRune(r)
		}
	}

	if sb.Len() > 0 {
		f(sb.String())
	}
}
