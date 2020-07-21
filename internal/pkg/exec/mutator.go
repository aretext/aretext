package exec

import "fmt"

// Mutator modifies the state of the cursor or text.
type Mutator interface {
	fmt.Stringer
	Mutate(state *State)
}

// cursorMutator updates the current location of the cursor.
type cursorMutator struct {
	loc Locator
}

// NewCursorMutator returns a mutator that updates the cursor location.
func NewCursorMutator(loc Locator) Mutator {
	return &cursorMutator{loc}
}

func (cpm *cursorMutator) Mutate(state *State) {
	state.cursor = cpm.loc.Locate(state)
}

func (cpm *cursorMutator) String() string {
	return fmt.Sprintf("MutateCursor(%s)", cpm.loc)
}

// insertRuneMutator inserts a rune at the current cursor location.
type insertRuneMutator struct {
	r rune
}

// NewInsertRuneMutator returns a mutator that inserts a rune at the current cursor location.
func NewInsertRuneMutator(r rune) Mutator {
	return &insertRuneMutator{r}
}

func (irm *insertRuneMutator) Mutate(state *State) {
	startPos := state.cursor.position
	if err := state.tree.InsertAtPosition(startPos, irm.r); err != nil {
		// Invalid UTF-8 character; ignore it.
		return
	}

	state.cursor.position = startPos + 1
}

func (irm *insertRuneMutator) String() string {
	return fmt.Sprintf("InsertRune(%c)", irm.r)
}
