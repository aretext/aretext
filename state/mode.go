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
	state.inputMode = mode
}
