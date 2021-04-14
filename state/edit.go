package state

import (
	"io"
	"log"
	"strings"
	"unicode/utf8"

	"github.com/aretext/aretext/cellwidth"
	"github.com/aretext/aretext/clipboard"
	"github.com/aretext/aretext/locate"
	"github.com/aretext/aretext/selection"
	"github.com/aretext/aretext/syntax/parser"
	"github.com/aretext/aretext/text"
	"github.com/aretext/aretext/text/segment"
	textUtf8 "github.com/aretext/aretext/text/utf8"
	"github.com/aretext/aretext/undo"
	"github.com/pkg/errors"
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

	for i, r := range s {
		if err := buffer.textTree.InsertAtPosition(pos+uint64(i), r); err != nil {
			return errors.Wrapf(err, "text.Tree.InsertAtPosition")
		}
	}

	edit := parser.Edit{
		Pos:         pos,
		NumInserted: uint64(utf8.RuneCountInString(s)),
	}
	if err := retokenizeAfterEdit(buffer, edit); err != nil {
		return errors.Wrapf(err, "retokenizeAfterEdit")
	}

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
	iter := segment.NewGraphemeClusterIterForTree(buffer.textTree, prevLineStartPos, text.ReadDirectionForward)
	seg := segment.Empty()
	numCols := uint64(0)
	for {
		eof := segment.NextOrEof(iter, seg)
		if eof {
			break
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

// InsertTab inserts a tab at the current cursor position.
func InsertTab(state *EditorState) {
	var cursorPos uint64
	if state.documentBuffer.tabExpand {
		cursorPos = insertSpacesForTab(state)
	} else {
		cursorPos = insertTabRune(state)
	}
	state.documentBuffer.cursor = cursorState{position: cursorPos}
}

func insertTabRune(state *EditorState) uint64 {
	cursorPos := state.documentBuffer.cursor.position
	mustInsertRuneAtPosition(state, '\t', cursorPos, true)
	return cursorPos + 1
}

func insertSpacesForTab(state *EditorState) uint64 {
	buffer := state.documentBuffer
	tabSize := buffer.tabSize
	offset := cursorOffsetInLine(state.documentBuffer)
	numSpaces := tabSize - (offset % tabSize)
	cursorPos := buffer.cursor.position
	for i := uint64(0); i < numSpaces; i++ {
		mustInsertRuneAtPosition(state, ' ', cursorPos, true)
		cursorPos++
	}
	return cursorPos
}

func cursorOffsetInLine(buffer *BufferState) uint64 {
	var offset uint64
	pos := locate.StartOfLineAtPos(buffer.textTree, buffer.cursor.position)
	iter := segment.NewGraphemeClusterIterForTree(buffer.textTree, pos, text.ReadDirectionForward)
	seg := segment.Empty()
	for pos < buffer.cursor.position {
		eof := segment.NextOrEof(iter, seg)
		if eof {
			break
		}
		offset += cellwidth.GraphemeClusterWidth(seg.Runes(), offset, buffer.tabSize)
		pos += seg.NumRunes()
	}
	return offset
}

// DeleteRunes deletes characters from the cursor position up to (but not including) the position returned by the locator.
// It can delete either forwards or backwards from the cursor.
// The cursor position will be set to the start of the deleted region,
// which could be on a newline character or past the end of the text.
func DeleteRunes(state *EditorState, loc Locator) {
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
		state.clipboard.Set(clipboard.PageDefault, clipboard.PageContent{
			Text:             deletedText,
			InsertOnNextLine: false,
		})
	}
}

// DeleteSelection deletes the currently selected region, if any.
func DeleteSelection(state *EditorState) {
	buffer := state.documentBuffer
	selectionMode := buffer.selector.Mode()
	if selectionMode == selection.ModeNone {
		return
	}

	r := buffer.SelectedRegion()
	MoveCursor(state, func(LocatorParams) uint64 { return r.StartPos })

	if selectionMode == selection.ModeChar {
		DeleteRunes(state, func(LocatorParams) uint64 { return r.EndPos })
	} else if selectionMode == selection.ModeLine {
		DeleteLines(state, func(LocatorParams) uint64 { return r.EndPos }, false)
	} else {
		panic("Unsupported selection mode")
	}
}

// DeleteLines deletes lines from the cursor's current line to the line of a target cursor.
// It moves the cursor to the start of the line following the last deleted line.
func DeleteLines(state *EditorState, targetLineLoc Locator, abortIfTargetIsCurrentLine bool) {
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

	numLinesToDelete := targetLine - currentLine + 1
	deletedLines := make([]string, 0, numLinesToDelete)
	for i := uint64(0); i < numLinesToDelete; i++ {
		s := deleteLine(state, currentLine)
		if s != "" {
			deletedLines = append(deletedLines, stripStartingAndTrailingNewlines(s))
		}
	}

	if len(deletedLines) > 0 {
		deletedText := strings.Join(deletedLines, "\n")
		state.clipboard.Set(clipboard.PageDefault, clipboard.PageContent{
			Text:             deletedText,
			InsertOnNextLine: true,
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

func deleteLine(state *EditorState, lineNum uint64) string {
	buffer := state.documentBuffer
	startOfLinePos := buffer.textTree.LineStartPosition(lineNum)
	startOfNextLinePos := buffer.textTree.LineStartPosition(lineNum + 1)

	isLastLine := lineNum+1 >= buffer.textTree.NumLines()
	if isLastLine && startOfLinePos > 0 {
		// The last line does not have a newline at the end, so delete the newline from the end of the previous line instead.
		startOfLinePos--
	}

	numToDelete := startOfNextLinePos - startOfLinePos
	deletedText := deleteRunes(state, startOfLinePos, numToDelete, true)

	buffer.cursor = cursorState{position: startOfLinePos}
	if buffer.cursor.position >= buffer.textTree.NumChars() {
		buffer.cursor = cursorState{
			position: locate.StartOfLastLine(buffer.textTree),
		}
	}

	return deletedText
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

	edit := parser.Edit{Pos: pos, NumDeleted: count}
	if err := retokenizeAfterEdit(buffer, edit); err != nil {
		// This should never happen when using a text tree reader.
		log.Fatalf("error deleting runes: %v\n", errors.Wrapf(err, "retokenizeAfterEdit"))
	}

	deletedText := string(deletedRunes)
	if updateUndoLog && deletedText != "" {
		op := undo.DeleteOp(pos, deletedText)
		buffer.undoLog.TrackOp(op)
	}

	return deletedText
}

// ReplaceChar replaces the character under the cursor with the specified string.
func ReplaceChar(state *EditorState, newText string) {
	buffer := state.documentBuffer
	cursorPos := state.documentBuffer.cursor.position
	nextCharPos := locate.NextCharInLine(buffer.textTree, 1, true, cursorPos)

	if nextCharPos == cursorPos {
		// No character under the cursor on the current line, so abort.
		return
	}

	numToDelete := nextCharPos - cursorPos
	deleteRunes(state, cursorPos, numToDelete, true)

	pos := cursorPos
	if err := insertTextAtPosition(state, newText, pos, true); err != nil {
		// invalid UTF-8 rune; ignore it.
		log.Printf("Error inserting text '%s': %v\n", newText, err)
	}
	pos += uint64(utf8.RuneCountInString(newText))

	if newText != "\n" {
		buffer.cursor.position = cursorPos
	} else {
		buffer.cursor.position = pos
	}
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

	forwardIter := segment.NewGraphemeClusterIterForTree(textTree, pos, text.ReadDirectionForward)

	// Consume the grapheme cluster on the position.
	segment.NextOrEof(forwardIter, seg)

	// Check the next grapheme cluster after the position.
	eof := segment.NextOrEof(forwardIter, seg)
	if eof || seg.HasNewline() {
		return true
	}

	// Check the grapheme cluster before the position.
	backwardIter := segment.NewGraphemeClusterIterForTree(textTree, pos, text.ReadDirectionBackward)
	eof = segment.NextOrEof(backwardIter, seg)
	if eof || seg.HasNewline() {
		return true
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

// ToggleCaseInSelection toggles the case of all characters in the current selection.
func ToggleCaseInSelection(state *EditorState) {
	buffer := state.documentBuffer
	selectionMode := buffer.selector.Mode()
	if selectionMode == selection.ModeNone {
		return
	}
	region := buffer.SelectedRegion()
	toggleCaseForRange(state, region.StartPos, region.EndPos)
	MoveCursor(state, func(p LocatorParams) uint64 { return region.StartPos })
}

// toggleCaseForRange changes the case of all characters in the range [startPos, endPos)
// It does NOT move the cursor.
func toggleCaseForRange(state *EditorState, startPos uint64, endPos uint64) {
	tree := state.documentBuffer.textTree
	newRunes := make([]rune, 0, 1)
	reader := tree.ReaderAtPosition(startPos, text.ReadDirectionForward)
	runeIter := text.NewCloneableForwardRuneIter(reader)
	for pos := startPos; pos < endPos; pos++ {
		r, err := runeIter.NextRune()
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

// CopyLine copies the line under the cursor to the default page in the clipboard.
func CopyLine(state *EditorState) {
	buffer := state.documentBuffer
	startPos := locate.StartOfLineAtPos(buffer.textTree, buffer.cursor.position)
	endPos := locate.NextLineBoundary(buffer.textTree, true, startPos)
	line := copyText(buffer.textTree, startPos, endPos-startPos)
	content := clipboard.PageContent{
		Text:             line,
		InsertOnNextLine: true,
	}
	state.clipboard.Set(clipboard.PageDefault, content)
}

// CopySelection copies the current selection to the clipboard.
func CopySelection(state *EditorState) {
	buffer := state.documentBuffer
	if buffer.selector.Mode() == selection.ModeNone {
		return
	}

	r := buffer.SelectedRegion()

	text := copyText(buffer.textTree, r.StartPos, r.EndPos-r.StartPos)
	if len(text) == 0 {
		return
	}

	content := clipboard.PageContent{Text: text}
	if buffer.selector.Mode() == selection.ModeLine {
		content.InsertOnNextLine = true
	}
	state.clipboard.Set(clipboard.PageDefault, content)

	MoveCursor(state, func(LocatorParams) uint64 { return r.StartPos })
}

// copyText copies part of the document text to a string.
func copyText(tree *text.Tree, pos uint64, numRunes uint64) string {
	var offset uint64
	var buf [256]byte
	textBytes := make([]byte, 0, numRunes)
	r := tree.ReaderAtPosition(pos, text.ReadDirectionForward)
	for offset < numRunes {
		n, err := r.Read(buf[:])
		if err == io.EOF {
			if n == 0 {
				break
			}
		} else if err != nil {
			panic(err) // Should never happen for a text tree.
		}

		for i := 0; i < n && offset < numRunes; i++ {
			b := buf[i]
			textBytes = append(textBytes, b)
			offset += uint64(textUtf8.StartByteIndicator[b])
		}
	}

	return string(textBytes)
}

// PasteAfterCursor inserts the text from the default page in the clipboard at the cursor position.
func PasteAfterCursor(state *EditorState) {
	content := state.clipboard.Get(clipboard.PageDefault)
	pos := state.documentBuffer.cursor.position
	if content.InsertOnNextLine {
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

	MoveCursor(state, func(LocatorParams) uint64 { return pos })
}
