package exec

import (
	"fmt"
	"log"

	"github.com/wedaly/aretext/internal/pkg/text"
	"github.com/wedaly/aretext/internal/pkg/text/segment"
)

// CursorLocator finds the position of the cursor according to some criteria.
type CursorLocator interface {
	fmt.Stringer

	// Locate finds the next position of the cursor based on the current state and criteria for this locator.
	Locate(state *BufferState) cursorState
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
func NewCharInLineLocator(direction text.ReadDirection, count uint64, includeEndOfLineOrFile bool) CursorLocator {
	if count == 0 {
		log.Fatalf("Count must be greater than zero")
	}
	return &charInLineLocator{direction, count, includeEndOfLineOrFile}
}

func (loc *charInLineLocator) String() string {
	return fmt.Sprintf("CharInLineLocator(%s, %d, %t)", directionString(loc.direction), loc.count, loc.includeEndOfLineOrFile)
}

// Locate finds a character to the right of the cursor on the current line.
func (loc *charInLineLocator) Locate(state *BufferState) cursorState {
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

func (loc *charInLineLocator) findPosition(state *BufferState) uint64 {
	if loc.direction == text.ReadDirectionBackward {
		return loc.findPositionBeforeCursor(state)
	}
	return loc.findPositionAfterCursor(state)
}

func (loc *charInLineLocator) findPositionBeforeCursor(state *BufferState) uint64 {
	startPos := state.cursor.position
	segmentIter := gcIterForTree(state.tree, startPos, text.ReadDirectionBackward)
	seg := segment.NewSegment()
	var offset uint64

	for i := uint64(0); i < loc.count; i++ {
		eof := nextSegmentOrEof(segmentIter, seg)
		if eof {
			break
		}

		if offset+seg.NumRunes() > startPos {
			return 0
		}

		if seg.HasNewline() {
			if loc.includeEndOfLineOrFile {
				offset += seg.NumRunes()
			}
			break
		}

		offset += seg.NumRunes()
	}

	return startPos - offset
}

func (loc *charInLineLocator) findPositionAfterCursor(state *BufferState) uint64 {
	startPos := state.cursor.position
	segmentIter := gcIterForTree(state.tree, startPos, text.ReadDirectionForward)
	seg := segment.NewSegment()
	var endOfLineOrFile bool
	var prevPrevOffset, prevOffset uint64

	for i := uint64(0); i <= loc.count; i++ {
		eof := nextSegmentOrEof(segmentIter, seg)
		if eof {
			endOfLineOrFile = true
			break
		}

		if seg.HasNewline() {
			endOfLineOrFile = true
			break
		}

		prevPrevOffset = prevOffset
		prevOffset += seg.NumRunes()
	}

	if endOfLineOrFile && loc.includeEndOfLineOrFile {
		return startPos + prevOffset
	}
	return startPos + prevPrevOffset
}

// ontoDocumentLocator finds a valid position within the document closest to the cursor.
type ontoDocumentLocator struct{}

func NewOntoDocumentLocator() CursorLocator {
	return &ontoDocumentLocator{}
}

// Locate finds the valid position within the document closest to the current cursor position.
// It handles POSIX end-of-file (line feed at end of document).
func (loc *ontoDocumentLocator) Locate(state *BufferState) cursorState {
	lastValidPos := state.tree.NumChars()
	if endsWithLineFeed(state.tree) {
		lastValidPos--
	}

	newPos := state.cursor.position
	if newPos > lastValidPos {
		newPos = lastValidPos
	}

	return cursorState{position: newPos}
}

func (loc *ontoDocumentLocator) String() string {
	return "OntoDocumentLocator()"
}

// ontoLineLocator finds the closest grapheme cluster on a line (not newline or past end of text).
// This is useful for "resetting" the cursor onto a line
// (for example, after deleting the last character on the line or exiting insert mode).
type ontoLineLocator struct{}

func NewOntoLineLocator() CursorLocator {
	return &ontoLineLocator{}
}

// Locate finds the closest grapheme cluster on a line (not newline or past end of text).
func (loc *ontoLineLocator) Locate(state *BufferState) cursorState {
	// If past the end of the text, return the start of the last grapheme cluster.
	numChars := state.tree.NumChars()
	if endsWithLineFeed(state.tree) {
		numChars--
	}

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
	segmentIter := gcIterForTree(tree, pos, text.ReadDirectionForward)
	seg := segment.NewSegment()
	eof := nextSegmentOrEof(segmentIter, seg)
	if eof {
		return false, 0
	}

	if seg.HasNewline() {
		return true, pos + seg.NumRunes()
	}

	return false, 0
}

func (loc *ontoLineLocator) findPrevGraphemeCluster(tree *text.Tree, pos uint64, count int) uint64 {
	segmentIter := gcIterForTree(tree, pos, text.ReadDirectionBackward)

	// Iterate backward by (count - 1) grapheme clusters.
	seg := segment.NewSegment()
	var offset uint64
	for i := 0; i < count-1; i++ {
		eof := nextSegmentOrEof(segmentIter, seg)
		if eof {
			break
		}

		offset += seg.NumRunes()
	}

	// Check the next grapheme cluster after (count - 1) grapheme clusters.
	eof := nextSegmentOrEof(segmentIter, seg)
	if eof {
		return 0
	}

	// If the immediately preceding cluster is a newline, then we're on
	// an empty line, in which case we shouldn't move the cursor.
	if seg.HasNewline() {
		return pos - offset
	}

	// Otherwise, move the cursor back a cluster to position it at the end of the previous line.
	return pos - offset - seg.NumRunes()
}

func (loc *ontoLineLocator) String() string {
	return "OntoLineLocator()"
}

// relativeLineStartLocator finds the start of a line above or below the cursor.
type relativeLineStartLocator struct {
	direction text.ReadDirection
	count     uint64
}

// NewRelativeLineStartLocator returns a locator that finds the start of a line above or below the cursor.
func NewRelativeLineStartLocator(direction text.ReadDirection, count uint64) CursorLocator {
	return &relativeLineStartLocator{direction, count}
}

func (loc *relativeLineStartLocator) Locate(state *BufferState) cursorState {
	newPos := loc.findStartOfLineAboveOrBelow(state.tree, state.cursor.position)
	return cursorState{position: newPos}
}

func (loc *relativeLineStartLocator) findStartOfLineAboveOrBelow(tree *text.Tree, pos uint64) uint64 {
	currentLineNum := tree.LineNumForPosition(pos)
	targetLineNum := loc.targetLineNum(currentLineNum)
	return tree.LineStartPosition(closestValidLineNum(tree, targetLineNum))
}

func (loc *relativeLineStartLocator) targetLineNum(currentLineNum uint64) uint64 {
	if loc.direction == text.ReadDirectionForward {
		return currentLineNum + loc.count
	}

	if currentLineNum < loc.count {
		return 0
	}

	return currentLineNum - loc.count
}

func (loc *relativeLineStartLocator) String() string {
	return fmt.Sprintf("RelativeLineStartLocator(%s, %d)", directionString(loc.direction), loc.count)
}

// relativeLineLocator finds a position at the same offset above or below the current line.
type relativeLineLocator struct {
	direction text.ReadDirection
	count     uint64
}

// NewRelativeLineLocator returns a locator for moving the cursor up or down by some number of lines.
// Count is the number of lines to move, and it must be at least one.
// Direction indicates whether to move up (ReadDirectionBackward) or down (ReadDirectionForward).
func NewRelativeLineLocator(direction text.ReadDirection, count uint64) CursorLocator {
	if count == 0 {
		log.Fatalf("Count must be greater than zero")
	}
	return &relativeLineLocator{direction, count}
}

// Locate returns a cursor position at the same offset above or below the current line.
// It does nothing when moving up from the first line or down from the last line.
// If the target line has fewer characters than the starting line, then the extra characters
// will be counted the cursor's logical offset.
// If the target line has more characters than the starting line, then the cursor will move
// as close as possible to the logical offset.
func (loc *relativeLineLocator) Locate(state *BufferState) cursorState {
	lineStartPos := loc.findLineStart(state.tree, state.cursor.position)
	targetLineStartPos := loc.findTargetLineStartPos(state)
	if targetLineStartPos == lineStartPos {
		return state.cursor
	}

	targetOffset := loc.findOffsetFromLineStart(state, lineStartPos)
	newPos, actualOffset := loc.advanceToOffset(state.tree, targetLineStartPos, targetOffset)
	return cursorState{
		position:      newPos,
		logicalOffset: targetOffset - actualOffset,
	}
}

func (loc *relativeLineLocator) findLineStart(tree *text.Tree, pos uint64) uint64 {
	lineNum := tree.LineNumForPosition(pos)
	return tree.LineStartPosition(lineNum)
}

func (loc *relativeLineLocator) findTargetLineStartPos(state *BufferState) uint64 {
	return NewRelativeLineStartLocator(loc.direction, loc.count).Locate(state).position
}

func (loc *relativeLineLocator) findStartOfLineAboveOrBelow(tree *text.Tree, pos uint64) uint64 {
	currentLineNum := tree.LineNumForPosition(pos)
	targetLineNum := loc.targetLineNum(currentLineNum)
	return tree.LineStartPosition(closestValidLineNum(tree, targetLineNum))
}

func (loc *relativeLineLocator) targetLineNum(currentLineNum uint64) uint64 {
	if loc.direction == text.ReadDirectionForward {
		return currentLineNum + loc.count
	}

	if currentLineNum < loc.count {
		return 0
	}

	return currentLineNum - loc.count
}

func (loc *relativeLineLocator) findOffsetFromLineStart(state *BufferState, lineStartPos uint64) uint64 {
	segmentIter := gcIterForTree(state.tree, lineStartPos, text.ReadDirectionForward)
	seg := segment.NewSegment()
	pos, offset := lineStartPos, uint64(0)

	for {
		eof := nextSegmentOrEof(segmentIter, seg)
		if eof || pos >= state.cursor.position {
			break
		}

		offset += GraphemeClusterWidth(seg.Runes(), offset)
		pos += seg.NumRunes()
	}

	return offset + state.cursor.logicalOffset
}

func (loc *relativeLineLocator) advanceToOffset(tree *text.Tree, lineStartPos uint64, targetOffset uint64) (newPos, actualOffset uint64) {
	segmentIter := gcIterForTree(tree, lineStartPos, text.ReadDirectionForward)
	seg := segment.NewSegment()
	var endOfLineOrFile bool
	var prevPosOffset, posOffset, cellOffset uint64

	for {
		eof := nextSegmentOrEof(segmentIter, seg)
		if eof {
			endOfLineOrFile = true
			break
		}

		if seg.HasNewline() {
			endOfLineOrFile = true
			break
		}

		gcWidth := GraphemeClusterWidth(seg.Runes(), cellOffset)
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

func (loc *relativeLineLocator) String() string {
	return fmt.Sprintf("RelativeLineLocator(%s, %d)", directionString(loc.direction), loc.count)
}

// lineBoundaryLocator locates the start or end of the current line.
type lineBoundaryLocator struct {
	direction              text.ReadDirection
	includeEndOfLineOrFile bool
}

// NewLineBoundaryLocator constructs a line boundary locator.
// Direction determines whether to locate the start (ReadDirectionBackward) or end (ReadDirectionForward) of the line.
// If includeEndOfLineOrFile is true, position the cursor at the newline or one past the last character in the text.
func NewLineBoundaryLocator(direction text.ReadDirection, includeEndOfLineOrFile bool) CursorLocator {
	return &lineBoundaryLocator{direction, includeEndOfLineOrFile}
}

func (loc *lineBoundaryLocator) String() string {
	return fmt.Sprintf("LineBoundaryLocator(%s, %t)", directionString(loc.direction), loc.includeEndOfLineOrFile)
}

// Locate the start or end of the current line.
// This assumes that the cursor is positioned on a line (not a newline character); if not, the result is undefined.
func (loc *lineBoundaryLocator) Locate(state *BufferState) cursorState {
	segmentIter := gcIterForTree(state.tree, state.cursor.position, loc.direction)
	seg := segment.NewSegment()
	var prevOffset, offset uint64

	for {
		eof := nextSegmentOrEof(segmentIter, seg)
		if eof || seg.HasNewline() {
			break
		}

		prevOffset = offset
		offset += seg.NumRunes()
	}

	newPosition := state.cursor.position
	if loc.direction == text.ReadDirectionForward {
		if loc.includeEndOfLineOrFile {
			newPosition += offset
		} else {
			newPosition += prevOffset
		}
	} else {
		newPosition -= offset
	}

	if newPosition == state.cursor.position {
		return state.cursor
	}

	return cursorState{position: newPosition}
}

// nonWhitespaceOrNewlineLocator finds the next non-whitespace character or newline on or after the cursor.
type nonWhitespaceOrNewlineLocator struct{}

func NewNonWhitespaceOrNewlineLocator() CursorLocator {
	return &nonWhitespaceOrNewlineLocator{}
}

func (loc *nonWhitespaceOrNewlineLocator) String() string {
	return "NonWhitespaceOrNewlineLocator()"
}

// Locate finds the nearest non-whitespace character or newline on or after the cursor.
func (loc *nonWhitespaceOrNewlineLocator) Locate(state *BufferState) cursorState {
	segmentIter := gcIterForTree(state.tree, state.cursor.position, text.ReadDirectionForward)
	seg := segment.NewSegment()
	var offset uint64

	for {
		eof := nextSegmentOrEof(segmentIter, seg)
		if eof || !seg.IsWhitespace() || seg.HasNewline() {
			break
		}

		offset += seg.NumRunes()
	}

	newPosition := state.cursor.position + offset
	if newPosition == state.cursor.position {
		return state.cursor
	}

	return cursorState{position: newPosition}
}

// lineNumLocator locates the start of a given line number.
type lineNumLocator struct {
	lineNum uint64
}

func NewLineNumLocator(lineNum uint64) CursorLocator {
	return &lineNumLocator{lineNum}
}

// Locate finds the start of the given line number.
func (loc *lineNumLocator) Locate(state *BufferState) cursorState {
	lineNum := closestValidLineNum(state.tree, loc.lineNum)
	pos := state.tree.LineStartPosition(lineNum)
	return cursorState{position: pos}
}

func (loc *lineNumLocator) String() string {
	return fmt.Sprintf("LineNumLocator(%d)", loc.lineNum)
}

// lastLineLocator finds the start of the last line.
type lastLineLocator struct{}

func NewLastLineLocator() CursorLocator {
	return &lastLineLocator{}
}

// locate returns the cursor position at the start of the last line.
func (loc *lastLineLocator) Locate(state *BufferState) cursorState {
	tree := state.tree
	lineNum := closestValidLineNum(tree, tree.NumLines())
	return cursorState{
		position: tree.LineStartPosition(lineNum),
	}
}

func (loc *lastLineLocator) String() string {
	return "LastLineLocator()"
}
