package display

import (
	"github.com/gdamore/tcell/v2"

	"github.com/aretext/aretext/editor/state"
)

// DrawSearchQuery draws the search query (if any) on the last line of the screen.
// This overwrites the status bar.
func DrawSearchQuery(screen tcell.Screen, palette *Palette, query string, direction state.SearchDirection) {
	screenWidth, screenHeight := screen.Size()
	if screenHeight == 0 {
		return
	}

	row := screenHeight - 1
	sr := NewScreenRegion(screen, 0, row, screenWidth, 1)
	sr.Fill(' ', tcell.StyleDefault)
	sr.SetContent(0, 0, searchPrefixForDirection(direction), nil, palette.StyleForSearchPrefix())
	col := drawStringNoWrap(sr, query, 1, 0, palette.StyleForSearchQuery())
	sr.ShowCursor(col, 0)
}

func searchPrefixForDirection(direction state.SearchDirection) rune {
	switch direction {
	case state.SearchDirectionForward:
		return '/'
	case state.SearchDirectionBackward:
		return '?'
	default:
		panic("unrecognized direction")
	}
}
