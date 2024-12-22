package selection

// Region represents a selection within a document.
// The zero value represents an empty selection.
type Region struct {
	StartPos uint64 // Inclusive
	EndPos   uint64 // Exclusive
}

var EmptyRegion = Region{}

// ContainsPosition returns whether the position is within the region.
func (r Region) ContainsPosition(pos uint64) bool {
	return r.StartPos <= pos && pos < r.EndPos
}

// Clip ensures that the region is within a document of length n.
func (r Region) Clip(n uint64) Region {
	if n == 0 {
		return EmptyRegion
	}

	if r.StartPos > n {
		r.StartPos = n
	}

	if r.EndPos > n {
		r.EndPos = n
	}

	return r
}
