package exec

import (
	"io"

	"github.com/wedaly/aretext/internal/pkg/text"
	"github.com/wedaly/aretext/internal/pkg/text/breaks"
)

// SegementIter iterates through rune slices segmented by a break iterator.
type segmentIter struct {
	runeIter  text.RuneIter
	breakIter breaks.BreakIter
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

// segmentHasNewline checks whether a segment contains a line feed rune.
func segmentHasNewline(segment []rune) bool {
	for _, r := range segment {
		if r == '\n' {
			return true
		}
	}
	return false
}
