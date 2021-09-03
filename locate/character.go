package locate

import (
	"io"

	"github.com/aretext/aretext/cellwidth"
	"github.com/aretext/aretext/text"
	"github.com/aretext/aretext/text/segment"
)

// NextCharInLine locates the next grapheme cluster in the current line.
func NextCharInLine(tree *text.Tree, count uint64, includeEndOfLineOrFile bool, pos uint64) uint64 {
	reader := tree.ReaderAtPosition(pos)
	segmentIter := segment.NewGraphemeClusterIter(reader)
	seg := segment.Empty()
	var endOfLineOrFile bool
	var prevPrevOffset, prevOffset uint64
	for i := uint64(0); i <= count; i++ {
		err := segmentIter.NextSegment(seg)
		if err == io.EOF {
			endOfLineOrFile = true
			break
		} else if err != nil {
			panic(err)
		}

		if seg.HasNewline() {
			endOfLineOrFile = true
			break
		}
		prevPrevOffset = prevOffset
		prevOffset += seg.NumRunes()
	}
	if endOfLineOrFile && includeEndOfLineOrFile {
		return pos + prevOffset
	}
	return pos + prevPrevOffset
}

// PrevCharInLine locates the previous grapheme cluster in the current line.
func PrevCharInLine(tree *text.Tree, count uint64, includeEndOfLineOrFile bool, pos uint64) uint64 {
	reader := tree.ReverseReaderAtPosition(pos)
	segmentIter := segment.NewReverseGraphemeClusterIter(reader)
	seg := segment.Empty()
	var offset uint64
	for i := uint64(0); i < count; i++ {
		err := segmentIter.NextSegment(seg)
		if err == io.EOF {
			break
		} else if err != nil {
			panic(err)
		}
		if offset+seg.NumRunes() > pos {
			return 0
		}
		if seg.HasNewline() {
			if includeEndOfLineOrFile {
				offset += seg.NumRunes()
			}
			break
		}
		offset += seg.NumRunes()
	}
	return pos - offset
}

// PrevChar locates the grapheme cluster before a position, which may be on a previous line.
func PrevChar(tree *text.Tree, count uint64, pos uint64) uint64 {
	reader := tree.ReverseReaderAtPosition(pos)
	iter := segment.NewReverseGraphemeClusterIter(reader)
	seg := segment.Empty()
	for i := uint64(0); i < count; i++ {
		err := iter.NextSegment(seg)
		if err == io.EOF {
			break
		} else if err != nil {
			panic(err)
		}
		pos -= seg.NumRunes()
	}
	return pos
}

// NextMatchingCharInLine locates the count'th next occurrence of a rune in the line.
func NextMatchingCharInLine(tree *text.Tree, char rune, count uint64, includeChar bool, pos uint64) (bool, uint64) {
	var matchCount uint64
	var offset, prevOffset uint64
	reader := tree.ReaderAtPosition(pos)
	segmentIter := segment.NewGraphemeClusterIter(reader)
	seg := segment.Empty()
	for {
		err := segmentIter.NextSegment(seg)
		if err == io.EOF || (err == nil && seg.HasNewline()) {
			// No match found before end of line or file.
			return false, 0
		} else if err != nil {
			panic(err)
		}

		if offset > 0 {
			for _, r := range seg.Runes() {
				if r == char {
					matchCount++
					if matchCount == count {
						if includeChar {
							return true, pos + offset
						} else {
							return true, pos + prevOffset
						}
					}
				}
			}
		}

		prevOffset = offset
		offset += seg.NumRunes()
	}
}

// PrevMatchingCharInLine locates the count'th previous occurrence of a rune in the line.
func PrevMatchingCharInLine(tree *text.Tree, char rune, count uint64, includeChar bool, pos uint64) (bool, uint64) {
	var matchCount uint64
	var offset, prevOffset uint64
	reader := tree.ReverseReaderAtPosition(pos)
	segmentIter := segment.NewReverseGraphemeClusterIter(reader)
	seg := segment.Empty()
	for {
		err := segmentIter.NextSegment(seg)
		if err == io.EOF || (err == nil && seg.HasNewline()) {
			// No match found before end of line or file.
			return false, 0
		} else if err != nil {
			panic(err)
		}

		prevOffset = offset
		offset += seg.NumRunes()

		for _, r := range seg.Runes() {
			if r == char {
				matchCount++
				if matchCount == count {
					if includeChar {
						return true, pos - offset
					} else {
						return true, pos - prevOffset
					}
				}
			}
		}
	}
}

// PrevAutoIndent locates the previous tab stop if autoIndent is enabled.
// If autoIndent is disabled or the characters before the cursor are not spaces/tabs, it returns the original position.
func PrevAutoIndent(tree *text.Tree, autoIndentEnabled bool, tabSize uint64, pos uint64) uint64 {
	if !autoIndentEnabled {
		return pos
	}

	prevTabAlignedPos := findPrevTabAlignedPos(tree, tabSize, pos)
	prevWhitespaceStartPos := findPrevWhitespaceStartPos(tree, tabSize, pos)
	if prevTabAlignedPos < prevWhitespaceStartPos {
		return prevWhitespaceStartPos
	} else {
		return prevTabAlignedPos
	}
}

func findPrevTabAlignedPos(tree *text.Tree, tabSize uint64, startPos uint64) uint64 {
	pos := StartOfLineAtPos(tree, startPos)
	reader := tree.ReaderAtPosition(pos)
	iter := segment.NewGraphemeClusterIter(reader)
	seg := segment.Empty()
	var offset uint64
	lastAlignedPos := pos
	for pos < startPos {
		if offset%tabSize == 0 {
			lastAlignedPos = pos
		}
		err := iter.NextSegment(seg)
		if err == io.EOF {
			break
		} else if err != nil {
			panic(err)
		}
		offset += cellwidth.GraphemeClusterWidth(seg.Runes(), offset, tabSize)
		pos += seg.NumRunes()
	}
	return lastAlignedPos
}

func findPrevWhitespaceStartPos(tree *text.Tree, tabSize uint64, pos uint64) uint64 {
	reader := tree.ReverseReaderAtPosition(pos)
	iter := segment.NewReverseGraphemeClusterIter(reader)
	seg := segment.Empty()
	for {
		err := iter.NextSegment(seg)
		if err == io.EOF {
			break
		} else if err != nil {
			panic(err)
		}

		r := seg.Runes()[0]
		if r != ' ' && r != '\t' {
			break
		}
		pos -= seg.NumRunes()
	}
	return pos
}

// NonWhitespaceOrNewline locates the next non-whitespace character or newline on or after a position.
func NextNonWhitespaceOrNewline(tree *text.Tree, pos uint64) uint64 {
	reader := tree.ReaderAtPosition(pos)
	segmentIter := segment.NewGraphemeClusterIter(reader)
	seg := segment.Empty()
	var offset uint64
	for {
		err := segmentIter.NextSegment(seg)
		if err == io.EOF || (err == nil && (!seg.IsWhitespace() || seg.HasNewline())) {
			break
		} else if err != nil {
			panic(err)
		}
		offset += seg.NumRunes()
	}
	return pos + offset
}

// NextNewline locates the next newline on or after the specified position.
// It returns both the positon of the newline as well as its length in runes,
// since the grapheme cluster could be either '\n' or '\r\n'.
func NextNewline(tree *text.Tree, pos uint64) (uint64, uint64, bool) {
	reader := tree.ReaderAtPosition(pos)
	segmentIter := segment.NewGraphemeClusterIter(reader)
	seg := segment.Empty()
	var offset uint64
	for {
		err := segmentIter.NextSegment(seg)
		if err == io.EOF {
			return 0, 0, false
		} else if err != nil {
			panic(err)
		} else if seg.HasNewline() {
			return pos + offset, seg.NumRunes(), true
		}
		offset += seg.NumRunes()
	}
}

// NumGraphemeClustersInRange counts the number of grapheme clusters from the start position (inclusive) to the end position (exclusive).
func NumGraphemeClustersInRange(tree *text.Tree, startPos, endPos uint64) uint64 {
	reader := tree.ReaderAtPosition(startPos)
	segmentIter := segment.NewGraphemeClusterIter(reader)
	seg := segment.Empty()
	var offset, count uint64
	for startPos+offset < endPos {
		err := segmentIter.NextSegment(seg)
		if err == io.EOF || (err == nil && seg.HasNewline()) {
			break
		} else if err != nil {
			panic(err)
		}
		count++
		offset += seg.NumRunes()
	}
	return count

}
