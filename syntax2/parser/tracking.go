package parser

import (
	"github.com/aretext/aretext/text"
)

// TrackingReader tracks the number of runes read by a reader iter and all its clones.
// It updates a shared counter, so clones of this iter should NOT be used in other threads.
// Copying the struct produces a new, independent iterator.
type TrackingRuneIter struct {
	reader  text.Reader
	numRead uint64
	maxRead *uint64
}

// NewTrackingRuneIter starts tracking an existing rune iter.
func NewTrackingRuneIter(reader text.Reader) TrackingRuneIter {
	var maxRead uint64
	return TrackingRuneIter{
		reader:  reader,
		maxRead: &maxRead,
	}
}

// NextRune returns the next rune from the underlying reader and advances the iterator.
func (iter *TrackingRuneIter) NextRune() (rune, error) {
	r, _, err := iter.reader.ReadRune()
	if err != nil {
		return r, err
	}

	iter.numRead++
	if iter.numRead > *iter.maxRead {
		*iter.maxRead = iter.numRead
	}

	return r, nil
}

// Skip advances the iterator by the specified number of positions or the end of the file, whichever comes first.
func (iter *TrackingRuneIter) Skip(n uint64) uint64 {
	for i := uint64(0); i < n; i++ {
		_, err := iter.NextRune()
		if err != nil {
			return i
		}
	}
	return n
}

// MaxRead returns the maximum number of runes read by this iter and all its clones.
func (iter *TrackingRuneIter) MaxRead() uint64 {
	return *iter.maxRead
}
