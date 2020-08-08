package display

import (
	runewidth "github.com/mattn/go-runewidth"
)

// GraphemeClusterWidth returns the width in cells of a grapheme cluster.
// Non-displayable characters are assigned a width of zero.
// Full-width east asian characters are assigned a width of two.
func GraphemeClusterWidth(gc []rune) uint64 {
	if len(gc) == 0 {
		return 0
	}

	return uint64(runewidth.RuneWidth(gc[0]))
}
