package display

import (
	"io"
	"log"
	"strconv"

	"github.com/gdamore/tcell/v2"

	"github.com/aretext/aretext/cellwidth"
	"github.com/aretext/aretext/selection"
	"github.com/aretext/aretext/state"
	"github.com/aretext/aretext/syntax/parser"
	"github.com/aretext/aretext/text"
	"github.com/aretext/aretext/text/segment"
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
	lineNumMargin := buffer.LineNumMarginWidth() // Zero if line numbers disabled.
	wrapWidth := uint64(width) - lineNumMargin
	wrapConfig := segment.NewLineWrapConfig(wrapWidth, gcWidthFunc)
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
		lineNum := textTree.LineNumForPosition(pos)
		lineStartPos := textTree.LineStartPosition(lineNum)
		drawLineAndSetCursor(
			sr,
			pos,
			row,
			int(wrapWidth),
			lineNum,
			lineNumMargin,
			lineStartPos,
			wrappedLine,
			tokenIter,
			cursorPos,
			selectedRegion,
			searchMatch,
			gcWidthFunc,
		)
		pos += wrappedLine.NumRunes()
	}

	// Text view is empty, with cursor positioned in the first cell.
	if pos-viewTextOrigin == 0 && pos == cursorPos {
		sr.ShowCursor(int(lineNumMargin), 0)
		drawLineNumIfNecessary(sr, 0, 0, lineNumMargin)
	}
}

func viewDimensions(buffer *state.BufferState) (int, int, int, int) {
	x, y := buffer.ViewOrigin()
	width, height := buffer.ViewSize()
	return int(x), int(y), int(width), int(height)
}

func drawLineAndSetCursor(
	sr *ScreenRegion,
	pos uint64,
	row int,
	maxLineWidth int,
	lineNum uint64,
	lineNumMargin uint64,
	lineStartPos uint64,
	wrappedLine *segment.Segment,
	tokenIter *parser.TokenIter,
	cursorPos uint64,
	selectedRegion selection.Region,
	searchMatch *state.SearchMatch,
	gcWidthFunc segment.GraphemeClusterWidthFunc,
) {
	startPos := pos
	runeIter := text.NewRuneIterForSlice(wrappedLine.Runes())
	gcIter := segment.NewGraphemeClusterIter(runeIter)
	gc := segment.Empty()
	totalWidth := uint64(0)
	col := 0
	var lastGcWasNewline bool

	if startPos == lineStartPos {
		drawLineNumIfNecessary(sr, row, lineNum, lineNumMargin)
	}
	col += int(lineNumMargin)

	for {
		err := gcIter.NextSegment(gc)
		if err == io.EOF {
			break
		} else if err != nil {
			log.Fatalf("%s", err)
		}

		gcRunes := gc.Runes()
		lastGcWasNewline = gc.HasNewline()
		gcWidth := gcWidthFunc(gcRunes, totalWidth)
		totalWidth += gcWidth

		if totalWidth > uint64(maxLineWidth) {
			// If there isn't enough space to show the line, skip it.
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
	}

	if gc != nil && lastGcWasNewline {
		// Draw line number for an empty final line.
		drawLineNumIfNecessary(sr, row+1, lineNum+1, lineNumMargin)
	}

	if pos == cursorPos {
		if gc != nil && (lastGcWasNewline || (pos-startPos) == uint64(maxLineWidth)) {
			// If the line ended on a newline or soft-wrapped line, show the cursor at the start of the next line.
			sr.ShowCursor(int(lineNumMargin), row+1)
		} else if pos == cursorPos {
			// Otherwise, show the cursor at the end of the current line.
			sr.ShowCursor(col, row)
		}
	}
}

func drawLineNumIfNecessary(sr *ScreenRegion, row int, lineNum uint64, lineNumMargin uint64) {
	if lineNumMargin == 0 {
		return
	}

	style := tcell.StyleDefault.Foreground(tcell.ColorOrange)
	lineNumStr := strconv.FormatUint(lineNum+1, 10)

	// Right-aligned in the margin, with one space of padding on the right.
	col := int(lineNumMargin) - 1 - len(lineNumStr)
	for _, r := range lineNumStr {
		sr.SetContent(col, row, r, nil, style)
		col++
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
		return s.Foreground(tcell.ColorTeal)
	case parser.TokenRoleComment, parser.TokenRoleCommentDelimiter:
		return s.Foreground(tcell.ColorBlue)
	default:
		return s
	}
}
