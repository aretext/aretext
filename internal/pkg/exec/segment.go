package exec

import (
	"io"
	"unicode"

	"github.com/wedaly/aretext/internal/pkg/text"
	"github.com/wedaly/aretext/internal/pkg/text/breaks"
)

// SegementIter iterates through rune slices segmented by a break iterator.
type segmentIter struct {
	runeIter  text.CloneableRuneIter
	breakIter breaks.CloneableBreakIter
	prevBreak uint64
}

// newGraphemeClusterSegmentIter constructs a segmentIter for grapheme clusters.
func newGraphemeClusterSegmentIter(tree *text.Tree, pos uint64, direction text.ReadDirection) *segmentIter {
	reader := tree.ReaderAtPosition(pos, direction)
	if direction == text.ReadDirectionBackward {
		runeIter := text.NewCloneableBackwardRuneIter(reader)
		gcBreakIter := breaks.NewReverseGraphemeClusterBreakIter(runeIter.Clone())
		return &segmentIter{
			runeIter:  runeIter,
			breakIter: gcBreakIter,
		}
	} else {
		runeIter := text.NewCloneableForwardRuneIter(reader)
		gcBreakIter := breaks.NewGraphemeClusterBreakIter(runeIter.Clone())
		return &segmentIter{
			runeIter:  runeIter,
			breakIter: gcBreakIter,
		}
	}
}

// nextSegment returns the next (non-empty) segment identified by the break iterator.
func (iter *segmentIter) nextSegment() (segment []rune, eof bool) {
	for {
		nextBreak, err := iter.breakIter.NextBreak()
		if err == io.EOF {
			return nil, true
		} else if err != nil {
			panic(err) // By construction, input must be valid UTF-8, so this should never happen.
		}

		segment := make([]rune, nextBreak-iter.prevBreak)
		for i := 0; i < len(segment); i++ {
			r, err := iter.runeIter.NextRune()
			if err != nil {
				// Since the break iter and rune iter are reading the same input,
				// and the break iter did not return an error, the rune iter should succeed as well.
				panic(err)
			}
			segment[i] = r
		}

		iter.prevBreak = nextBreak

		if len(segment) > 0 {
			return segment, false
		}
	}
}

// clone returns a new, independent iterator at the same position as the original iterator.
func (iter *segmentIter) clone() *segmentIter {
	return &segmentIter{
		runeIter:  iter.runeIter.Clone(),
		breakIter: iter.breakIter.Clone(),
		prevBreak: iter.prevBreak,
	}
}

// segmentHasNewline checks whether a segment contains a line feed rune.
func segmentHasNewline(segment []rune) bool {
	for _, r := range segment {
		if r == '\n' {
			return true
		}
	}
	return false
}

// segmentIsWhitespace checks whether a segment contains all whitespace runes (spaces, tabs, etc).
func segmentIsWhitespace(segment []rune) bool {
	for _, r := range segment {
		if !unicode.IsSpace(r) {
			return false
		}
	}
	return true
}
