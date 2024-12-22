package display

import (
	"unicode"
	"unicode/utf8"

	"github.com/gdamore/tcell/v2"

	"github.com/aretext/aretext/editor/cellwidth"
	"github.com/aretext/aretext/editor/config"
	"github.com/aretext/aretext/editor/text/segment"
)

func drawStringNoWrap(sr *ScreenRegion, s string, col int, row int, style tcell.Style) int {
	maxLineWidth, _ := sr.Size()
	var gcBreaker segment.GraphemeClusterBreaker
	gcRunes := []rune{'\x00', '\x00', '\x00', '\x00'}[:0] // Stack-allocate runes for the last grapheme cluster.
	var i int
	for {
		if i >= len(s) && len(gcRunes) == 0 {
			break
		}

		r, rsize := utf8.DecodeRuneInString(s[i:])
		i += rsize
		canBreakBefore := gcBreaker.ProcessRune(r)
		if canBreakBefore && len(gcRunes) > 0 {
			gcWidth := cellwidth.GraphemeClusterWidth(gcRunes, uint64(col), config.DefaultTabSize)
			if uint64(col)+gcWidth > uint64(maxLineWidth) {
				break
			}
			drawGraphemeCluster(sr, col, row, gcRunes, int(gcWidth), style, false, false)
			col += int(gcWidth) // Safe to downcast because there's a limit on the number of cells a grapheme cluster can occupy.
			gcRunes = gcRunes[:0]
		}

		if rsize > 0 {
			gcRunes = append(gcRunes, r)
		}
	}

	return col
}

func drawGraphemeCluster(
	sr *ScreenRegion,
	col, row int,
	gc []rune,
	gcWidth int,
	style tcell.Style,
	showTabs bool,
	showSpaces bool,
) {
	startCol := col

	// Style whitespace (newlines, tabs, etc.) but don't set any runes.
	// This prevents drawing artifacts with '\r\n' where tcell
	// sends the combining character ('\n') to the terminal.
	if unicode.IsSpace(gc[0]) {
		// Always style at least one cell, even when the gcWidth is zero (carriage returns).
		// Tabs will usually style multiple cells.
		for col == startCol || col < startCol+gcWidth {
			sr.SetContent(col, row, ' ', nil, style)
			col++
		}

		// Draw a special character to represent a tab.
		if gc[0] == '\t' && showTabs {
			sr.SetContent(startCol, row, tcell.RuneRArrow, nil, style.Dim(true))
		}

		// Draw a special character to represent a space.
		if gc[0] == ' ' && showSpaces {
			sr.SetContent(startCol, row, tcell.RuneBullet, nil, style.Dim(true))
		}

		return
	}

	// Emoji and regional indicator sequences are usually rendered using the
	// width of the first rune.  This won't support every terminal, but it's probably
	// the best we can do without knowing how the terminal will render the glyphs.
	if segment.GraphemeClusterIsEmoji(gc) || segment.GraphemeClusterIsRegionalIndicator(gc) {
		sr.SetContent(col, row, gc[0], gc[1:], style)
		return
	}

	// For other sequences, we break the grapheme cluster into cells.
	// Each cell starts with a main rune, followed by zero or more combining runes.
	// In most cases, the entire grapheme cluster will fit in a single cell,
	// but there are exceptions (for example, some Thai sequences).
	i := 0
	for i < len(gc) {
		j := i + 1
		for j < len(gc) {
			r := gc[j]
			if cellwidth.RuneWidth(r) > 0 {
				break
			}
			j++
		}
		sr.SetContent(col, row, gc[i], gc[i+1:j], style)
		col += int(cellwidth.RuneWidth(gc[i]))
		i = j
	}
}
