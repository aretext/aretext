package display

import (
	"github.com/aretext/aretext/exec"
	"github.com/aretext/aretext/file"
	"github.com/gdamore/tcell"
)

// DrawStatusBar draws a status bar on the last line of the screen.
func DrawStatusBar(screen tcell.Screen, statusMsg exec.StatusMsg, inputMode exec.InputMode, filePath string) {
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

func statusBarContent(statusMsg exec.StatusMsg, inputMode exec.InputMode, filePath string) (string, tcell.Style) {
	if len(statusMsg.Text) > 0 {
		return statusMsg.Text, styleForStatusMsg(statusMsg)
	}

	switch inputMode {
	case exec.InputModeInsert:
		return "-- INSERT --", tcell.StyleDefault.Bold(true)
	default:
		relPath := file.RelativePathCwd(filePath)
		return relPath, tcell.StyleDefault
	}
}

func styleForStatusMsg(statusMsg exec.StatusMsg) tcell.Style {
	s := tcell.StyleDefault
	switch statusMsg.Style {
	case exec.StatusMsgStyleSuccess:
		return s.Foreground(tcell.ColorGreen).Bold(true)
	case exec.StatusMsgStyleError:
		return s.Background(tcell.ColorRed).Foreground(tcell.ColorWhite).Bold(true)
	default:
		return s
	}
}
