package display

import (
	"io"
	"log"
	"unicode"

	"github.com/gdamore/tcell/v2"

	"github.com/aretext/aretext/cellwidth"
	"github.com/aretext/aretext/config"
	"github.com/aretext/aretext/text"
	"github.com/aretext/aretext/text/segment"
)

func drawStringNoWrap(sr *ScreenRegion, s string, col int, row int, style tcell.Style) int {
	maxLineWidth, _ := sr.Size()
	runeIter := text.NewRuneIterForSlice([]rune(s))
	gcIter := segment.NewGraphemeClusterIter(runeIter)
	gc := segment.Empty()
	for {
		err := gcIter.NextSegment(gc)
		if err == io.EOF {
			break
		} else if err != nil {
			log.Fatalf("%s", err)
		}

		gcRunes := gc.Runes()
		gcWidth := cellwidth.GraphemeClusterWidth(gcRunes, uint64(col), config.DefaultTabSize)
		if uint64(col)+gcWidth > uint64(maxLineWidth) {
			break
		}

		drawGraphemeCluster(sr, col, row, gc.Runes(), int(gcWidth), style)
		col += int(gcWidth) // Safe to downcast because there's a limit on the number of cells a grapheme cluster can occupy.
	}

	return col
}

func drawGraphemeCluster(sr *ScreenRegion, col, row int, gc []rune, gcWidth int, style tcell.Style) {
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
