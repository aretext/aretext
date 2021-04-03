package state

import (
	"log"
	"unicode/utf8"

	"github.com/aretext/aretext/cellwidth"
	"github.com/aretext/aretext/locate"
	"github.com/aretext/aretext/syntax/parser"
	"github.com/aretext/aretext/text"
	"github.com/aretext/aretext/text/segment"
	"github.com/pkg/errors"
)

// InsertRune inserts a rune at the current cursor location.
func InsertRune(state *EditorState, r rune) {
	buffer := state.documentBuffer
	startPos := buffer.cursor.position
	if err := insertTextAtPosition(state, string(r), startPos); err != nil {
		log.Printf("Error inserting rune: %v\n", err)
		return
	}
	buffer.cursor.position = startPos + 1
}

// insertTextAtPosition inserts text into the document.
// It also updates the syntax tokens and unsaved changes flag.
// It does NOT move the cursor.
func insertTextAtPosition(state *EditorState, s string, pos uint64) error {
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

	state.hasUnsavedChanges = true

	return nil
}

func mustInsertRuneAtPosition(state *EditorState, r rune, pos uint64) {
	err := insertTextAtPosition(state, string(r), pos)
	if err != nil {
		panic(err)
	}
}

// InsertNewline inserts a newline at the current cursor position.
func InsertNewline(state *EditorState) {
	cursorPos := state.documentBuffer.cursor.position
	mustInsertRuneAtPosition(state, '\n', cursorPos)
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
	deleteRunes(state, startPos, count)
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
			mustInsertRuneAtPosition(state, '\t', pos)
			i += tabSize
		} else {
			mustInsertRuneAtPosition(state, ' ', pos)
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
	mustInsertRuneAtPosition(state, '\t', cursorPos)
	return cursorPos + 1
}

func insertSpacesForTab(state *EditorState) uint64 {
	buffer := state.documentBuffer
	tabSize := buffer.tabSize
	offset := cursorOffsetInLine(state.documentBuffer)
	numSpaces := tabSize - (offset % tabSize)
	cursorPos := buffer.cursor.position
	for i := uint64(0); i < numSpaces; i++ {
		mustInsertRuneAtPosition(state, ' ', cursorPos)
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

	if startPos < deleteToPos {
		deleteRunes(state, startPos, deleteToPos-startPos)
		buffer.cursor = cursorState{position: startPos}
	} else if startPos > deleteToPos {
		deleteRunes(state, deleteToPos, startPos-deleteToPos)
		buffer.cursor = cursorState{position: deleteToPos}
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
	for i := uint64(0); i < numLinesToDelete; i++ {
		deleteLine(state, currentLine)
	}
}

func deleteLine(state *EditorState, lineNum uint64) {
	buffer := state.documentBuffer
	startOfLinePos := buffer.textTree.LineStartPosition(lineNum)
	startOfNextLinePos := buffer.textTree.LineStartPosition(lineNum + 1)

	isLastLine := lineNum+1 >= buffer.textTree.NumLines()
	if isLastLine && startOfLinePos > 0 {
		// The last line does not have a newline at the end, so delete the newline from the end of the previous line instead.
		startOfLinePos--
	}

	numToDelete := startOfNextLinePos - startOfLinePos
	for i := uint64(0); i < numToDelete; i++ {
		buffer.textTree.DeleteAtPosition(startOfLinePos)
	}

	edit := parser.Edit{Pos: startOfLinePos, NumDeleted: numToDelete}
	if err := retokenizeAfterEdit(buffer, edit); err != nil {
		log.Printf("Error retokenizing doument: %v\n", err)
	}

	buffer.cursor = cursorState{position: startOfLinePos}
	if buffer.cursor.position >= buffer.textTree.NumChars() {
		buffer.cursor = cursorState{
			position: locate.StartOfLastLine(buffer.textTree),
		}
	}

	state.hasUnsavedChanges = state.hasUnsavedChanges || numToDelete > 0
}

// deleteRunes deletes text from the document.
// It also updates the syntax token and unsaved changes flag.
// It does NOT move the cursor.
func deleteRunes(state *EditorState, pos uint64, count uint64) {
	buffer := state.documentBuffer
	for i := uint64(0); i < count; i++ {
		buffer.textTree.DeleteAtPosition(pos)
	}
	edit := parser.Edit{Pos: pos, NumDeleted: count}
	if err := retokenizeAfterEdit(buffer, edit); err != nil {
		// This should never happen when using a text tree reader.
		log.Fatalf("error deleting runes: %v\n", errors.Wrapf(err, "retokenizeAfterEdit"))
	}
	state.hasUnsavedChanges = true
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
	deleteRunes(state, cursorPos, numToDelete)

	pos := cursorPos
	if err := insertTextAtPosition(state, newText, pos); err != nil {
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
	deleteRunes(state, nextNewlinePos, endOfIndentationPos-nextNewlinePos)

	// Replace the newline with a space and move the cursor there.
	mustInsertRuneAtPosition(state, ' ', nextNewlinePos)
	MoveCursor(state, func(LocatorParams) uint64 { return nextNewlinePos })

	// If the space is adjacent to a newline, delete it.
	if isAdjacentToNewlineOrEof(buffer.textTree, nextNewlinePos) {
		deleteRunes(state, nextNewlinePos, 1)
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
