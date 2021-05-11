package cellwidth

import (
	"unicode"

	runewidth "github.com/mattn/go-runewidth"

	"github.com/aretext/aretext/text/segment"
)

// RuneWidth returns the width in cells of an individual rune.
// Non-displayable characters and non-spacing marks are assigned a width of zero.
// Full-width East Asian characters are assigned a width of one.
func RuneWidth(r rune) uint64 {
	// Skip non-spacing marks.
	if unicode.Is(unicode.Mn, r) {
		return 0
	}

	// The go-runewidth library handles East Asian characters.
	// tcell also uses this library internally to calculate the cell width,
	// and it's important that we are consistent with tcell (otherwise strange
	// display artifacts can occur).
	return uint64(runewidth.RuneWidth(r))
}

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

	if segment.GraphemeClusterIsEmoji(gc) || segment.GraphemeClusterIsRegionalIndicator(gc) {
		return RuneWidth(gc[0])
	}

	w := uint64(0)
	for _, r := range gc {
		w += RuneWidth(r)
	}
	return w
}
