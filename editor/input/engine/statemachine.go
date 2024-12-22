package engine

// stateId is a unique identifier for a state.
type stateId uint64

// stateMachine is a discrete finite automaton (DFA) for parsing input commands.
type StateMachine struct {
	numStates   uint64
	startState  stateId
	acceptCmd   map[stateId]CmdId
	transitions map[stateId][]transition
}

// eventRange is a range of input events (inclusive).
type eventRange struct {
	start, end Event
}

// transition represents a transition from one state to another in the DFA.
type transition struct {
	eventRange eventRange
	nextState  stateId
	captures   map[CmdId]CaptureId
}
