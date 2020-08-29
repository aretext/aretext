package display

import (
	"io"
	"log"

	"github.com/gdamore/tcell"
	"github.com/wedaly/aretext/internal/pkg/exec"
	"github.com/wedaly/aretext/internal/pkg/text"
	"github.com/wedaly/aretext/internal/pkg/text/segment"
)

// DrawBuffer draws text buffer in the screen.
func DrawBuffer(screen tcell.Screen, bufferState *exec.BufferState) {
	x, y, width, height := viewDimensions(bufferState)
	screenRegion := NewScreenRegion(screen, x, y, width, height)
	tree := bufferState.Tree()
	cursorPos := bufferState.CursorPosition()
	viewTextOrigin := bufferState.ViewTextOrigin()
	pos := viewTextOrigin
	reader := tree.ReaderAtPosition(pos, text.ReadDirectionForward)
	runeIter := text.NewCloneableForwardRuneIter(reader)
	wrapConfig := segment.NewLineWrapConfig(uint64(width), exec.GraphemeClusterWidth)
	wrappedLineIter := segment.NewWrappedLineIter(runeIter, wrapConfig)
	wrappedLine := segment.NewSegment()

	screenRegion.HideCursor()
	for row := 0; row < height; row++ {
		err := wrappedLineIter.NextSegment(wrappedLine)
		if err == io.EOF {
			break
		} else if err != nil {
			log.Fatalf("%s", err)
		}
		drawLineAndSetCursor(screenRegion, pos, row, width, wrappedLine, cursorPos)
		pos += wrappedLine.NumRunes()
	}

	// Text view is empty, with cursor positioned in the first cell.
	if pos-viewTextOrigin == 0 && pos == cursorPos {
		screenRegion.ShowCursor(0, 0)
	}
}

func viewDimensions(bufferState *exec.BufferState) (int, int, int, int) {
	x, y := bufferState.ViewOrigin()
	width, height := bufferState.ViewSize()
	return int(x), int(y), int(width), int(height)
}

func drawLineAndSetCursor(screenRegion *ScreenRegion, pos uint64, row int, maxLineWidth int, wrappedLine *segment.Segment, cursorPos uint64) {
	runeIter := text.NewRuneIterForSlice(wrappedLine.Runes())
	gcIter := segment.NewGraphemeClusterIter(runeIter)
	gc := segment.NewSegment()
	totalWidth := uint64(0)
	col := 0
	var lastGcWasNewline bool

	for {
		err := gcIter.NextSegment(gc)
		if err == io.EOF {
			break
		} else if err != nil {
			log.Fatalf("%s", err)
		}

		gcRunes := gc.Runes()
		gcWidth := exec.GraphemeClusterWidth(gcRunes, totalWidth)
		totalWidth += gcWidth

		if totalWidth > uint64(maxLineWidth) {
			// If there isn't enough space to show the line, fill it with a placeholder.
			drawLineTooLong(screenRegion, row, maxLineWidth)
			return
		}

		drawGraphemeCluster(screenRegion, col, row, gcRunes, tcell.StyleDefault)

		if pos == cursorPos {
			screenRegion.ShowCursor(col, row)
		}

		pos += gc.NumRunes()
		col += int(gcWidth) // Safe to downcast because there's a limit on the number of cells a grapheme cluster can occupy.
		lastGcWasNewline = gc.HasNewline()
	}

	if pos == cursorPos {
		if gc != nil && (lastGcWasNewline || pos == uint64(maxLineWidth)) {
			// If the line ended on a newline or soft-wrapped line, show the cursor at the start of the next line.
			screenRegion.ShowCursor(0, row+1)
		} else {
			// Otherwise, show the cursor at the end of the current line.
			screenRegion.ShowCursor(col, row)
		}
	}
}

func drawGraphemeCluster(screenRegion *ScreenRegion, col, row int, gc []rune, style tcell.Style) {
	// Emoji and regional indicator sequences are usually rendered using the
	// width of the first rune.  This won't support every terminal, but it's probably
	// the best we can do without knowing how the terminal will render the glyphs.
	if segment.GraphemeClusterIsEmoji(gc) || segment.GraphemeClusterIsRegionalIndicator(gc) {
		screenRegion.SetContent(col, row, gc[0], gc[1:], style)
		return
	}

	// For other sequences, we break the grapheme cluster into cells.
	// Each cell starts with a main rune, followed by zero or more combining runes.
	// In most cases, the entire grapheme cluster will fit in a single cell,
	// but there are exceptions (for example, some Thai sequences).
	i := 0
	for i < len(gc) {
		j := i + 1
		for j < len(gc) {
			r := gc[j]
			if exec.RuneWidth(r) > 0 {
				break
			}
			j++
		}
		screenRegion.SetContent(col, row, gc[i], gc[i+1:j], style)
		col += int(exec.RuneWidth(gc[i]))
		i = j
	}
}

func drawLineTooLong(screenRegion *ScreenRegion, row int, maxLineWidth int) {
	for col := 0; col < maxLineWidth; col++ {
		screenRegion.SetContent(col, row, '~', nil, tcell.StyleDefault.Dim(true))
	}
}
