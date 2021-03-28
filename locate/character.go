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
		eof := segment.NextSegmentOrEof(segmentIter, seg)
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
		eof := segment.NextSegmentOrEof(segmentIter, seg)
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
		eof := segment.NextSegmentOrEof(iter, seg)
		if eof {
			break
		}
		pos -= seg.NumRunes()
	}
	return pos
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
		eof := segment.NextSegmentOrEof(iter, seg)
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
		eof := segment.NextSegmentOrEof(iter, seg)
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
		eof := segment.NextSegmentOrEof(segmentIter, seg)
		if eof || !seg.IsWhitespace() || seg.HasNewline() {
			break
		}
		offset += seg.NumRunes()
	}
	return pos + offset
}
