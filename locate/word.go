package locate

import (
	"github.com/aretext/aretext/text"
	"github.com/aretext/aretext/text/segment"
)

// NextWordStart locates the start of the next word after the cursor.
// Word boundaries occur:
//  1) at the first non-whitespace after a whitespace
//  2) at the start of an empty line
func NextWordStart(textTree *text.Tree, pos uint64, includeEndOfFile bool) uint64 {
	segmentIter := segment.NewGraphemeClusterIterForTree(textTree, pos, text.ReadDirectionForward)
	seg := segment.Empty()
	var whitespaceFlag, newlineFlag bool
	var offset, prevOffset uint64
	for {
		eof := segment.NextOrEof(segmentIter, seg)
		if eof {
			if includeEndOfFile {
				// Stop after the last character in the document.
				return pos + offset
			} else {
				// Stop on (not after) the last character in the document.
				return pos + prevOffset
			}
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
	return pos + offset
}

// NextWordStartInLine locates the start of the next word or the end of the line, whichever comes first.
func NextWordStartInLine(textTree *text.Tree, pos uint64) uint64 {
	nextWordStart := NextWordStart(textTree, pos, true)
	endOfLine := NextLineBoundary(textTree, true, pos)
	if nextWordStart < endOfLine {
		return nextWordStart
	} else {
		return endOfLine
	}
}

// PrevWordStart locates the start of the word before the cursor.
// It uses the same word break rules as NextWordStart.
func PrevWordStart(textTree *text.Tree, pos uint64) uint64 {
	segmentIter := segment.NewGraphemeClusterIterForTree(textTree, pos, text.ReadDirectionBackward)
	seg := segment.Empty()
	var nonwhitespaceFlag, newlineFlag bool
	var offset uint64
	for {
		eof := segment.NextOrEof(segmentIter, seg)
		if eof {
			// Start of the document.
			return 0
		}
		if seg.HasNewline() {
			if newlineFlag {
				// An empty line is a word boundary.
				return pos - offset
			}
			newlineFlag = true
		}
		if seg.IsWhitespace() {
			if nonwhitespaceFlag {
				// A whitespace after a nonwhitespace is a word boundary.
				return pos - offset
			}
		} else {
			nonwhitespaceFlag = true
		}
		offset += seg.NumRunes()
	}
}

// NextWordEnd locates the next word-end boundary after the cursor.
// The word break rules are the same as for NextWordStart, except
// that empty lines are NOT treated as word boundaries.
func NextWordEnd(textTree *text.Tree, pos uint64) uint64 {
	segmentIter := segment.NewGraphemeClusterIterForTree(textTree, pos, text.ReadDirectionForward)
	seg := segment.Empty()
	var prevWasNonwhitespace bool
	var offset, prevOffset uint64
	for {
		eof := segment.NextOrEof(segmentIter, seg)
		if eof {
			// Stop on (not after) the last character in the document.
			return pos + prevOffset
		}

		if seg.IsWhitespace() {
			if prevWasNonwhitespace && offset > 1 {
				// Nonwhitespace followed by whitespace should stop at the nonwhitespace.
				return pos + prevOffset
			}
			prevWasNonwhitespace = false
		} else {
			prevWasNonwhitespace = true
		}

		prevOffset = offset
		offset += seg.NumRunes()
	}
}

// CurrentWordStart locates the start of the word or whitespace under the cursor.
// Word boundaries are determined by whitespace.
func CurrentWordStart(textTree *text.Tree, pos uint64) uint64 {
	if isWhitespaceAtPos(textTree, pos) {
		return findEndOfWordBeforeWhitespace(textTree, pos)
	} else {
		return findStartOfCurrentWord(textTree, pos)
	}
}

func findEndOfWordBeforeWhitespace(textTree *text.Tree, pos uint64) uint64 {
	segmentIter := segment.NewGraphemeClusterIterForTree(textTree, pos, text.ReadDirectionBackward)
	seg := segment.Empty()
	var offset uint64
	for {
		eof := segment.NextOrEof(segmentIter, seg)
		if eof {
			return 0
		} else if seg.HasNewline() || !seg.IsWhitespace() {
			return pos - offset
		}
		offset += seg.NumRunes()
	}
}

func findStartOfCurrentWord(textTree *text.Tree, pos uint64) uint64 {
	segmentIter := segment.NewGraphemeClusterIterForTree(textTree, pos, text.ReadDirectionBackward)
	seg := segment.Empty()
	var offset uint64
	for {
		eof := segment.NextOrEof(segmentIter, seg)
		if eof {
			return 0
		} else if seg.HasNewline() || seg.IsWhitespace() {
			return pos - offset
		}
		offset += seg.NumRunes()
	}
}

// CurrentWordEnd locates the end of the word or whitespace under the cursor.
// The returned position is one past the last character in the word or whitespace,
// so this can be used to delete all the characters in the word.
// Word boundaries are determined by whitespace.
func CurrentWordEnd(textTree *text.Tree, pos uint64) uint64 {
	if isWhitespaceAtPos(textTree, pos) {
		return findStartOfWordAfterWhitespace(textTree, pos)
	} else {
		return findEndOfCurrentWord(textTree, pos)
	}
}

func findStartOfWordAfterWhitespace(textTree *text.Tree, pos uint64) uint64 {
	nextWordStartPos := NextWordStart(textTree, pos, false)
	nextLineBoundaryPos := NextLineBoundary(textTree, true, pos)
	if nextWordStartPos < nextLineBoundaryPos {
		return nextWordStartPos
	} else {
		return nextLineBoundaryPos
	}
}

func findEndOfCurrentWord(textTree *text.Tree, pos uint64) uint64 {
	segmentIter := segment.NewGraphemeClusterIterForTree(textTree, pos, text.ReadDirectionForward)
	seg := segment.Empty()
	var offset uint64
	for {
		eof := segment.NextOrEof(segmentIter, seg)
		if eof {
			break
		} else if seg.HasNewline() || seg.IsWhitespace() {
			break
		}
		offset += seg.NumRunes()
	}
	return pos + offset
}

// CurrentWordEndWithTrailingWhitespace returns the end of the whitespace after current word.
// It uses the same word break rules as NextWordStart.
func CurrentWordEndWithTrailingWhitespace(textTree *text.Tree, pos uint64) uint64 {
	nextWordPos := NextWordStart(textTree, pos, false)
	endOfLinePos := NextLineBoundary(textTree, true, pos)
	onLastWordInDocument := bool(nextWordPos+1 == textTree.NumChars())
	if onLastWordInDocument || endOfLinePos < nextWordPos {
		return endOfLinePos
	} else {
		return nextWordPos
	}
}

// isWhitespace returns whether the character at the position is whitespace.
func isWhitespaceAtPos(tree *text.Tree, pos uint64) bool {
	segmentIter := segment.NewGraphemeClusterIterForTree(tree, pos, text.ReadDirectionForward)
	seg := segment.Empty()
	eof := segment.NextOrEof(segmentIter, seg)
	return !eof && seg.IsWhitespace()
}
