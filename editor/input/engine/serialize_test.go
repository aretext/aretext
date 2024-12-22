package engine

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSerialize(t *testing.T) {
	sm := &StateMachine{
		numStates:  3,
		startState: 1,
		acceptCmd: map[stateId]CmdId{
			0: 2,
			1: 3,
			2: 4,
		},
		transitions: map[stateId][]transition{
			0: {
				{
					eventRange: eventRange{
						start: 10,
						end:   17,
					},
					nextState: 2,
					captures: map[CmdId]CaptureId{
						1: 19,
						2: 27,
					},
				},
				{
					eventRange: eventRange{
						start: 99,
						end:   104,
					},
					nextState: 0,
				},
			},

			1: {
				{
					eventRange: eventRange{
						start: 77,
						end:   79,
					},
					nextState: 0,
				},
			},
		},
	}

	data := Serialize(sm)
	deserialized, err := Deserialize(data)
	require.NoError(t, err)
	assert.Equal(t, sm, deserialized)
}
