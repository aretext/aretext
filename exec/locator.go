package exec

import (
	"fmt"
	"log"
	"strings"

	"github.com/aretext/aretext/syntax/parser"
	"github.com/aretext/aretext/text"
	"github.com/aretext/aretext/text/segment"
)

// CursorLocator finds the position of the cursor according to some criteria.
type CursorLocator interface {
	fmt.Stringer

	// Locate finds the next position of the cursor based on the current state and criteria for this locator.
	Locate(state *BufferState) cursorState
}

// currentCursorLocator locates the current cursor position.
type currentCursorLocator struct{}

func NewCurrentCursorLocator() CursorLocator {
	return &currentCursorLocator{}
}

func (loc *currentCursorLocator) Locate(state *BufferState) cursorState {
	return state.cursor
}

func (loc *currentCursorLocator) String() string {
	return "CurrentCursorLocator()"
}

// mnPosLocator returns the cursor with the smallest position.
type minPosLocator struct {
	childLocators []CursorLocator
}

func NewMinPosLocator(childLocators []CursorLocator) CursorLocator {
	return &minPosLocator{childLocators}
}

func (loc *minPosLocator) Locate(state *BufferState) cursorState {
	min := state.cursor
	for i, child := range loc.childLocators {
		c := child.Locate(state)
		if i == 0 || c.position < min.position {
			min = c
		}
	}
	return min
}

func (loc *minPosLocator) String() string {
	childStrings := make([]string, 0, len(loc.childLocators))
	for _, child := range loc.childLocators {
		childStrings = append(childStrings, child.String())
	}

	return fmt.Sprintf("MinPosLocator(%s)", strings.Join(childStrings, ","))
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
	return fmt.Sprintf("CharInLineLocator(%s, %d, %t)", loc.direction, loc.count, loc.includeEndOfLineOrFile)
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
	segmentIter := gcIterForTree(state.textTree, startPos, text.ReadDirectionBackward)
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
	segmentIter := gcIterForTree(state.textTree, startPos, text.ReadDirectionForward)
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

// prevCharLocator returns the position before the cursor, which may be on a previous line.
type prevCharLocator struct {
	count uint64
}

func NewPrevCharLocator(count uint64) CursorLocator {
	return &prevCharLocator{count}
}

func (loc *prevCharLocator) Locate(state *BufferState) cursorState {
	pos := state.cursor.position
	iter := gcIterForTree(state.textTree, pos, text.ReadDirectionBackward)
	seg := segment.NewSegment()
	for i := uint64(0); i < loc.count; i++ {
		eof := nextSegmentOrEof(iter, seg)
		if eof {
			break
		}
		pos -= seg.NumRunes()
	}
	return cursorState{position: pos}
}

func (loc *prevCharLocator) String() string {
	return fmt.Sprintf("PrevCharLocator(%d)", loc.count)
}

// prevAutoIndentLocator returns the location of the previous tab stop if autoIndent is enabled.
// It returns the current cursor position if autoIndent is disabled or the characters before the cursor are not spaces/tabs.
type prevAutoIndentLocator struct{}

func NewPrevAutoIndentLocator() CursorLocator {
	return &prevAutoIndentLocator{}
}

func (loc *prevAutoIndentLocator) Locate(state *BufferState) cursorState {
	if !state.autoIndent {
		return state.cursor
	}

	prevTabAlignedPos := loc.findPrevTabAlignedPos(state)
	prevWhitespaceStartPos := loc.findPrevWhitespaceStartPos(state)
	if prevTabAlignedPos < prevWhitespaceStartPos {
		return cursorState{position: prevWhitespaceStartPos}
	}
	return cursorState{position: prevTabAlignedPos}
}

func (loc *prevAutoIndentLocator) findPrevTabAlignedPos(state *BufferState) uint64 {
	tabSize := state.TabSize()
	pos := lineStartPos(state.textTree, state.cursor.position)
	iter := gcIterForTree(state.textTree, pos, text.ReadDirectionForward)
	seg := segment.NewSegment()

	var offset uint64
	lastAlignedPos := pos
	for pos < state.cursor.position {
		if offset%tabSize == 0 {
			lastAlignedPos = pos
		}

		eof := nextSegmentOrEof(iter, seg)
		if eof {
			break
		}

		offset += GraphemeClusterWidth(seg.Runes(), offset, tabSize)
		pos += seg.NumRunes()
	}

	return lastAlignedPos
}

func (loc *prevAutoIndentLocator) findPrevWhitespaceStartPos(state *BufferState) uint64 {
	pos := state.cursor.position
	iter := gcIterForTree(state.textTree, pos, text.ReadDirectionBackward)
	seg := segment.NewSegment()
	for {
		eof := nextSegmentOrEof(iter, seg)
		if eof {
			break
		}

		r := seg.Runes()[0]
		if r != ' ' && r != '\t' {
			break
		}
		pos -= seg.NumRunes()
	}
	return pos
}

func (loc *prevAutoIndentLocator) String() string {
	return "PrevAutoIndentLocator()"
}

// ontoDocumentLocator finds a valid position within the document closest to the cursor.
type ontoDocumentLocator struct{}

func NewOntoDocumentLocator() CursorLocator {
	return &ontoDocumentLocator{}
}

// Locate finds the valid position within the document closest to the current cursor position.
func (loc *ontoDocumentLocator) Locate(state *BufferState) cursorState {
	n := state.textTree.NumChars()
	if n == 0 {
		return cursorState{position: 0}
	}

	lastValidPos := n - 1
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
	numChars := state.textTree.NumChars()
	if state.cursor.position >= numChars {
		newPos := loc.findPrevGraphemeCluster(state.textTree, numChars, 1)
		return cursorState{position: newPos}
	}

	// If on a grapheme cluster with a newline (either "\n" or "\r\n"), return the start
	// of the last grapheme cluster before the current grapheme cluster.
	if hasNewline, afterNewlinePos := loc.findNewlineAtPos(state.textTree, state.cursor.position); hasNewline {
		newPos := loc.findPrevGraphemeCluster(state.textTree, afterNewlinePos, 2)
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
	newPos := loc.findStartOfLineAboveOrBelow(state.textTree, state.cursor.position)
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
	return fmt.Sprintf("RelativeLineStartLocator(%s, %d)", loc.direction, loc.count)
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
	lineStartPos := lineStartPos(state.textTree, state.cursor.position)
	targetLineStartPos := loc.findTargetLineStartPos(state)
	if targetLineStartPos == lineStartPos {
		return state.cursor
	}

	targetOffset := loc.findOffsetFromLineStart(state, lineStartPos)
	newPos, actualOffset := loc.advanceToOffset(state.textTree, targetLineStartPos, targetOffset, state.TabSize())
	return cursorState{
		position:      newPos,
		logicalOffset: targetOffset - actualOffset,
	}
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
	segmentIter := gcIterForTree(state.textTree, lineStartPos, text.ReadDirectionForward)
	seg := segment.NewSegment()
	pos, offset := lineStartPos, uint64(0)

	for {
		eof := nextSegmentOrEof(segmentIter, seg)
		if eof || pos >= state.cursor.position {
			break
		}

		offset += GraphemeClusterWidth(seg.Runes(), offset, state.TabSize())
		pos += seg.NumRunes()
	}

	return offset + state.cursor.logicalOffset
}

func (loc *relativeLineLocator) advanceToOffset(tree *text.Tree, lineStartPos uint64, targetOffset uint64, tabSize uint64) (newPos, actualOffset uint64) {
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

		gcWidth := GraphemeClusterWidth(seg.Runes(), cellOffset, tabSize)
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
	return fmt.Sprintf("RelativeLineLocator(%s, %d)", loc.direction, loc.count)
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
	return fmt.Sprintf("LineBoundaryLocator(%s, %t)", loc.direction, loc.includeEndOfLineOrFile)
}

// Locate the start or end of the current line.
// This assumes that the cursor is positioned on a line (not a newline character); if not, the result is undefined.
func (loc *lineBoundaryLocator) Locate(state *BufferState) cursorState {
	segmentIter := gcIterForTree(state.textTree, state.cursor.position, loc.direction)
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

// nonWhitespaceOrNewlineLocator finds the next non-whitespace character or newline on or after a locator position.
type nonWhitespaceOrNewlineLocator struct {
	childLocator CursorLocator
}

func NewNonWhitespaceOrNewlineLocator(childLocator CursorLocator) CursorLocator {
	return &nonWhitespaceOrNewlineLocator{childLocator}
}

func (loc *nonWhitespaceOrNewlineLocator) String() string {
	return fmt.Sprintf("NonWhitespaceOrNewlineLocator(%s)", loc.childLocator)
}

// Locate finds the nearest non-whitespace character or newline on or after the cursor.
func (loc *nonWhitespaceOrNewlineLocator) Locate(state *BufferState) cursorState {
	startPos := loc.childLocator.Locate(state).position
	segmentIter := gcIterForTree(state.textTree, startPos, text.ReadDirectionForward)
	seg := segment.NewSegment()
	var offset uint64

	for {
		eof := nextSegmentOrEof(segmentIter, seg)
		if eof || !seg.IsWhitespace() || seg.HasNewline() {
			break
		}

		offset += seg.NumRunes()
	}

	newPosition := startPos + offset
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
	lineNum := closestValidLineNum(state.textTree, loc.lineNum)
	pos := state.textTree.LineStartPosition(lineNum)
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
	tree := state.textTree
	lineNum := closestValidLineNum(tree, tree.NumLines())
	return cursorState{
		position: tree.LineStartPosition(lineNum),
	}
}

func (loc *lastLineLocator) String() string {
	return "LastLineLocator()"
}

// nextWordStartLocator finds the start of the next word after the cursor.
// Word boundaries occur:
//  1) at the first non-whitespace after a whitespace
//  2) at the start of an empty line
//  3) at the start of a non-empty syntax token
type nextWordStartLocator struct{}

func NewNextWordStartLocator() CursorLocator {
	return &nextWordStartLocator{}
}

func (loc *nextWordStartLocator) Locate(state *BufferState) cursorState {
	pos := loc.findBasedOnWhitespace(state)
	if state.tokenTree != nil {
		nextTokenPos := loc.findBasedOnSyntaxTokens(state)
		if nextTokenPos < pos {
			pos = nextTokenPos
		}
	}
	return cursorState{position: pos}
}

func (loc *nextWordStartLocator) findBasedOnWhitespace(state *BufferState) uint64 {
	startPos := state.cursor.position
	segmentIter := gcIterForTree(state.textTree, startPos, text.ReadDirectionForward)
	seg := segment.NewSegment()
	var whitespaceFlag, newlineFlag bool
	var offset, prevOffset uint64
	for {
		eof := nextSegmentOrEof(segmentIter, seg)
		if eof {
			// Stop on (not after) the last character in the document.
			return startPos + prevOffset
		}

		if seg.HasNewline() {
			if newlineFlag {
				// An empty line is a word boundary.
				break
			}
			newlineFlag = true
		}

		if seg.IsWhitespace() {
			whitespaceFlag = true
		} else if whitespaceFlag {
			// Non-whitespace after whitespace is a word boundary.
			break
		}

		prevOffset = offset
		offset += seg.NumRunes()
	}
	return startPos + offset
}

func (loc *nextWordStartLocator) findBasedOnSyntaxTokens(state *BufferState) uint64 {
	startPos := state.cursor.position
	pos := startPos
	iter := state.tokenTree.IterFromPosition(pos, parser.IterDirectionForward)
	var tok parser.Token
	for iter.Get(&tok) {
		pos = tok.StartPos
		if tok.Role != parser.TokenRoleNone && pos > startPos {
			break
		}
		pos = tok.EndPos
		iter.Advance()
	}
	return pos
}

func (loc *nextWordStartLocator) String() string {
	return "NextWordStart()"
}

// prevWordStartLocator finds the start of the word before the cursor.
// It uses the same word break rules as nextWordStartLocator.
type prevWordStartLocator struct{}

func NewPrevWordStartLocator() CursorLocator {
	return &prevWordStartLocator{}
}

func (loc *prevWordStartLocator) Locate(state *BufferState) cursorState {
	pos := loc.findBasedOnWhitespace(state)
	if state.tokenTree != nil {
		prevTokenPos := loc.findBasedOnSyntaxTokens(state)
		if prevTokenPos > pos {
			pos = prevTokenPos
		}
	}
	return cursorState{position: pos}
}

func (loc *prevWordStartLocator) findBasedOnWhitespace(state *BufferState) uint64 {
	startPos := state.cursor.position
	segmentIter := gcIterForTree(state.textTree, startPos, text.ReadDirectionBackward)
	seg := segment.NewSegment()
	var nonwhitespaceFlag, newlineFlag bool
	var offset uint64
	for {
		eof := nextSegmentOrEof(segmentIter, seg)
		if eof {
			// Start of the document.
			return 0
		}

		if seg.HasNewline() {
			if newlineFlag {
				// An empty line is a word boundary.
				return startPos - offset
			}
			newlineFlag = true
		}

		if seg.IsWhitespace() {
			if nonwhitespaceFlag {
				// A whitespace after a nonwhitespace is a word boundary.
				return startPos - offset
			}
		} else {
			nonwhitespaceFlag = true
		}

		offset += seg.NumRunes()
	}
}

func (loc *prevWordStartLocator) findBasedOnSyntaxTokens(state *BufferState) uint64 {
	startPos := state.cursor.position
	pos := startPos
	iter := state.tokenTree.IterFromPosition(pos, parser.IterDirectionBackward)
	var tok parser.Token
	for iter.Get(&tok) {
		pos = tok.StartPos
		if tok.Role != parser.TokenRoleNone && pos < startPos {
			break
		}
		iter.Advance()
	}
	return pos
}

func (loc *prevWordStartLocator) String() string {
	return "PrevWordStart()"
}
