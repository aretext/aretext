package segment

import (
	"unicode"
	"unicode/utf8"
	"iter"
)

// Segment represents a sequence of runes from a larger text (for example a grapheme cluster).
type Segment struct {
	bytes        []byte
	hasNewline   bool
	isWhitespace bool
}

// Empty returns a new, empty segment.
func Empty() *Segment {
	return &Segment{bytes: make([]byte, 0, 1)}
}

// Clear removes all runes from the segment.
func (seg *Segment) Clear() *Segment {
	seg.bytes = seg.bytes[:0]
	seg.hasNewline = false
	seg.isWhitespace = false
	return seg
}

// Append appends a single rune to the end of the segment.
func (seg *Segment) Append(r rune) *Segment {
	var b [utf8.UTFMax]byte
	n := utf8.EncodeRune(b[:], r)
	seg.hasNewline = seg.hasNewline || bool(r == '\n')
	seg.isWhitespace = unicode.IsSpace(r) && (len(seg.bytes) == 0 || seg.isWhitespace)
	seg.bytes = append(seg.bytes, b[:n]...)
	return seg
}

// ReverseRunes reverses the order of the runes in the segment.
func (seg *Segment) ReverseRunes() *Segment {
	reversed := make([]byte, len(seg.bytes))
	i := 0
	for i < len(seg.bytes) {
		r, n := utf8.DecodeRune(seg.bytes[i:])
		utf8.EncodeRune(reversed[len(reversed)-i-n:], r)
		i += n
	}
	seg.bytes = reversed
	return seg
}

// NumRunes returns the number of runes in the segment.
func (seg *Segment) NumRunes() uint64 {
	return uint64(utf8.RuneCount(seg.bytes))
}

// Bytes returns the UTF-8 encoded bytes of the segment.
// Callers should not modify the returned slice.
func (seg *Segment) Bytes() []byte {
	return seg.bytes
}

// IterRunes iterates over the runes in the segment.
// This can be used in a for loop with range.
func (s *Segment) IterRunes() iter.Seq[rune] {
	return func(yield func(rune) bool) {
		i := 0
		for i < len(s.bytes) {
			r, n := utf8.DecodeRune(s.bytes[i:])
			i += n
			if !yield(r) {
				return
			}
		}
	}
}

// HasNewline checks whether a segment contains a line feed rune.
func (seg *Segment) HasNewline() bool {
	return seg.hasNewline
}

// IsWhitespace checks whether the segment contains only whitespace runes (spaces, tabs, etc).
func (seg *Segment) IsWhitespace() bool {
	return seg.isWhitespace
}
