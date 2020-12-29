package display

import (
	"github.com/gdamore/tcell"
	"github.com/wedaly/aretext/internal/pkg/input"
)

// DrawStatusBar draws a status bar on the last line of the screen.
func DrawStatusBar(screen tcell.Screen, inputMode input.ModeType, filePath string) {
	screenWidth, screenHeight := screen.Size()
	if screenHeight == 0 {
		return
	}

	row := screenHeight - 1
	sr := NewScreenRegion(screen, 0, row, screenWidth, 1)
	text, style := statusBarContent(inputMode, filePath)
	sr.Fill(' ', tcell.StyleDefault)
	drawStringNoWrap(sr, text, 0, 0, style)
}

func statusBarContent(inputMode input.ModeType, filePath string) (string, tcell.Style) {
	switch inputMode {
	case input.ModeTypeInsert:
		return "-- INSERT --", tcell.StyleDefault.Bold(true)
	default:
		return filePath, tcell.StyleDefault
	}
}
