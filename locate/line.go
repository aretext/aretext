package locate

import (
	"github.com/aretext/aretext/text"
	"github.com/aretext/aretext/text/segment"
)

// ClosestCharOnLine locates the closest grapheme cluster on a line (not newline or past end of text).
// This is useful for "resetting" the cursor onto a line
// (for example, after deleting the last character on the line or exiting insert mode).
func ClosestCharOnLine(tree *text.Tree, pos uint64) uint64 {
	// If past the end of the text, return the start of the last grapheme cluster.
	n := tree.NumChars()
	if pos >= n {
		return findPrevGraphemeCluster(tree, n, 1)
	}

	// If on a grapheme cluster with a newline (either "\n" or "\r\n"), return the start
	// of the last grapheme cluster before the current grapheme cluster.
	if hasNewline, afterNewlinePos := findNewlineAtPos(tree, pos); hasNewline {
		return findPrevGraphemeCluster(tree, afterNewlinePos, 2)
	}

	// The cursor is already on a line, so do nothing.
	return pos
}

func findNewlineAtPos(tree *text.Tree, pos uint64) (bool, uint64) {
	segmentIter := segment.NewGraphemeClusterIterForTree(tree, pos, text.ReadDirectionForward)
	seg := segment.Empty()
	eof := segment.NextOrEof(segmentIter, seg)
	if eof {
		return false, 0
	}

	if seg.HasNewline() {
		return true, pos + seg.NumRunes()
	}

	return false, 0
}

func findPrevGraphemeCluster(tree *text.Tree, pos uint64, count int) uint64 {
	segmentIter := segment.NewGraphemeClusterIterForTree(tree, pos, text.ReadDirectionBackward)

	// Iterate backward by (count - 1) grapheme clusters.
	seg := segment.Empty()
	var offset uint64
	for i := 0; i < count-1; i++ {
		eof := segment.NextOrEof(segmentIter, seg)
		if eof {
			break
		}

		offset += seg.NumRunes()
	}

	// Check the next grapheme cluster after (count - 1) grapheme clusters.
	eof := segment.NextOrEof(segmentIter, seg)
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

// ClosestValidLineNum returns the line number in the text that is closest to the target.
func ClosestValidLineNum(tree *text.Tree, targetLineNum uint64) uint64 {
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

// StartOfCurrentLine locates the start of the line for a given position.
func StartOfLineAtPos(tree *text.Tree, pos uint64) uint64 {
	lineNum := tree.LineNumForPosition(pos)
	return tree.LineStartPosition(lineNum)
}

// StartOfLineNum locates the start of a given line number.
func StartOfLineNum(tree *text.Tree, lineNum uint64) uint64 {
	return tree.LineStartPosition(ClosestValidLineNum(tree, lineNum))
}

// StartOfLastLine locates the start of the last line.
func StartOfLastLine(tree *text.Tree) uint64 {
	lineNum := ClosestValidLineNum(tree, tree.NumLines())
	return tree.LineStartPosition(lineNum)
}

// StartOfLineAbove locates the start of a line above the cursor.
func StartOfLineAbove(tree *text.Tree, count uint64, pos uint64) uint64 {
	currentLineNum := tree.LineNumForPosition(pos)
	targetLineNum := uint64(0)
	if currentLineNum >= count {
		targetLineNum = currentLineNum - count
	}
	return tree.LineStartPosition(ClosestValidLineNum(tree, targetLineNum))
}

// StartOfLineBelow locates the start of a line below the cursor.
func StartOfLineBelow(tree *text.Tree, count uint64, pos uint64) uint64 {
	currentLineNum := tree.LineNumForPosition(pos)
	targetLineNum := currentLineNum + count
	return tree.LineStartPosition(ClosestValidLineNum(tree, targetLineNum))
}

// NextLineBoundary locates the start of the current line.
// This assumes that the start position is on a line (not a newline character); if not, the result is undefined.
func NextLineBoundary(tree *text.Tree, includeEndOfLineOrFile bool, pos uint64) uint64 {
	segmentIter := segment.NewGraphemeClusterIterForTree(tree, pos, text.ReadDirectionForward)
	seg := segment.Empty()
	var prevOffset, offset uint64

	for {
		eof := segment.NextOrEof(segmentIter, seg)
		if eof || seg.HasNewline() {
			break
		}
		prevOffset = offset
		offset += seg.NumRunes()
	}

	if includeEndOfLineOrFile {
		return pos + offset
	} else {
		return pos + prevOffset
	}
}

// PrevLineBoundary locates the end of the current line.
// This assumes that the start position is on a line (not a newline character); if not, the result is undefined.
func PrevLineBoundary(tree *text.Tree, pos uint64) uint64 {
	segmentIter := segment.NewGraphemeClusterIterForTree(tree, pos, text.ReadDirectionBackward)
	seg := segment.Empty()
	var offset uint64
	for {
		eof := segment.NextOrEof(segmentIter, seg)
		if eof || seg.HasNewline() {
			break
		}
		offset += seg.NumRunes()
	}
	return pos - offset
}
