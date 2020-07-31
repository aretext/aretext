package segment

import (
	"unicode"
)

// Segment represents a sequence of runes from a larger text (for example a grapheme cluster).
type Segment struct {
	runes []rune
}

// NewSegment returns a new, empty segment.
func NewSegment() *Segment {
	return &Segment{runes: make([]rune, 0, 1)}
}

// Append adds a rune to the end of the segment.
func (seg *Segment) Append(r rune) *Segment {
	seg.runes = append(seg.runes, r)
	return seg
}

// ReverseRunes reverses the order of the runes in the segment.
func (seg *Segment) ReverseRunes() *Segment {
	i := 0
	j := len(seg.runes) - 1
	for i < j {
		seg.runes[i], seg.runes[j] = seg.runes[j], seg.runes[i]
		i++
		j--
	}
	return seg
}

// NumRunes returns the number of runes in the segment.
func (seg *Segment) NumRunes() uint64 {
	return uint64(len(seg.runes))
}

// Clone returns a copy of the segment.
func (seg *Segment) Clone() *Segment {
	runes := make([]rune, len(seg.runes))
	copy(runes, seg.runes)
	return &Segment{runes}
}

// Runes returns the runes contained in the segment.
// Callers should not modify the returned slice.
func (seg *Segment) Runes() []rune {
	return seg.runes
}

// HasNewline checks whether a segment contains a line feed rune.
func (seg *Segment) HasNewline() bool {
	for _, r := range seg.runes {
		if r == '\n' {
			return true
		}
	}
	return false
}

// SegmentIsWhitespace checks whether a segment contains all whitespace runes (spaces, tabs, etc).
func (seg *Segment) IsWhitespace() bool {
	for _, r := range seg.runes {
		if !unicode.IsSpace(r) {
			return false
		}
	}
	return true
}
