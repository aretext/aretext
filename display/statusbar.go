package display

import (
	"github.com/aretext/aretext/file"
	"github.com/aretext/aretext/state"
	"github.com/gdamore/tcell/v2"
)

// DrawStatusBar draws a status bar on the last line of the screen.
func DrawStatusBar(screen tcell.Screen, statusMsg state.StatusMsg, inputMode state.InputMode, filePath string) {
	screenWidth, screenHeight := screen.Size()
	if screenHeight == 0 {
		return
	}

	row := screenHeight - 1
	sr := NewScreenRegion(screen, 0, row, screenWidth, 1)
	sr.Fill(' ', tcell.StyleDefault)
	text, style := statusBarContent(statusMsg, inputMode, filePath)
	drawStringNoWrap(sr, text, 0, 0, style)
}

func statusBarContent(statusMsg state.StatusMsg, inputMode state.InputMode, filePath string) (string, tcell.Style) {
	if len(statusMsg.Text) > 0 {
		return statusMsg.Text, styleForStatusMsg(statusMsg)
	}

	switch inputMode {
	case state.InputModeInsert:
		return "-- INSERT --", tcell.StyleDefault.Bold(true)
	case state.InputModeVisual:
		return "-- VISUAL --", tcell.StyleDefault.Bold(true)
	default:
		relPath := file.RelativePathCwd(filePath)
		return relPath, tcell.StyleDefault
	}
}

func styleForStatusMsg(statusMsg state.StatusMsg) tcell.Style {
	s := tcell.StyleDefault
	switch statusMsg.Style {
	case state.StatusMsgStyleSuccess:
		return s.Foreground(tcell.ColorGreen).Bold(true)
	case state.StatusMsgStyleError:
		return s.Background(tcell.ColorRed).Foreground(tcell.ColorWhite).Bold(true)
	default:
		return s
	}
}
