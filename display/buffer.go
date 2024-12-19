package display

import (
	"io"
	"log"
	"strconv"

	"github.com/gdamore/tcell/v2"

	"github.com/aretext/aretext/config"
	"github.com/aretext/aretext/selection"
	"github.com/aretext/aretext/state"
	"github.com/aretext/aretext/syntax/parser"
	"github.com/aretext/aretext/text/segment"
)

// DrawBuffer draws text buffer in the screen.
func DrawBuffer(screen tcell.Screen, palette *Palette, buffer *state.BufferState, inputMode state.InputMode) {
	width, height := viewSize(buffer)
	sr := NewScreenRegion(screen, 0, 0, width, height)
	textTree := buffer.TextTree()
	cursorPos := buffer.CursorPosition()
	selectedRegion := buffer.SelectedRegion()
	viewTextOrigin := buffer.ViewTextOrigin()
	pos := viewTextOrigin
	showTabs := buffer.ShowTabs()
	showSpaces := buffer.ShowSpaces()
	lineNumMargin := buffer.LineNumMarginWidth() // Zero if line numbers disabled.
	lineNumberMode := buffer.LineNumberMode()
	cursorLine := textTree.LineNumForPosition(cursorPos)
	wrapConfig := buffer.LineWrapConfig()
	wrappedLineIter := segment.NewWrappedLineIter(wrapConfig, textTree, pos)
	wrappedLine := segment.Empty()
	searchMatch := buffer.SearchMatch()

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
		wrappedLineRunes := wrappedLine.Runes()
		syntaxTokens := buffer.SyntaxTokensIntersectingRange(pos, pos+uint64(len(wrappedLineRunes)))
		drawLineAndSetCursor(
			sr,
			palette,
			inputMode,
			pos,
			row,
			int(wrapConfig.MaxLineWidth),
			lineNum,
			lineNumMargin,
			lineStartPos,
			lineNumberMode,
			cursorLine,
			wrappedLineRunes,
			syntaxTokens,
			cursorPos,
			selectedRegion,
			searchMatch,
			wrapConfig.WidthFunc,
			showTabs,
			showSpaces,
		)
		pos += wrappedLine.NumRunes()
	}

	// Text view is empty, with cursor positioned in the first cell.
	if pos-viewTextOrigin == 0 && pos == cursorPos {
		showCursorInBuffer(sr, int(lineNumMargin), 0, palette, inputMode)
		drawLineNumIfNecessary(sr, palette, 0, 0, lineNumMargin, lineNumberMode, cursorLine)
	}
}

func viewSize(buffer *state.BufferState) (int, int) {
	width, height := buffer.ViewSize()
	return int(width), int(height)
}

func drawLineAndSetCursor(
	sr *ScreenRegion,
	palette *Palette,
	inputMode state.InputMode,
	pos uint64,
	row int,
	maxLineWidth int,
	lineNum uint64,
	lineNumMargin uint64,
	lineStartPos uint64,
	lineNumberMode config.LineNumberMode,
	cursorLine uint64,
	wrappedLineRunes []rune,
	syntaxTokens []parser.Token,
	cursorPos uint64,
	selectedRegion selection.Region,
	searchMatch *state.SearchMatch,
	gcWidthFunc segment.GraphemeClusterWidthFunc,
	showTabs bool,
	showSpaces bool,
) {
	startPos := pos
	gcRunes := []rune{'\x00', '\x00', '\x00', '\x00'}[:0] // Stack-allocate runes for the last grapheme cluster.
	totalWidth := uint64(0)
	col := 0
	var gcBreaker segment.GraphemeClusterBreaker
	var lastGcWasNewline bool

	if startPos == lineStartPos {
		drawLineNumIfNecessary(sr, palette, row, lineNum, lineNumMargin, lineNumberMode, cursorLine)
	}
	col += int(lineNumMargin)

	var i int
	for i < len(wrappedLineRunes) || len(gcRunes) > 0 {
		for _, r := range wrappedLineRunes[i:] {
			canBreakBefore := gcBreaker.ProcessRune(r)
			if canBreakBefore && len(gcRunes) > 0 {
				break
			}
			lastGcWasNewline = (r == '\n')
			gcRunes = append(gcRunes, r)
		}
		gcWidth := gcWidthFunc(gcRunes, totalWidth)
		totalWidth += gcWidth

		if totalWidth > uint64(maxLineWidth) {
			// If there isn't enough space to show the line, skip it.
			return
		}

		style := tcell.StyleDefault
		if selectedRegion.ContainsPosition(pos) {
			style = palette.StyleForSelection()
		} else if searchMatch.ContainsPosition(pos) {
			style = palette.StyleForSearchMatch()
		} else {
			for len(syntaxTokens) > 0 {
				token := syntaxTokens[0]
				if token.StartPos <= pos && token.EndPos > pos {
					style = palette.StyleForTokenRole(token.Role)
					break
				} else if token.StartPos > pos {
					break
				}
				syntaxTokens = syntaxTokens[1:]
			}
		}

		drawGraphemeCluster(sr, col, row, gcRunes, int(gcWidth), style, showTabs, showSpaces)

		if pos-startPos == uint64(maxLineWidth) {
			// This occurs when there are maxLineWidth characters followed by a line feed.
			break
		}

		if pos == cursorPos {
			showCursorInBuffer(sr, col, row, palette, inputMode)
		}

		i += len(gcRunes)
		pos += uint64(len(gcRunes))
		col += int(gcWidth) // Safe to downcast because there's a limit on the number of cells a grapheme cluster can occupy.
		gcRunes = gcRunes[:0]
	}

	if lastGcWasNewline {
		// Draw line number for an empty final line.
		drawLineNumIfNecessary(sr, palette, row+1, lineNum+1, lineNumMargin, lineNumberMode, cursorLine)
	}

	if pos == cursorPos {
		if lastGcWasNewline || (pos-startPos) == uint64(maxLineWidth) {
			// If the line ended on a newline or soft-wrapped line, show the cursor at the start of the next line.
			showCursorInBuffer(sr, int(lineNumMargin), row+1, palette, inputMode)
		} else if pos == cursorPos {
			// Otherwise, show the cursor at the end of the current line.
			showCursorInBuffer(sr, col, row, palette, inputMode)
		}
	}
}

func drawLineNumIfNecessary(sr *ScreenRegion, palette *Palette, row int, lineNum uint64, lineNumMargin uint64, lineNumberMode config.LineNumberMode, cursorLine uint64) {
	if lineNumMargin == 0 {
		return
	}

	style := palette.StyleForLineNum()
	lineNumStr := strconv.FormatUint(displayLineNum(lineNumberMode, lineNum, cursorLine), 10)

	// Right-aligned in the margin, with one space of padding on the right.
	col := int(lineNumMargin) - 1 - len(lineNumStr)
	for _, r := range lineNumStr {
		sr.SetContent(col, row, r, nil, style)
		col++
	}
}

func showCursorInBuffer(sr *ScreenRegion, col int, row int, palette *Palette, inputMode state.InputMode) {
	if inputMode == state.InputModeSearch {
		// In search mode, the terminal cursor will appear in the search query at the bottom of the screen.
		// Highlight the cursor position in the document with another style so the user knows where it is.
		sr.SetStyleInCell(col, row, palette.StyleForSearchCursor())
	} else {
		sr.ShowCursor(col, row)
	}
}

func displayLineNum(lineNumberMode config.LineNumberMode, lineNum uint64, cursorLine uint64) uint64 {
	switch lineNumberMode {
	case config.LineNumberModeAbsolute:
		return lineNum + 1
	case config.LineNumberModeRelative:
		if lineNum < cursorLine {
			return cursorLine - lineNum
		} else {
			return lineNum - cursorLine
		}
	default:
		panic("Unrecognized line number mode")
	}
}
