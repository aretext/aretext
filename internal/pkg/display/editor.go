package display

import (
	"github.com/gdamore/tcell"
	"github.com/wedaly/aretext/internal/pkg/exec"
)

// DrawEditor draws the editor in the screen.
func DrawEditor(screen tcell.Screen, editorState *exec.EditorState) {
	screen.Clear()
	switch editorState.Layout() {
	case exec.LayoutDocumentOnly:
		DrawBuffer(screen, editorState.DocumentBuffer())
	case exec.LayoutDocumentAndRepl:
		DrawBuffer(screen, editorState.DocumentBuffer())
		DrawBuffer(screen, editorState.ReplBuffer())
		drawBorderAbove(screen, editorState.ReplBuffer())
	default:
		panic("Unrecognized layout")
	}
}

func drawBorderAbove(screen tcell.Screen, buffer *exec.BufferState) {
	width, _ := screen.Size()
	_, y := buffer.ViewOrigin()
	if y > 0 {
		sr := NewScreenRegion(screen, 0, int(y-1), width, 1)
		sr.Fill(tcell.RuneHLine, tcell.StyleDefault.Dim(true))
	}
}
