package segment

import (
	"io"
	"log"

	"github.com/aretext/aretext/text"
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
		log.Fatalf("maxLineWidth (%d) must be greater than zero", maxLineWidth)
	}

	return LineWrapConfig{maxLineWidth, widthFunc}
}

// WrappedLineIter iterates through soft- and hard-wrapped lines.
type WrappedLineIter struct {
	wrapConfig   LineWrapConfig
	gcIter       GraphemeClusterIter
	gcSegment    *Segment
	buffer       []rune
	currentWidth uint64
}

// NewWrappedLineIter constructs a segment iterator for soft- and hard-wrapped lines.
func NewWrappedLineIter(reader text.Reader, wrapConfig LineWrapConfig) WrappedLineIter {
	return WrappedLineIter{
		wrapConfig: wrapConfig,
		gcIter:     NewGraphemeClusterIter(reader),
		gcSegment:  Empty(),
		buffer:     make([]rune, 0, 256),
	}
}

// NextSegment retrieves the next soft- or hard-wrapped line.
// For hard-wrapped lines, the grapheme cluster containing the newline will be included at the end of the line.
// If a segment is too long to fit on any line, put it in its own line.
func (iter *WrappedLineIter) NextSegment(segment *Segment) error {
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
				gcIterClone := iter.gcIter
				err := gcIterClone.NextSegment(iter.gcSegment)
				if err == nil && iter.gcSegment.HasNewline() {
					// There's a newline grapheme cluster after the too-large grapheme cluster.
					// Include the newline in this line so we don't accidentally introduce an empty line.
					if err := iter.gcIter.NextSegment(iter.gcSegment); err != nil {
						log.Fatalf("%s", err)
					}
					segment.Extend(iter.gcSegment.Runes())
				}
				return nil
			}

			// Adding the next grapheme cluster would exceed the max line length, so break at the end
			// of the current line and start a new line with the next grapheme cluster.
			segment.Extend(iter.buffer)
			iter.buffer = iter.buffer[:0]
			iter.buffer = append(iter.buffer, iter.gcSegment.Runes()...)
			iter.currentWidth = gcWidth
			return nil
		}

		// The next grapheme cluster fits in the current line.
		iter.buffer = append(iter.buffer, iter.gcSegment.Runes()...)
		iter.currentWidth += gcWidth
	}
}
