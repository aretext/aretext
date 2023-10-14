package engine

import (
	"encoding/binary"
	"fmt"
	"io"
	"slices"
)

// Serialize encodes the state machine as binary.
func Serialize(sm *StateMachine) []byte {
	var data []byte
	writeInt := func(i uint64) {
		data = binary.AppendUvarint(data, i)
	}

	writeInt(uint64(sm.numStates))
	writeInt(uint64(sm.startState))
	writeInt(uint64(len(sm.acceptCmd)))
	for _, stateId := range sortedAcceptCmdKeys(sm) {
		cmdId := sm.acceptCmd[stateId]
		writeInt(uint64(stateId))
		writeInt(uint64(cmdId))
	}

	writeInt(uint64(len(sm.transitions)))
	for _, stateId := range sortedTransitionKeys(sm) {
		transitions := sm.transitions[stateId]
		writeInt(uint64(stateId))
		writeInt(uint64(len(transitions)))
		for _, t := range transitions {
			writeInt(uint64(t.eventRange.start))
			writeInt(uint64(t.eventRange.end))
			writeInt(uint64(t.nextState))

			writeInt(uint64(len(t.captures)))
			for _, cmdId := range sortedCaptureKeys(t) {
				captureId := t.captures[cmdId]
				writeInt(uint64(cmdId))
				writeInt(uint64(captureId))
			}
		}
	}

	return data
}

// Deserialize constructs a state machine from serialized binary.
func Deserialize(data []byte) (*StateMachine, error) {
	var i int
	readInt := func() (uint64, error) {
		if i >= len(data) {
			return 0, io.EOF
		}
		x, n := binary.Uvarint(data[i:])
		if n <= 0 {
			return 0, fmt.Errorf("varint decode error")
		}
		i += n
		return x, nil
	}

	sm := &StateMachine{}

	numStates, err := readInt()
	if err != nil {
		return nil, fmt.Errorf("deserialize numStates: %w", err)
	}
	sm.numStates = numStates

	startState, err := readInt()
	if err != nil {
		return nil, fmt.Errorf("deserialize startState: %w", err)
	}
	sm.startState = stateId(startState)

	numAcceptCmd, err := readInt()
	if err != nil {
		return nil, fmt.Errorf("deserialize numAcceptCmd: %w", err)
	}

	sm.acceptCmd = make(map[stateId]CmdId, numAcceptCmd)
	for i := uint64(0); i < numAcceptCmd; i++ {
		stateKey, err := readInt()
		if err != nil {
			return nil, fmt.Errorf("deserialize acceptCmd stateKey: %w", err)
		}
		cmdVal, err := readInt()
		if err != nil {
			return nil, fmt.Errorf("deserialize acceptCmd cmdVal: %w", err)
		}
		sm.acceptCmd[stateId(stateKey)] = CmdId(cmdVal)
	}

	numTransitions, err := readInt()
	if err != nil {
		return nil, fmt.Errorf("deserialize numTransitions: %w", err)
	}

	sm.transitions = make(map[stateId][]transition, numTransitions)
	for i := uint64(0); i < numTransitions; i++ {
		stateKey, err := readInt()
		if err != nil {
			return nil, fmt.Errorf("deserialize transition stateKey: %w", err)
		}

		numTransitionsForState, err := readInt()
		if err != nil {
			return nil, fmt.Errorf("deserialize transition numTransitionsForState: %w", err)
		}

		sm.transitions[stateId(stateKey)] = make([]transition, 0, numTransitionsForState)
		for j := uint64(0); j < numTransitionsForState; j++ {
			eventStartVal, err := readInt()
			if err != nil {
				return nil, fmt.Errorf("deserialize transition eventStartVal: %w", err)
			}

			eventEndVal, err := readInt()
			if err != nil {
				return nil, fmt.Errorf("deserialize transition eventEndVal: %w", err)
			}

			nextStateVal, err := readInt()
			if err != nil {
				return nil, fmt.Errorf("deserialize transition nextStateVal: %w", err)
			}

			numCaptures, err := readInt()
			if err != nil {
				return nil, fmt.Errorf("deserialize transition numCaptures: %w", err)
			}

			t := transition{
				eventRange: eventRange{
					start: Event(eventStartVal),
					end:   Event(eventEndVal),
				},
				nextState: stateId(nextStateVal),
			}

			if numCaptures > 0 {
				t.captures = make(map[CmdId]CaptureId, numCaptures)
				for k := uint64(0); k < numCaptures; k++ {
					captureCmdKey, err := readInt()
					if err != nil {
						return nil, fmt.Errorf("deserialize capture cmdKey: %w", err)
					}

					captureIdVal, err := readInt()
					if err != nil {
						return nil, fmt.Errorf("deserialize capture captureIdVal: %w", err)
					}

					t.captures[CmdId(captureCmdKey)] = CaptureId(captureIdVal)
				}
			}

			sm.transitions[stateId(stateKey)] = append(sm.transitions[stateId(stateKey)], t)
		}
	}

	return sm, nil
}

func sortedAcceptCmdKeys(sm *StateMachine) []stateId {
	keys := make([]stateId, 0, len(sm.acceptCmd))
	for stateId := range sm.acceptCmd {
		keys = append(keys, stateId)
	}
	slices.Sort(keys)
	return keys
}

func sortedTransitionKeys(sm *StateMachine) []stateId {
	keys := make([]stateId, 0, len(sm.transitions))
	for stateId := range sm.transitions {
		keys = append(keys, stateId)
	}
	slices.Sort(keys)
	return keys
}

func sortedCaptureKeys(t transition) []CmdId {
	keys := make([]CmdId, 0, len(t.captures))
	for cmdId := range t.captures {
		keys = append(keys, cmdId)
	}
	slices.Sort(keys)
	return keys
}
