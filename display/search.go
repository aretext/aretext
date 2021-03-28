package display

import (
	"github.com/aretext/aretext/state"
	"github.com/aretext/aretext/text"
	"github.com/gdamore/tcell/v2"
)

// DrawSearchQuery draws the search query (if any) on the last line of the screen.
// This overwrites the status bar.
func DrawSearchQuery(screen tcell.Screen, inputMode state.InputMode, query string, direction text.ReadDirection) {
	if inputMode != state.InputModeSearch {
		return
	}

	screenWidth, screenHeight := screen.Size()
	if screenHeight == 0 {
		return
	}

	row := screenHeight - 1
	sr := NewScreenRegion(screen, 0, row, screenWidth, 1)
	sr.Fill(' ', tcell.StyleDefault)
	sr.SetContent(0, 0, searchPrefixForDirection(direction), nil, tcell.StyleDefault)
	col := drawStringNoWrap(sr, query, 1, 0, tcell.StyleDefault)
	sr.ShowCursor(col, 0)
}

func searchPrefixForDirection(direction text.ReadDirection) rune {
	switch direction {
	case text.ReadDirectionForward:
		return '/'
	case text.ReadDirectionBackward:
		return '?'
	default:
		panic("unrecognized direction")
	}
}
