package exec

import (
	"github.com/wedaly/aretext/internal/pkg/text"
)

// EditorState represents the current state of the editor.
type EditorState struct {
	documentBuffer *BufferState
}

func NewEditorState(documentBuffer *BufferState) *EditorState {
	return &EditorState{documentBuffer}
}

func (s *EditorState) FocusedBuffer() *BufferState {
	return s.documentBuffer
}

// BufferState represents the current state of a text buffer.
type BufferState struct {
	tree   *text.Tree
	cursor cursorState
	view   viewState
}

func NewBufferState(tree *text.Tree, cursorPosition, viewWidth, viewHeight uint64) *BufferState {
	return &BufferState{
		tree:   tree,
		cursor: cursorState{position: cursorPosition},
		view:   viewState{0, viewWidth, viewHeight},
	}
}

func (s *BufferState) Tree() *text.Tree {
	return s.tree
}

func (s *BufferState) CursorPosition() uint64 {
	return s.cursor.position
}

func (s *BufferState) ViewOrigin() uint64 {
	return s.view.origin
}

func (s *BufferState) SetViewSize(width, height uint64) {
	s.view.width = width
	s.view.height = height
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

// viewState represents the current view of the document.
type viewState struct {
	// origin is the location in the text tree of the first visible character.
	origin uint64

	// width and height are the visible width (in columns) and height (in rows) of the document.
	width, height uint64
}
