package display

import (
	"io"
	"unicode/utf8"

	runewidth "github.com/mattn/go-runewidth"
)

const maxBytesPerToken = 64
const zeroWidthJoiner = '\u200D'

// splitUtf8Cells splits UTF-8 characters into "cells" in a terminal display.
//
// Cells are defined as:
//     1) A character with width > 0, optionally followed by one or more zero-width characters.
//        Example: lower-case e followed by an accent
//     2) Two or more characters with width > 0 interleaved with zero-width joiners.
//        Example: emoji of man + woman + child
//     3) Zero-width characters at the start of the string (each one is put in its own cell).
//     4) Newline characters, optionally followed by one or more zero-width characters (such as line feeds).
//
// The function assumes that the input characters are valid UTF-8.
func splitUtf8Cells(data []byte, atEOF bool) (advance int, token []byte, err error) {
	prevRune := '\x00'
	i := 0

	for i < len(data) {
		r, n := utf8.DecodeRune(data[i:])
		if n == 1 && r == utf8.RuneError {
			// Received invalid UTF-8, so wait for more bytes.
			return 0, nil, nil
		}

		if i == 0 && isCombiningChar(r) {
			// Combining char at the start of the input gets output as a separate cell.
			// For example, a file that starts with a zero-width joiner or accent character.
			// These characters can't be combined with anything, so output them individually.
			i += n
			return i, data[:i], nil
		}

		if i > 0 && !isCombiningChar(r) && prevRune != zeroWidthJoiner {
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

func isCombiningChar(r rune) bool {
	return runewidth.RuneWidth(r) == 0 && r != '\n'
}
