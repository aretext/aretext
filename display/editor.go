package display

import (
	"github.com/gdamore/tcell/v2"

	"github.com/aretext/aretext/state"
)

// DrawEditor draws the editor in the screen.
func DrawEditor(screen tcell.Screen, editorState *state.EditorState) {
	screen.Clear()
	DrawBuffer(screen, editorState.DocumentBuffer())
	DrawMenu(screen, editorState.Menu())
	DrawStatusBar(screen, editorState.StatusMsg(), editorState.InputMode(), editorState.FileWatcher().Path())
	searchQuery, searchDirection := editorState.DocumentBuffer().SearchQueryAndDirection()
	DrawSearchQuery(screen, editorState.InputMode(), searchQuery, searchDirection)
}
