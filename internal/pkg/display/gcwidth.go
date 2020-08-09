package display

import (
	runewidth "github.com/mattn/go-runewidth"
)

const tabWidth = uint64(4)

// GraphemeClusterWidth returns the width in cells of a grapheme cluster.
// Non-displayable characters are assigned a width of zero.
// Full-width east asian characters are assigned a width of two.
func GraphemeClusterWidth(gc []rune, offsetInLine uint64) uint64 {
	if len(gc) == 0 {
		return 0
	}

	if gc[0] == '\t' {
		nextTabPosition := ((offsetInLine / tabWidth) + 1) * tabWidth
		return nextTabPosition - offsetInLine
	}

	return uint64(runewidth.RuneWidth(gc[0]))
}
