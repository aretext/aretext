package parser

import (
	"io"
	"math"

	"github.com/aretext/aretext/text"
)

// TrackingReader tracks the number of runes read by a reader iter and all its clones.
// It updates a shared counter, so clones of this iter should NOT be used in other threads.
// Copying the struct produces a new, independent iterator.
type TrackingRuneIter struct {
	reader  text.Reader
	eof     bool
	limit   uint64
	numRead uint64
	maxRead *uint64
}

// NewTrackingRuneIter starts tracking an existing rune iter.
func NewTrackingRuneIter(reader text.Reader) TrackingRuneIter {
	var maxRead uint64
	return TrackingRuneIter{
		reader:  reader,
		limit:   math.MaxUint64,
		maxRead: &maxRead,
	}
}

// NextRune returns the next rune from the underlying reader and advances the iterator.
func (iter *TrackingRuneIter) NextRune() (rune, error) {
	if iter.limit == 0 {
		return '\x00', io.EOF
	}

	r, _, err := iter.reader.ReadRune()

	if err == nil && iter.limit > 0 {
		iter.limit--
	}

	if err == nil || (err == io.EOF && !iter.eof) {
		iter.eof = bool(err == io.EOF)
		iter.numRead++
		if iter.numRead > *iter.maxRead {
			*iter.maxRead = iter.numRead
		}
	}

	return r, err
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

// Limit sets the maximum number of runes this reader can produce.
func (iter *TrackingRuneIter) Limit(n uint64) {
	iter.limit = n
}

// MaxRead returns the maximum number of runes read by this iter and all its clones.
func (iter *TrackingRuneIter) MaxRead() uint64 {
	return *iter.maxRead
}
