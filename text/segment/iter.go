package segment

import (
	"io"
	"log"
)

// SegmentIter iterates through segments from a larger text.
type SegmentIter interface {
	// NextSegment finds the next segment in the text.
	// If there are no more segments to return, err will be io.EOF and the segment will empty.
	// Once NextSegment returns io.EOF, every subsequent call will also return io.EOF.
	NextSegment(segment *Segment) error
}

// CloneableSegmentIter is a segment iterator that can be cloned to produce a new, independent iterator.
// This can be used to "look ahead" in the stream of segments.
type CloneableSegmentIter interface {
	SegmentIter

	// Clone returns an independent copy of the iterator.
	Clone() CloneableSegmentIter
}

// NextOrEof finds the next segment and returns a flag indicating end of file.
// If an error occurs (e.g. due to invalid UTF-8), it exits with an error.
func NextOrEof(segmentIter SegmentIter, seg *Segment) (eof bool) {
	err := segmentIter.NextSegment(seg)
	if err == io.EOF {
		return true
	}

	if err != nil {
		log.Fatalf("%s", err)
	}

	return false
}
