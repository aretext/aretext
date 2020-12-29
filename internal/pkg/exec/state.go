package exec

import (
	"github.com/pkg/errors"
	"github.com/wedaly/aretext/internal/pkg/file"
	"github.com/wedaly/aretext/internal/pkg/syntax"
	"github.com/wedaly/aretext/internal/pkg/syntax/parser"
	"github.com/wedaly/aretext/internal/pkg/text"
)

// EditorState represents the current state of the editor.
type EditorState struct {
	screenWidth, screenHeight uint64
	documentBuffer            *BufferState
	menu                      *MenuState
	fileWatcher               file.Watcher
	quitFlag                  bool
}

func NewEditorState(screenWidth, screenHeight uint64) *EditorState {
	var documentBufferHeight uint64
	if screenHeight > 0 {
		// Leave one line for the status bar at the bottom.
		documentBufferHeight = screenHeight - 1
	}
	return &EditorState{
		screenWidth:    screenWidth,
		screenHeight:   screenHeight,
		documentBuffer: NewBufferState(text.NewTree(), 0, 0, 0, screenWidth, documentBufferHeight),
		menu:           &MenuState{},
		fileWatcher:    file.NewEmptyWatcher(),
	}
}

func (s *EditorState) ScreenSize() (uint64, uint64) {
	return s.screenWidth, s.screenHeight
}

func (s *EditorState) SetScreenSize(width, height uint64) {
	s.screenWidth = width
	s.screenHeight = height
}

func (s *EditorState) DocumentBuffer() *BufferState {
	return s.documentBuffer
}

func (s *EditorState) Menu() *MenuState {
	return s.menu
}

func (s *EditorState) FileWatcher() file.Watcher {
	return s.fileWatcher
}

func (s *EditorState) QuitFlag() bool {
	return s.quitFlag
}

// BufferState represents the current state of a text buffer.
type BufferState struct {
	textTree       *text.Tree
	cursor         cursorState
	view           viewState
	syntaxLanguage syntax.Language
	tokenTree      *parser.TokenTree
	tokenizer      *parser.Tokenizer
}

func NewBufferState(textTree *text.Tree, cursorPosition, viewX, viewY, viewWidth, viewHeight uint64) *BufferState {
	return &BufferState{
		textTree: textTree,
		cursor:   cursorState{position: cursorPosition},
		view: viewState{
			textOrigin: 0,
			x:          viewX,
			y:          viewY,
			width:      viewWidth,
			height:     viewHeight,
		},
		syntaxLanguage: syntax.LanguageUndefined,
		tokenTree:      nil,
		tokenizer:      nil,
	}
}

func (s *BufferState) TextTree() *text.Tree {
	return s.textTree
}

func (s *BufferState) TokenTree() *parser.TokenTree {
	return s.tokenTree
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

func (s *BufferState) SetSyntax(language syntax.Language) error {
	s.syntaxLanguage = language
	s.tokenizer = syntax.TokenizerForLanguage(language)

	if s.tokenizer == nil {
		s.tokenTree = nil
		return nil
	}

	r := s.textTree.ReaderAtPosition(0, text.ReadDirectionForward)
	textLen := s.textTree.NumChars()
	tokenTree, err := s.tokenizer.TokenizeAll(r, textLen)
	if err != nil {
		return err
	}

	s.tokenTree = tokenTree
	return nil
}

func (s *BufferState) retokenizeAfterEdit(edit parser.Edit) error {
	if s.tokenizer == nil {
		return nil
	}

	textLen := s.textTree.NumChars()
	readerAtPos := func(pos uint64) parser.InputReader {
		return s.textTree.ReaderAtPosition(pos, text.ReadDirectionForward)
	}
	updatedTokenTree, err := s.tokenizer.RetokenizeAfterEdit(s.tokenTree, edit, textLen, readerAtPos)
	if err != nil {
		return errors.Wrapf(err, "RetokenizeAfterEdit()")
	}

	s.tokenTree = updatedTokenTree
	return nil
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

// MenuState represents the menu for searching and selecting items.
type MenuState struct {
	// visible indicates whether the menu is currently displayed.
	visible bool

	// prompt is a user-facing description of the menu contents.
	prompt string

	// search controls which items are visible based on the user's current search query.
	search *MenuSearch

	// selectedResultIdx is the index of the currently selected search result.
	// If there are no results, this is set to zero.
	// If there are results, this must be less than the number of results.
	selectedResultIdx int
}

func (m *MenuState) Visible() bool {
	return m.visible
}

func (m *MenuState) Prompt() string {
	return m.prompt
}

func (m *MenuState) SearchQuery() string {
	if m.search == nil {
		return ""
	}
	return m.search.Query()
}

func (m *MenuState) SearchResults() (results []MenuItem, selectedResultIdx int) {
	if m.search == nil {
		return nil, 0
	}
	return m.search.Results(), m.selectedResultIdx
}

// MenuItem represents an item in the editor's menu.
type MenuItem struct {
	// Name is the displayed name of the item.
	// This is also used when searching for menu items.
	Name string

	// Action is the action to perform when the user selects the menu item.
	Action Mutator
}
