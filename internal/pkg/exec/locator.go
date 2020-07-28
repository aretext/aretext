package exec

import (
	"fmt"

	"github.com/wedaly/aretext/internal/pkg/text"
)

// Locator finds the position of the cursor according to some criteria.
type Locator interface {
	fmt.Stringer

	// Locate finds the next position of the cursor based on the current state and criteria for this locator.
	Locate(state *State) cursorState
}

// charInLineLocator locates a character (grapheme cluster) in the current line.
type charInLineLocator struct {
	direction              text.ReadDirection
	count                  uint64
	includeEndOfLineOrFile bool
}

// NewCharInLineLocator builds a new locator for a character on the same line as the cursor.
// The direction arg indicates whether to read forward or backwards from the cursor.
// The count arg is the maximum number of characters to move the cursor.
func NewCharInLineLocator(direction text.ReadDirection, count uint64, includeEndOfLineOrFile bool) Locator {
	if count == 0 {
		panic("Count must be greater than zero")
	}
	return &charInLineLocator{direction, count, includeEndOfLineOrFile}
}

func (loc *charInLineLocator) String() string {
	return fmt.Sprintf("CharInLineLocator(%s, %d, %t)", directionString(loc.direction), loc.count, loc.includeEndOfLineOrFile)
}

// Locate finds a character to the right of the cursor on the current line.
func (loc *charInLineLocator) Locate(state *State) cursorState {
	newPosition := loc.findPosition(state)

	logicalOffset := uint64(0)
	if newPosition == state.cursor.position {
		// This handles the case where the user is moving the cursor up to a shorter line,
		// then tries to move the cursor to the right at the end of the line.
		// The cursor doesn't actually move, so when the user moves up another line,
		// it should use the offset from the longest line.
		logicalOffset = state.cursor.logicalOffset
	}

	return cursorState{
		position:      newPosition,
		logicalOffset: logicalOffset,
	}
}

func (loc *charInLineLocator) findPosition(state *State) uint64 {
	if loc.direction == text.ReadDirectionBackward {
		return loc.findPositionBeforeCursor(state)
	}
	return loc.findPositionAfterCursor(state)
}

func (loc *charInLineLocator) findPositionBeforeCursor(state *State) uint64 {
	startPos := state.cursor.position
	segmentIter := newGraphemeClusterSegmentIter(state.tree, startPos, text.ReadDirectionBackward)

	var offset uint64
	for i := uint64(0); i < loc.count; i++ {
		segment, eof := segmentIter.nextSegment()
		if eof {
			break
		}

		if offset+uint64(len(segment)) > startPos {
			return 0
		}

		if segmentHasNewline(segment) {
			if loc.includeEndOfLineOrFile {
				offset += uint64(len(segment))
			}
			break
		}

		offset += uint64(len(segment))
	}

	return startPos - offset
}

func (loc *charInLineLocator) findPositionAfterCursor(state *State) uint64 {
	startPos := state.cursor.position
	segmentIter := newGraphemeClusterSegmentIter(state.tree, startPos, text.ReadDirectionForward)

	var endOfLineOrFile bool
	var prevPrevOffset, prevOffset uint64
	for i := uint64(0); i <= loc.count; i++ {
		segment, eof := segmentIter.nextSegment()
		if eof {
			endOfLineOrFile = true
			break
		}

		if segmentHasNewline(segment) {
			endOfLineOrFile = true
			break
		}

		prevPrevOffset = prevOffset
		prevOffset += uint64(len(segment))
	}

	if endOfLineOrFile && loc.includeEndOfLineOrFile {
		return startPos + prevOffset
	}
	return startPos + prevPrevOffset
}

// ontoLineLocator finds the closest grapheme cluster on a line (not newline or past end of text).
// This is useful for "resetting" the cursor onto a line
// (for example, after deleting the last character on the line or exiting insert mode).
type ontoLineLocator struct {
}

func NewOntoLineLocator() Locator {
	return &ontoLineLocator{}
}

// Locate finds the closest grapheme cluster on a line (not newline or past end of text).
func (loc *ontoLineLocator) Locate(state *State) cursorState {
	// If past the end of the text, return the start of the last grapheme cluster.
	numChars := state.tree.NumChars()
	if state.cursor.position >= numChars {
		newPos := loc.findPrevGraphemeCluster(state.tree, numChars, 1)
		return cursorState{position: newPos}
	}

	// If on a grapheme cluster with a newline (either "\n" or "\r\n"), return the start
	// of the last grapheme cluster before the current grapheme cluster.
	if hasNewline, afterNewlinePos := loc.findNewlineAtPos(state.tree, state.cursor.position); hasNewline {
		newPos := loc.findPrevGraphemeCluster(state.tree, afterNewlinePos, 2)
		return cursorState{position: newPos}
	}

	// The cursor is already on a line, so do nothing.
	return cursorState{position: state.cursor.position}
}

func (loc *ontoLineLocator) findNewlineAtPos(tree *text.Tree, pos uint64) (bool, uint64) {
	segmentIter := newGraphemeClusterSegmentIter(tree, pos, text.ReadDirectionForward)
	segment, eof := segmentIter.nextSegment()
	if eof {
		return false, 0
	}

	if segmentHasNewline(segment) {
		return true, pos + uint64(len(segment))
	}

	return false, 0
}

func (loc *ontoLineLocator) findPrevGraphemeCluster(tree *text.Tree, pos uint64, count int) uint64 {
	segmentIter := newGraphemeClusterSegmentIter(tree, pos, text.ReadDirectionBackward)

	// Iterate backward by (count - 1) grapheme clusters.
	var offset uint64
	for i := 0; i < count-1; i++ {
		segment, eof := segmentIter.nextSegment()
		if eof {
			break
		}

		offset += uint64(len(segment))
	}

	// Check the next grapheme cluster after (count - 1) grapheme clusters.
	segment, eof := segmentIter.nextSegment()
	if eof {
		return 0
	}

	// If the immediately preceding cluster is a newline, then we're on
	// an empty line, in which case we shouldn't move the cursor.
	if segmentHasNewline(segment) {
		return pos - offset
	}

	// Otherwise, move the cursor back a cluster to position it at the end of the previous line.
	return pos - offset - uint64(len(segment))
}

func (loc *ontoLineLocator) String() string {
	return "OntoLineLocator()"
}

// relativeLineLocator finds a position at the same offset above or below the current line.
type relativeLineLocator struct {
	direction text.ReadDirection
	count     uint64
}

// NewRelativeLineLocator returns a locator for moving the cursor up or down by some number of lines.
// Count is the number of lines to move, and it must be at least one.
// Direction indicates whether to move up (ReadDirectionBackward) or down (ReadDirectionForward).
func NewRelativeLineLocator(direction text.ReadDirection, count uint64) Locator {
	if count == 0 {
		panic("Count must be greater than zero")
	}
	return &relativeLineLocator{direction, count}
}

// Locate returns a cursor position at the same offset above or below the current line.
// It does nothing when moving up from the first line or down from the last line.
// If the target line has fewer characters than the starting line, then the extra characters
// will be counted the cursor's logical offset.
// If the target line has more characters than the starting line, then the cursor will move
// as close as possible to the logical offset.
func (loc *relativeLineLocator) Locate(state *State) cursorState {
	targetOffset := loc.findOffsetFromLineStart(state)
	lineStartPos, newlineCount := loc.findStartOfLine(state.tree, state.cursor.position)
	if newlineCount == 0 {
		return state.cursor
	}

	newPos, actualOffset := loc.advanceToOffset(state.tree, lineStartPos, targetOffset)
	return cursorState{
		position:      newPos,
		logicalOffset: targetOffset - actualOffset,
	}
}

func (loc *relativeLineLocator) findOffsetFromLineStart(state *State) uint64 {
	segmentIter := newGraphemeClusterSegmentIter(state.tree, state.cursor.position, text.ReadDirectionBackward)

	var offset uint64
	for {
		segment, eof := segmentIter.nextSegment()
		if eof {
			break
		}

		if segmentHasNewline(segment) {
			break
		}

		offset++
	}

	return offset + state.cursor.logicalOffset
}

func (loc *relativeLineLocator) findStartOfLine(tree *text.Tree, pos uint64) (lineStartPos, newlineCount uint64) {
	if loc.direction == text.ReadDirectionBackward {
		return loc.findStartOfLineAbove(tree, pos)
	} else {
		return loc.findStartOfLineBelow(tree, pos)
	}
}

func (loc *relativeLineLocator) findStartOfLineAbove(tree *text.Tree, pos uint64) (lineStartPos, newlineCount uint64) {
	segmentIter := newGraphemeClusterSegmentIter(tree, pos, text.ReadDirectionBackward)

	var offset uint64
	for {
		segment, eof := segmentIter.nextSegment()
		if eof {
			break
		}

		if segmentHasNewline(segment) {
			newlineCount++
			if newlineCount > loc.count {
				break
			}
		}

		offset += uint64(len(segment))
	}

	return pos - offset, newlineCount
}

func (loc *relativeLineLocator) findStartOfLineBelow(tree *text.Tree, pos uint64) (lineStartPos, newlineCount uint64) {
	segmentIter := newGraphemeClusterSegmentIter(tree, pos, text.ReadDirectionForward)

	// Lookahead one grapheme cluster.
	nextSegmentIter := segmentIter.clone()
	nextSegmentIter.nextSegment()

	var offset uint64
	for newlineCount < loc.count {
		segment, eof := segmentIter.nextSegment()
		_, lookaheadEof := nextSegmentIter.nextSegment()

		// POSIX allows the last newline to be treated as EOF,
		// so if the current segment is a newline and the next segment is EOF
		// then stop advancing the cursor.
		if eof || (segmentHasNewline(segment) && lookaheadEof) {
			break
		}

		if segmentHasNewline(segment) {
			newlineCount++
		}

		offset += uint64(len(segment))
	}

	return pos + offset, newlineCount
}

func (loc *relativeLineLocator) advanceToOffset(tree *text.Tree, lineStartPos uint64, targetOffset uint64) (newPos, actualOffset uint64) {
	segmentIter := newGraphemeClusterSegmentIter(tree, lineStartPos, text.ReadDirectionForward)

	var endOfLineOrFile bool
	var prevPosOffset, posOffset, gcOffset uint64
	for gcOffset < targetOffset {
		segment, eof := segmentIter.nextSegment()
		if eof {
			endOfLineOrFile = true
			break
		}

		if segmentHasNewline(segment) {
			endOfLineOrFile = true
			break
		}

		gcOffset++
		prevPosOffset = posOffset
		posOffset += uint64(len(segment))
	}

	if endOfLineOrFile {
		if gcOffset > 0 {
			gcOffset--
		}
		return lineStartPos + prevPosOffset, gcOffset
	}

	return lineStartPos + posOffset, gcOffset
}

func (loc *relativeLineLocator) String() string {
	return fmt.Sprintf("RelativeLineLocator(%s, %d)", directionString(loc.direction), loc.count)
}

func directionString(direction text.ReadDirection) string {
	switch direction {
	case text.ReadDirectionForward:
		return "forward"
	case text.ReadDirectionBackward:
		return "backward"
	default:
		panic("Unrecognized direction")
	}
}
