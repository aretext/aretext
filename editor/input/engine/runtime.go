package engine

type Decision uint8

const (
	DecisionWait = Decision(iota)
	DecisionReject
	DecisionAccept
)

// Result is the result of processing a single input event.
type Result struct {
	Decision Decision
	CmdId    CmdId                 // The command that was accepted.
	Captures map[CaptureId][]Event // Captures for the accepted command.
}

type Runtime struct {
	sm           *StateMachine
	currentState stateId
	inputEvents  []Event
	maxInputLen  int
}

func NewRuntime(sm *StateMachine, maxInputLen int) *Runtime {
	if maxInputLen < 1 {
		panic("Runtime maxInputLen must be > 0")
	}

	return &Runtime{
		sm:           sm,
		currentState: sm.startState,
		inputEvents:  make([]Event, 0, maxInputLen),
		maxInputLen:  maxInputLen,
	}
}

// ProcessEvent processes a single input event according to the runtime's program.
// If any command accepts the input, the runtime accepts the input and resets.
// If all commands reject the input, the runtime resets.
// Otherwise, the runtime waits for more input before making a decision.
func (r *Runtime) ProcessEvent(event Event) Result {
	r.inputEvents = append(r.inputEvents, event)
	transition := r.nextTransition(r.currentState, event)
	if transition == nil {
		// No transition from this state based on the input event, so reject the input.
		r.reset()
		return Result{Decision: DecisionReject}
	}

	r.currentState = transition.nextState
	acceptCmd, accepted := r.sm.acceptCmd[r.currentState]
	if accepted {
		// Reached an accept state.
		captures := r.findCapturesForCmd(acceptCmd)
		r.reset()
		return Result{
			Decision: DecisionAccept,
			CmdId:    acceptCmd,
			Captures: captures,
		}
	}

	if len(r.inputEvents) > r.maxInputLen {
		// Reached maximum number of input events, so reject and reset.
		r.reset()
		return Result{Decision: DecisionReject}
	}

	// Wait for more input before making a decision.
	return Result{Decision: DecisionWait}
}

func (r *Runtime) nextTransition(state stateId, event Event) *transition {
	transitions := r.sm.transitions[state]
	lo, hi := 0, len(transitions)-1
	for lo <= hi {
		mid := lo + (hi-lo)/2
		t := transitions[mid]
		if event < t.eventRange.start {
			hi = mid - 1
		} else if event > t.eventRange.end {
			lo = mid + 1
		} else {
			return &t
		}
	}
	return nil
}

func (r *Runtime) reset() {
	r.currentState = r.sm.startState
	r.inputEvents = r.inputEvents[:0]
}

func (r *Runtime) findCapturesForCmd(cmdId CmdId) map[CaptureId][]Event {
	// Replay the events through the state machine to find captures for the accepted command.
	var captures map[CaptureId][]Event
	state := stateId(0)
	for _, event := range r.inputEvents {
		t := r.nextTransition(state, event)
		id, ok := t.captures[cmdId]
		if ok {
			if captures == nil {
				captures = make(map[CaptureId][]Event)
			}
			captures[id] = append(captures[id], event)
		}

		state = t.nextState
	}

	return captures
}
