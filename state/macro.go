package state

// MacroAction is a transformation of editor state that can be recorded and replayed.
type MacroAction func(*EditorState)

// MacroState stores recorded macros.
// The "last action" macro is used to repeat the last logical action
// (using the "." command in normal mode).
type MacroState struct {
	lastActions []MacroAction
}

// AddToLastActionMacro adds an action to the "last action" macro.
func AddToLastActionMacro(s *EditorState, action MacroAction) {
	s.macroState.lastActions = append(s.macroState.lastActions, action)
}

// ClearLastActionMacro resets the "last action" macro.
func ClearLastActionMacro(s *EditorState) {
	s.macroState.lastActions = nil
}

// ReplayLastActionMacro executes the actions recorded in the "last action" macro.
func ReplayLastActionMacro(s *EditorState) {
	for _, action := range s.macroState.lastActions {
		action(s)
	}
}
