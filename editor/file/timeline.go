package file

// TimelineState represents the state of a file loaded in the editor
// before or after a transition in the timeline.
type TimelineState struct {
	Path    string
	LineNum uint64
	Col     uint64
}

// Empty checks whether the timeline state is empty.
// An empty state represents the beginning or end of the timeline.
func (s TimelineState) Empty() bool {
	return s.Path == ""
}

// Timeline represents the history of files loaded in the editor.
// It provides the ability to move forward or backward through loaded files.
type Timeline struct {
	pastStates   []TimelineState
	futureStates []TimelineState
}

// NewTimeline returns a new, empty timeline.
func NewTimeline() *Timeline {
	return &Timeline{
		pastStates:   make([]TimelineState, 0, 16),
		futureStates: make([]TimelineState, 0, 16),
	}
}

// TransitionFrom moves the timeline forward from the given state
// and invalidates any future states in the timeline.
// This happens when the user loads a new document in the editor.
func (t *Timeline) TransitionFrom(fromState TimelineState) {
	t.futureStates = t.futureStates[:0]
	t.pastStates = append(t.pastStates, fromState)
}

// TransitionBackwardFrom moves the timeline backward to the previous state.
func (t *Timeline) TransitionBackwardFrom(fromState TimelineState) {
	if len(t.pastStates) == 0 {
		return
	}
	t.pastStates = t.pastStates[:len(t.pastStates)-1]
	t.futureStates = append(t.futureStates, fromState)
}

// TransitionForwardFrom moves the timeline forward to the next state.
func (t *Timeline) TransitionForwardFrom(fromState TimelineState) {
	if len(t.futureStates) == 0 {
		return
	}
	t.futureStates = t.futureStates[:len(t.futureStates)-1]
	t.pastStates = append(t.pastStates, fromState)
}

// PeekBackward returns the state immediately before the current state.
func (t *Timeline) PeekBackward() TimelineState {
	if len(t.pastStates) == 0 {
		return TimelineState{}
	}
	return t.pastStates[len(t.pastStates)-1]
}

// PeekForward returns the state immediately after the current state.
func (t *Timeline) PeekForward() TimelineState {
	if len(t.futureStates) == 0 {
		return TimelineState{}
	}
	return t.futureStates[len(t.futureStates)-1]
}
