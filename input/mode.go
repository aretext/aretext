package input

import (
	"log"

	"github.com/aretext/aretext/locate"
	"github.com/aretext/aretext/state"
	"github.com/gdamore/tcell/v2"
)

// Mode represents an input mode, which is a way of interpreting key events.
type Mode interface {
	// ProcessKeyEvent interprets the key event according to this mode.
	// It will return any user-initiated action resulting from the keypress
	ProcessKeyEvent(event *tcell.EventKey, config Config) Action
}

// normalMode is used for navigating text.
type normalMode struct {
	parser *Parser
}

func newNormalMode() Mode {
	parser := NewParser(normalModeRules)
	return &normalMode{parser}
}

func (m *normalMode) ProcessKeyEvent(event *tcell.EventKey, config Config) Action {
	result := m.parser.ProcessInput(event)
	if !result.Accepted {
		return EmptyAction
	}

	log.Printf("Parser accepted input for rule '%s'\n", result.Rule.Name)
	action := result.Rule.ActionBuilder(result.Input, result.Count, config)
	return thenScrollViewToCursor(action)
}

// insertMode is used for inserting characters into text.
type insertMode struct{}

func newInsertMode() Mode {
	return &insertMode{}
}

func (m *insertMode) ProcessKeyEvent(event *tcell.EventKey, config Config) Action {
	action := m.processKeyEvent(event)
	return thenScrollViewToCursor(action)
}

func (m *insertMode) processKeyEvent(event *tcell.EventKey) Action {
	switch event.Key() {
	case tcell.KeyRune:
		return m.insertRune(event.Rune())
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		return m.deletePrevChar()
	case tcell.KeyEnter:
		return m.insertNewline()
	case tcell.KeyTab:
		return m.insertTab()
	case tcell.KeyLeft:
		return CursorLeft(nil, nil, Config{})
	case tcell.KeyRight:
		return CursorRight(nil, nil, Config{})
	case tcell.KeyUp:
		return CursorUp(nil, nil, Config{})
	case tcell.KeyDown:
		return CursorDown(nil, nil, Config{})
	default:
		return m.returnToNormalMode()
	}
}

func (m *insertMode) insertRune(r rune) Action {
	return func(s *state.EditorState) {
		state.InsertRune(s, r)
	}
}

func (m *insertMode) insertNewline() Action {
	return state.InsertNewline
}

func (m *insertMode) insertTab() Action {
	return state.InsertTab
}

func (m *insertMode) deletePrevChar() Action {
	return func(s *state.EditorState) {
		state.DeleteRunes(s, func(params state.LocatorParams) uint64 {
			prevInLinePos := locate.PrevCharInLine(params.TextTree, 1, true, params.CursorPos)
			prevAutoIndentPos := locate.PrevAutoIndent(
				params.TextTree,
				params.AutoIndentEnabled,
				params.TabSize,
				params.CursorPos)
			if prevInLinePos < prevAutoIndentPos {
				return prevInLinePos
			} else {
				return prevAutoIndentPos
			}
		})
	}
}

func (m *insertMode) returnToNormalMode() Action {
	return func(s *state.EditorState) {
		state.MoveCursor(s, func(params state.LocatorParams) uint64 {
			return locate.ClosestCharOnLine(params.TextTree, params.CursorPos)
		})
		state.SetInputMode(s, state.InputModeNormal)
	}
}

// menuMode allows the user to search for and select items in a menu.
type menuMode struct{}

func newMenuMode() Mode {
	return &menuMode{}
}

func (m *menuMode) ProcessKeyEvent(event *tcell.EventKey, config Config) Action {
	switch event.Key() {
	case tcell.KeyEscape:
		return m.closeMenu()
	case tcell.KeyEnter:
		return m.executeSelectedMenuItem()
	case tcell.KeyUp:
		return m.menuSelectionUp()
	case tcell.KeyDown:
		return m.menuSelectionDown()
	case tcell.KeyTab:
		return m.menuSelectionDown()
	case tcell.KeyRune:
		return m.appendMenuSearch(event.Rune())
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		return m.deleteMenuSearch()
	default:
		return EmptyAction
	}
}

func (m *menuMode) closeMenu() Action {
	// Returns to normal mode.
	return state.HideMenu
}

func (m *menuMode) executeSelectedMenuItem() Action {
	// Hides the menu, then executes the menu item action.
	// This usually returns to normal mode, unless the menu item action sets a different mode.
	return state.ExecuteSelectedMenuItem
}

func (m *menuMode) menuSelectionUp() Action {
	return func(s *state.EditorState) {
		state.MoveMenuSelection(s, -1)
	}
}

func (m *menuMode) menuSelectionDown() Action {
	return func(s *state.EditorState) {
		state.MoveMenuSelection(s, 1)
	}
}

func (m *menuMode) appendMenuSearch(r rune) Action {
	return func(s *state.EditorState) {
		state.AppendRuneToMenuSearch(s, r)
	}
}

func (m *menuMode) deleteMenuSearch() Action {
	return state.DeleteRuneFromMenuSearch
}

// thenScrollViewToCursor executes the action, then scrolls the view so the cursor is visible.
func thenScrollViewToCursor(f Action) Action {
	return func(s *state.EditorState) {
		f(s)
		state.ScrollViewToCursor(s)
	}
}

// searchMode is used to search the text for a substring.
type searchMode struct {
}

func newSearchMode() Mode {
	return &searchMode{}
}

func (m *searchMode) ProcessKeyEvent(event *tcell.EventKey, config Config) Action {
	switch event.Key() {
	case tcell.KeyEscape:
		// This returns the input mode to normal.
		return func(s *state.EditorState) {
			commit := false
			state.CompleteSearch(s, commit)
		}
	case tcell.KeyEnter:
		// This returns the input mode to normal.
		return func(s *state.EditorState) {
			commit := true
			state.CompleteSearch(s, commit)
		}
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		// This returns the input mode to normal if the search query is empty.
		return state.DeleteRuneFromSearchQuery
	case tcell.KeyRune:
		r := event.Rune()
		return func(s *state.EditorState) {
			state.AppendRuneToSearchQuery(s, r)
		}
	default:
		return EmptyAction
	}
}
