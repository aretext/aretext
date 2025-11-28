package display

import (
	"unicode"

	"github.com/gdamore/tcell/v2"
)

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
			sr.Put(col, row, " ", style)
			col++
		}

		// Draw a special character to represent a tab.
		if gc[0] == '\t' && showTabs {
			sr.Put(startCol, row, string(tcell.RuneRArrow), style.Dim(true))
		}

		// Draw a special character to represent a space.
		if gc[0] == ' ' && showSpaces {
			sr.Put(startCol, row, string(tcell.RuneBullet), style.Dim(true))
		}
	} else {
		// For everything else, tcell expects the grapheme cluster contents to be
		// written into the first of the cell(s) occupied by the gc.
		sr.Put(col, row, string(gc), style)
	}
}
