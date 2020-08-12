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
	wrapConfig   LineWrapConfig
	gcIter       CloneableSegmentIter
	gcSegment    *Segment
	buffer       []rune
	currentWidth uint64
}

// NewWrappedLineIter constructs a segment iterator for soft- and hard-wrapped lines.
func NewWrappedLineIter(runeIter text.CloneableRuneIter, wrapConfig LineWrapConfig) CloneableSegmentIter {
	return &wrappedLineIter{
		wrapConfig: wrapConfig,
		gcIter:     NewGraphemeClusterIter(runeIter),
		gcSegment:  NewSegment(),
		buffer:     make([]rune, 0, 256),
	}
}

// NextSegment implements SegmentIter#NextSegment().
// For hard-wrapped lines, the grapheme cluster containing the newline will be included at the end of the line.
// If a segment is too long to fit on any line, put it in its own line.
func (iter *wrappedLineIter) NextSegment(segment *Segment) error {
	segment.Clear()
	for {
		err := iter.gcIter.NextSegment(iter.gcSegment)
		if err == io.EOF {
			if len(iter.buffer) > 0 {
				// There are runes left in the current segment, so return that.
				segment.Extend(iter.buffer)
				iter.buffer = iter.buffer[:0]
				return nil
			}

			// No runes left to process, so return EOF.
			return io.EOF
		}

		if err != nil {
			return err
		}

		if iter.gcSegment.HasNewline() {
			// Hard line-break on newline.
			segment.Extend(iter.buffer).Extend(iter.gcSegment.Runes())
			iter.buffer = iter.buffer[:0]
			iter.currentWidth = 0
			return nil
		}

		gcWidth := iter.wrapConfig.widthFunc(iter.gcSegment.Runes(), iter.currentWidth)
		if iter.currentWidth+gcWidth > iter.wrapConfig.maxLineWidth {
			if iter.currentWidth == 0 {
				segment.Extend(iter.gcSegment.Runes())

				// This grapheme cluster is too large to fit on the line, so give it its own line.
				err := iter.gcIter.Clone().NextSegment(iter.gcSegment)
				if err == nil && iter.gcSegment.HasNewline() {
					// There's a newline grapheme cluster after the too-large grapheme cluster.
					// Include the newline in this line so we don't accidentally introduce an empty line.
					if err := iter.gcIter.NextSegment(iter.gcSegment); err != nil {
						panic(err)
					}
					segment.Extend(iter.gcSegment.Runes())
				}
				return nil
			}

			// Adding the next grapheme cluster would exceed the max line length, so break at the end
			// of the current line and start a new line with the next grapheme cluster.
			segment.Extend(iter.buffer)
			iter.buffer = iter.buffer[:0]
			for _, r := range iter.gcSegment.Runes() {
				iter.buffer = append(iter.buffer, r)
			}
			iter.currentWidth = gcWidth
			return nil
		}

		// The next grapheme cluster fits in the current line.
		for _, r := range iter.gcSegment.Runes() {
			iter.buffer = append(iter.buffer, r)
		}
		iter.currentWidth += gcWidth
	}
}

// Clone implements CloneableSegmentIter#Clone()
func (iter *wrappedLineIter) Clone() CloneableSegmentIter {
	buffer := make([]rune, len(iter.buffer))
	copy(buffer, iter.buffer)
	return &wrappedLineIter{
		wrapConfig:   iter.wrapConfig,
		gcIter:       iter.gcIter.Clone(),
		gcSegment:    iter.gcSegment.Clone(),
		buffer:       buffer,
		currentWidth: iter.currentWidth,
	}
}
