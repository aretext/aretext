package display

import (
	"github.com/aretext/aretext/internal/pkg/exec"
	"github.com/gdamore/tcell"
)

// DrawEditor draws the editor in the screen.
func DrawEditor(screen tcell.Screen, editorState *exec.EditorState) {
	screen.Clear()
	DrawBuffer(screen, editorState.DocumentBuffer())
	DrawMenu(screen, editorState.Menu())
	DrawStatusBar(screen, editorState.StatusMsg(), editorState.InputMode(), editorState.FileWatcher().Path())
}
