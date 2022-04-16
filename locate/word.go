package locate

import (
	"io"

	"github.com/aretext/aretext/text"
	"github.com/aretext/aretext/text/segment"
)

// NextWordStart locates the start of the next word after the cursor.
// Word boundaries occur:
//  1) at the first non-whitespace after a whitespace
//  2) at the start of an empty line
//  3) between punctuation and non-punctuation
func NextWordStart(textTree *text.Tree, pos uint64) uint64 {
	return nextWordBoundary(textTree, pos, func(gcOffset uint64, s1, s2 *segment.Segment) wordBoundaryDecision {
		if s2.NumRunes() == 0 {
			// Stop before EOF.
			return boundaryBefore
		}

		if gcOffset == 0 {
			// Skip the first boundary so the cursor doesn't get stuck
			// at the start of the current word.
			return noBoundary
		}

		s1ws, s2ws := s1.IsWhitespace(), s2.IsWhitespace()
		if s1ws && !s2ws {
			// Stop on first non-whitespace after whitespace.
			return boundaryAfter
		}

		if s1.HasNewline() && s2.HasNewline() {
			// Stop on empty line.
			return boundaryAfter
		}

		if !s1ws && !s2ws && isPunct(s1) != isPunct(s2) {
			// Stop after punctuation -> non-punctuation
			// and non-punctuation -> punctuation.
			return boundaryAfter
		}

		return noBoundary
	})
}

// NextWordStartInLine locates the start of the next word or the end of the line, whichever comes first.
func NextWordStartInLine(textTree *text.Tree, pos uint64) uint64 {
	return nextWordBoundary(textTree, pos, func(gcOffset uint64, s1, s2 *segment.Segment) wordBoundaryDecision {
		if s2.NumRunes() == 0 {
			// Stop after EOF.
			return boundaryAfter
		}

		if s2.HasNewline() {
			// Stop at end of line.
			return boundaryAfter
		}

		if gcOffset == 0 {
			// Skip the first boundary so the cursor doesn't get stuck
			// at the start of the current word.
			return noBoundary
		}

		s1ws, s2ws := s1.IsWhitespace(), s2.IsWhitespace()
		if s1ws && !s2ws {
			// Stop on first non-whitespace after whitespace.
			return boundaryAfter
		}

		if !s1ws && !s2ws && isPunct(s1) != isPunct(s2) {
			// Stop after punctuation -> non-punctuation
			// and non-punctuation -> punctuation.
			return boundaryAfter
		}

		return noBoundary
	})
}

// NextWordStartInLineOrAfterEmptyLine is the same as NextWordStartInLine, except it includes
// the newline at the end of an empty line.
func NextWordStartInLineOrAfterEmptyLine(textTree *text.Tree, pos uint64) uint64 {
	nextPos := NextWordStartInLine(textTree, pos)

	// The cursor didn't move, so check if we're on an empty line.
	// If so, inlude the newline at the end of the empty line.
	if nextPos == pos {
		prevSeg := prevAdjacentGraphemeCluster(textTree, pos)
		nextSeg := nextAdjacentGraphemeCluster(textTree, pos)
		if prevSeg.HasNewline() && nextSeg.HasNewline() {
			nextPos = pos + nextSeg.NumRunes()
		}
	}

	return nextPos
}

// PrevWordStart locates the start of the word before the cursor.
// It uses the same word break rules as NextWordStart.
func PrevWordStart(textTree *text.Tree, pos uint64) uint64 {
	return prevWordBoundary(textTree, pos, func(gcOffset uint64, s1, s2 *segment.Segment) wordBoundaryDecision {
		if gcOffset == 0 {
			// Skip the first boundary so the cursor doesn't get stuck
			// at the start of the current word.
			return noBoundary
		}

		s1ws, s2ws := s1.IsWhitespace(), s2.IsWhitespace()
		if s1ws && !s2ws {
			// Stop on last non-whitespace before whitespace.
			return boundaryAfter
		}

		if s1.HasNewline() && s2.HasNewline() {
			// Stop on empty line.
			return boundaryAfter
		}

		if !s1ws && !s2ws && isPunct(s1) != isPunct(s2) {
			// Stop after punctuation -> non-punctuation
			// and non-punctuation -> punctuation.
			return boundaryAfter
		}

		return noBoundary
	})
}

// NextWordEnd locates the next word-end boundary after the cursor.
// The word break rules are the same as for NextWordStart, except
// that empty lines are NOT treated as word boundaries.
func NextWordEnd(textTree *text.Tree, pos uint64) uint64 {
	return nextWordBoundary(textTree, pos, func(gcOffset uint64, s1, s2 *segment.Segment) wordBoundaryDecision {
		if s2.NumRunes() == 0 {
			// Stop before EOF.
			return boundaryBefore
		}

		if gcOffset <= 1 {
			// Skip the first boundary so the cursor doesn't get stuck
			// at the end of the current word.
			return noBoundary
		}

		s1ws, s2ws := s1.IsWhitespace(), s2.IsWhitespace()
		if !s1ws && s2ws {
			// Stop on last non-whitespace before whitespace.
			return boundaryBefore
		}

		if !s1ws && !s2ws && isPunct(s1) != isPunct(s2) {
			// Stop after punctuation -> non-punctuation
			// and non-punctuation -> punctuation.
			return boundaryBefore
		}

		return noBoundary
	})
}

// CurrentWordStart locates the start of the word or whitespace under the cursor.
// Word boundaries are determined by whitespace and punctuation.
func CurrentWordStart(textTree *text.Tree, pos uint64) uint64 {
	return prevWordBoundary(textTree, pos, func(gcOffset uint64, s1, s2 *segment.Segment) wordBoundaryDecision {
		if s1.HasNewline() {
			// Stop at end of line.
			return boundaryAfter
		}

		s1ws, s2ws := s1.IsWhitespace(), s2.IsWhitespace()
		if s1ws != s2ws {
			// Stop at whitespace / non-whitespace boundary.
			return boundaryAfter
		}

		if !s1ws && !s2ws && isPunct(s1) != isPunct(s2) {
			// Stop at punctuation boundary.
			return boundaryAfter
		}

		return noBoundary
	})
}

// CurrentWordEnd locates the end of the word or whitespace under the cursor.
// The returned position is one past the last character in the word or whitespace,
// so this can be used to delete all the characters in the word.
// Word boundaries are determined by whitespace and punctuation.
func CurrentWordEnd(textTree *text.Tree, pos uint64) uint64 {
	return nextWordBoundary(textTree, pos, func(gcOffset uint64, s1, s2 *segment.Segment) wordBoundaryDecision {
		if s2.NumRunes() == 0 {
			// Stop after EOF.
			return boundaryAfter
		}

		if s2.HasNewline() {
			// Stop at newline.
			return boundaryAfter
		}

		if gcOffset == 0 {
			// Skip the first boundary check (character before and the initial cursor)
			// to avoid getting stuck at the current position.
			return noBoundary
		}

		s1ws, s2ws := s1.IsWhitespace(), s2.IsWhitespace()
		if s1ws != s2ws {
			// Stop after non-whitespace followed by whitespace.
			return boundaryAfter
		}

		if !s1ws && !s2ws && isPunct(s1) != isPunct(s2) {
			// Stop at punctuation boundary.
			return boundaryAfter
		}

		return noBoundary
	})
}

// CurrentWordEndWithTrailingWhitespace returns the end of the whitespace after current word.
func CurrentWordEndWithTrailingWhitespace(textTree *text.Tree, pos uint64) uint64 {
	// Find the end of the current word.
	endOfWordPos := CurrentWordEnd(textTree, pos)

	// Continue to the next non-whitespace or line boundary.
	return nextWordBoundary(textTree, endOfWordPos, func(gcOffset uint64, s1, s2 *segment.Segment) wordBoundaryDecision {
		if s2.NumRunes() == 0 {
			// Stop after EOF.
			return boundaryAfter
		}

		if s2.HasNewline() {
			// Stop at newline.
			return boundaryAfter
		}

		if !s2.IsWhitespace() {
			// Stop at first non-whitespace.
			return boundaryAfter
		}

		return noBoundary
	})
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

type wordBoundaryDecision int

const (
	// noBoundary means that there is NO boundary between two grapheme clusters.
	noBoundary = wordBoundaryDecision(iota)

	// boundaryBefore means that the cursor should be placed on the FIRST grapheme cluster (before the boundary).
	boundaryBefore

	// boundaryAfter means that the cursor should be placed on the SECOND grapheme cluster (after the boundary).
	boundaryAfter
)

// wordBoundaryFunc decides whether a boundary exists between two grapheme clusters.
// The grapheme clusters are passed in the order they appear in the text,
// regardless of the read direction.
type wordBoundaryFunc func(gcOffset uint64, s1, s2 *segment.Segment) wordBoundaryDecision

// nextWordBoundary finds a position based on the next word boundary after a given position.
// It starts with the boundary between the grapheme clusters before and after the given position.
func nextWordBoundary(textTree *text.Tree, pos uint64, f wordBoundaryFunc) uint64 {
	var offset, prevOffset, gcOffset uint64
	reader := textTree.ReaderAtPosition(pos)
	segmentIter := segment.NewGraphemeClusterIter(reader)
	seg := segment.Empty()
	prevSeg := prevAdjacentGraphemeCluster(textTree, pos)
	for {
		var eof bool
		err := segmentIter.NextSegment(seg)
		if err == io.EOF {
			// Let the wordBoundaryFunc decide whether to place the cursor
			// on or after the last position in the document.
			seg = segment.Empty()
			eof = true
		} else if err != nil {
			panic(err)
		}

		switch f(gcOffset, prevSeg, seg) {
		case noBoundary:
			if eof {
				// Can't go past the end of the text.
				return pos + offset
			}
			break
		case boundaryBefore:
			return pos + prevOffset
		case boundaryAfter:
			return pos + offset
		}

		prevOffset = offset
		offset += seg.NumRunes()
		gcOffset++
		seg, prevSeg = prevSeg, seg
	}
}

// prevWordBoundary finds a position based on the word boundary before a given position.
// It starts with the boundary between the grapheme clusters before and after the given position.
func prevWordBoundary(textTree *text.Tree, pos uint64, f wordBoundaryFunc) uint64 {
	var offset, gcOffset uint64
	reader := textTree.ReverseReaderAtPosition(pos)
	segmentIter := segment.NewReverseGraphemeClusterIter(reader)
	seg := segment.Empty()
	prevSeg := nextAdjacentGraphemeCluster(textTree, pos)
	for {
		err := segmentIter.NextSegment(seg)
		if err == io.EOF {
			// Start of document
			return 0
		} else if err != nil {
			panic(err)
		}

		// Pass the grapheme clusters in the order they appear in the text
		// (reverse of the order in which we read them).
		switch f(gcOffset, seg, prevSeg) {
		case noBoundary:
			break
		case boundaryBefore:
			return pos - offset - seg.NumRunes()
		case boundaryAfter:
			return pos - offset
		}

		offset += seg.NumRunes()
		gcOffset++
		seg, prevSeg = prevSeg, seg
	}
}

// nextAdjacentGraphemeCluster finds a grapheme cluster after a position.
// If at the end of the text, this returns an empty segment.
func nextAdjacentGraphemeCluster(tree *text.Tree, pos uint64) *segment.Segment {
	reader := tree.ReaderAtPosition(pos)
	segmentIter := segment.NewGraphemeClusterIter(reader)
	seg := segment.Empty()
	err := segmentIter.NextSegment(seg)
	if err == io.EOF {
		return segment.Empty()
	} else if err != nil {
		panic(err)
	}
	return seg
}

// prevAdjacentGraphemeCluster finds a grapheme cluster before a position.
// If at the start of the text, this returns an empty segment.
func prevAdjacentGraphemeCluster(tree *text.Tree, pos uint64) *segment.Segment {
	reader := tree.ReverseReaderAtPosition(pos)
	segmentIter := segment.NewReverseGraphemeClusterIter(reader)
	seg := segment.Empty()
	err := segmentIter.NextSegment(seg)
	if err == io.EOF {
		return segment.Empty()
	} else if err != nil {
		panic(err)
	}
	return seg
}
