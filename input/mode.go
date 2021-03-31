package input

import (
	"log"

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

func newNormalMode() *normalMode {
	parser := NewParser(normalModeRules)
	return &normalMode{parser}
}

func (m *normalMode) ProcessKeyEvent(event *tcell.EventKey, config Config) Action {
	result := m.parser.ProcessInput(event)
	if !result.Accepted {
		return EmptyAction
	}

	log.Printf("Parser accepted input for rule '%s'\n", result.Rule.Name)
	action := result.Rule.ActionBuilder(ActionBuilderParams{
		InputEvents: result.Input,
		CountArg:    result.Count,
		Config:      config,
	})
	return thenScrollViewToCursor(action)
}

// insertMode is used for inserting characters into text.
type insertMode struct{}

func (m *insertMode) ProcessKeyEvent(event *tcell.EventKey, config Config) Action {
	action := m.processKeyEvent(event)
	return thenScrollViewToCursor(action)
}

func (m *insertMode) processKeyEvent(event *tcell.EventKey) Action {
	switch event.Key() {
	case tcell.KeyRune:
		return InsertRune(event.Rune())
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		return DeletePrevChar
	case tcell.KeyEnter:
		return InsertNewline
	case tcell.KeyTab:
		return InsertTab
	case tcell.KeyLeft:
		return CursorLeft
	case tcell.KeyRight:
		return CursorRight
	case tcell.KeyUp:
		return CursorUp
	case tcell.KeyDown:
		return CursorDown
	default:
		return ReturnToNormalMode
	}
}

// menuMode allows the user to search for and select items in a menu.
type menuMode struct{}

func (m *menuMode) ProcessKeyEvent(event *tcell.EventKey, config Config) Action {
	switch event.Key() {
	case tcell.KeyEscape:
		return HideMenuAndReturnToNormalMode
	case tcell.KeyEnter:
		return ExecuteSelectedMenuItem
	case tcell.KeyUp:
		return MenuSelectionUp
	case tcell.KeyDown:
		return MenuSelectionDown
	case tcell.KeyTab:
		return MenuSelectionDown
	case tcell.KeyRune:
		return AppendRuneToMenuSearch(event.Rune())
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		return DeleteRuneFromMenuSearch
	default:
		return EmptyAction
	}
}

// thenScrollViewToCursor executes the action, then scrolls the view so the cursor is visible.
func thenScrollViewToCursor(f Action) Action {
	return func(s *state.EditorState) {
		f(s)
		state.ScrollViewToCursor(s)
	}
}

// searchMode is used to search the text for a substring.
type searchMode struct{}

func (m *searchMode) ProcessKeyEvent(event *tcell.EventKey, config Config) Action {
	switch event.Key() {
	case tcell.KeyEscape:
		return AbortSearchAndReturnToNormalMode
	case tcell.KeyEnter:
		return CommitSearchAndReturnToNormalMode
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		// This returns the input mode to normal if the search query is empty.
		return DeleteRuneFromSearchQuery
	case tcell.KeyRune:
		return AppendRuneToSearchQuery(event.Rune())
	default:
		return EmptyAction
	}
}
