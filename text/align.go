package text

import (
	"crypto/md5"
	"fmt"
	"io"
	"math"
	"sort"
)

// LineMatch represents a line matched between two documents.
type LineMatch struct {
	LeftLineNum  uint64
	RightLineNum uint64
}

// Align matches lines in the left document with identical lines in the right document.
func Align(leftReader, rightReader io.Reader) ([]LineMatch, error) {
	leftLineHashes, rightLineHashes, err := hashLeftAndRightLines(leftReader, rightReader)
	if err != nil {
		return nil, err
	}

	leftLineCount, rightLineCount := uint64(len(leftLineHashes)), uint64(len(rightLineHashes))

	if allLineHashesMatch(leftLineHashes, rightLineHashes) {
		// Fast path for exact match.
		matches := make([]LineMatch, len(leftLineHashes))
		for i := 0; i < len(leftLineHashes); i++ {
			matches[i] = LineMatch{LeftLineNum: uint64(i), RightLineNum: uint64(i)}
		}
		return matches, nil
	}

	// Align lines that occur exactly once in each document.
	// This should skip "insignificant" lines like whitespace and braces,
	// which are unlikely to be unique within a document.
	// Then expand to include adjacent matching lines.
	// (This is similar to the "patience diff" algorithm used in some version control systems,
	// except it doesn't attempt to recursively diff unmatched regions.)
	var matches []LineMatch
	var lastMatch LineMatch
	for _, match := range longestCommonSubsequence(uniqueSharedLines(leftLineHashes, rightLineHashes)) {
		i, j := match.LeftLineNum, match.RightLineNum

		if i > 0 && i <= lastMatch.LeftLineNum {
			// Already covered this match by extending a previous match.
			continue
		}

		// Extend to previous adjacent lines that exactly match.
		for i > lastMatch.LeftLineNum+1 && j > lastMatch.RightLineNum+1 && leftLineHashes[i-1] == rightLineHashes[j-1] {
			i--
			j--
		}

		// Include every adjacent line that exactly matches.
		for i < leftLineCount && j < rightLineCount && leftLineHashes[i] == rightLineHashes[j] {
			matches = append(matches, LineMatch{LeftLineNum: i, RightLineNum: j})
			i++
			j++
		}

		if len(matches) > 0 {
			// Keep track of the last match so we can skip overlapping matches.
			lastMatch = matches[len(matches)-1]
		}
	}

	return matches, nil
}

// lineHash represents a hash of a line in a document.
type lineHash [md5.Size]byte

// lineHashFromBytes returns a fixed-width hash from a byte slice.
// It assumes that the byte slice has the expected length.
func lineHashFromBytes(b []byte) lineHash {
	var lh lineHash
	for i := 0; i < md5.Size; i++ {
		lh[i] = b[i]
	}
	return lh
}

// hashLeftAndRightLines hashes the lines of two documents.
func hashLeftAndRightLines(leftReader, rightReader io.Reader) ([]lineHash, []lineHash, error) {
	leftLineHashes, err := hashLines(leftReader)
	if err != nil {
		return nil, nil, fmt.Errorf("hashLines(leftReader): %w", err)
	}

	rightLineHashes, err := hashLines(rightReader)
	if err != nil {
		return nil, nil, fmt.Errorf("hashLines(rightReader): %w", err)
	}

	return leftLineHashes, rightLineHashes, nil
}

// hashLines hashes the lines in a document.
func hashLines(r io.Reader) ([]lineHash, error) {
	h := md5.New()
	var result []lineHash
	var buf [256]byte
	var hashBytesWritten int
	for {
		n, err := r.Read(buf[:])
		if err != nil && err != io.EOF {
			return nil, fmt.Errorf("Read: %w", err)
		}

		var i int
		for j := 0; j < n; j++ {
			if buf[j] == '\n' {
				// Output a hash for bytes from the previous line up to and including the line feed.
				h.Write(buf[i : j+1])
				result = append(result, lineHashFromBytes(h.Sum(nil)))
				h.Reset()
				hashBytesWritten = 0
				i = j + 1
			}
		}
		h.Write(buf[i:n])
		hashBytesWritten += (n - i)

		if err == io.EOF {
			if hashBytesWritten > 0 {
				// Output a hash for bytes from the previous line to end-of-file.
				result = append(result, lineHashFromBytes(h.Sum(nil)))
			}
			break
		}
	}
	return result, nil
}

func allLineHashesMatch(left, right []lineHash) bool {
	if len(left) != len(right) {
		return false
	}

	for i := 0; i < len(left); i++ {
		if left[i] != right[i] {
			return false
		}
	}

	return true
}

// uniqueSharedLine represents a line that appears exactly once in each of two documents.
type uniqueSharedLine struct {
	hash         lineHash
	leftLineNum  int
	rightLineNum int
}

func uniqueSharedLines(leftLineHashes, rightLineHashes []lineHash) []uniqueSharedLine {
	leftLineCounts := countLinesByHash(leftLineHashes)
	rightLineCounts := countLinesByHash(rightLineHashes)

	var results []uniqueSharedLine
	for h, lc := range leftLineCounts {
		rc := rightLineCounts[h]
		if lc.count == 1 && rc.count == 1 {
			results = append(results, uniqueSharedLine{
				hash:         h,
				leftLineNum:  lc.lastLineNum,
				rightLineNum: rc.lastLineNum,
			})
		}
	}
	return results
}

type lineCount struct {
	count       int
	lastLineNum int
}

// countLinesByHash counts each distinct line in a document.
func countLinesByHash(lineHashes []lineHash) map[lineHash]lineCount {
	result := make(map[lineHash]lineCount, 0)
	for i, h := range lineHashes {
		prevCount := result[h].count
		result[h] = lineCount{
			lastLineNum: i,
			count:       prevCount + 1,
		}
	}
	return result
}

// longestCommonSubsequence returns the LCS for unique shared lines.
// This is the algorithm from:
//
//	Szymanski, T. G. (1975) A special case of the maximal common subsequence problem.
//	Computer Science Lab TR-170, Princeton University.
//
// The algorithm assumes that the two sequences have the same elements, possibly in a different
// order, and runs in time O(n log n).
func longestCommonSubsequence(lines []uniqueSharedLine) []LineMatch {
	// Sort lines ascending by left line number.
	sort.Slice(lines, func(i, j int) bool { return lines[i].leftLineNum < lines[j].leftLineNum })

	// lengths[i] represents the length of the LCS whose last element occurs at lines[i]
	// Intitially these are all zero.
	lengths := make([]int, len(lines))

	// thresholds[k] represents the shortest prefix of the right sequence
	// that has an LCS of length k with some prefix of the left sequence.
	// We build this up incrementally by adding each line from the left sequence.
	thresholds := make([]int, len(lines)+1)
	for k := 1; k < len(thresholds); k++ {
		thresholds[k] = math.MaxInt
	}

	// Calculate the thresholds and lengths arrays incrementally,
	// updating for each line from the left document.
	var maxLen int
	for i := 0; i < len(lines); i++ {
		j := lines[i].rightLineNum

		// Find k such that thresholds[k-1] < j < thresholds[k]
		//
		// By construction, right[0:thresholds[k-1]] has an LCS of length k-1 with left[0:i-1].
		// So when we "add" one more matching element from the left sequence,
		//   left[0:i-1]   =>   left[0:i]
		//   k-1           =>   k
		//
		// Thresholds are sorted ascending, so we can binary search to find the value of k,
		// which always exists and is unique (see Szymanksi (1975) for the proof).
		k := sort.Search(len(thresholds), func(k int) bool {
			// Return the smallest k for which thresholds[k] > j.
			return thresholds[k] > j
		})

		thresholds[k] = j
		lengths[i] = k
		if k > maxLen {
			maxLen = k
		}
	}

	// Reconstruct a reversed least common subsequence by working backwards
	// from longer prefixes to shorter prefixes of the left document.
	// At each iteration, check if the prefix ended with an LCS of the necessary length
	// (k=4, k=3, k=2, ...) and that the matching prefixes in the right sequence are
	// descending in length.
	results := make([]LineMatch, 0, len(lines))
	lastRightLineNum := math.MaxInt
	k := maxLen
	for i := len(lines) - 1; i >= 0; i-- {
		j := lines[i].rightLineNum
		if lengths[i] == k && j < lastRightLineNum {
			results = append(results, LineMatch{
				LeftLineNum:  uint64(lines[i].leftLineNum),
				RightLineNum: uint64(lines[i].rightLineNum),
			})
			lastRightLineNum = j
			k--
		}
	}

	// The results sequence is backwards, so reverse it to recover the LCS.
	for i, j := 0, len(results)-1; i < j; i, j = i+1, j-1 {
		results[i], results[j] = results[j], results[i]
	}

	return results
}
