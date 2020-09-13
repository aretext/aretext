package display

import (
	"log"

	"github.com/gdamore/tcell"
	"github.com/wedaly/aretext/internal/pkg/exec"
)

// DrawEditor draws the editor in the screen.
func DrawEditor(screen tcell.Screen, editorState *exec.EditorState) {
	screen.Clear()
	switch layout := editorState.Layout(); layout {
	case exec.LayoutDocumentOnly:
		DrawBuffer(screen, editorState.Buffer(exec.BufferIdDocument))
	case exec.LayoutDocumentAndRepl:
		DrawBuffer(screen, editorState.Buffer(exec.BufferIdDocument))
		DrawBuffer(screen, editorState.Buffer(exec.BufferIdRepl))
		drawBorderAbove(screen, editorState.Buffer(exec.BufferIdRepl))
	default:
		log.Fatalf("Unrecognized layout %d", layout)
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
