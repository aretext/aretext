package exec

import (
	"fmt"
	"io"

	"github.com/wedaly/aretext/internal/pkg/text"
	"github.com/wedaly/aretext/internal/pkg/text/segment"
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

func (loc *charInLineLocator) findPositionAfterCursor(state *State) uint64 {
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
	lineStartPos, newlineCount := loc.findStartOfLineAboveOrBelow(state.tree, state.cursor.position)
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
	lineStartPos := loc.findLineStart(state.tree, state.cursor.position)
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

func (loc *relativeLineLocator) findLineStart(tree *text.Tree, pos uint64) uint64 {
	reader := tree.ReaderAtPosition(pos, text.ReadDirectionBackward)
	runeIter := text.NewCloneableBackwardRuneIter(reader)
	for {
		r, err := runeIter.NextRune()
		if err == io.EOF {
			break
		} else if err != nil {
			panic(err)
		} else if r == '\n' {
			break
		}
		pos--
	}
	return pos
}

func (loc *relativeLineLocator) findStartOfLineAboveOrBelow(tree *text.Tree, pos uint64) (lineStartPos, newlineCount uint64) {
	if loc.direction == text.ReadDirectionBackward {
		return loc.findStartOfLineAbove(tree, pos)
	} else {
		return loc.findStartOfLineBelow(tree, pos)
	}
}

func (loc *relativeLineLocator) findStartOfLineAbove(tree *text.Tree, pos uint64) (lineStartPos, newlineCount uint64) {
	segmentIter := gcIterForTree(tree, pos, text.ReadDirectionBackward)
	seg := segment.NewSegment()
	var offset uint64

	for {
		eof := nextSegmentOrEof(segmentIter, seg)
		if eof {
			break
		}

		if seg.HasNewline() {
			newlineCount++
			if newlineCount > loc.count {
				break
			}
		}

		offset += seg.NumRunes()
	}

	return pos - offset, newlineCount
}

func (loc *relativeLineLocator) findStartOfLineBelow(tree *text.Tree, pos uint64) (lineStartPos, newlineCount uint64) {
	segmentIter := gcIterForTree(tree, pos, text.ReadDirectionForward)

	// Lookahead one grapheme cluster.
	seg, lookaheadSeg := segment.NewSegment(), segment.NewSegment()
	nextSegmentIter := segmentIter.Clone()
	nextSegmentIter.NextSegment(seg)

	var offset uint64
	for newlineCount < loc.count {
		eof := nextSegmentOrEof(segmentIter, seg)
		lookaheadEof := nextSegmentOrEof(nextSegmentIter, lookaheadSeg)

		// POSIX allows the last newline to be treated as EOF,
		// so if the current segment is a newline and the next segment is EOF
		// then stop advancing the cursor.
		if eof || (seg.HasNewline() && lookaheadEof) {
			break
		}

		if seg.HasNewline() {
			newlineCount++
		}

		offset += seg.NumRunes()
	}

	return pos + offset, newlineCount
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
func NewLineBoundaryLocator(direction text.ReadDirection, includeEndOfLineOrFile bool) Locator {
	return &lineBoundaryLocator{direction, includeEndOfLineOrFile}
}

func (loc *lineBoundaryLocator) String() string {
	return fmt.Sprintf("LineBoundaryLocator(%s, %t)", directionString(loc.direction), loc.includeEndOfLineOrFile)
}

// Locate the start or end of the current line.
// This assumes that the cursor is positioned on a line (not a newline character); if not, the result is undefined.
func (loc *lineBoundaryLocator) Locate(state *State) cursorState {
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

// nonWhitespaceLocator finds the nearest non-whitespace character in the specified direction.
type nonWhitespaceLocator struct {
	direction text.ReadDirection
}

func NewNonWhitespaceLocator(direction text.ReadDirection) Locator {
	return &nonWhitespaceLocator{direction}
}

func (loc *nonWhitespaceLocator) String() string {
	return fmt.Sprintf("NonWhitespaceLocator(%s)", directionString(loc.direction))
}

// Locate finds the nearest non-whitespace character in the specified direction.
func (loc *nonWhitespaceLocator) Locate(state *State) cursorState {
	segmentIter := gcIterForTree(state.tree, state.cursor.position, loc.direction)
	seg := segment.NewSegment()
	var offset uint64

	for {
		eof := nextSegmentOrEof(segmentIter, seg)
		if eof || !seg.IsWhitespace() {
			break
		}

		offset += seg.NumRunes()
	}

	newPosition := state.cursor.position
	if loc.direction == text.ReadDirectionForward {
		newPosition += offset
	} else {
		if offset > 0 {
			// When iterating backward, need to advance an additional segment
			// to position the cursor at the start of the non-whitespace character.
			eof := nextSegmentOrEof(segmentIter, seg)
			if !eof {
				offset += seg.NumRunes()
			}
		}

		newPosition -= offset
	}

	if newPosition == state.cursor.position {
		return state.cursor
	}

	return cursorState{position: newPosition}
}

// lineNumLocator locates the start of a given line number.
type lineNumLocator struct {
	lineNum uint64
}

func NewLineNumLocator(lineNum uint64) Locator {
	return &lineNumLocator{lineNum}
}

// Locate finds the start of the given line number.
func (loc *lineNumLocator) Locate(state *State) cursorState {
	lineNum := closestValidLineNum(state.tree, loc.lineNum)
	pos := state.tree.LineStartPosition(lineNum)
	return cursorState{position: pos}
}

func (loc *lineNumLocator) String() string {
	return fmt.Sprintf("LineNumLocator(%d)", loc.lineNum)
}

// lastLineLocator finds the start of the last line.
type lastLineLocator struct{}

func NewLastLineLocator() Locator {
	return &lastLineLocator{}
}

// locate returns the cursor position at the start of the last line.
func (loc *lastLineLocator) Locate(state *State) cursorState {
	tree := state.tree
	lineNum := closestValidLineNum(tree, tree.NumLines())
	return cursorState{
		position: tree.LineStartPosition(lineNum),
	}
}

func (loc *lastLineLocator) String() string {
	return "LastLineLocator()"
}

// gcIterForTree constructs a grapheme cluster iterator for the tree.
func gcIterForTree(tree *text.Tree, pos uint64, direction text.ReadDirection) segment.CloneableSegmentIter {
	reader := tree.ReaderAtPosition(pos, direction)
	if direction == text.ReadDirectionBackward {
		runeIter := text.NewCloneableBackwardRuneIter(reader)
		return segment.NewReverseGraphemeClusterIter(runeIter)
	} else {
		runeIter := text.NewCloneableForwardRuneIter(reader)
		return segment.NewGraphemeClusterIter(runeIter)
	}
}

// nextSegmentOrEof finds the next segment and returns a flag indicating end of file.
// If an error occurs (e.g. due to invalid UTF-8), it panics.
func nextSegmentOrEof(segmentIter segment.SegmentIter, seg *segment.Segment) (eof bool) {
	err := segmentIter.NextSegment(seg)
	if err == io.EOF {
		return true
	}

	if err != nil {
		panic(err)
	}

	return false
}

// closestValidLineNum finds the line number in the text that is closest to the target.
// It correctly interprets a line feed at the end of the file as a POSIX EOF marker, not a newline.
func closestValidLineNum(tree *text.Tree, targetLineNum uint64) uint64 {
	numLines := tree.NumLines()
	if numLines == 0 {
		return 0
	}

	lastRealLine := numLines - 1
	if endsWithLineFeed(tree) {
		// POSIX end-of-file marker is not considered the start of a new line.
		lastRealLine--
	}

	if targetLineNum > lastRealLine {
		return lastRealLine
	}
	return targetLineNum
}

// endsWithLineFeed returns whether the text ends with a line feed character.
func endsWithLineFeed(tree *text.Tree) bool {
	reader := tree.ReaderAtPosition(tree.NumChars(), text.ReadDirectionBackward)
	var buf [1]byte
	if n, err := reader.Read(buf[:]); err != nil || n == 0 {
		return false
	}
	return buf[0] == '\n'
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
