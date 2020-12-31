package display

import (
	"github.com/gdamore/tcell"
	"github.com/wedaly/aretext/internal/pkg/exec"
)

// DrawEditor draws the editor in the screen.
func DrawEditor(screen tcell.Screen, editorState *exec.EditorState) {
	screen.Clear()
	DrawBuffer(screen, editorState.DocumentBuffer())
	DrawMenu(screen, editorState.Menu())
	DrawStatusBar(screen, editorState.StatusMsg(), editorState.InputMode(), editorState.FileWatcher().Path())
}
