package display

import (
	"github.com/gdamore/tcell/v2"

	"github.com/aretext/aretext/file"
	"github.com/aretext/aretext/state"
)

// DrawStatusBar draws a status bar on the last line of the screen.
func DrawStatusBar(
	screen tcell.Screen,
	palette *Palette,
	statusMsg state.StatusMsg,
	inputMode state.InputMode,
	inputBufferString string,
	isRecordingUserMacro bool,
	filePath string,
) {
	screenWidth, screenHeight := screen.Size()
	if screenHeight == 0 {
		return
	}

	row := screenHeight - 1
	sr := NewScreenRegion(screen, 0, row, screenWidth, 1)
	sr.Fill(' ', tcell.StyleDefault)
	text, style := statusBarContent(
		palette,
		statusMsg,
		inputMode,
		inputBufferString,
		isRecordingUserMacro,
		filePath)
	drawStringNoWrap(sr, text, 0, 0, style)
}

func statusBarContent(
	palette *Palette,
	statusMsg state.StatusMsg,
	inputMode state.InputMode,
	inputBufferString string,
	isRecordingUserMacro bool,
	filePath string,
) (string, tcell.Style) {
	if len(inputBufferString) > 0 {
		return inputBufferString, palette.StyleForStatusInputBuffer()
	}

	if len(statusMsg.Text) > 0 {
		return statusMsg.Text, palette.StyleForStatusMsg(statusMsg.Style)
	}

	if isRecordingUserMacro {
		return "Recording macro...", palette.StyleForStatusRecordingMacro()
	}

	switch inputMode {
	case state.InputModeInsert:
		return "-- INSERT --", palette.StyleForStatusInputMode()
	case state.InputModeVisual:
		return "-- VISUAL --", palette.StyleForStatusInputMode()
	case state.InputModeTask:
		return "Running... press ESC to abort", palette.StyleForStatusInputMode()
	default:
		relPath := file.RelativePathCwd(filePath)
		return relPath, palette.StyleForStatusFilePath()
	}
}
