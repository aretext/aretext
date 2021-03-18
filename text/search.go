package text

import (
	"io"

	"github.com/aretext/aretext/text/utf8"
	"github.com/pkg/errors"
)

// Search finds the first occurrence of a query in a text.
// If it finds a match, it returns the offset (in rune positions) from the start of the reader.
// This uses the Knuth-Morris-Pratt algorithm, which runs in O(n+m) time, where n is the length
// of the text and m is the length of the query.
func Search(q string, r io.Reader) (bool, uint64, error) {
	if len(q) == 0 {
		return false, 0, nil
	}

	var i int
	var offsetToEnd uint64
	var buf [256]byte
	prefixTable := buildPrefixTable(q)
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
			if q[i] != buf[j] {
				if i > 0 {
					// Backtrack to the next-longest prefix and retry.
					i = prefixTable[i-1]
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

			if i == len(q) {
				// Found a substring match, so calculate the offset (in rune positions) and return.
				offset := offsetToEnd
				for k := 0; k < len(q); k++ {
					offset -= uint64(utf8.StartByteIndicator[q[k]])
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
