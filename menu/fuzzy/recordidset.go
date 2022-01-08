package fuzzy

import (
	"math"
	"sort"
)

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

func (r recordIdSet) toSlice() []int {
	recordIds := make([]int, 0, len(r.ids))
	for id := range r.ids {
		recordIds = append(recordIds, id)
	}
	sort.Ints(recordIds)
	return recordIds
}
