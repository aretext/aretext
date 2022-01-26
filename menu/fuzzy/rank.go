package fuzzy

import (
	"container/heap"
	"sync"
	"unicode"
	"unicode/utf8"
)

const (
	minRecordsPerPartition            = 64  // each goroutine will be assigned at least this many records.
	maxNumPartitions                  = 128 // maximum number of goroutines used to score records.
	deleteQueryCharCost               = 1.5
	insertQueryCharCost               = 1.0
	replaceQueryCharCost              = 1.0
	matchCharScore                    = 1.0
	matchCharDifferentCaseScore       = 0.2
	alignAtStartOrAfterSeparatorBonus = 0.5
)

// candidateRecord is a record that can be scored by the ranking algorithm.
type candidateRecord struct {
	recordId int
	record   string
}

// rankRecords scores records against a query, then returns the IDs of the top records in descending order by score.
// It uses an O(nm) dynamic programming algorithm to score records, where n and m are the lengths
// of a record and query respectively.
// The order of candidates does not matter; any order will return the same result.
func rankRecords(candidates []candidateRecord, query string, limit int) []int {
	return topRecordsDescByScore(scoreAllRecords(candidates, query), limit)
}

// scoreAllRecords assigns scores to every candidate record based on the query.
func scoreAllRecords(candidates []candidateRecord, query string) []scoredRecord {
	scoredRecords := make([]scoredRecord, len(candidates))
	for i := 0; i < len(scoredRecords); i++ {
		scoredRecords[i].recordId = candidates[i].recordId
		scoredRecords[i].record = candidates[i].record
	}

	numPartitions := numPartitions(len(scoredRecords))
	if numPartitions == 1 {
		// If the number of records is small, avoid the overhead of starting a separate goroutine.
		scoreRecordsPartition(scoredRecords, query)
		return scoredRecords
	}

	// Start multiple goroutines to score records in parallel.
	// Each goroutine is assigned a non-overlapping partition of candidate records.
	var wg sync.WaitGroup
	recordsPerPartition := len(scoredRecords)/numPartitions + 1
	for start := 0; start < len(scoredRecords); start += recordsPerPartition {
		end := start + recordsPerPartition
		if end > len(scoredRecords) {
			end = len(scoredRecords)
		}

		wg.Add(1)
		go func(partition []scoredRecord, query string) {
			defer wg.Done()
			scoreRecordsPartition(partition, query)
		}(scoredRecords[start:end], query)
	}
	wg.Wait()

	return scoredRecords
}

func numPartitions(numRecords int) int {
	n := numRecords / minRecordsPerPartition
	if n < 1 {
		return 1
	} else if n > maxNumPartitions {
		return maxNumPartitions
	} else {
		return n
	}
}

// scoreRecordsPartition assigns scores to all records in a partition.
func scoreRecordsPartition(partition []scoredRecord, query string) {
	// We are using a dynamic programming algorithm to score each record.
	// Conceptually, the algorithm "fills in" a scoring matrix, where each column
	// represents a prefix of the query, and each row represents a prefix of the
	// record. To reduce the space complexity, store only the current and previous
	// rows of the scoring matrix.
	numCols := utf8.RuneCountInString(query) + 1
	prevRow := make([]float64, numCols)
	currentRow := make([]float64, numCols)
	for i := 0; i < len(partition); i++ {
		record := partition[i].record

		// Initialize the dynamic programming score matrix rows.
		for col := 0; col < numCols; col++ {
			// Penalize deletion of query characters.
			prevRow[col] = -1.0 * deleteQueryCharCost * float64(col)

			// Reset the current row in case there are leftover values from the last record.
			// This isn't strictly necessary since we overwrite these values anyway, but it's safer
			// to always start from a clean slate.
			currentRow[col] = 0.0
		}

		var prevRecordRune rune
		row := 1
		for _, recordRune := range record {
			// No penalty to skip characters at the beginning of the record.
			// This allows the query to "align" at any position in the record.
			currentRow[0] = 0
			col := 1
			for _, queryRune := range query {
				// Score deleting a character from the query.
				// This is expensive!
				currentRow[col] = currentRow[col-1] - deleteQueryCharCost

				// Score inserting a character into the query.
				// Also expensive, but not quite as bad as deleting a character.
				insertScore := prevRow[col] - insertQueryCharCost
				if insertScore > currentRow[col] {
					currentRow[col] = insertScore
				}

				// Calculate similarity between the query rune and record rune.
				// It's measurably faster to inline this to avoid the cost of a function call.
				var runeSimilarity float64
				if queryRune == recordRune {
					runeSimilarity = matchCharScore
				} else {
					// Check for a case-insensitive match between query rune and record rune.
					// We special-case ASCII runes to avoid an expensive call to unicode.ToLower.
					caseInsensitiveMatch := (('A' <= queryRune && queryRune <= 'Z' && (queryRune+32) == recordRune) ||
						('A' <= recordRune && recordRune <= 'Z' && (recordRune+32) == queryRune) ||
						(queryRune > unicode.MaxASCII && recordRune > unicode.MaxASCII && (unicode.ToLower(queryRune) == unicode.ToLower(recordRune))))
					if caseInsensitiveMatch {
						runeSimilarity = matchCharDifferentCaseScore
					}
				}

				if runeSimilarity > 0.0 {
					// Query rune and record rune match!
					matchScore := prevRow[col-1] + runeSimilarity
					if col == 1 && (row == 1 || (isSeparator(prevRecordRune) && !isSeparator(recordRune))) {
						// Break ties in favor of matches at the start of words or path components.
						matchScore += alignAtStartOrAfterSeparatorBonus
					}
					if matchScore > currentRow[col] {
						currentRow[col] = matchScore
					}
				} else {
					// Replace the query character with a different character to match the record.
					replaceScore := prevRow[col-1] - replaceQueryCharCost
					if replaceScore > currentRow[col] {
						currentRow[col] = replaceScore
					}
				}

				col++
			}

			// Update the "best" score for this record.
			// The score at currentRow[numCols-1] represents the best score for a record prefix
			// in which every character in the query (but not necessarily the record) has been consumed.
			rowScore := currentRow[numCols-1]
			if row == 1 {
				// Always take the first row's score, since no other score has been set yet.
				partition[i].score = rowScore
			} else if rowScore > partition[i].score {
				// This row's score is the best score so far.
				partition[i].score = rowScore
			}

			prevRow, currentRow = currentRow, prevRow
			prevRecordRune = recordRune
			row++
		}

		// Normalize scores so that the maximum possible score is 1.0, representing an exact match
		// between the record and the query. If two records match the same number of query characters,
		// the shorter string will rank higher because a larger percentage of its characters matched the query.
		partition[i].score /= (alignAtStartOrAfterSeparatorBonus + matchCharScore*float64(row))
	}
}

func isSeparator(r rune) bool {
	return r == '/' || r == ' ' || r == '\t'
}

func topRecordsDescByScore(scoredRecords []scoredRecord, limit int) []int {
	numResults := len(scoredRecords)
	if numResults > limit {
		numResults = limit
	}
	result := make([]int, 0, numResults)
	h := scoredRecordHeap(scoredRecords)
	heap.Init(&h)
	for i := 0; i < numResults; i++ {
		sr := heap.Pop(&h).(scoredRecord)
		result = append(result, sr.recordId)
	}
	return result
}

type scoredRecord struct {
	recordId int
	record   string
	score    float64
}

// scoredRecordHeap is a min heap of scored records.
// It implements both heap.Interface and sort.Interface
type scoredRecordHeap []scoredRecord

func (h scoredRecordHeap) Len() int {
	return len(h)
}

func (h scoredRecordHeap) Less(i, j int) bool {
	if h[i].score != h[j].score {
		// min-heaps always sorts in ascending order, and we want
		// the first item to have the highest score, so we treat
		// the higher-scoring item as "less" than the lower-scoring item.
		return h[i].score > h[j].score
	} else {
		return h[i].record < h[j].record
	}
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
