package text

import (
	"io"

	"github.com/aretext/aretext/text/utf8"
	"github.com/pkg/errors"
)

func SearchNextInReader(q string, r io.Reader) (bool, uint64, error) {
	return NewSearcher(q).NextInReader(r)
}

// Searcher searches for an exact match of a query.
// It uses the Knuth-Morris-Pratt algorithm, which runs in O(n+m) time, where n is the length
// of the text and m is the length of the query.
type Searcher struct {
	query       string
	prefixTable []int
}

func NewSearcher(query string) Searcher {
	return Searcher{
		query:       query,
		prefixTable: buildPrefixTable(query),
	}
}

// NextInReader finds the next occurrence of a query in the text produced by an io.Reader.
// If it finds a match, it returns the offset (in rune positions) from the start of the reader.
func (s Searcher) NextInReader(r io.Reader) (bool, uint64, error) {
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
			return false, 0, errors.Wrapf(err, "Read")
		}

		var j int
		for j < n {
			if s.query[i] != buf[j] {
				if i > 0 {
					// Backtrack to the next-longest prefix and retry.
					i = s.prefixTable[i-1]
				} else {
					// No possible match at this index, so continue searching at the next character.
					offsetToEnd += uint64(utf8.StartByteIndicator[buf[j]])
					j++
				}
			} else {
				// This character matches the query, so check the next character.
				offsetToEnd += uint64(utf8.StartByteIndicator[buf[j]])
				i++
				j++
			}

			if i == len(s.query) {
				// Found a substring match, so calculate the offset (in rune positions) and return.
				offset := offsetToEnd
				for k := 0; k < len(s.query); k++ {
					offset -= uint64(utf8.StartByteIndicator[s.query[k]])
				}
				return true, offset, nil
			}
		}
	}
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
