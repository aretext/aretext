package text

import (
	"io"

	"github.com/pkg/errors"

	"github.com/aretext/aretext/text/utf8"
)

func SearchNextInReader(q string, r io.Reader) (bool, uint64, error) {
	return NewSearcher(q).NextInReader(r)
}

func SearchAllInString(q string, text string) []uint64 {
	return NewSearcher(q).AllInString(text, nil)
}

// Searcher searches for an exact match of a query.
// It uses the Knuth-Morris-Pratt algorithm, which runs in O(n+m) time, where n is the length
// of the text and m is the length of the query.
type Searcher struct {
	query               string
	queryStartByteCount uint64
	prefixTable         []int
	offsetLimit         *uint64
}

func NewSearcher(query string) *Searcher {
	// Count the number of UTF8 start bytes in the query.
	// For forward search, this is equivalent to the rune length;
	// for backward search, however, the query bytes are reversed
	// so the query won't necessarily be valid UTF-8.
	var queryStartByteCount uint64
	for i := 0; i < len(query); i++ {
		queryStartByteCount += uint64(utf8.StartByteIndicator[query[i]])
	}

	return &Searcher{
		query:               query,
		queryStartByteCount: queryStartByteCount,
		prefixTable:         buildPrefixTable(query),
	}
}

// Limit sets the maximum offset (in rune positions) for the end of a match.
// For example, a limit of 3 would allow matches that end on the second
// rune from the reader, but not on the following runes.
func (s *Searcher) Limit(offset uint64) *Searcher {
	s.offsetLimit = &offset
	return s
}

// NextInReader finds the next occurrence of a query in the text produced by an io.Reader.
// If it finds a match, it returns the offset (in rune positions) from the start of the reader.
func (s *Searcher) NextInReader(r io.Reader) (bool, uint64, error) {
	if len(s.query) == 0 {
		return false, 0, nil
	}

	var i int
	var offsetToEnd uint64
	var buf [256]byte
	for {
		n, err := r.Read(buf[:])
		if err == io.EOF {
			if n == 0 {
				return false, 0, nil
			}
		} else if err != nil {
			return false, 0, errors.Wrap(err, "Read")
		}

		var j int
		for j < n {
			i, j, offsetToEnd = s.advance(i, j, offsetToEnd, buf[j])
			if s.offsetLimit != nil && offsetToEnd > *s.offsetLimit {
				// No match found within offset limit.
				return false, 0, nil
			}

			if i == len(s.query) {
				// Found a substring match, so calculate the offset (in rune positions) and return.
				offsetToStart := offsetToEnd - s.queryStartByteCount
				return true, offsetToStart, nil
			}
		}
	}
}

// AllInString finds all (possibly overlapping) matches of the query in a string.
// It returns the rune positions for the start of each match.
// If not nil, the matchPositions slice will be used to store the results
// (avoids allocating a new slice for each call).
func (s *Searcher) AllInString(text string, matchPositions []uint64) []uint64 {
	if len(s.query) == 0 {
		return nil
	}

	if matchPositions != nil {
		matchPositions = matchPositions[:0]
	}

	var i, j int
	var offsetToEnd uint64
	for j < len(text) {
		i, j, offsetToEnd = s.advance(i, j, offsetToEnd, text[j])
		if s.offsetLimit != nil && offsetToEnd > *s.offsetLimit {
			// No match found within offset limit.
			break
		}

		if i == len(s.query) {
			// Found a substring match, so calculate the offset (in rune positions) and add it to the result set.
			offsetToStart := offsetToEnd - s.queryStartByteCount
			matchPositions = append(matchPositions, offsetToStart)
			offsetToEnd = offsetToStart + uint64(utf8.StartByteIndicator[text[j-i]])
			j = j - i + 1
			i = 0
		}
	}
	return matchPositions
}

func (s *Searcher) advance(i int, j int, offsetToEnd uint64, textByte byte) (int, int, uint64) {
	if s.query[i] != textByte {
		if i > 0 {
			// Backtrack to the next-longest prefix and retry.
			i = s.prefixTable[i-1]
		} else {
			// No possible match at this index, so continue searching at the next character.
			offsetToEnd += uint64(utf8.StartByteIndicator[textByte])
			j++
		}
	} else {
		// This character matches the query, so check the next character.
		offsetToEnd += uint64(utf8.StartByteIndicator[textByte])
		i++
		j++
	}

	return i, j, offsetToEnd
}

func buildPrefixTable(q string) []int {
	prefixTable := make([]int, len(q))
	i, j := 0, 1
	for j < len(q) {
		if q[i] != q[j] {
			// Next character in the suffix does not match the next character after the prefix.
			if i > 0 {
				// Backtrack to next-longest prefix and retry.
				i = prefixTable[i-1]
			} else {
				// No prefix matches this suffix.
				prefixTable[j] = 0
				j++
			}
		} else {
			// Next character in the suffix matches the next character after the last prefix,
			// so continue the last prefix.
			prefixLen := i + 1
			prefixTable[j] = prefixLen
			i++
			j++
		}
	}
	return prefixTable
}
