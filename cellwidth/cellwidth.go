package cellwidth

import (
	"os"
	"strings"

	"github.com/clipperhouse/displaywidth"

	"github.com/aretext/aretext/text"
)

// Copied from https://github.com/gdamore/tcell/blob/29b8586f6141d93aac712fd29fec38224b624625/internal/widthutil/widthutil.go
func textWidthOptionsTcell() displaywidth.Options {
	if rw := strings.ToLower(os.Getenv("RUNEWIDTH_EASTASIAN")); rw == "1" || rw == "true" || rw == "yes" {
		return displaywidth.Options{EastAsianWidth: true}
	}
	return displaywidth.Options{}
}

var textWidthOptions = textWidthOptionsTcell()

// Sizer determines the cellwidth of grapheme clusters.
type Sizer interface {
	GraphemeClusterWidth(gc []rune, offsetInLine uint64) uint64
}

type sizer struct {
	tabSize     uint64
	showUnicode bool
	escaper     *text.Escaper
}

// New constructs a new Sizer with the specified configuration.
func New(tabSize uint64, showUnicode bool) Sizer {
	return &sizer{
		tabSize:     tabSize,
		showUnicode: showUnicode,
		escaper:     &text.Escaper{},
	}
}

// GraphemeClusterWidth returns the width in cells of a grapheme cluster.
// It attempts to handle combining characters, emoji, and regional indicators reasonably,
// but can't be 100% accurate without knowing how the terminal will render the glyphs.
// Tab width is determined based on the position within the line.
func (s *sizer) GraphemeClusterWidth(gc []rune, offsetInLine uint64) uint64 {
	if len(gc) == 0 {
		return 0
	}

	// Tab width depends on offset in the line.
	// For example, a tab at the start of the line occupies 4 spaces,
	// but a tab at the second character occupies only 2 spaces.
	// This ensures that characters after the tab "line up" at even multiples
	// of the tab size.
	if gc[0] == '\t' {
		nextTabPosition := ((offsetInLine / s.tabSize) + 1) * s.tabSize
		return nextTabPosition - offsetInLine
	}

	// If displaying non-ascii unicode codepoints, calculate the width for
	// the unicode sequence (e.g. "<U+1F9D4,U+200D,U+2642,U+FE0F>")
	if s.showUnicode && !(len(gc) == 1 && gc[0] <= 127) {
		return uint64(s.escaper.RunesToStrLen(gc))
	}

	// It is very important that the cell width matches what tcell expects.
	// Since version 3.3.0, tcell uses clipperhouse/displaywidth to determine cell width,
	// so we do the same. Unfortunately, this requires converting the
	// grapheme cluster []rune to string, but there's currently no other mechanism available.
	width := textwidthoptions.String(string(gc))

	return uint64(width)
}
