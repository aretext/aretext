package display

import (
	"github.com/gdamore/tcell/v2"

	"github.com/aretext/aretext/state"
)

// DrawEditor draws the editor in the screen.
func DrawEditor(screen tcell.Screen, palette *Palette, editorState *state.EditorState, inputBufferString string) {
	screen.Clear()
	DrawBuffer(screen, palette, editorState.DocumentBuffer())
	DrawMenu(screen, palette, editorState.Menu())
	DrawStatusBar(
		screen,
		palette,
		editorState.StatusMsg(),
		editorState.InputMode(),
		inputBufferString,
		editorState.IsRecordingUserMacro(),
		editorState.FileWatcher().Path(),
	)
	searchQuery, searchDirection := editorState.DocumentBuffer().SearchQueryAndDirection()
	DrawSearchQuery(
		screen,
		palette,
		editorState.InputMode(),
		searchQuery,
		searchDirection,
	)
}
