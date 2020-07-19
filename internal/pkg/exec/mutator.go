package exec

// Mutator modifies the state of the cursor or text.
type Mutator interface {
	Mutate(state *State) error
}

// cursorMutator updates the current location of the cursor.
type cursorMutator struct {
	loc Locator
}

// NewCursorMutator returns a mutator that updates the cursor location.
func NewCursorMutator(loc Locator) Mutator {
	return &cursorMutator{loc}
}

func (cpm *cursorMutator) Mutate(state *State) error {
	state.cursor = cpm.loc.Locate(state)
	return nil
}
