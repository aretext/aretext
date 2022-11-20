package state

import (
	"io"
	"log"
	"strings"
	"unicode/utf8"

	"github.com/pkg/errors"

	"github.com/aretext/aretext/cellwidth"
	"github.com/aretext/aretext/clipboard"
	"github.com/aretext/aretext/locate"
	"github.com/aretext/aretext/selection"
	"github.com/aretext/aretext/syntax/parser"
	"github.com/aretext/aretext/text"
	"github.com/aretext/aretext/text/segment"
	"github.com/aretext/aretext/undo"
)

// InsertRune inserts a rune at the current cursor location.
func InsertRune(state *EditorState, r rune) {
	buffer := state.documentBuffer
	startPos := buffer.cursor.position
	if err := insertTextAtPosition(state, string(r), startPos, true); err != nil {
		log.Printf("Error inserting rune: %v\n", err)
		return
	}
	buffer.cursor.position = startPos + 1
}

// insertTextAtPosition inserts text into the document.
// It also updates the syntax tokens and unsaved changes flag.
// It does NOT move the cursor.
func insertTextAtPosition(state *EditorState, s string, pos uint64, updateUndoLog bool) error {
	buffer := state.documentBuffer

	var n uint64
	for _, r := range s {
		if err := buffer.textTree.InsertAtPosition(pos+n, r); err != nil {
			return errors.Wrap(err, "text.Tree.InsertAtPosition")
		}
		n++
	}

	edit := parser.NewInsertEdit(pos, n)
	retokenizeAfterEdit(buffer, edit)

	if updateUndoLog && len(s) > 0 {
		op := undo.InsertOp(pos, s)
		buffer.undoLog.TrackOp(op)
	}

	return nil
}

func mustInsertTextAtPosition(state *EditorState, text string, pos uint64, updateUndoLog bool) {
	err := insertTextAtPosition(state, text, pos, updateUndoLog)
	if err != nil {
		panic(err)
	}
}

func mustInsertRuneAtPosition(state *EditorState, r rune, pos uint64, updateUndoLog bool) {
	mustInsertTextAtPosition(state, string(r), pos, updateUndoLog)
}

// InsertNewline inserts a newline at the current cursor position.
func InsertNewline(state *EditorState) {
	cursorPos := state.documentBuffer.cursor.position
	mustInsertRuneAtPosition(state, '\n', cursorPos, true)
	cursorPos++

	buffer := state.documentBuffer
	if buffer.autoIndent {
		deleteToNextNonWhitespace(state, cursorPos)
		numCols := numColsIndentedPrevLine(buffer, cursorPos)
		cursorPos = indentFromPos(state, cursorPos, numCols)
	}

	buffer.cursor = cursorState{position: cursorPos}
}

func deleteToNextNonWhitespace(state *EditorState, startPos uint64) {
	pos := locate.NextNonWhitespaceOrNewline(state.documentBuffer.textTree, startPos)
	count := pos - startPos
	deleteRunes(state, startPos, count, true)
}

func numColsIndentedPrevLine(buffer *BufferState, cursorPos uint64) uint64 {
	tabSize := buffer.tabSize
	lineNum := buffer.textTree.LineNumForPosition(cursorPos)
	if lineNum == 0 {
		return 0
	}

	prevLineStartPos := buffer.textTree.LineStartPosition(lineNum - 1)
	reader := buffer.textTree.ReaderAtPosition(prevLineStartPos)
	iter := segment.NewGraphemeClusterIter(reader)
	seg := segment.Empty()
	numCols := uint64(0)
	for {
		err := iter.NextSegment(seg)
		if err == io.EOF {
			break
		} else if err != nil {
			panic(err)
		}

		gc := seg.Runes()
		if gc[0] != '\t' && gc[0] != ' ' {
			break
		}

		numCols += cellwidth.GraphemeClusterWidth(gc, numCols, tabSize)
	}

	return numCols
}

func indentFromPos(state *EditorState, pos uint64, numCols uint64) uint64 {
	tabSize := state.documentBuffer.tabSize
	tabExpand := state.documentBuffer.tabExpand

	i := uint64(0)
	for i < numCols {
		if !tabExpand && numCols-i >= tabSize {
			mustInsertRuneAtPosition(state, '\t', pos, true)
			i += tabSize
		} else {
			mustInsertRuneAtPosition(state, ' ', pos, true)
			i++
		}
		pos++
	}
	return pos
}

// ClearAutoIndentWhitespaceLine clears a line consisting of only whitespace characters when autoindent is enabled.
// This is used to remove whitespace introduced by autoindent (for example, when inserting consecutive newlines).
func ClearAutoIndentWhitespaceLine(state *EditorState, startOfLineLoc Locator) {
	if !state.documentBuffer.autoIndent {
		return
	}

	params := locatorParamsForBuffer(state.documentBuffer)
	startOfLinePos := startOfLineLoc(params)
	endOfLinePos := locate.NextLineBoundary(params.TextTree, true, startOfLinePos)
	firstNonWhitespacePos := locate.NextNonWhitespaceOrNewline(params.TextTree, startOfLinePos)

	if endOfLinePos > startOfLinePos && firstNonWhitespacePos == endOfLinePos {
		numDeleted := endOfLinePos - startOfLinePos
		deleteRunes(state, startOfLinePos, numDeleted, true)
		MoveCursor(state, func(params LocatorParams) uint64 {
			if params.CursorPos > endOfLinePos {
				return params.CursorPos - numDeleted
			} else if params.CursorPos > startOfLinePos {
				return startOfLinePos
			} else {
				return params.CursorPos
			}
		})
	}
}

// InsertTab inserts a tab at the current cursor position.
func InsertTab(state *EditorState) {
	cursorPos := state.documentBuffer.cursor.position
	newCursorPos := insertTabsAtPos(state, cursorPos, tabText(state, 1))
	state.documentBuffer.cursor = cursorState{position: newCursorPos}
}

func tabText(state *EditorState, count uint64) string {
	var buf []byte
	if state.documentBuffer.tabExpand {
		buf = make([]byte, count*state.documentBuffer.tabSize)
		for i := 0; i < len(buf); i++ {
			buf[i] = ' '
		}
	} else {
		buf = make([]byte, count)
		for i := 0; i < len(buf); i++ {
			buf[i] = '\t'
		}
	}
	return string(buf)
}

func insertTabsAtPos(state *EditorState, pos uint64, tabs string) uint64 {
	n := uint64(len(tabs))

	if state.documentBuffer.tabExpand {
		// Inserted tab should end at a tab stop (aligned to multiples of tabSize from start of line).
		offset := offsetInLine(state.documentBuffer, pos) % state.documentBuffer.tabSize
		if offset > 0 {
			n -= offset
		}
	}

	mustInsertTextAtPosition(state, tabs[:n], pos, true)
	return pos + n
}

func offsetInLine(buffer *BufferState, startPos uint64) uint64 {
	var offset uint64
	textTree := buffer.textTree
	pos := locate.StartOfLineAtPos(textTree, startPos)
	reader := textTree.ReaderAtPosition(pos)
	iter := segment.NewGraphemeClusterIter(reader)
	seg := segment.Empty()
	for pos < startPos {
		err := iter.NextSegment(seg)
		if err == io.EOF {
			break
		} else if err != nil {
			panic(err)
		}
		offset += cellwidth.GraphemeClusterWidth(seg.Runes(), offset, buffer.tabSize)
		pos += seg.NumRunes()
	}
	return offset
}

// DeleteToPos deletes characters from the cursor position up to (but not including) the position returned by the locator.
// It can delete either forwards or backwards from the cursor.
// The cursor position will be set to the start of the deleted region,
// which could be on a newline character or past the end of the text.
func DeleteToPos(state *EditorState, loc Locator, clipboardPage clipboard.PageId) {
	buffer := state.documentBuffer
	startPos := buffer.cursor.position
	deleteToPos := loc(locatorParamsForBuffer(buffer))

	var deletedText string
	if startPos < deleteToPos {
		deletedText = deleteRunes(state, startPos, deleteToPos-startPos, true)
		buffer.cursor = cursorState{position: startPos}
	} else if startPos > deleteToPos {
		deletedText = deleteRunes(state, deleteToPos, startPos-deleteToPos, true)
		buffer.cursor = cursorState{position: deleteToPos}
	}

	if deletedText != "" {
		state.clipboard.Set(clipboardPage, clipboard.PageContent{
			Text:     deletedText,
			Linewise: false,
		})
	}
}

// DeleteRange deletes all characters in a range (for example, a word or selection).
// This moves the cursor to the start position of the range.
func DeleteRange(state *EditorState, loc RangeLocator, clipboardPage clipboard.PageId) (uint64, uint64) {
	buffer := state.documentBuffer
	startPos, endPos := loc(locatorParamsForBuffer(buffer))
	startLoc := func(LocatorParams) uint64 { return startPos }
	endLoc := func(LocatorParams) uint64 { return endPos }
	MoveCursor(state, startLoc)
	DeleteToPos(state, endLoc, clipboardPage)
	return startPos, endPos
}

// DeleteLines deletes lines from the cursor's current line to the line of a target cursor.
// It moves the cursor to the start of the line following the last deleted line.
func DeleteLines(state *EditorState, targetLineLoc Locator, abortIfTargetIsCurrentLine bool, replaceWithEmptyLine bool, clipboardPage clipboard.PageId) {
	buffer := state.documentBuffer
	currentLine := buffer.textTree.LineNumForPosition(buffer.cursor.position)
	targetPos := targetLineLoc(locatorParamsForBuffer(buffer))
	targetLine := buffer.textTree.LineNumForPosition(targetPos)

	if targetLine == currentLine && abortIfTargetIsCurrentLine {
		return
	}

	if targetLine < currentLine {
		currentLine, targetLine = targetLine, currentLine
	}

	startPos := buffer.textTree.LineStartPosition(currentLine)
	if startPos > 0 && targetLine+1 >= buffer.textTree.NumLines() {
		// The last line does not have a newline at the end, so delete the newline from the end of the previous line instead.
		startPos--
	}

	endOfLastLineToDelPos := locate.NextLineBoundary(buffer.textTree, true, buffer.textTree.LineStartPosition(targetLine))
	if endOfLastLineToDelPos < buffer.textTree.NumChars() {
		endOfLastLineToDelPos++ // Add one to include the newline at the end of the line, if it exists.
	}
	deletedLastLine := targetLine+1 >= buffer.textTree.NumLines()

	if replaceWithEmptyLine && buffer.textTree.NumChars() > 0 {
		mustInsertRuneAtPosition(state, '\n', endOfLastLineToDelPos, true)
	}

	numToDelete := endOfLastLineToDelPos - startPos
	deletedText := deleteRunes(state, startPos, numToDelete, true)

	buffer.cursor = cursorState{position: startPos}

	if deletedLastLine && replaceWithEmptyLine {
		// Special case: if we deleted the last line and inserted a newline, we want to place the cursor *after* the newline.
		buffer.cursor.position++
	}

	if buffer.cursor.position >= buffer.textTree.NumChars() {
		buffer.cursor = cursorState{
			position: locate.StartOfLastLine(buffer.textTree),
		}
	}

	if len(deletedText) > 0 {
		state.clipboard.Set(clipboardPage, clipboard.PageContent{
			Text:     stripStartingAndTrailingNewlines(deletedText),
			Linewise: true,
		})
	}
}

func stripStartingAndTrailingNewlines(s string) string {
	if len(s) > 0 && s[0] == '\n' {
		s = s[1:]
	}

	if len(s) > 0 && s[len(s)-1] == '\n' {
		s = s[0 : len(s)-1]
	}

	return s
}

// deleteRunes deletes text from the document.
// It also updates the syntax token and undo log.
// It does NOT move the cursor.
func deleteRunes(state *EditorState, pos uint64, count uint64, updateUndoLog bool) string {
	deletedRunes := make([]rune, 0, count)
	buffer := state.documentBuffer
	for i := uint64(0); i < count; i++ {
		didDelete, r := buffer.textTree.DeleteAtPosition(pos)
		if didDelete {
			deletedRunes = append(deletedRunes, r)
		}
	}

	edit := parser.NewDeleteEdit(pos, count)
	retokenizeAfterEdit(buffer, edit)

	deletedText := string(deletedRunes)
	if updateUndoLog && deletedText != "" {
		op := undo.DeleteOp(pos, deletedText)
		buffer.undoLog.TrackOp(op)
	}

	return deletedText
}

// ReplaceChar replaces the character under the cursor.
func ReplaceChar(state *EditorState, newChar rune) {
	buffer := state.documentBuffer
	pos := state.documentBuffer.cursor.position
	nextCharPos := locate.NextCharInLine(buffer.textTree, 1, true, pos)

	if nextCharPos == pos {
		// No character under the cursor on the current line, so abort.
		return
	}

	numToDelete := nextCharPos - pos
	deleteRunes(state, pos, numToDelete, true)

	switch newChar {
	case '\n':
		InsertNewline(state)
	case '\t':
		InsertTab(state)
		MoveCursor(state, func(p LocatorParams) uint64 {
			return locate.PrevCharInLine(p.TextTree, 1, false, p.CursorPos)
		})
	default:
		newText := string(newChar)
		if err := insertTextAtPosition(state, newText, pos, true); err != nil {
			// invalid UTF-8 rune; ignore it.
			log.Printf("Error inserting text %q: %v\n", newText, err)
		}
		MoveCursor(state, func(p LocatorParams) uint64 {
			return pos
		})
	}
}

// BeginNewLineAbove starts a new line above the current line, positioning the cursor at the end of the new line.
func BeginNewLineAbove(state *EditorState) {
	autoIndent := state.documentBuffer.autoIndent
	MoveCursor(state, func(params LocatorParams) uint64 {
		pos := locate.PrevLineBoundary(params.TextTree, params.CursorPos)
		if autoIndent {
			return locate.NextNonWhitespaceOrNewline(params.TextTree, pos)
		} else {
			return pos
		}
	})

	InsertNewline(state)

	MoveCursor(state, func(params LocatorParams) uint64 {
		pos := locate.StartOfLineAbove(params.TextTree, 1, params.CursorPos)
		if autoIndent {
			return locate.NextNonWhitespaceOrNewline(params.TextTree, pos)
		} else {
			return pos
		}
	})
}

// JoinLines joins the next line with the current line.
// This matches vim's behavior, which has some subtle edge cases
// involving empty lines and indentation at the beginning of lines.
func JoinLines(state *EditorState) {
	buffer := state.documentBuffer
	cursorPos := buffer.cursor.position

	nextNewlinePos, newlineLen, foundNewline := locate.NextNewline(buffer.textTree, cursorPos)
	if !foundNewline {
		// If we're on the last line, do nothing.
		return
	}

	// Delete newline and any indentation at start of next line.
	startOfNextLinePos := nextNewlinePos + newlineLen
	endOfIndentationPos := locate.NextNonWhitespaceOrNewline(buffer.textTree, startOfNextLinePos)
	deleteRunes(state, nextNewlinePos, endOfIndentationPos-nextNewlinePos, true)

	// Replace the newline with a space and move the cursor there.
	mustInsertRuneAtPosition(state, ' ', nextNewlinePos, true)
	MoveCursor(state, func(LocatorParams) uint64 { return nextNewlinePos })

	// If the space is adjacent to a newline, delete it.
	if isAdjacentToNewlineOrEof(buffer.textTree, nextNewlinePos) {
		deleteRunes(state, nextNewlinePos, 1, true)
	}

	// Move the cursor onto the line if necessary.
	MoveCursor(state, func(p LocatorParams) uint64 {
		return locate.ClosestCharOnLine(p.TextTree, p.CursorPos)
	})
}

func isAdjacentToNewlineOrEof(textTree *text.Tree, pos uint64) bool {
	seg := segment.Empty()

	reader := textTree.ReaderAtPosition(pos)
	forwardIter := segment.NewGraphemeClusterIter(reader)

	// Consume the grapheme cluster on the position.
	if err := forwardIter.NextSegment(seg); err != nil && err != io.EOF {
		panic(err)
	}

	// Check the next grapheme cluster after the position.
	err := forwardIter.NextSegment(seg)
	if err == io.EOF || (err == nil && seg.HasNewline()) {
		return true
	} else if err != nil {
		panic(err)
	}

	// Check the grapheme cluster before the position.
	reverseReader := textTree.ReverseReaderAtPosition(pos)
	backwardIter := segment.NewReverseGraphemeClusterIter(reverseReader)
	err = backwardIter.NextSegment(seg)
	if err == io.EOF || (err == nil && seg.HasNewline()) {
		return true
	} else if err != nil {
		panic(err)
	}

	return false
}

// ToggleCaseAtCursor changes the character under the cursor from upper-to-lowercase or vice-versa.
func ToggleCaseAtCursor(state *EditorState) {
	buffer := state.documentBuffer
	startPos := buffer.cursor.position
	endPos := locate.NextCharInLine(buffer.textTree, 1, true, startPos)
	toggleCaseForRange(state, startPos, endPos)
	MoveCursor(state, func(p LocatorParams) uint64 {
		return locate.NextCharInLine(buffer.textTree, 1, false, p.CursorPos)
	})
}

// ToggleCaseInSelection toggles the case of all characters in the region
// from the cursor position to the position found by selectionEndLoc.
func ToggleCaseInSelection(state *EditorState, selectionEndLoc Locator) {
	buffer := state.documentBuffer
	cursorPos := buffer.cursor.position
	endPos := selectionEndLoc(locatorParamsForBuffer(buffer))
	toggleCaseForRange(state, cursorPos, endPos)
}

// toggleCaseForRange changes the case of all characters in the range [startPos, endPos)
// It does NOT move the cursor.
func toggleCaseForRange(state *EditorState, startPos uint64, endPos uint64) {
	tree := state.documentBuffer.textTree
	newRunes := make([]rune, 0, 1)
	reader := tree.ReaderAtPosition(startPos)
	for pos := startPos; pos < endPos; pos++ {
		r, _, err := reader.ReadRune()
		if err == io.EOF {
			break
		} else if err != nil {
			panic(err) // Should never happen because the document is valid UTF-8.
		}
		newRunes = append(newRunes, text.ToggleRuneCase(r))
	}
	deleteRunes(state, startPos, uint64(len(newRunes)), true)
	mustInsertTextAtPosition(state, string(newRunes), startPos, true)
}

// IndentLines indents every line from the current cursor position to the position found by targetLineLoc.
func IndentLines(state *EditorState, targetLineLoc Locator, count uint64) {
	tabs := tabText(state, count) // Allocate once for all lines.
	changeIndentationOfLines(state, targetLineLoc, func(state *EditorState, lineNum uint64) {
		buffer := state.documentBuffer
		startOfLinePos := locate.StartOfLineNum(buffer.textTree, lineNum)
		endOfLinePos := locate.NextLineBoundary(buffer.textTree, true, startOfLinePos)
		if startOfLinePos < endOfLinePos {
			// Indent if line is non-empty.
			insertTabsAtPos(state, startOfLinePos, tabs)
		}
	})
}

// OutdentLines outdents every line from the current cursor position to the position found by targetLineLoc.
func OutdentLines(state *EditorState, targetLineLoc Locator, count uint64) {
	changeIndentationOfLines(state, targetLineLoc, func(state *EditorState, lineNum uint64) {
		buffer := state.documentBuffer
		startOfLinePos := locate.StartOfLineNum(buffer.textTree, lineNum)
		numToDelete := numRunesInIndent(buffer, startOfLinePos, count)
		deleteRunes(state, startOfLinePos, numToDelete, true)
	})
}

func changeIndentationOfLines(state *EditorState, targetLineLoc Locator, f func(*EditorState, uint64)) {
	buffer := state.documentBuffer
	currentLine := buffer.textTree.LineNumForPosition(buffer.cursor.position)
	targetPos := targetLineLoc(locatorParamsForBuffer(buffer))
	targetLine := buffer.textTree.LineNumForPosition(targetPos)
	if targetLine < currentLine {
		currentLine, targetLine = targetLine, currentLine
	}

	for lineNum := currentLine; lineNum <= targetLine; lineNum++ {
		f(state, lineNum)
	}

	startOfFirstLinePos := locate.StartOfLineNum(buffer.textTree, currentLine)
	newCursorPos := locate.NextNonWhitespaceOrNewline(buffer.textTree, startOfFirstLinePos)
	buffer.cursor = cursorState{position: newCursorPos}
}

func numRunesInIndent(buffer *BufferState, startOfLinePos uint64, count uint64) uint64 {
	var offset uint64
	pos := startOfLinePos
	endOfIndentPos := locate.NextNonWhitespaceOrNewline(buffer.textTree, startOfLinePos)
	reader := buffer.textTree.ReaderAtPosition(pos)
	iter := segment.NewGraphemeClusterIter(reader)
	seg := segment.Empty()
	for pos < endOfIndentPos && offset < (buffer.tabSize*count) {
		err := iter.NextSegment(seg)
		if err == io.EOF {
			break
		} else if err != nil {
			panic(err)
		}
		offset += cellwidth.GraphemeClusterWidth(seg.Runes(), offset, buffer.tabSize)
		pos += seg.NumRunes()
	}

	return pos - startOfLinePos
}

// CopyRange copies the characters in a range to the default page in the clipboard.
func CopyRange(state *EditorState, page clipboard.PageId, loc RangeLocator) {
	startPos, endPos := loc(locatorParamsForBuffer(state.documentBuffer))
	if startPos >= endPos {
		return
	}
	text := copyText(state.documentBuffer.textTree, startPos, endPos-startPos)
	state.clipboard.Set(page, clipboard.PageContent{Text: text})
}

// CopyLine copies the line under the cursor to the default page in the clipboard.
func CopyLine(state *EditorState, page clipboard.PageId) {
	buffer := state.documentBuffer
	startPos := locate.StartOfLineAtPos(buffer.textTree, buffer.cursor.position)
	endPos := locate.NextLineBoundary(buffer.textTree, true, startPos)
	line := copyText(buffer.textTree, startPos, endPos-startPos)
	content := clipboard.PageContent{
		Text:     line,
		Linewise: true,
	}
	state.clipboard.Set(page, content)
}

// CopySelection copies the current selection to the clipboard.
func CopySelection(state *EditorState, page clipboard.PageId) {
	buffer := state.documentBuffer
	text, r := copySelectionText(buffer)
	content := clipboard.PageContent{Text: text}
	if buffer.selector.Mode() == selection.ModeLine {
		content.Linewise = true
	}
	state.clipboard.Set(page, content)

	MoveCursor(state, func(LocatorParams) uint64 { return r.StartPos })
}

// copyText copies part of the document text to a string.
func copyText(tree *text.Tree, pos uint64, numRunes uint64) string {
	var sb strings.Builder
	var offset uint64
	reader := tree.ReaderAtPosition(pos)
	for offset < numRunes {
		r, _, err := reader.ReadRune()
		if err == io.EOF {
			break
		} else if err != nil {
			panic(err) // should never happen because text should be valid UTF-8
		}
		sb.WriteRune(r)
		offset++
	}
	return sb.String()
}

// copySelectionText copies the currently selected text.
// If no text is selected, it returns an empty string.
func copySelectionText(buffer *BufferState) (string, selection.Region) {
	if buffer.selector.Mode() == selection.ModeNone {
		return "", selection.EmptyRegion
	}
	r := buffer.SelectedRegion()
	text := copyText(buffer.textTree, r.StartPos, r.EndPos-r.StartPos)
	return text, r
}

// PasteAfterCursor inserts the text from the clipboard after the cursor position.
func PasteAfterCursor(state *EditorState, page clipboard.PageId) {
	content := state.clipboard.Get(page)
	pos := state.documentBuffer.cursor.position
	if content.Linewise {
		pos = locate.NextLineBoundary(state.documentBuffer.textTree, true, pos)
		mustInsertRuneAtPosition(state, '\n', pos, true)
		pos++
	} else {
		pos = locate.NextCharInLine(state.documentBuffer.textTree, 1, true, pos)
	}

	err := insertTextAtPosition(state, content.Text, pos, true)
	if err != nil {
		log.Printf("Error pasting text: %v\n", err)
		return
	}

	if content.Linewise {
		MoveCursor(state, func(LocatorParams) uint64 { return pos })
	} else {
		MoveCursor(state, func(params LocatorParams) uint64 {
			posAfterInsert := pos + uint64(utf8.RuneCountInString(content.Text))
			return locate.PrevCharInLine(params.TextTree, 1, false, posAfterInsert)
		})
	}
}

// PasteBeforeCursor inserts the text from the clipboard before the cursor position.
func PasteBeforeCursor(state *EditorState, page clipboard.PageId) {
	content := state.clipboard.Get(page)
	pos := state.documentBuffer.cursor.position
	if content.Linewise {
		pos = locate.StartOfLineAtPos(state.documentBuffer.textTree, pos)
		mustInsertRuneAtPosition(state, '\n', pos, true)
	}

	err := insertTextAtPosition(state, content.Text, pos, true)
	if err != nil {
		log.Printf("Error pasting text: %v\n", err)
		return
	}

	if content.Linewise {
		MoveCursor(state, func(LocatorParams) uint64 { return pos })
	} else {
		MoveCursor(state, func(params LocatorParams) uint64 {
			posAfterInsert := pos + uint64(utf8.RuneCountInString(content.Text))
			newPos := locate.PrevChar(params.TextTree, 1, posAfterInsert)
			return locate.ClosestCharOnLine(params.TextTree, newPos)
		})
	}
}
