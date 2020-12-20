package display

import (
	"github.com/gdamore/tcell"
	"github.com/wedaly/aretext/internal/pkg/exec"
)

// DrawEditor draws the editor in the screen.
func DrawEditor(screen tcell.Screen, editorState *exec.EditorState) {
	screen.Clear()
	buffer := editorState.DocumentBuffer()
	DrawBuffer(screen, buffer)
}
