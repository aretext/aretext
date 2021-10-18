package input

import (
	"log"

	"github.com/gdamore/tcell/v2"

	"github.com/aretext/aretext/state"
)

// Mode represents an input mode, which is a way of interpreting key events.
type Mode interface {
	// ProcessKeyEvent interprets the key event according to this mode.
	// It will return any user-initiated action resulting from the keypress
	ProcessKeyEvent(event *tcell.EventKey, config Config) Action

	// InputBufferString returns a string describing buffered input events.
	// It can be displayed to the user to help them understand the input state.
	InputBufferString() string
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

	log.Printf("Normal mode parser accepted input for rule '%s'\n", result.Rule.Name)
	action := result.Rule.ActionBuilder(ActionBuilderParams{
		InputEvents:          result.Input,
		CountArg:             result.Count,
		ClipboardPageNameArg: result.ClipboardPageName,
		Config:               config,
	})

	action = thenScrollViewToCursor(thenClearStatusMsg(action))

	clearLastActionMacro := result.Rule.AddToLastActionMacro
	action = thenAddToMacros(
		action,
		clearLastActionMacro,
		result.Rule.AddToLastActionMacro,
		result.Rule.AddToUserMacro,
	)

	return firstCheckpointUndoLog(action)
}

func (m *normalMode) InputBufferString() string {
	return m.parser.InputBufferString()
}

// insertMode is used for inserting characters into text.
type insertMode struct{}

func (m *insertMode) ProcessKeyEvent(event *tcell.EventKey, config Config) Action {
	action := m.processKeyEvent(event)
	action = thenScrollViewToCursor(action)
	action = thenAddToMacros(action, false, true, true)
	return action
}

func (m *insertMode) processKeyEvent(event *tcell.EventKey) Action {
	switch event.Key() {
	case tcell.KeyRune:
		return InsertRune(event.Rune())
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		return DeletePrevChar(nil)
	case tcell.KeyEnter:
		return InsertNewlineAndUpdateAutoIndentWhitespace
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
	case tcell.KeyEscape:
		return ReturnToNormalModeAfterInsert
	default:
		return EmptyAction
	}
}

func (m *insertMode) InputBufferString() string {
	return ""
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

func (m *menuMode) InputBufferString() string {
	return ""
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

func (m *searchMode) InputBufferString() string {
	return ""
}

// visualMode is used to visually select a region of the document.
type visualMode struct {
	parser *Parser
}

func newVisualMode() *visualMode {
	parser := NewParser(visualModeRules)
	return &visualMode{parser}
}

func (m *visualMode) ProcessKeyEvent(event *tcell.EventKey, config Config) Action {
	result := m.parser.ProcessInput(event)
	if !result.Accepted {
		return EmptyAction
	}

	log.Printf("Visual mode parser accepted input for rule '%s'\n", result.Rule.Name)
	action := result.Rule.ActionBuilder(ActionBuilderParams{
		InputEvents:          result.Input,
		CountArg:             result.Count,
		ClipboardPageNameArg: result.ClipboardPageName,
		Config:               config,
	})
	action = thenScrollViewToCursor(thenClearStatusMsg(action))

	action = thenAddToMacros(
		action,
		false,
		result.Rule.AddToLastActionMacro,
		result.Rule.AddToUserMacro,
	)

	return action
}

func (m *visualMode) InputBufferString() string {
	return m.parser.InputBufferString()
}

// taskMode is used while a task is running asynchronously.
// This allows the user to cancel the task if it takes too long.
type taskMode struct{}

func (m *taskMode) ProcessKeyEvent(event *tcell.EventKey, config Config) Action {
	switch event.Key() {
	case tcell.KeyEscape:
		return state.CancelTaskIfRunning
	default:
		return EmptyAction
	}
}

func (m *taskMode) InputBufferString() string {
	return ""
}

// firstCheckpointUndoLog sets a checkpoint in the undo log before executing the action.
func firstCheckpointUndoLog(f Action) Action {
	return func(s *state.EditorState) {
		// This ensures that an undo after the action returns the document
		// to the state BEFORE the action was executed.
		// For example, if the user deletes a line (dd), then the next undo should
		// restore the deleted line.
		state.CheckpointUndoLog(s)
		f(s)
	}
}

// thenScrollViewToCursor executes the action, then scrolls the view so the cursor is visible.
func thenScrollViewToCursor(f Action) Action {
	return func(s *state.EditorState) {
		f(s)
		state.ScrollViewToCursor(s)
	}
}

// thenClearStatusMsg executes the action, then clears the status message.
func thenClearStatusMsg(f Action) Action {
	return func(s *state.EditorState) {
		f(s)
		state.SetStatusMsg(s, state.StatusMsg{})
	}
}

// thenAddToMacros executes the action, then records it in macros.
func thenAddToMacros(
	f Action,
	clearLastActionMacro bool,
	addToLastActionMacro bool,
	addToRecordingUserMacro bool,
) Action {
	return func(s *state.EditorState) {
		f(s)

		if clearLastActionMacro {
			state.ClearLastActionMacro(s)
		}

		if addToLastActionMacro {
			state.AddToLastActionMacro(s, state.MacroAction(f))
		}

		if addToRecordingUserMacro {
			state.AddToRecordingUserMacro(s, state.MacroAction(f))
		}
	}
}
