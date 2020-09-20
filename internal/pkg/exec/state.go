package exec

import (
	"log"

	"github.com/wedaly/aretext/internal/pkg/file"
	"github.com/wedaly/aretext/internal/pkg/text"
)

// EditorState represents the current state of the editor.
type EditorState struct {
	screenWidth, screenHeight uint64
	layout                    Layout
	focusedBufferId           BufferId
	documentBuffer            *BufferState
	replBuffer                *BufferState
	replInputStartPos         uint64
	fileWatcher               file.Watcher
	quitFlag                  bool
}

func NewEditorState(screenWidth, screenHeight uint64) *EditorState {
	return &EditorState{
		screenWidth:       screenWidth,
		screenHeight:      screenHeight,
		layout:            LayoutDocumentOnly,
		focusedBufferId:   BufferIdDocument,
		documentBuffer:    NewBufferState(text.NewTree(), 0, 0, 0, screenWidth, screenHeight),
		replBuffer:        NewBufferState(text.NewTree(), 0, 0, 0, 0, 0),
		replInputStartPos: 0,
		fileWatcher:       file.NewEmptyWatcher(),
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

func (s *EditorState) Buffer(id BufferId) *BufferState {
	switch id {
	case BufferIdDocument:
		return s.documentBuffer
	case BufferIdRepl:
		return s.replBuffer
	default:
		log.Fatalf("Unrecognized buffer ID %d\n", id)
		return nil
	}
}

func (s *EditorState) FocusedBufferId() BufferId {
	return s.focusedBufferId
}

func (s *EditorState) ReplInputStartPos() uint64 {
	return s.replInputStartPos
}

func (s *EditorState) SetReplInputStartPos(pos uint64) {
	endPos := s.replBuffer.tree.NumChars()
	if pos > endPos {
		pos = endPos
	}

	s.replInputStartPos = pos
}

func (s *EditorState) FileWatcher() file.Watcher {
	return s.fileWatcher
}

func (s *EditorState) QuitFlag() bool {
	return s.quitFlag
}

// BufferState represents the current state of a text buffer.
type BufferState struct {
	tree   *text.Tree
	cursor cursorState
	view   viewState
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

// BufferId identifies a buffer.
type BufferId int

const (
	// BufferIdDocument identifies the main document being edited.
	BufferIdDocument = BufferId(iota)

	// BufferIdRepl identifies the buffer for REPL input/output.
	BufferIdRepl
)

func (b BufferId) String() string {
	switch b {
	case BufferIdDocument:
		return "document"
	case BufferIdRepl:
		return "repl"
	default:
		log.Fatalf("Unrecognized buffer ID %d\n", b)
		return ""
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

func (l Layout) String() string {
	if l == LayoutDocumentOnly {
		return "DocumentOnly"
	} else if l == LayoutDocumentAndRepl {
		return "DocumentAndRepl"
	} else {
		log.Fatalf("Unrecognized layout: %d\n", l)
		return ""
	}
}
