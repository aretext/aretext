package locate

import (
	"github.com/aretext/aretext/text"
	"github.com/aretext/aretext/text/segment"
)

// NextWordStart locates the start of the next word after the cursor.
// Word boundaries occur:
//  1) at the first non-whitespace after a whitespace
//  2) at the start of an empty line
//  3) between punctuation and non-punctuation
func NextWordStart(textTree *text.Tree, pos uint64, includeEndOfFile bool) uint64 {
	return nextWordBoundary(textTree, pos, includeEndOfFile, true, wordStartBoundary)
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
	return prevWordBoundary(textTree, pos, true, wordStartBoundary)
}

// NextWordEnd locates the next word-end boundary after the cursor.
// The word break rules are the same as for NextWordStart, except
// that empty lines are NOT treated as word boundaries.
func NextWordEnd(textTree *text.Tree, pos uint64) uint64 {
	return nextWordBoundary(textTree, pos, false, false, wordEndBoundary)
}

// CurrentWordStart locates the start of the word or whitespace under the cursor.
// Word boundaries are determined by whitespace and punctuation.
func CurrentWordStart(textTree *text.Tree, pos uint64) uint64 {
	if isWhitespaceAtPos(textTree, pos) {
		return prevWordBoundary(textTree, pos, false, whitespaceStartBoundary)
	} else {
		return prevWordBoundary(textTree, pos, false, currentWordStartBoundary)
	}
}

// CurrentWordEnd locates the end of the word or whitespace under the cursor.
// The returned position is one past the last character in the word or whitespace,
// so this can be used to delete all the characters in the word.
// Word boundaries are determined by whitespace and punctuation.
func CurrentWordEnd(textTree *text.Tree, pos uint64) uint64 {
	if isWhitespaceAtPos(textTree, pos) {
		return findStartOfWordAfterWhitespace(textTree, pos)
	} else {
		return nextWordBoundary(textTree, pos, true, true, currentWordEndBoundary)
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

// wordBoundaryFunc returns whether a word boundary occurs between two grapheme clusters.
// The segments s1 and s2 are given in the order they appear in the text, even when reading backwards.
type wordBoundaryFunc func(s1, s2 *segment.Segment) bool

// nextWordBoundary finds the next word boundary after the current position.
func nextWordBoundary(textTree *text.Tree, pos uint64, includeEndOfFile bool, includeBoundary bool, f wordBoundaryFunc) uint64 {
	var offset, prevOffset uint64
	segmentIter := segment.NewGraphemeClusterIterForTree(textTree, pos, text.ReadDirectionForward)
	seg, prevSeg := segment.Empty(), segment.Empty()
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

		if f(prevSeg, seg) {
			if includeBoundary && offset > 0 {
				// Stop at the position after the boundary.
				return pos + offset
			} else if prevOffset > 0 {
				// Stop at the position before the boundary.
				return pos + prevOffset
			}
		}

		prevOffset = offset
		offset += seg.NumRunes()
		seg, prevSeg = prevSeg, seg
	}
}

// prevWordBoundary finds the word boundary on or before the current position.
func prevWordBoundary(textTree *text.Tree, pos uint64, skipStartPos bool, f wordBoundaryFunc) uint64 {
	var offset uint64
	segmentIter := segment.NewGraphemeClusterIterForTree(textTree, pos, text.ReadDirectionBackward)
	seg, prevSeg := segment.Empty(), segment.Empty()
	for {
		eof := segment.NextOrEof(segmentIter, seg)
		if eof {
			// Start of the document.
			return 0
		}

		if f(seg, prevSeg) && !(skipStartPos && offset == 0) {
			return pos - offset
		}

		offset += seg.NumRunes()
		seg, prevSeg = prevSeg, seg
	}
}

func wordStartBoundary(s1, s2 *segment.Segment) bool {
	if s1.NumRunes() == 0 || s2.NumRunes() == 0 {
		return false
	}

	if s1.HasNewline() && s2.HasNewline() {
		return true
	}

	s1ws, s2ws := s1.IsWhitespace(), s2.IsWhitespace()
	if s1ws && !s2ws {
		return true
	}

	if !s1ws && !s2ws && isPunct(s1) != isPunct(s2) {
		return true
	}

	return false
}

func wordEndBoundary(s1, s2 *segment.Segment) bool {
	if s1.NumRunes() == 0 || s2.NumRunes() == 0 {
		return false
	}

	s1ws, s2ws := s1.IsWhitespace(), s2.IsWhitespace()

	if !s1ws && s2ws {
		return true
	}

	if !s1ws && !s2ws && isPunct(s1) != isPunct(s2) {
		return true
	}

	return false
}

func whitespaceStartBoundary(s1, s2 *segment.Segment) bool {
	return s1.HasNewline() || !s1.IsWhitespace()
}

func currentWordStartBoundary(s1, s2 *segment.Segment) bool {
	return s1.HasNewline() || s1.IsWhitespace() || isPunct(s1)
}

func currentWordEndBoundary(s1, s2 *segment.Segment) bool {
	return s2.HasNewline() || s2.IsWhitespace() || isPunct(s2)
}

// isPunct returns whether a grapheme cluster should be treated as punctuation for determining word boundaries.
func isPunct(seg *segment.Segment) bool {
	if seg.NumRunes() != 1 {
		return false
	}

	r := seg.Runes()[0]

	// These ranges are the same as the unicode punctuation class for ASCII characters, except that:
	// * underscores ('_') are NOT treated as punctuation
	// * the following chars ARE treated as punctuation: '$', '+', '<', '=', '>', '^', '`', '|', '~'
	return (r >= '!' && r <= '/') || (r >= ':' && r <= '@') || (r >= '[' && r <= '^') || (r == '`' || r >= '{' && r <= '~')
}

// isWhitespace returns whether the character at the position is whitespace.
func isWhitespaceAtPos(tree *text.Tree, pos uint64) bool {
	segmentIter := segment.NewGraphemeClusterIterForTree(tree, pos, text.ReadDirectionForward)
	seg := segment.Empty()
	eof := segment.NextOrEof(segmentIter, seg)
	return !eof && seg.IsWhitespace()
}
