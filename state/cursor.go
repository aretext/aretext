package state

import (
	"io"

	"github.com/aretext/aretext/cellwidth"
	"github.com/aretext/aretext/locate"
	"github.com/aretext/aretext/selection"
	"github.com/aretext/aretext/text"
	"github.com/aretext/aretext/text/segment"
)

// cursorState is the current state of the cursor.
type cursorState struct {
	// position is a position within the text tree where the cursor appears.
	position uint64

	// logicalOffset is the number of cells after the end of the line
	// for the cursor's logical (not necessarily visible) position.
	// This is used for navigating up/down.
	// For example, consider this text, where [m] is the current cursor position.
	//     1: the quick
	//     2: brown
	//     3: fox ju[m]ped over the lazy dog
	// If the user then navigates up one line, then we'd see:
	//     1: the quick
	//     2: brow[n]  [*]
	//     3: fox jumped over the lazy dog
	// where [n] is the visible position and [*] is the logical position,
	// with logicalOffset = 2.
	// If the user then navigates up one line again, we'd see:
	//     1: the qu[i]ck
	//     2: brown
	//     3: fox jumped over the lazy dog
	// where [i] is the character directly above the logical position.
	logicalOffset uint64
}

// MoveCursor moves the cursor to the specified position in the document.
func MoveCursor(state *EditorState, loc Locator) {
	buffer := state.documentBuffer
	cursorPos := buffer.cursor.position
	newPos := loc(locatorParamsForBuffer(buffer))

	// Limit the position to within the document.
	if n := buffer.textTree.NumChars(); newPos > n {
		if n == 0 {
			newPos = 0
		} else {
			newPos = n - 1
		}
	}

	var logicalOffset uint64
	if newPos == cursorPos {
		// This handles the case where the user is moving the cursor up to a shorter line,
		// then tries to move the cursor to the right at the end of the line.
		// The cursor doesn't actually move, so when the user moves up another line,
		// it should use the offset from the longest line.
		logicalOffset = buffer.cursor.logicalOffset
	}

	buffer.cursor = cursorState{
		position:      newPos,
		logicalOffset: logicalOffset,
	}
}

// MoveCursorToLineAbove moves the cursor up by the specified number of lines, preserving the offset within the line.
func MoveCursorToLineAbove(state *EditorState, count uint64) {
	buffer := state.documentBuffer
	targetLineStartPos := locate.StartOfLineAbove(buffer.textTree, count, buffer.cursor.position)
	moveCursorToLine(buffer, targetLineStartPos)
}

// MoveCursorToLineBelow moves the cursor down by the specified number of lines, preserving the offset within the line.
func MoveCursorToLineBelow(state *EditorState, count uint64) {
	buffer := state.documentBuffer
	targetLineStartPos := locate.StartOfLineBelow(buffer.textTree, count, buffer.cursor.position)
	moveCursorToLine(buffer, targetLineStartPos)
}

func moveCursorToLine(buffer *BufferState, targetLineStartPos uint64) {
	lineStartPos := locate.StartOfLineAtPos(buffer.textTree, buffer.cursor.position)
	if targetLineStartPos == lineStartPos {
		return
	}

	targetOffset := findOffsetFromLineStart(
		buffer.textTree,
		lineStartPos,
		buffer.cursor,
		buffer.tabSize)

	newPos, actualOffset := advanceToOffset(
		buffer.textTree,
		targetLineStartPos,
		targetOffset,
		buffer.tabSize)

	buffer.cursor = cursorState{
		position:      newPos,
		logicalOffset: targetOffset - actualOffset,
	}
}

func findOffsetFromLineStart(textTree *text.Tree, lineStartPos uint64, cursor cursorState, tabSize uint64) uint64 {
	reader := textTree.ReaderAtPosition(lineStartPos)
	segmentIter := segment.NewGraphemeClusterIter(reader)
	seg := segment.Empty()
	pos, offset := lineStartPos, uint64(0)

	for {
		err := segmentIter.NextSegment(seg)
		if err == io.EOF || (err == nil && pos >= cursor.position) {
			break
		} else if err != nil {
			panic(err)
		}

		offset += cellwidth.GraphemeClusterWidth(seg.Runes(), offset, tabSize)
		pos += seg.NumRunes()
	}

	return offset + cursor.logicalOffset
}

func advanceToOffset(textTree *text.Tree, lineStartPos uint64, targetOffset uint64, tabSize uint64) (uint64, uint64) {
	reader := textTree.ReaderAtPosition(lineStartPos)
	segmentIter := segment.NewGraphemeClusterIter(reader)
	seg := segment.Empty()
	var endOfLineOrFile bool
	var prevPosOffset, posOffset, cellOffset uint64

	for {
		err := segmentIter.NextSegment(seg)
		if err == io.EOF {
			endOfLineOrFile = true
			break
		} else if err != nil {
			panic(err)
		}

		if seg.HasNewline() {
			endOfLineOrFile = true
			break
		}

		gcWidth := cellwidth.GraphemeClusterWidth(seg.Runes(), cellOffset, tabSize)
		if cellOffset+gcWidth > targetOffset {
			break
		}

		cellOffset += gcWidth
		prevPosOffset = posOffset
		posOffset += seg.NumRunes()
	}

	if endOfLineOrFile {
		if cellOffset > 0 {
			cellOffset--
		}
		return lineStartPos + prevPosOffset, cellOffset
	}

	return lineStartPos + posOffset, cellOffset
}

// MoveCursorToStartOfSelection moves the cursor to the start of the current selection.
// If nothing is selected, this does nothing.
func MoveCursorToStartOfSelection(state *EditorState) {
	if state.documentBuffer.SelectionMode() == selection.ModeNone {
		return
	}
	selectedRegion := state.documentBuffer.SelectedRegion()
	MoveCursor(state, func(p LocatorParams) uint64 {
		return selectedRegion.StartPos
	})
}

// SelectRange selects a given range in charwise mode.
// This will clear any prior selection and move the cursor to the end of the new selection.
func SelectRange(state *EditorState, loc RangeLocator) {
	startPos, endPos := loc(locatorParamsForBuffer(state.documentBuffer))
	selector := state.documentBuffer.selector
	selector.Clear()
	selector.Start(selection.ModeChar, startPos)
	MoveCursor(state, func(p LocatorParams) uint64 {
		if endPos > 0 {
			return endPos - 1
		} else {
			return 0
		}
	})
}
