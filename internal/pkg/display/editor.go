package display

import (
	"log"

	"github.com/gdamore/tcell"
	"github.com/wedaly/aretext/internal/pkg/exec"
)

// DrawEditor draws the editor in the screen.
func DrawEditor(screen tcell.Screen, editorState *exec.EditorState) {
	screen.Clear()

	drawBuffer := func(bufferId exec.BufferId) {
		buffer := editorState.Buffer(bufferId)
		hasFocus := editorState.FocusedBufferId() == bufferId
		DrawBuffer(screen, buffer, hasFocus)
	}

	drawBorderAbove := func() {
		replBuffer := editorState.Buffer(exec.BufferIdRepl)
		width, _ := screen.Size()
		_, y := replBuffer.ViewOrigin()
		if y > 0 {
			sr := NewScreenRegion(screen, 0, int(y-1), width, 1)
			sr.Fill(tcell.RuneHLine, tcell.StyleDefault.Dim(true))
		}
	}

	switch layout := editorState.Layout(); layout {
	case exec.LayoutDocumentOnly:
		drawBuffer(exec.BufferIdDocument)
	case exec.LayoutDocumentAndRepl:
		drawBuffer(exec.BufferIdDocument)
		drawBuffer(exec.BufferIdRepl)
		drawBorderAbove()
	default:
		log.Fatalf("Unrecognized layout %d", layout)
	}
}
