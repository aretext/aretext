package cellwidth

import (
	"github.com/rivo/uniseg"
)

// GraphemeClusterWidth returns the width in cells of a grapheme cluster.
// It attempts to handle combining characters, emoji, and regional indicators reasonably,
// but can't be 100% accurate without knowing how the terminal will render the glyphs.
// Tab width is determined based on the position within the line.
func GraphemeClusterWidth(gc []rune, offsetInLine uint64, tabSize uint64) uint64 {
	if len(gc) == 0 {
		return 0
	}

	// Tab width depends on offset in the line.
	// For example, a tab at the start of the line occupies 4 spaces,
	// but a tab at the second character occupies only 2 spaces.
	// This ensures that characters after the tab "line up" at even multiples
	// of the tab size.
	if gc[0] == '\t' {
		nextTabPosition := ((offsetInLine / tabSize) + 1) * tabSize
		return nextTabPosition - offsetInLine
	}

	// It is very important that the cell width matches what tcell expects.
	// Since version 2.11, tcell uses rivo/uniseg to determine cell width,
	// so we do the same. Unfortunately, this requires converting the
	// grapheme cluster []rune to string (which uniseg then decodes back to
	// runes internally), but there's currently no other mechanism available.
	width := uniseg.StringWidth(string(gc))
	return uint64(width)
}
