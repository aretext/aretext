package display

import (
	"io"
	"unicode/utf8"

	runewidth "github.com/mattn/go-runewidth"
)

const maxBytesPerToken = 64
const zeroWidthJoiner = '\u200D'

// SplitNonCombiningCharacters splits UTF-8 bytes before each non-combining character.
// For example, "\u0065\u0301\u007A" (lowercase e + combining acute accent + lowercase z)
// would be split into "\u0065\u0301" and "\u007A", because the accent is a combining character.
// SplitNonCombiningCharacters is meant to be used as the split function for a bufio.Scanner.
// It assumes the input characters are valid UTF-8, although multi-byte characters may be split between reads.
// If too many zero-width characters are received consecutively, the split func will output a token without waiting for a non-combining character.
func SplitNonCombiningCharacters(data []byte, atEOF bool) (advance int, token []byte, err error) {
	prevRune := '\x00'
	i := 0

	for i < len(data) {
		r, n := utf8.DecodeRune(data[i:])
		if n == 1 && r == utf8.RuneError {
			// Received invalid UTF-8, so wait for more bytes.
			return 0, nil, nil
		}

		if i > 0 && runewidth.RuneWidth(r) > 0 && prevRune != zeroWidthJoiner {
			// Consume up to the first non-combining character.
			return i, data[:i], nil
		}

		i += n

		if i >= maxBytesPerToken {
			// If we've received too many non-combining characters, return what we have.
			// Otherwise, bufio.Scanner might panic because we aren't advancing the input.
			return i, data[:i], nil
		}

		prevRune = r
	}

	if atEOF {
		if len(data) > 0 {
			// Consume remaining characters at end of file.
			return len(data), data[:], nil
		} else {
			return 0, nil, io.EOF
		}
	}

	// wait for more characters to arrive
	return 0, nil, nil
}
