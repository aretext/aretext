package exec

import (
	"io"
	"log"

	"github.com/wedaly/aretext/internal/pkg/text"
	"github.com/wedaly/aretext/internal/pkg/text/segment"
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
		log.Fatalf("Unrecognized direction: %d", direction)
		return ""
	}
}
