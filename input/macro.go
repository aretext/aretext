package input

import "github.com/aretext/aretext/state"

// MacroRecorder records actions to be replayed later.
type MacroRecorder struct {
	actions []Action
}

// NewMacroRecorder constructs a new macro recorder.
func NewMacroRecorder() *MacroRecorder {
	return &MacroRecorder{
		actions: make([]Action, 0, 256),
	}
}

// RecordAction records an action to be replayed later.
func (m *MacroRecorder) RecordAction(f Action) {
	m.actions = append(m.actions, f)
}

// ClearLastAction resets the "last action" to be replayed.
func (m *MacroRecorder) ClearLastAction() {
	m.actions = m.actions[:0]
}

// LastAction returns an action that replays the last action.
// The "last" action is the sequence of all actions recorded since the recorder
// was created or last cleared.
// For example, if the user might press "o" in normal mode to start a new line,
// type "abc" in insert mode, then return to normal mode.  The last action
// in this case would BOTH start a new line AND insert the text "abc".
func (m *MacroRecorder) LastAction() Action {
	// Copy the slice so it won't be affected by later mutations.
	actions := make([]Action, len(m.actions))
	copy(actions, m.actions)

	// Construct a new action that executes each of the recorded actions sequentially.
	combinedAction := func(s *state.EditorState) {
		for _, f := range actions {
			f(s)
		}
	}

	return combinedAction
}
