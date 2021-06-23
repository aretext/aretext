package state

import (
	"github.com/aretext/aretext/clipboard"
	"github.com/aretext/aretext/config"
	"github.com/aretext/aretext/file"
	"github.com/aretext/aretext/menu"
	"github.com/aretext/aretext/selection"
	"github.com/aretext/aretext/syntax"
	"github.com/aretext/aretext/syntax/parser"
	"github.com/aretext/aretext/text"
	"github.com/aretext/aretext/undo"
)

// EditorState represents the current state of the editor.
type EditorState struct {
	screenWidth, screenHeight uint64
	configRuleSet             config.RuleSet
	configVersion             int
	inputMode                 InputMode
	prevInputMode             InputMode
	documentBuffer            *BufferState
	clipboard                 *clipboard.C
	fileWatcher               *file.Watcher
	fileTimeline              *file.Timeline
	menu                      *MenuState
	customMenuItems           []menu.Item
	dirPatternsToHide         []string
	styles                    map[string]config.StyleConfig
	statusMsg                 StatusMsg
	suspendScreenFunc         SuspendScreenFunc
	quitFlag                  bool
}

func NewEditorState(screenWidth, screenHeight uint64, configRuleSet config.RuleSet, suspendScreenFunc SuspendScreenFunc) *EditorState {
	var documentBufferHeight uint64
	if screenHeight > 0 {
		// Leave one line for the status bar at the bottom.
		documentBufferHeight = screenHeight - 1
	}

	buffer := &BufferState{
		textTree: text.NewTree(),
		cursor:   cursorState{},
		selector: &selection.Selector{},
		view: viewState{
			textOrigin: 0,
			x:          0,
			y:          0,
			width:      screenWidth,
			height:     documentBufferHeight,
		},
		search:         searchState{},
		undoLog:        undo.NewLog(),
		syntaxLanguage: syntax.LanguageUndefined,
		tokenTree:      nil,
		tokenizer:      nil,
		tabSize:        uint64(config.DefaultTabSize),
		tabExpand:      config.DefaultTabExpand,
		showTabs:       config.DefaultShowTabs,
		autoIndent:     config.DefaultAutoIndent,
	}

	return &EditorState{
		screenWidth:       screenWidth,
		screenHeight:      screenHeight,
		configRuleSet:     configRuleSet,
		documentBuffer:    buffer,
		clipboard:         clipboard.New(),
		fileWatcher:       file.NewEmptyWatcher(),
		fileTimeline:      file.NewTimeline(),
		menu:              &MenuState{},
		customMenuItems:   nil,
		dirPatternsToHide: nil,
		statusMsg:         StatusMsg{},
		styles:            nil,
		suspendScreenFunc: suspendScreenFunc,
	}
}

func (s *EditorState) ScreenSize() (uint64, uint64) {
	return s.screenWidth, s.screenHeight
}

func (s *EditorState) ConfigVersion() int {
	return s.configVersion
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

func (s *EditorState) DirPatternsToHide() []string {
	return s.dirPatternsToHide
}

func (s *EditorState) StatusMsg() StatusMsg {
	return s.statusMsg
}

func (s *EditorState) Styles() map[string]config.StyleConfig {
	return s.styles
}

func (s *EditorState) FileWatcher() *file.Watcher {
	return s.fileWatcher
}

func (s *EditorState) QuitFlag() bool {
	return s.quitFlag
}

// BufferState represents the current state of a text buffer.
type BufferState struct {
	textTree       *text.Tree
	cursor         cursorState
	selector       *selection.Selector
	view           viewState
	search         searchState
	undoLog        *undo.Log
	syntaxLanguage syntax.Language
	tokenTree      *parser.TokenTree
	tokenizer      *parser.Tokenizer
	tabSize        uint64
	tabExpand      bool
	showTabs       bool
	autoIndent     bool
	showLineNum    bool
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

func (s *BufferState) SelectedRegion() selection.Region {
	return s.selector.Region(s.textTree, s.cursor.position)
}

func (s *BufferState) SelectionMode() selection.Mode {
	return s.selector.Mode()
}

func (s *BufferState) SelectionEndLocator() Locator {
	return SelectionEndLocator(s.textTree, s.cursor.position, s.selector)
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

func (s *BufferState) ShowTabs() bool {
	return s.showTabs
}

func (s *BufferState) LineNumMarginWidth() uint64 {
	if !s.showLineNum {
		return 0
	}

	// One column for each digit in the last line number,
	// plus one space, with a minimum of three cols.
	width := uint64(1)
	n := s.textTree.NumLines()
	for n > 0 {
		width++
		n /= 10
	}
	if width < 3 {
		width = 3
	}

	// Collapse the line margin column if there isn't enough
	// space for at least one column of document text.
	if width >= s.view.width {
		return 0
	}

	return width
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
