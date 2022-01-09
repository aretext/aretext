package fuzzy

import (
	"container/heap"
	"math"
	"sort"
	"strings"
)

// ranker orders records by relevance to a search query.
type ranker struct {
	// The query against which records are ranked.
	query string

	// The maximum number of records to return.
	limit int

	// Used in the dynamic programming algorithm for calculating longest-common subsequence.
	numCols             int
	prevRow, currentRow []int

	// Store scores of records that have been ranked.
	// The dirty flag is set whenever a new record is ranked
	// and unset whenever the slice is sorted.
	scoredRecords scoredRecordHeap
	dirty         bool
}

// newRanker returns a new, empty ranker for a query.
// Limit controls the maximum number of records returned from the ranker.
func newRanker(query string, limit int) *ranker {
	numCols := len(query) + 1
	return &ranker{
		query:         query,
		limit:         limit,
		numCols:       numCols,
		prevRow:       make([]int, numCols),
		currentRow:    make([]int, numCols),
		scoredRecords: make([]scoredRecord, 0, limit),
	}
}

// addRecord ranks a record against the search query.
func (r *ranker) addRecord(recordId int, record string) {
	score, ok := r.scoreExactSubstringMatch(record)
	if !ok {
		score = r.scorePartialSubstringMatch(record)
	}

	sr := scoredRecord{recordId, record, score}
	if len(r.scoredRecords) < r.limit {
		heap.Push(&r.scoredRecords, sr)
		r.dirty = true
	} else if !r.scoredRecords[0].Less(sr) { // same order as scoredRecordHeap.Less()
		heap.Remove(&r.scoredRecords, 0)
		heap.Push(&r.scoredRecords, sr)
		r.dirty = true
	}
}

// rankedRecordIds returns a slice of all records, ordered from most- to least-relevant.
func (r *ranker) rankedRecordIds() []int {
	if r.dirty {
		sort.SliceStable(r.scoredRecords, func(i, j int) bool {
			return r.scoredRecords[i].Less(r.scoredRecords[j])
		})
		r.dirty = false
	}

	rankedRecordIds := make([]int, len(r.scoredRecords))
	for i, sr := range r.scoredRecords {
		rankedRecordIds[i] = sr.recordId
	}
	return rankedRecordIds
}

func (r *ranker) scoreExactSubstringMatch(record string) (int, bool) {
	if strings.HasPrefix(record, r.query) {
		if len(record) == len(r.query) {
			return math.MaxInt, true
		} else {
			return math.MaxInt - 1, true
		}
	} else if strings.Contains(record, r.query) {
		return math.MaxInt - 2, true
	} else {
		return 0, false
	}
}

func (r *ranker) scorePartialSubstringMatch(record string) int {
	// Clear the score matrix rows.
	for i := 0; i < r.numCols; i++ {
		r.prevRow[i], r.currentRow[i] = 0, 0
	}

	// Calculate the length of the longest-common subsequence
	// between the record and the query.
	// It uses a dynamic programming algorithm, keeping track
	// of the previous row in the DP table.
	var bestScore int
	for i := 0; i < len(record); i++ {
		for j := 0; j < len(r.query); j++ {
			col := j + 1
			var score int
			if record[i] == r.query[j] {
				score = r.prevRow[col-1] + 1
			} else {
				score = maxScore(r.currentRow[col-1], r.prevRow[col])
			}
			r.currentRow[col] = score
			bestScore = maxScore(bestScore, score)
		}
		r.prevRow, r.currentRow = r.currentRow, r.prevRow
	}

	return bestScore
}

func maxScore(s1, s2 int) int {
	if s1 > s2 {
		return s1
	} else {
		return s2
	}
}

// scoredRecord represents a record that has been assigned a score.
type scoredRecord struct {
	recordId int
	record   string
	score    int
}

// Order descending by score, then ascending by record string length,
// then ascending by lexicographic order.
func (r scoredRecord) Less(other scoredRecord) bool {
	if r.score != other.score {
		// slice.Sort always sorts in ascending order, and we want
		// the first item to have the highest score, so we treat
		// the higher-scoring item as "less" than the lower-scoring item.
		return r.score > other.score
	} else if len(r.record) != len(other.record) {
		return len(r.record) < len(other.record)
	} else {
		return r.record < other.record
	}
}

// scoredRecordHeap is a min heap of scored records.
// It implements both heap.Interface and sort.Interface
type scoredRecordHeap []scoredRecord

func (h scoredRecordHeap) Len() int {
	return len(h)
}

func (h scoredRecordHeap) Less(i, j int) bool {
	// We want the lowest scoring item to be first so we can replace it with a higher scoring item.
	// This is the the opposite of the usual sort order, so negate Less().
	return !h[i].Less(h[j])
}

func (h scoredRecordHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
}

func (h *scoredRecordHeap) Push(x interface{}) {
	*h = append(*h, x.(scoredRecord))
}

func (h *scoredRecordHeap) Pop() interface{} {
	x := (*h)[len(*h)-1]
	*h = (*h)[0 : len(*h)-1]
	return x
}
