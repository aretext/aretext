package state

import (
	"github.com/aretext/aretext/locate"
)

// Scroll direction represents the direction of the scroll (forward or backward).
type ScrollDirection int

const (
	ScrollDirectionForward = iota
	ScrollDirectionBackward
)

// ResizeView resizes the view to the specified width and height.
func ResizeView(state *EditorState, width, height uint64) {
	state.screenWidth = width
	state.screenHeight = height
	state.documentBuffer.view.x = 0
	state.documentBuffer.view.y = 0
	state.documentBuffer.view.width = state.screenWidth
	state.documentBuffer.view.height = 0
	if height > 0 {
		// Leave one line for the status bar at the bottom.
		state.documentBuffer.view.height = height - 1
	}
}

// ScrollViewToCursor moves the view origin so that the cursor is visible.
func ScrollViewToCursor(state *EditorState) {
	buffer := state.documentBuffer
	scrollViewToPosition(buffer, buffer.cursor.position)
}

func scrollViewToPosition(buffer *BufferState, pos uint64) {
	buffer.view.textOrigin = locate.ViewOriginAfterScroll(
		pos,
		buffer.textTree,
		buffer.view.textOrigin,
		buffer.view.width-buffer.LineNumMarginWidth(),
		buffer.view.height,
		buffer.tabSize)
}

// ScrollViewByNumLines moves the view origin up or down by the specified number of lines.
func ScrollViewByNumLines(state *EditorState, direction ScrollDirection, numLines uint64) {
	buffer := state.documentBuffer
	lineNum := buffer.textTree.LineNumForPosition(buffer.view.textOrigin)
	if direction == ScrollDirectionForward {
		lineNum += numLines
	} else if lineNum >= numLines {
		lineNum -= numLines
	} else {
		lineNum = 0
	}

	lineNum = locate.ClosestValidLineNum(buffer.textTree, lineNum)

	// When scrolling to the end of the file, we want most of the last lines to remain visible.
	// To achieve this, set the view origin (viewHeight - scrollMargin) lines above
	// the last line.  This will leave a few blank lines past the end of the document
	// (the scroll margin) for consistency with ScrollToCursor.
	lastLineNum := locate.ClosestValidLineNum(buffer.textTree, buffer.textTree.NumLines())
	if lastLineNum-lineNum < buffer.view.height {
		if lastLineNum+locate.ScrollMargin+1 > buffer.view.height {
			lineNum = lastLineNum + locate.ScrollMargin + 1 - buffer.view.height
		} else {
			lineNum = 0
		}
	}

	buffer.view.textOrigin = buffer.textTree.LineStartPosition(lineNum)
}
