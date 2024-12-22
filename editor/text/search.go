package text

import (
	"fmt"
	"io"

	"github.com/aretext/aretext/editor/text/utf8"
)

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

// NoLimit removes any limit set on the Searcher.
func (s *Searcher) NoLimit() *Searcher {
	s.offsetLimit = nil
	return s
}

// NextInReader finds the next occurrence of a query in the text produced by an io.Reader.
// If it finds a match, it returns the offset (in rune positions) from the start of the reader.
func (s *Searcher) NextInReader(r io.Reader) (bool, uint64, error) {
	return s.searchInReader(r, searchModeFirstMatch)
}

// LastInReader finds the last occurrence of a query in the text produced by an io.Reader.
// If it finds a match, it returns the offset (in rune positions) from the start of the reader.
func (s *Searcher) LastInReader(r io.Reader) (bool, uint64, error) {
	return s.searchInReader(r, searchModeLastMatch)
}

// searchMode controls whether to return the first or last match.
type searchMode int

const (
	searchModeFirstMatch = searchMode(iota)
	searchModeLastMatch
)

func (s *Searcher) searchInReader(r io.Reader, mode searchMode) (bool, uint64, error) {
	if len(s.query) == 0 {
		return false, 0, nil
	}

	var i int
	var offsetToEnd uint64
	var foundMatch bool
	var matchOffsetToStart uint64
	var buf [256]byte
	for {
		n, err := r.Read(buf[:])
		if err == io.EOF {
			if n == 0 {
				goto done
			}
		} else if err != nil {
			return false, 0, fmt.Errorf("Read: %w", err)
		}

		var j int
		for j < n {
			i, j, offsetToEnd = s.advance(i, j, offsetToEnd, buf[j])
			if s.offsetLimit != nil && offsetToEnd > *s.offsetLimit {
				// Past limit set on the searcher.
				goto done
			}

			if i == len(s.query) {
				// Found a substring match.
				foundMatch = true
				matchOffsetToStart = offsetToEnd - s.queryStartByteCount
				switch mode {
				case searchModeFirstMatch:
					goto done // Return the first match found.
				case searchModeLastMatch:
					i = 0 // Keep searching for a later match.
				default:
					panic("invalid search mode")
				}
			}
		}
	}

done:
	return foundMatch, matchOffsetToStart, nil
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
