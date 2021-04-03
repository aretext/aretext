package state

// InputMode controls how the editor interprets input events.
type InputMode int

const (
	InputModeNormal = InputMode(iota)
	InputModeInsert
	InputModeMenu
	InputModeSearch
)

func (im InputMode) String() string {
	switch im {
	case InputModeNormal:
		return "normal"
	case InputModeInsert:
		return "insert"
	case InputModeMenu:
		return "menu"
	case InputModeSearch:
		return "search"
	default:
		panic("invalid input mode")
	}
}

// SetInputMode sets the editor input mode.
func SetInputMode(state *EditorState, mode InputMode) {
	if state.inputMode != mode && mode == InputModeNormal {
		// Transition back to normal mode should set an undo checkpoint.
		// For example, suppose a user adds text in insert mode, then returns to normal mode,
		// then deletes a line.  The next undo should restore the deleted line, returning to
		// the checkpoint AFTER the user changed from insert->normal mode.
		CheckpointUndoLog(state)
	}

	state.inputMode = mode
}
