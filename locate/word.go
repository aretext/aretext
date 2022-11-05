package locate

import (
	"unicode"

	"github.com/aretext/aretext/text"
	"github.com/aretext/aretext/text/segment"
)

// NextWordStart locates the start of the next word after the cursor.
// Word boundaries occur:
//  1. at the first non-whitespace after a whitespace
//  2. at the start of an empty line
//  3. between punctuation and non-punctuation (unless withPunctuation=true)
func NextWordStart(textTree *text.Tree, pos uint64, targetCount uint64, withPunctuation, stopAtEndOfLastLine bool) uint64 {
	if targetCount == 0 {
		return pos
	}

	reader := textTree.ReaderAtPosition(pos)
	gcIter := segment.NewGraphemeClusterIter(reader)
	gc := segment.Empty()

	// Read the first gc to check if we're on
	// a newline, whitespace, or punct.
	err := gcIter.NextSegment(gc)
	if err != nil {
		return pos
	}
	prevHasNewline := gc.HasNewline()
	prevWasWhitespace := gc.IsWhitespace()
	prevWasPunct := isPunct(gc)

	if stopAtEndOfLastLine && targetCount == 1 && prevHasNewline {
		return pos
	}

	pos += gc.NumRunes()

	// Read subsequent runes to find the next word boundary.
	var count uint64
	for {
		err = gcIter.NextSegment(gc)
		if err != nil {
			break
		}

		isWhitespace := gc.IsWhitespace()
		hasNewline := gc.HasNewline()
		isPunct := isPunct(gc)

		if (prevWasWhitespace && !isWhitespace) ||
			(!withPunctuation && prevWasPunct && !isPunct && !isWhitespace) ||
			(!withPunctuation && !prevWasPunct && isPunct) ||
			(prevHasNewline && hasNewline) {
			count++
		}

		if stopAtEndOfLastLine && count+1 == targetCount && hasNewline {
			break
		}

		if count == targetCount {
			break
		}

		pos += gc.NumRunes()
		prevHasNewline = hasNewline
		prevWasWhitespace = isWhitespace
		prevWasPunct = isPunct
	}

	return pos
}

// PrevWordStart locates the start of the word before the cursor.
// It is the inverse of NextWordStart.
func PrevWordStart(textTree *text.Tree, pos uint64, targetCount uint64, withPunctuation bool) uint64 {
	if targetCount == 0 {
		return pos
	}

	reader := textTree.ReverseReaderAtPosition(pos)
	gcIter := segment.NewReverseGraphemeClusterIter(reader)
	gc := segment.Empty()

	// Read the gc before pos to check if we're on
	// a newline, whitespace, or punct.
	err := gcIter.NextSegment(gc)
	if err != nil {
		return 0 // io.EOF means we're at the start of the document.
	}
	prevHasNewline := gc.HasNewline()
	prevWasWhitespace := gc.IsWhitespace()
	prevWasPunct := isPunct(gc)
	pos -= gc.NumRunes()

	// Read backwards until we find a boundary.
	var count uint64
	for {
		err = gcIter.NextSegment(gc)
		if err != nil {
			return 0 // io.EOF means we're at the start of the document.
		}

		isWhitespace := gc.IsWhitespace()
		hasNewline := gc.HasNewline()
		isPunct := isPunct(gc)

		if (isWhitespace && !prevWasWhitespace) ||
			(!withPunctuation && isPunct && !prevWasPunct && !prevWasWhitespace) ||
			(!withPunctuation && !isPunct && prevWasPunct) ||
			(hasNewline && prevHasNewline) {
			count++
		}

		if count == targetCount {
			break
		}

		pos -= gc.NumRunes()
		prevHasNewline = hasNewline
		prevWasWhitespace = isWhitespace
		prevWasPunct = isPunct
	}

	return pos
}

// NextWordEnd locates the next word-end boundary after the cursor.
// The word break rules are the same as for NextWordStart, except
// that empty lines are NOT treated as word boundaries.
func NextWordEnd(textTree *text.Tree, pos uint64, targetCount uint64, withPunctuation bool) uint64 {
	if targetCount == 0 {
		return pos
	}

	reader := textTree.ReaderAtPosition(pos)
	gcIter := segment.NewGraphemeClusterIter(reader)
	gc := segment.Empty()

	// Discard the first gc.
	// This ensures that we advance even if we start
	// at the end of a word.
	err := gcIter.NextSegment(gc)
	if err != nil {
		return pos
	}
	prevPos := pos
	pos += gc.NumRunes()

	// Read the second gc to check if we're on
	// a newline, whitespace, or punct.
	err = gcIter.NextSegment(gc)
	if err != nil {
		return prevPos
	}
	prevWasWhitespace := gc.IsWhitespace()
	prevWasPunct := isPunct(gc)
	prevPos = pos
	pos += gc.NumRunes()

	// Read subsequent runes to find the next word boundary.
	var count uint64
	for {
		err = gcIter.NextSegment(gc)
		if err != nil {
			break
		}

		isWhitespace := gc.IsWhitespace()
		isPunct := isPunct(gc)

		if (!prevWasWhitespace && isWhitespace) ||
			(!withPunctuation && prevWasPunct != isPunct) {
			count++
		}

		if count == targetCount {
			break
		}

		prevPos = pos
		pos += gc.NumRunes()
		prevWasWhitespace = isWhitespace
		prevWasPunct = isPunct
	}

	// Return the previous position to ensure that we stop on,
	// not after, the end of word.
	return prevPos
}

// WordObject returns the start and end positions of the word object under the cursor.
// If the cursor is on whitespace, include it as leading whitespace.
// Otherwise, include trailing whitespace.
// This is equivalent to vim's "aw" ("a word") object.
func WordObject(textTree *text.Tree, pos uint64, targetCount uint64) (uint64, uint64) {
	if targetCount == 0 {
		return pos, pos
	}

	// Lookahead one rune to detect whether we're in whitespace or not.
	reader := textTree.ReaderAtPosition(pos)
	r, _, err := reader.ReadRune()
	if err != nil {
		// This can only occur in an empty document.
		return pos, pos
	}

	if unicode.IsSpace(r) {
		// If we're in whitespace, treat it as leading whitespace
		// and move to the following word.
		return wordObjectWithLeadingWhitespace(textTree, pos, targetCount)
	} else {
		// Otherwise, move past the end of the word and
		// any trailing whitespace.
		return wordObjectWithTrailingWhitespace(textTree, pos, targetCount)
	}
}

func wordObjectWithLeadingWhitespace(textTree *text.Tree, pos uint64, targetCount uint64) (uint64, uint64) {
	startPos, endPos := pos, pos

	// Scan backwards to the start of leading whitespace.
	gc := segment.Empty()
	reverseReader := textTree.ReverseReaderAtPosition(pos)
	reverseGcIter := segment.NewReverseGraphemeClusterIter(reverseReader)
	for {
		err := reverseGcIter.NextSegment(gc)
		if err != nil || gc.HasNewline() || !gc.IsWhitespace() {
			break
		}
		startPos -= gc.NumRunes()
	}

	// Skip the next gc, since we already know it's whitespace.
	reader := textTree.ReaderAtPosition(pos)
	gcIter := segment.NewGraphemeClusterIter(reader)
	err := gcIter.NextSegment(gc)
	if err != nil {
		// Should never happen, because the caller validated that there's at least one rune.
		panic(err)
	}
	endPos += gc.NumRunes()

	// Scan forward to the end of the word after leading whitespace.
	prevWasWhitespace, prevWasPunct := true, false
	var count uint64
	for {
		err := gcIter.NextSegment(gc)
		if err != nil {
			break
		}

		isWhitespace := gc.IsWhitespace()
		isPunct := isPunct(gc)
		if (!prevWasWhitespace && isWhitespace) ||
			(!prevWasPunct && !prevWasWhitespace && isPunct) ||
			(prevWasPunct && !isPunct && !isWhitespace) {
			count++
		}

		if count == targetCount {
			break
		}

		endPos += gc.NumRunes()
		prevWasWhitespace = isWhitespace
		prevWasPunct = isPunct
	}

	return startPos, endPos
}

func wordObjectWithTrailingWhitespace(textTree *text.Tree, pos uint64, targetCount uint64) (uint64, uint64) {
	startPos, endPos := pos, pos
	reader := textTree.ReaderAtPosition(pos)
	gcIter := segment.NewGraphemeClusterIter(reader)
	gc := segment.Empty()

	// Lookahead one gc to see if we're in punctuation.
	err := gcIter.NextSegment(gc)
	if err != nil {
		// Should never happen, because the caller validated that there's at least one rune.
		panic(err)
	}
	firstIsPunct := isPunct(gc)
	firstIsWhitespace := gc.IsWhitespace()
	endPos += gc.NumRunes()

	// Scan backwards to the previous word boundary.
	reverseReader := textTree.ReverseReaderAtPosition(pos)
	reverseGcIter := segment.NewReverseGraphemeClusterIter(reverseReader)
	for {
		err = reverseGcIter.NextSegment(gc)
		if err != nil ||
			gc.IsWhitespace() ||
			gc.HasNewline() ||
			(firstIsPunct != isPunct(gc)) {
			break
		}
		startPos -= gc.NumRunes()
	}

	// Scan forward to the end of word.
	prevWasWhitespace := firstIsWhitespace
	prevWasPunct := firstIsPunct
	var count uint64
	for {
		err = gcIter.NextSegment(gc)
		if err != nil {
			break
		}

		isWhitespace := gc.IsWhitespace()
		isPunct := isPunct(gc)
		if (!prevWasWhitespace && isWhitespace) ||
			(!prevWasPunct && !prevWasWhitespace && isPunct) ||
			(prevWasPunct && !isPunct && !isWhitespace) {
			count++
		}

		if count == targetCount {
			break
		}

		prevWasWhitespace = isWhitespace
		prevWasPunct = isPunct
		endPos += gc.NumRunes()
	}

	// If we're at the end of the line or the next char isn't whitespace, we're done.
	if gc.HasNewline() || !gc.IsWhitespace() {
		return startPos, endPos
	}

	// Count the whitespace character we already scanned.
	endPos += gc.NumRunes()

	// Otherwise, keep scanning to end of trailing whitespace.
	for {
		err = gcIter.NextSegment(gc)
		if err != nil || !gc.IsWhitespace() || gc.HasNewline() {
			break
		}
		endPos += gc.NumRunes()
	}

	return startPos, endPos
}

// InnerWordObject returns the start and end positions of the word object or whitespace regions under the cursor.
// This is similar to WordObject, except that whitespace regions are counted as if they were words.
// This is equivalent to vim's "iw" ("inner word") object.
func InnerWordObject(textTree *text.Tree, pos uint64, targetCount uint64) (uint64, uint64) {
	if targetCount == 0 {
		return pos, pos
	}

	startPos, endPos := pos, pos
	reader := textTree.ReaderAtPosition(pos)
	gcIter := segment.NewGraphemeClusterIter(reader)
	gc := segment.Empty()

	// Lookahead one gc to see if we're in whitespace or punctuation.
	err := gcIter.NextSegment(gc)
	if err != nil {
		// This can occur only in an empty document.
		return pos, pos
	}
	firstNumRunes := gc.NumRunes()
	firstHasNewline := gc.HasNewline()
	firstIsWhitespace := gc.IsWhitespace()
	firstIsPunct := isPunct(gc)

	// Scan backwards for a word boundary.
	reverseReader := textTree.ReverseReaderAtPosition(pos)
	reverseGcIter := segment.NewReverseGraphemeClusterIter(reverseReader)
	for {
		err = reverseGcIter.NextSegment(gc)
		if err != nil ||
			(firstIsWhitespace != gc.IsWhitespace()) ||
			(firstIsPunct != isPunct(gc)) ||
			gc.HasNewline() {
			break
		}
		startPos -= gc.NumRunes()
	}

	// If the next gc is a newline, then stop there.
	if targetCount == 1 && firstHasNewline {
		return startPos, endPos
	}

	endPos += firstNumRunes

	prevHasNewline := firstHasNewline
	prevWasWhitespace := firstIsWhitespace
	prevWasPunct := firstIsPunct

	// Otherwise, scan forward to the next boundary.
	var count uint64
	for {
		err = gcIter.NextSegment(gc)
		if err != nil {
			break
		}

		hasNewline := gc.HasNewline()
		isWhitespace := gc.IsWhitespace()
		isPunct := isPunct(gc)

		if (!prevWasWhitespace && isWhitespace) ||
			(prevWasWhitespace && !prevHasNewline && !isWhitespace) ||
			(prevWasPunct != isPunct) {
			count++
		}

		if count == targetCount {
			break
		}

		endPos += gc.NumRunes()
		prevHasNewline = hasNewline
		prevWasWhitespace = isWhitespace
		prevWasPunct = isPunct
	}

	return startPos, endPos
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
