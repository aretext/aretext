package fuzzy

import (
	"math"
)

const initialNumSlots = 16
const maxLoadFactor = float64(0.7)

// intSet represents a set of non-negative integers, with efficient insertion and membership checking.
type intSet struct {
	n     int
	slots intSetSlots
}

func newIntSet() intSet {
	return intSet{
		n:     0,
		slots: make(intSetSlots, initialNumSlots),
	}
}

func (s *intSet) add(x int) {
	if x < 0 {
		panic("intSet elements cannot be negative")
	}

	idx, exists := s.slots.probe(x)
	if exists {
		return
	}

	s.slots.set(idx, x)
	s.n++

	loadFactor := float64(s.n) / float64(len(s.slots))
	if loadFactor > maxLoadFactor {
		s.grow()
	}
}

func (s intSet) contains(x int) bool {
	_, exists := s.slots.probe(x)
	return exists
}

func (s intSet) forEach(f func(int)) {
	for i := 0; i < len(s.slots); i++ {
		if x, ok := s.slots.get(i); ok {
			f(x)
		}
	}
}

func (s *intSet) grow() {
	slots := make(intSetSlots, 2*len(s.slots))
	s.forEach(func(x int) {
		idx, exists := slots.probe(x)
		if !exists {
			slots.set(idx, x)
		}
	})
	s.slots = slots
}

const slotFilledMask = uint64(1) << 63

// intSetSlots represent slots in a hash table for non-zero integer keys.
// The high-order bit of each slot indicates whether the slot is filled,
// and the remaining bits store the integer value.
type intSetSlots []uint64

func (ss intSetSlots) set(idx int, x int) {
	ss[idx] = uint64(x) | slotFilledMask
}

func (ss intSetSlots) get(idx int) (int, bool) {
	filled := ss[idx]&slotFilledMask > 0
	x := int(ss[idx] & ^slotFilledMask)
	return x, filled
}

func (ss intSetSlots) probe(target int) (int, bool) {
	// Use double-hashing to probe the hash table.
	// We are assuming that there are always 2^n slots, so as long as the second hash function is odd
	// we will eventually probe every slot.
	// The constant multipliers were chosen randomly; the exact values are not significant.
	h1 := (uint64(target) * 0x13205676652c1c3a)
	h2 := (uint64(target) * 0xd23491f24a15d7ee) | 0x1
	m := uint64(len(ss))
	for i := uint64(0); i < m; i++ {
		idx := int((h1 + (i * h2)) % m)
		if x, ok := ss.get(idx); !ok {
			return idx, false
		} else if x == target {
			return idx, true
		}
	}

	// This should never happen because we grow the hash table whenever it exceeds a target load threshold.
	panic("all slots are full")
}

// recordIdSet represents a set of record IDs.
type recordIdSet struct {
	ids      map[int]struct{}
	min, max int // valid only if the set contains at least one ID.
}

func newRecordIdSet() *recordIdSet {
	return &recordIdSet{
		ids: make(map[int]struct{}, 0),
		min: math.MaxInt,
		max: 0,
	}
}

func (r *recordIdSet) add(id int) {
	r.ids[id] = struct{}{}
	if id < r.min {
		r.min = id
	}
	if id > r.max {
		r.max = id
	}
}

func (r recordIdSet) contains(id int) bool {
	_, ok := r.ids[id]
	return ok
}

func (r recordIdSet) forEach(f func(int)) {
	for id := range r.ids {
		f(id)
	}
}
