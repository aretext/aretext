package exec

import (
	"github.com/wedaly/aretext/internal/pkg/text"
)

// State is the current state of the text and cursor.
type State struct {
	tree   *text.Tree
	cursor cursorState
}

func NewState(tree *text.Tree, cursorPosition uint64) *State {
	return &State{
		tree:   tree,
		cursor: cursorState{position: cursorPosition},
	}
}

func (s *State) Tree() *text.Tree {
	return s.tree
}

func (s *State) CursorPosition() uint64 {
	return s.cursor.position
}

// cursorState is the current state of the cursor.
type cursorState struct {
	// position is a position within the text tree where the cursor appears.
	position uint64

	// logicalOffset is the number of cells after the end of the line
	// for the cursor's logical (not necessarily visible) position.
	// This is used for navigating up/down.
	// For example, consider this text, where [m] is the current cursor position.
	//     1: the quick
	//     2: brown
	//     3: fox ju[m]ped over the lazy dog
	// If the user then navigates up one line, then we'd see:
	//     1: the quick
	//     2: brow[n]  [*]
	//     3: fox jumped over the lazy dog
	// where [n] is the visible position and [*] is the logical position,
	// with logicalOffset = 2.
	// If the user then navigates up one line again, we'd see:
	//     1: the qu[i]ck
	//     2: brown
	//     3: fox jumped over the lazy dog
	// where [i] is the character directly above the logical position.
	logicalOffset uint64
}
