package display

import (
	"github.com/gdamore/tcell"
	"github.com/wedaly/aretext/internal/pkg/exec"
	"github.com/wedaly/aretext/internal/pkg/input"
)

// DrawEditor draws the editor in the screen.
func DrawEditor(screen tcell.Screen, editorState *exec.EditorState, inputMode input.ModeType) {
	screen.Clear()
	DrawBuffer(screen, editorState.DocumentBuffer())
	DrawMenu(screen, editorState.Menu())
	DrawStatusBar(screen, editorState.StatusMsg(), inputMode, editorState.FileWatcher().Path())
}
