package segment

import (
	"io"

	"github.com/wedaly/aretext/internal/pkg/text"
)

// GraphemeClusterWidthFunc returns the width in cells for a given grapheme cluster.
type GraphemeClusterWidthFunc func(gc []rune, offsetInLine uint64) uint64

// LineWrapConfig controls how lines should be soft-wrapped.
type LineWrapConfig struct {
	maxLineWidth uint64
	widthFunc    GraphemeClusterWidthFunc
}

// NewLineWrapConfig constructs a configuration for soft-wrapping lines.
// maxLineWidth is the maximum number of cells per line, which must be at least one.
// widthFunc returns the width in cells for a given grapheme cluster.
func NewLineWrapConfig(maxLineWidth uint64, widthFunc GraphemeClusterWidthFunc) LineWrapConfig {
	if maxLineWidth == 0 {
		panic("maxLineWidth must be greater than zero")
	}

	return LineWrapConfig{maxLineWidth, widthFunc}
}

// wrappedLineIter iterates through soft- and hard-wrapped lines.
type wrappedLineIter struct {
	wrapConfig     LineWrapConfig
	gcIter         CloneableSegmentIter
	currentSegment *Segment
	currentWidth   uint64
}

// NewWrappedLineIter constructs a segment iterator for soft- and hard-wrapped lines.
func NewWrappedLineIter(runeIter text.CloneableRuneIter, wrapConfig LineWrapConfig) CloneableSegmentIter {
	return &wrappedLineIter{
		wrapConfig:     wrapConfig,
		gcIter:         NewGraphemeClusterIter(runeIter),
		currentSegment: NewSegment(),
	}
}

// NextSegment implements SegmentIter#NextSegment().
// For hard-wrapped lines, the grapheme cluster containing the newline will be included at the end of the line.
// If a segment is too long to fit on any line, put it in its own line.
func (iter *wrappedLineIter) NextSegment() (*Segment, error) {
	for {
		gc, err := iter.gcIter.NextSegment()
		if err == io.EOF {
			if iter.currentSegment != nil && iter.currentSegment.NumRunes() > 0 {
				// There are runes left in the current segment, so return that.
				seg := iter.currentSegment
				iter.currentSegment = nil
				return seg, nil
			}

			// No runes left to process, so return EOF.
			return nil, io.EOF
		}

		if err != nil {
			return nil, err
		}

		if gc.HasNewline() {
			// Hard line-break on newline.
			seg := iter.currentSegment.Extend(gc.Runes())
			iter.currentSegment = NewSegment()
			iter.currentWidth = 0
			return seg, nil
		}

		gcWidth := iter.wrapConfig.widthFunc(gc.Runes(), iter.currentWidth)
		if iter.currentWidth+gcWidth > iter.wrapConfig.maxLineWidth {
			if iter.currentWidth == 0 {
				// This grapheme cluster is too large to fit on the line, so give it its own line.
				lookaheadGc, err := iter.gcIter.Clone().NextSegment()
				if err == nil && lookaheadGc.HasNewline() {
					// There's a newline grapheme cluster after the too-large grapheme cluster.
					// Include the newline in this line so we don't accidentally introduce an empty line.
					newlineGc, _ := iter.gcIter.NextSegment()
					return NewSegment().Extend(gc.Runes()).Extend(newlineGc.Runes()), nil
				}
				return NewSegment().Extend(gc.Runes()), nil
			}

			// Adding the next grapheme cluster would exceed the max line length, so break at the end
			// of the current line and start a new line with the next grapheme cluster.
			seg := iter.currentSegment
			iter.currentSegment = NewSegment().Extend(gc.Runes())
			iter.currentWidth = gcWidth
			return seg, nil
		}

		// The next grapheme cluster fits in the current line.
		iter.currentSegment.Extend(gc.Runes())
		iter.currentWidth += gcWidth
	}
}

// Clone implements CloneableSegmentIter#Clone()
func (iter *wrappedLineIter) Clone() CloneableSegmentIter {
	return &wrappedLineIter{
		wrapConfig:     iter.wrapConfig,
		gcIter:         iter.gcIter.Clone(),
		currentSegment: iter.currentSegment.Clone(),
		currentWidth:   iter.currentWidth,
	}
}
