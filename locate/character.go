package locate

import (
	"github.com/aretext/aretext/cellwidth"
	"github.com/aretext/aretext/text"
	"github.com/aretext/aretext/text/segment"
)

// NextCharInLine locates the next grapheme cluster in the current line.
func NextCharInLine(tree *text.Tree, count uint64, includeEndOfLineOrFile bool, pos uint64) uint64 {
	segmentIter := segment.NewGraphemeClusterIterForTree(tree, pos, text.ReadDirectionForward)
	seg := segment.Empty()
	var endOfLineOrFile bool
	var prevPrevOffset, prevOffset uint64
	for i := uint64(0); i <= count; i++ {
		eof := segment.NextOrEof(segmentIter, seg)
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
	if endOfLineOrFile && includeEndOfLineOrFile {
		return pos + prevOffset
	}
	return pos + prevPrevOffset
}

// PrevCharInLine locates the previous grapheme cluster in the current line.
func PrevCharInLine(tree *text.Tree, count uint64, includeEndOfLineOrFile bool, pos uint64) uint64 {
	segmentIter := segment.NewGraphemeClusterIterForTree(tree, pos, text.ReadDirectionBackward)
	seg := segment.Empty()
	var offset uint64
	for i := uint64(0); i < count; i++ {
		eof := segment.NextOrEof(segmentIter, seg)
		if eof {
			break
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
	iter := segment.NewGraphemeClusterIterForTree(tree, pos, text.ReadDirectionBackward)
	seg := segment.Empty()
	for i := uint64(0); i < count; i++ {
		eof := segment.NextOrEof(iter, seg)
		if eof {
			break
		}
		pos -= seg.NumRunes()
	}
	return pos
}

// NextMatchingCharInLine locates the count'th next occurrence of a rune in the line.
// If no match is found, the original position is returned.
func NextMatchingCharInLine(tree *text.Tree, char rune, count uint64, includeChar bool, pos uint64) uint64 {
	var matchCount uint64
	var offset, prevOffset uint64
	startPos := pos
	segmentIter := segment.NewGraphemeClusterIterForTree(tree, pos, text.ReadDirectionForward)
	seg := segment.Empty()
	for {
		eof := segment.NextOrEof(segmentIter, seg)
		if eof || seg.HasNewline() {
			// No match found before end of line or file.
			return startPos
		}

		for _, r := range seg.Runes() {
			if r == char {
				matchCount++
				if matchCount == count {
					if includeChar {
						return pos + offset
					} else {
						return pos + prevOffset
					}
				}
			}
		}

		prevOffset = offset
		offset += seg.NumRunes()
	}
}

// PrevMatchingCharInLine locates the count'th previous occurrence of a rune in the line.
// If no match is found, the original position is returned.
func PrevMatchingCharInLine(tree *text.Tree, char rune, count uint64, includeChar bool, pos uint64) uint64 {
	var matchCount uint64
	var offset, prevOffset uint64
	startPos := pos
	segmentIter := segment.NewGraphemeClusterIterForTree(tree, pos, text.ReadDirectionBackward)
	seg := segment.Empty()
	for {
		eof := segment.NextOrEof(segmentIter, seg)
		if eof || seg.HasNewline() {
			// No match found before end of line or file.
			return startPos
		}

		prevOffset = offset
		offset += seg.NumRunes()

		for _, r := range seg.Runes() {
			if r == char {
				matchCount++
				if matchCount == count {
					if includeChar {
						return pos - offset
					} else {
						return pos - prevOffset
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
	iter := segment.NewGraphemeClusterIterForTree(tree, pos, text.ReadDirectionForward)
	seg := segment.Empty()
	var offset uint64
	lastAlignedPos := pos
	for pos < startPos {
		if offset%tabSize == 0 {
			lastAlignedPos = pos
		}
		eof := segment.NextOrEof(iter, seg)
		if eof {
			break
		}
		offset += cellwidth.GraphemeClusterWidth(seg.Runes(), offset, tabSize)
		pos += seg.NumRunes()
	}
	return lastAlignedPos
}

func findPrevWhitespaceStartPos(tree *text.Tree, tabSize uint64, pos uint64) uint64 {
	iter := segment.NewGraphemeClusterIterForTree(tree, pos, text.ReadDirectionBackward)
	seg := segment.Empty()
	for {
		eof := segment.NextOrEof(iter, seg)
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

// NonWhitespaceOrNewline locates the next non-whitespace character or newline on or after a position.
func NextNonWhitespaceOrNewline(tree *text.Tree, pos uint64) uint64 {
	segmentIter := segment.NewGraphemeClusterIterForTree(tree, pos, text.ReadDirectionForward)
	seg := segment.Empty()
	var offset uint64
	for {
		eof := segment.NextOrEof(segmentIter, seg)
		if eof || !seg.IsWhitespace() || seg.HasNewline() {
			break
		}
		offset += seg.NumRunes()
	}
	return pos + offset
}

// NextNewline locates the next newline on or after the specified position.
// It returns both the positon of the newline as well as its length in runes,
// since the grapheme cluster could be either '\n' or '\r\n'.
func NextNewline(tree *text.Tree, pos uint64) (uint64, uint64, bool) {
	segmentIter := segment.NewGraphemeClusterIterForTree(tree, pos, text.ReadDirectionForward)
	seg := segment.Empty()
	var offset uint64
	for {
		eof := segment.NextOrEof(segmentIter, seg)
		if eof {
			return 0, 0, false
		} else if seg.HasNewline() {
			return pos + offset, seg.NumRunes(), true
		}
		offset += seg.NumRunes()
	}
}
