package parser

import (
	"github.com/aretext/aretext/text"
)

// TrackingRuneIter tracks the number of runes read by an iter and all its clones.
// It updates a shared counter, so clones of this iter should NOT be used in other threads.
type TrackingRuneIter struct {
	subIter text.CloneableRuneIter
	numRead uint64
	maxRead *uint64
}

// NewTrackingRuneIter starts tracking an existing rune iter.
func NewTrackingRuneIter(subIter text.CloneableRuneIter) *TrackingRuneIter {
	var maxRead uint64
	return &TrackingRuneIter{
		subIter: subIter,
		maxRead: &maxRead,
	}
}

// NextRune implements text.CloneableRuneIter#NextRune()
func (iter *TrackingRuneIter) NextRune() (rune, error) {
	r, err := iter.subIter.NextRune()
	if err != nil {
		return r, err
	}

	iter.numRead++
	if iter.numRead > *iter.maxRead {
		*iter.maxRead = iter.numRead
	}

	return r, nil
}

// Clone implements text.CloneableRuneIter#Clone()
func (iter *TrackingRuneIter) Clone() text.CloneableRuneIter {
	return &TrackingRuneIter{
		subIter: iter.subIter.Clone(),
		numRead: iter.numRead,
		maxRead: iter.maxRead, // All clones point to the same counter.
	}
}

// MaxRead returns the maximum number of runes read by this iter and all its clones.
func (iter *TrackingRuneIter) MaxRead() uint64 {
	return *iter.maxRead
}
