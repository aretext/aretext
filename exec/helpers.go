package exec

import (
	"io"
	"log"

	"github.com/aretext/aretext/text"
	"github.com/aretext/aretext/text/segment"
)

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
// If an error occurs (e.g. due to invalid UTF-8), it exits with an error.
func nextSegmentOrEof(segmentIter segment.SegmentIter, seg *segment.Segment) (eof bool) {
	err := segmentIter.NextSegment(seg)
	if err == io.EOF {
		return true
	}

	if err != nil {
		log.Fatalf("%s", err)
	}

	return false
}

// closestValidLineNum finds the line number in the text that is closest to the target.
func closestValidLineNum(tree *text.Tree, targetLineNum uint64) uint64 {
	numLines := tree.NumLines()
	if numLines == 0 {
		return 0
	}

	lastRealLine := numLines - 1
	if targetLineNum > lastRealLine {
		return lastRealLine
	}
	return targetLineNum
}

// lineStartPos returns the position at the start of the current line.
func lineStartPos(tree *text.Tree, cursorPos uint64) uint64 {
	lineNum := tree.LineNumForPosition(cursorPos)
	return tree.LineStartPosition(lineNum)
}
