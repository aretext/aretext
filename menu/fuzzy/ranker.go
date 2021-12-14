package fuzzy

import (
	"math"
	"sort"
	"strings"
)

// ranker orders records by relevance to a search query.
type ranker struct {
	// The query against which records are ranked.
	query string

	// Used in the dynamic programming algorithm for calculating longest-common subsequence.
	numCols             int
	prevRow, currentRow []int

	// Store scores of records that have been ranked.
	// The dirty flag is set whenever a new record is ranked
	// and unset whenever the slice is sorted.
	scoredRecords []scoredRecord
	dirty         bool
}

type scoredRecord struct {
	recordId int
	record   string
	score    int
}

// newRanker returns a new, empty ranker for a query.
func newRanker(query string, capacity int) *ranker {
	numCols := len(query) + 1
	return &ranker{
		query:         query,
		numCols:       numCols,
		prevRow:       make([]int, numCols),
		currentRow:    make([]int, numCols),
		scoredRecords: make([]scoredRecord, 0, capacity),
	}
}

// addRecord ranks a record against the search query.
func (r *ranker) addRecord(recordId int, record string) {
	score, ok := r.scoreExactSubstringMatch(record)
	if !ok {
		score = r.scorePartialSubstringMatch(record)
	}
	r.scoredRecords = append(r.scoredRecords, scoredRecord{
		recordId: recordId,
		record:   record,
		score:    score,
	})
	r.dirty = true
}

// rankedRecordIds returns a slice of all records, ordered from most- to least-relevant.
func (r *ranker) rankedRecordIds() []int {
	if r.dirty {
		// Sort descending by score, then ascending by record string length,
		// then ascending by lexicographic order.
		sort.SliceStable(r.scoredRecords, func(i, j int) bool {
			if r.scoredRecords[i].score != r.scoredRecords[j].score {
				return r.scoredRecords[i].score > r.scoredRecords[j].score
			} else if len(r.scoredRecords[i].record) != len(r.scoredRecords[j].record) {
				return len(r.scoredRecords[i].record) < len(r.scoredRecords[j].record)
			} else {
				return r.scoredRecords[i].record < r.scoredRecords[j].record
			}
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
