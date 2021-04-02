package state

import (
	"github.com/aretext/aretext/config"
	"github.com/aretext/aretext/file"
	"github.com/aretext/aretext/menu"
	"github.com/aretext/aretext/syntax"
	"github.com/aretext/aretext/syntax/parser"
	"github.com/aretext/aretext/text"
)

// EditorState represents the current state of the editor.
type EditorState struct {
	screenWidth, screenHeight uint64
	configRuleSet             config.RuleSet
	inputMode                 InputMode
	documentBuffer            *BufferState
	fileWatcher               *file.Watcher
	menu                      *MenuState
	customMenuItems           []menu.Item
	dirNamesToHide            map[string]struct{}
	statusMsg                 StatusMsg
	hasUnsavedChanges         bool
	scheduledShellCmd         string
	quitFlag                  bool
}

func NewEditorState(screenWidth, screenHeight uint64, configRuleSet config.RuleSet) *EditorState {
	var documentBufferHeight uint64
	if screenHeight > 0 {
		// Leave one line for the status bar at the bottom.
		documentBufferHeight = screenHeight - 1
	}

	buffer := &BufferState{
		textTree: text.NewTree(),
		cursor:   cursorState{},
		view: viewState{
			textOrigin: 0,
			x:          0,
			y:          0,
			width:      screenWidth,
			height:     documentBufferHeight,
		},
		search:         searchState{},
		syntaxLanguage: syntax.LanguageUndefined,
		tokenTree:      nil,
		tokenizer:      nil,
		tabSize:        uint64(config.DefaultTabSize),
		tabExpand:      config.DefaultTabExpand,
		autoIndent:     config.DefaultAutoIndent,
	}

	return &EditorState{
		screenWidth:     screenWidth,
		screenHeight:    screenHeight,
		configRuleSet:   configRuleSet,
		documentBuffer:  buffer,
		fileWatcher:     file.NewEmptyWatcher(),
		menu:            &MenuState{},
		customMenuItems: nil,
		dirNamesToHide:  nil,
		statusMsg:       StatusMsg{},
	}
}

func (s *EditorState) ScreenSize() (uint64, uint64) {
	return s.screenWidth, s.screenHeight
}

func (s *EditorState) SetScreenSize(width, height uint64) {
	s.screenWidth = width
	s.screenHeight = height
}

func (s *EditorState) InputMode() InputMode {
	return s.inputMode
}

func (s *EditorState) DocumentBuffer() *BufferState {
	return s.documentBuffer
}

func (s *EditorState) Menu() *MenuState {
	return s.menu
}

func (s *EditorState) DirNamesToHide() map[string]struct{} {
	return s.dirNamesToHide
}

func (s *EditorState) StatusMsg() StatusMsg {
	return s.statusMsg
}

func (s *EditorState) FileWatcher() *file.Watcher {
	return s.fileWatcher
}

func (s *EditorState) ScheduledShellCmd() string {
	return s.scheduledShellCmd
}

func (s *EditorState) ClearScheduledShellCmd() {
	s.scheduledShellCmd = ""
}

func (s *EditorState) QuitFlag() bool {
	return s.quitFlag
}

// BufferState represents the current state of a text buffer.
type BufferState struct {
	textTree       *text.Tree
	cursor         cursorState
	view           viewState
	search         searchState
	syntaxLanguage syntax.Language
	tokenTree      *parser.TokenTree
	tokenizer      *parser.Tokenizer
	tabSize        uint64
	tabExpand      bool
	autoIndent     bool
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

func (s *BufferState) SearchQueryAndDirection() (string, text.ReadDirection) {
	return s.search.query, s.search.direction
}

func (s *BufferState) SearchMatch() *SearchMatch {
	return s.search.match
}

func (s *BufferState) SetViewSize(width, height uint64) {
	s.view.width = width
	s.view.height = height
}

func (s *BufferState) TabSize() uint64 {
	return s.tabSize
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