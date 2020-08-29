package exec

import (
	"log"

	"github.com/wedaly/aretext/internal/pkg/text"
)

// EditorState represents the current state of the editor.
type EditorState struct {
	screenWidth, screenHeight uint64
	layout                    Layout
	documentBuffer            *BufferState
	replBuffer                *BufferState
}

func NewEditorState(screenWidth, screenHeight uint64, documentBuffer *BufferState) *EditorState {
	documentBuffer.focus = true
	return &EditorState{
		screenWidth:    screenWidth,
		screenHeight:   screenHeight,
		layout:         LayoutDocumentOnly,
		documentBuffer: documentBuffer,
		replBuffer:     NewBufferState(text.NewTree(), 0, 0, 0, 0, 0),
	}
}

func (s *EditorState) ScreenSize() (uint64, uint64) {
	return s.screenWidth, s.screenHeight
}

func (s *EditorState) SetScreenSize(width, height uint64) {
	s.screenWidth = width
	s.screenHeight = height
}

func (s *EditorState) Layout() Layout {
	return s.layout
}

func (s *EditorState) DocumentBuffer() *BufferState {
	return s.documentBuffer
}

func (s *EditorState) ReplBuffer() *BufferState {
	return s.replBuffer
}

func (s *EditorState) FocusedBuffer() *BufferState {
	if s.documentBuffer.focus {
		return s.documentBuffer
	} else if s.replBuffer.focus {
		return s.replBuffer
	} else {
		log.Fatalf("No buffer in focus")
		return nil
	}
}

// Layout controls how buffers are displayed in the editor.
type Layout int

const (
	// LayoutDocumentOnly means that only the document is displayed; the REPL is hidden.
	LayoutDocumentOnly = Layout(iota)

	// LayoutDocumentAndRepl means that both the document and REPL are displayed.
	LayoutDocumentAndRepl
)

// BufferState represents the current state of a text buffer.
type BufferState struct {
	tree   *text.Tree
	cursor cursorState
	view   viewState
	focus  bool
}

func NewBufferState(tree *text.Tree, cursorPosition, viewX, viewY, viewWidth, viewHeight uint64) *BufferState {
	return &BufferState{
		tree:   tree,
		cursor: cursorState{position: cursorPosition},
		view: viewState{
			textOrigin: 0,
			x:          viewX,
			y:          viewY,
			width:      viewWidth,
			height:     viewHeight,
		},
		focus: false,
	}
}

func (s *BufferState) Tree() *text.Tree {
	return s.tree
}

func (s *BufferState) CursorPosition() uint64 {
	return s.cursor.position
}

func (s *BufferState) ViewTextOrigin() uint64 {
	return s.view.textOrigin
}

func (s *BufferState) ViewOrigin() (uint64, uint64) {
	return s.view.x, s.view.y
}

func (s *BufferState) ViewSize() (uint64, uint64) {
	return s.view.width, s.view.height
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
	// textOrigin is the location in the text tree of the first visible character.
	textOrigin uint64

	// x and y are the screen coordinates of the top-left corner
	x, y uint64

	// width and height are the visible width (in columns) and height (in rows) of the document.
	width, height uint64
}
