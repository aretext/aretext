package display

import (
	"io"
	"log"
	"strconv"
	"unicode"

	"github.com/gdamore/tcell/v2"

	"github.com/aretext/aretext/cellwidth"
	"github.com/aretext/aretext/config"
	"github.com/aretext/aretext/selection"
	"github.com/aretext/aretext/state"
	"github.com/aretext/aretext/syntax/parser"
	"github.com/aretext/aretext/text"
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
	showUnicode := buffer.ShowUnicode()
	lineNumMargin := buffer.LineNumMarginWidth() // Zero if line numbers disabled.
	lineNumberMode := buffer.LineNumberMode()
	cursorLine := textTree.LineNumForPosition(cursorPos)
	wrapConfig := buffer.LineWrapConfig()
	wrappedLineIter := segment.NewWrappedLineIter(wrapConfig, textTree, pos)
	wrappedLine := segment.Empty()
	searchMatch := buffer.SearchMatch()
	textEscaper := &text.Escaper{}

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
			wrapConfig.CellWidthSizer,
			showTabs,
			showSpaces,
			showUnicode,
			textEscaper,
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
	cellWidthSizer cellwidth.Sizer,
	showTabs bool,
	showSpaces bool,
	showUnicode bool,
	textEscaper *text.Escaper,
) {
	startPos := pos
	gcRunes := []rune{'\x00', '\x00', '\x00', '\x00'}[:0] // Stack-allocate runes for the last grapheme cluster.
	totalWidth := uint64(0)
	col := 0
	var lastGcWasNewline bool
	var tokenForPos *parser.Token

	if startPos == lineStartPos {
		drawLineNumIfNecessary(sr, palette, row, lineNum, lineNumMargin, lineNumberMode, cursorLine)
	}
	col += int(lineNumMargin)

	var i int
	for i < len(wrappedLineRunes) || len(gcRunes) > 0 {
		var gcBreaker segment.GraphemeClusterBreaker
		for _, r := range wrappedLineRunes[i:] {
			canBreakBefore := gcBreaker.ProcessRune(r)
			if canBreakBefore && len(gcRunes) > 0 {
				break
			}
			lastGcWasNewline = (r == '\n')
			gcRunes = append(gcRunes, r)
		}
		gcWidth := cellWidthSizer.GraphemeClusterWidth(gcRunes, totalWidth)
		totalWidth += gcWidth

		if totalWidth > uint64(maxLineWidth) {
			// If there isn't enough space to show the line, skip it.
			return
		}

		// If "show unicode" enabled and grapheme cluster is non-ASCII unicode,
		// then escape unicode in this grapheme cluster.
		// This MUST match the same criteria in cellwidth.
		escapeUnicode := showUnicode && !(len(gcRunes) == 1 && gcRunes[0] <= 127)

		tokenForPos, syntaxTokens = consumeSyntaxTokensUntilPos(syntaxTokens, pos)
		style := styleForGraphemeCluster(pos, palette, selectedRegion, searchMatch, tokenForPos, escapeUnicode)
		drawGraphemeCluster(sr, col, row, gcRunes, int(gcWidth), style, showTabs, showSpaces, escapeUnicode, textEscaper)

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
	sr.PutStrStyled(col, row, lineNumStr, style)
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

func consumeSyntaxTokensUntilPos(syntaxTokens []parser.Token, pos uint64) (*parser.Token, []parser.Token) {
	for i, token := range syntaxTokens {
		if token.StartPos <= pos && token.EndPos > pos {
			return &token, syntaxTokens[i:]
		} else if token.StartPos > pos {
			return nil, syntaxTokens[i:]
		}
	}
	return nil, nil
}

func styleForGraphemeCluster(
	pos uint64,
	palette *Palette,
	selectedRegion selection.Region,
	searchMatch *state.SearchMatch,
	syntaxToken *parser.Token,
	escapeUnicode bool,
) tcell.Style {
	if selectedRegion.ContainsPosition(pos) {
		return palette.StyleForSelection()
	}

	if searchMatch.ContainsPosition(pos) {
		return palette.StyleForSearchMatch()
	}

	if escapeUnicode {
		return palette.StyleForEscapedUnicode()
	}

	if syntaxToken != nil {
		return palette.StyleForTokenRole(syntaxToken.Role)
	}

	return tcell.StyleDefault
}

func drawGraphemeCluster(
	sr *ScreenRegion,
	col, row int,
	gc []rune,
	gcWidth int,
	style tcell.Style,
	showTabs bool,
	showSpaces bool,
	escapeUnicode bool,
	textEscaper *text.Escaper,
) {
	startCol := col

	// Style whitespace (newlines, tabs, etc.) but don't set any runes.
	// This prevents drawing artifacts with '\r\n' where tcell
	// sends the combining character ('\n') to the terminal.
	if isGraphemeClusterWhitespace(gc) {
		// Always style at least one cell, even when the gcWidth is zero (carriage returns).
		// Tabs will usually style multiple cells.
		for col == startCol || col < startCol+gcWidth {
			sr.Put(col, row, " ", style)
			col++
		}

		// Draw a special character to represent a tab.
		if gc[0] == '\t' && showTabs {
			sr.Put(startCol, row, string(tcell.RuneRArrow), style.Dim(true))
		}

		// Draw a special character to represent a space.
		if gc[0] == ' ' && showSpaces {
			sr.Put(startCol, row, string(tcell.RuneBullet), style.Dim(true))
		}
	} else if escapeUnicode {
		sr.PutStrStyled(col, row, textEscaper.RunesToStr(gc), style)
	} else {
		// For everything else, tcell expects the grapheme cluster contents to be
		// written into the first of the cell(s) occupied by the gc.
		sr.Put(col, row, string(gc), style)
	}
}

// isGraphemeClusterWhitespace returns true if EVERY rune in the grapheme cluster
// is a unicode space. It's important to check every rune because the unicode
// grapheme cluster segmentation algorithm allows "extend" runes to follow
// space characters (for example, to display a combining mark by itself).
//
// Tabs or newlines will always be their own grapheme cluster without modifiers
// because unicode grapheme cluster segmentation rules GB4 and GB5 say
// you MUST break before and after a control character (tab) or CR/LF (newline).
func isGraphemeClusterWhitespace(gc []rune) bool {
	for _, r := range gc {
		if !unicode.IsSpace(r) {
			return false
		}
	}
	return true
}
