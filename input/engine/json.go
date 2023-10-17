package engine

import "encoding/json"

type JsonStateMachine struct {
	NumStates    uint64
	StartState   uint64
	AcceptStates map[uint64]JsonCommand
	Transitions  map[uint64][]JsonTransition
}

type JsonCommand struct {
	CommandId uint64
	Name      string
}

type JsonCapture struct {
	CaptureId uint64
	CommandId uint64
	Name      string
}

type JsonTransition struct {
	Event     interface{}
	NextState uint64
	Captures  []JsonCapture
}

type Enricher interface {
	NameForCmd(CmdId) string
	NameForCapture(CaptureId) string
	Event(start, end Event) interface{}
}

func SerializeJson(sm *StateMachine, enricher Enricher) ([]byte, error) {
	x := &JsonStateMachine{
		NumStates:      uint64(sm.numStates),
		StartState:     uint64(sm.startState),
		AcceptStates: make(map[uint64]JsonCommand, len(sm.acceptCmd)),
		Transitions:    make(map[uint64][]JsonTransition, len(sm.transitions)),
	}

	for state, cmdId := range sm.acceptCmd {
		x.AcceptStates[uint64(state)] = JsonCommand{
			CommandId:   uint64(cmdId),
			Name: enricher.NameForCmd(cmdId),
		}
	}

	for state, transitions := range sm.transitions {
		jsonTransitions := make([]JsonTransition, 0, len(transitions))
		for _, t := range transitions {
			jt := JsonTransition{
				Event:     enricher.Event(t.eventRange.start, t.eventRange.end),
				NextState: uint64(t.nextState),
				Captures:  make([]JsonCapture, 0, len(t.captures)),
			}

			for _, cmdId := range sortedCaptureKeys(t) {
				captureId := t.captures[cmdId]
				jt.Captures = append(jt.Captures, JsonCapture{
					CaptureId:        uint64(captureId),
					CommandId: uint64(cmdId),
					Name:      enricher.NameForCapture(captureId),
				})
			}

			jsonTransitions = append(jsonTransitions, jt)
		}
		x.Transitions[uint64(state)] = jsonTransitions
	}

	return json.Marshal(x)
}
