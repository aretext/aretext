package parser

// State represents an arbitrary state of the parser.
type State interface {
	// Equals returns whether two states are equal.
	Equals(other State) bool
}

// EmptyState represents an empty state.
// An empty state is considered equal to every other empty state.
type EmptyState struct{}

func (s EmptyState) Equals(other State) bool {
	_, ok := other.(EmptyState)
	return ok
}
