package display

import (
	"io"

	"github.com/gdamore/tcell"
	"github.com/wedaly/aretext/internal/pkg/exec"
	"github.com/wedaly/aretext/internal/pkg/text"
	"github.com/wedaly/aretext/internal/pkg/text/segment"
)

// TextView displays text in a terminal, clipping and scrolling as necessary.
type TextView struct {
	execState    *exec.State
	screenRegion *ScreenRegion
	origin       uint64 // Position in the text displayed in the top-left corner of the view.
}

// NewTextView initializes a text view for a text tree and screen.
func NewTextView(execState *exec.State, screenRegion *ScreenRegion) *TextView {
	return &TextView{execState, screenRegion, 0}
}

// Resize notifies the text view that the terminal size has changed.
func (v *TextView) Resize(width, height int) {
	v.screenRegion.Resize(width, height)
}

// ScrollToCursor adjusts the view origin such that the cursor is visible.
func (v *TextView) ScrollToCursor() {
	width, height := v.screenRegion.Size()
	v.origin = ScrollToCursor(v.execState.CursorPosition(), v.execState.Tree(), v.origin, width, height)
}

// Draw draws text to the screen.
func (v *TextView) Draw() {
	v.screenRegion.HideCursor()
	width, height := v.screenRegion.Size()

	pos := v.origin
	reader := v.execState.Tree().ReaderAtPosition(pos, text.ReadDirectionForward)
	runeIter := text.NewCloneableForwardRuneIter(reader)
	wrapConfig := segment.NewLineWrapConfig(uint64(width), exec.GraphemeClusterWidth)
	wrappedLineIter := segment.NewWrappedLineIter(runeIter, wrapConfig)

	for row := 0; row < height; row++ {
		wrappedLine, err := wrappedLineIter.NextSegment()
		if err == io.EOF {
			break
		} else if err != nil {
			panic(err)
		}
		v.drawLineAndSetCursor(pos, row, width, wrappedLine)
		pos += wrappedLine.NumRunes()
	}

	// Text view is empty, with cursor positioned in the first cell.
	if pos-v.origin == 0 && pos == v.execState.CursorPosition() {
		v.screenRegion.ShowCursor(0, 0)
	}
}

func (v *TextView) drawLineAndSetCursor(pos uint64, row int, maxLineWidth int, wrappedLine *segment.Segment) {
	runeIter := text.NewRuneIterForSlice(wrappedLine.Runes())
	gcIter := segment.NewGraphemeClusterIter(runeIter)
	var lastGc *segment.Segment
	totalWidth := uint64(0)
	col := 0

	for {
		gc, err := gcIter.NextSegment()
		if err == io.EOF {
			break
		} else if err != nil {
			panic(err)
		}

		gcRunes := gc.Runes()
		gcWidth := exec.GraphemeClusterWidth(gcRunes, totalWidth)
		totalWidth += gcWidth

		if totalWidth > uint64(maxLineWidth) {
			// If there isn't enough space to show the line, fill it with a placeholder.
			v.drawLineTooLong(row, maxLineWidth)
			return
		}

		v.screenRegion.SetContent(col, row, gcRunes[0], gcRunes[1:], tcell.StyleDefault)

		if pos == v.execState.CursorPosition() {
			v.screenRegion.ShowCursor(col, row)
		}

		pos += gc.NumRunes()
		col += int(gcWidth) // Safe to downcast because there's a limit on the number of cells a grapheme cluster can occupy.
		lastGc = gc
	}

	if pos == v.execState.CursorPosition() {
		if lastGc != nil && (lastGc.HasNewline() || pos == uint64(maxLineWidth)) {
			// If the line ended on a newline or soft-wrapped line, show the cursor at the start of the next line.
			v.screenRegion.ShowCursor(0, row+1)
		} else {
			// Otherwise, show the cursor at the end of the current line.
			v.screenRegion.ShowCursor(col, row)
		}
	}
}

func (v *TextView) drawLineTooLong(row int, maxLineWidth int) {
	for col := 0; col < maxLineWidth; col++ {
		v.screenRegion.SetContent(col, row, '~', nil, tcell.StyleDefault.Dim(true))
	}
}
