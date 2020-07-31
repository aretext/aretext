package segment

// SegmentIter iterates through segments from a larger text.
type SegmentIter interface {
	// NextSegment returns the next segment in the text.
	// If there are no more segments to return, err will be io.EOF and the segment will be nil.
	// Once NextSegment returns io.EOF, every subsequent call will also return io.EOF.
	NextSegment() (*Segment, error)
}

// CloneableSegmentIter is a segment iterator that can be cloned to produce a new, independent iterator.
// This can be used to "look ahead" in the stream of segments.
type CloneableSegmentIter interface {
	SegmentIter

	// Clone returns an independent copy of the iterator.
	Clone() CloneableSegmentIter
}
