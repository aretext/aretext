package display

import (
	"io"
	"log"

	"github.com/aretext/aretext/cellwidth"
	"github.com/aretext/aretext/selection"
	"github.com/aretext/aretext/state"
	"github.com/aretext/aretext/syntax/parser"
	"github.com/aretext/aretext/text"
	"github.com/aretext/aretext/text/segment"
	"github.com/gdamore/tcell/v2"
)

// DrawBuffer draws text buffer in the screen.
func DrawBuffer(screen tcell.Screen, buffer *state.BufferState) {
	x, y, width, height := viewDimensions(buffer)
	sr := NewScreenRegion(screen, x, y, width, height)
	textTree := buffer.TextTree()
	cursorPos := buffer.CursorPosition()
	selectedRegion := buffer.SelectedRegion()
	viewTextOrigin := buffer.ViewTextOrigin()
	pos := viewTextOrigin
	reader := textTree.ReaderAtPosition(pos, text.ReadDirectionForward)
	runeIter := text.NewCloneableForwardRuneIter(reader)
	gcWidthFunc := func(gc []rune, offsetInLine uint64) uint64 {
		return cellwidth.GraphemeClusterWidth(gc, offsetInLine, buffer.TabSize())
	}
	wrapConfig := segment.NewLineWrapConfig(uint64(width), gcWidthFunc)
	wrappedLineIter := segment.NewWrappedLineIter(runeIter, wrapConfig)
	wrappedLine := segment.Empty()
	searchMatch := buffer.SearchMatch()
	tokenIter := buffer.TokenTree().IterFromPosition(pos, parser.IterDirectionForward)

	sr.HideCursor()

	for row := 0; row < height; row++ {
		err := wrappedLineIter.NextSegment(wrappedLine)
		if err == io.EOF {
			break
		} else if err != nil {
			log.Fatalf("%s", err)
		}
		drawLineAndSetCursor(sr, pos, row, width, wrappedLine, tokenIter, cursorPos, selectedRegion, searchMatch, gcWidthFunc)
		pos += wrappedLine.NumRunes()
	}

	// Text view is empty, with cursor positioned in the first cell.
	if pos-viewTextOrigin == 0 && pos == cursorPos {
		sr.ShowCursor(0, 0)
	}
}

func viewDimensions(buffer *state.BufferState) (int, int, int, int) {
	x, y := buffer.ViewOrigin()
	width, height := buffer.ViewSize()
	return int(x), int(y), int(width), int(height)
}

func drawLineAndSetCursor(sr *ScreenRegion, pos uint64, row int, maxLineWidth int, wrappedLine *segment.Segment, tokenIter *parser.TokenIter, cursorPos uint64, selectedRegion selection.Region, searchMatch *state.SearchMatch, gcWidthFunc segment.GraphemeClusterWidthFunc) {
	startPos := pos
	runeIter := text.NewRuneIterForSlice(wrappedLine.Runes())
	gcIter := segment.NewGraphemeClusterIter(runeIter)
	gc := segment.Empty()
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
		gcWidth := gcWidthFunc(gcRunes, totalWidth)
		totalWidth += gcWidth

		if totalWidth > uint64(maxLineWidth) {
			// If there isn't enough space to show the line, fill it with a placeholder.
			drawLineTooLong(sr, row, maxLineWidth)
			return
		}

		style := styleAtPosition(pos, selectedRegion, searchMatch, tokenIter)
		drawGraphemeCluster(sr, col, row, gcRunes, int(gcWidth), style)

		if pos-startPos == uint64(maxLineWidth) {
			// This occurs when there are maxLineWidth characters followed by a line feed.
			break
		}

		if pos == cursorPos {
			sr.ShowCursor(col, row)
		}

		pos += gc.NumRunes()
		col += int(gcWidth) // Safe to downcast because there's a limit on the number of cells a grapheme cluster can occupy.
		lastGcWasNewline = gc.HasNewline()
	}

	if pos == cursorPos {
		if gc != nil && (lastGcWasNewline || (pos-startPos) == uint64(maxLineWidth)) {
			// If the line ended on a newline or soft-wrapped line, show the cursor at the start of the next line.
			sr.ShowCursor(0, row+1)
		} else {
			// Otherwise, show the cursor at the end of the current line.
			sr.ShowCursor(col, row)
		}
	}
}

func drawLineTooLong(sr *ScreenRegion, row int, maxLineWidth int) {
	for col := 0; col < maxLineWidth; col++ {
		sr.SetContent(col, row, '~', nil, tcell.StyleDefault.Dim(true))
	}
}

func styleAtPosition(pos uint64, selectedRegion selection.Region, searchMatch *state.SearchMatch, tokenIter *parser.TokenIter) tcell.Style {
	if selectedRegion.ContainsPosition(pos) {
		return tcell.StyleDefault.Reverse(true).Dim(true)
	}

	if searchMatch.ContainsPosition(pos) {
		return tcell.StyleDefault.Reverse(true)
	}

	var token parser.Token
	for tokenIter.Get(&token) {
		if token.StartPos <= pos && token.EndPos > pos {
			return styleForTokenRole(token.Role)
		}

		if token.StartPos > pos {
			break
		}

		tokenIter.Advance()
	}

	return tcell.StyleDefault
}

func styleForTokenRole(tokenRole parser.TokenRole) tcell.Style {
	s := tcell.StyleDefault
	switch tokenRole {
	case parser.TokenRoleOperator:
		return s.Foreground(tcell.ColorFuchsia)
	case parser.TokenRoleKeyword:
		return s.Foreground(tcell.ColorOrange)
	case parser.TokenRoleNumber:
		return s.Foreground(tcell.ColorGreen)
	case parser.TokenRoleString, parser.TokenRoleStringQuote:
		return s.Foreground(tcell.ColorRed)
	case parser.TokenRoleKey:
		return s.Foreground(tcell.ColorNavy)
	case parser.TokenRoleComment, parser.TokenRoleCommentDelimiter:
		return s.Foreground(tcell.ColorBlue)
	default:
		return s
	}
}
