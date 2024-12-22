package display

import (
	"github.com/gdamore/tcell/v2"

	"github.com/aretext/aretext/editor/state"
)

// DrawEditor draws the editor in the screen.
func DrawEditor(screen tcell.Screen, palette *Palette, editorState *state.EditorState, inputBufferString string) {
	screen.Fill(' ', tcell.StyleDefault)

	DrawBuffer(screen, palette, editorState.DocumentBuffer(), editorState.InputMode())

	DrawStatusBar(
		screen,
		palette,
		editorState.StatusMsg(),
		editorState.InputMode(),
		inputBufferString,
		editorState.IsRecordingUserMacro(),
		editorState.FileWatcher().Path(),
	)

	switch editorState.InputMode() {
	case state.InputModeMenu:
		DrawMenu(screen, palette, editorState.Menu())
	case state.InputModeSearch:
		searchQuery, searchDirection := editorState.DocumentBuffer().SearchQueryAndDirection()
		DrawSearchQuery(screen, palette, searchQuery, searchDirection)
	case state.InputModeTextField:
		DrawTextField(screen, palette, editorState.TextField())
	}
}
