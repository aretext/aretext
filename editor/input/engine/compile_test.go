package engine

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCompile(t *testing.T) {
	testCases := []struct {
		name     string
		cmdExprs []CmdExpr
		expected *StateMachine
	}{
		{
			name:     "empty",
			cmdExprs: []CmdExpr{},
			expected: &StateMachine{
				numStates:   1,
				acceptCmd:   map[stateId]CmdId{},
				transitions: map[stateId][]transition{},
			},
		},
		{
			name: "EventExpr",
			cmdExprs: []CmdExpr{
				{
					CmdId: 0,
					Expr:  EventExpr{Event: 99},
				},
			},
			expected: &StateMachine{
				numStates: 2,
				acceptCmd: map[stateId]CmdId{
					1: 0,
				},
				transitions: map[stateId][]transition{
					0: {
						{
							eventRange: eventRange{start: 99, end: 99},
							nextState:  1,
						},
					},
				},
			},
		},
		{
			name: "EventRangeExpr",
			cmdExprs: []CmdExpr{
				{
					CmdId: 0,
					Expr: EventRangeExpr{
						StartEvent: 23,
						EndEvent:   79,
					},
				},
			},
			expected: &StateMachine{
				numStates: 2,
				acceptCmd: map[stateId]CmdId{
					1: 0,
				},
				transitions: map[stateId][]transition{
					0: {
						{
							eventRange: eventRange{start: 23, end: 79},
							nextState:  1,
						},
					},
				},
			},
		},
		{
			name: "ConcatExpr",
			cmdExprs: []CmdExpr{
				{
					CmdId: 0,
					Expr: ConcatExpr{
						Children: []Expr{
							EventExpr{Event: 12},
							EventExpr{Event: 34},
						},
					},
				},
			},
			expected: &StateMachine{
				numStates: 3,
				acceptCmd: map[stateId]CmdId{
					1: 0,
				},
				transitions: map[stateId][]transition{
					0: {
						{
							eventRange: eventRange{start: 12, end: 12},
							nextState:  2,
						},
					},
					2: {
						{
							eventRange: eventRange{start: 34, end: 34},
							nextState:  1,
						},
					},
				},
			},
		},
		{
			name: "AltExpr",
			cmdExprs: []CmdExpr{
				{
					CmdId: 0,
					Expr: AltExpr{
						Children: []Expr{
							EventExpr{Event: 12},
							EventExpr{Event: 34},
						},
					},
				},
			},
			expected: &StateMachine{
				numStates: 2,
				acceptCmd: map[stateId]CmdId{
					1: 0,
				},
				transitions: map[stateId][]transition{
					0: {
						{
							eventRange: eventRange{start: 12, end: 12},
							nextState:  1,
						},
						{
							eventRange: eventRange{start: 34, end: 34},
							nextState:  1,
						},
					},
				},
			},
		},
		{
			name: "OptionExpr",
			cmdExprs: []CmdExpr{
				{
					CmdId: 0,
					Expr: OptionExpr{
						Child: EventExpr{Event: 99},
					},
				},
			},
			expected: &StateMachine{
				numStates: 2,
				acceptCmd: map[stateId]CmdId{
					0: 0,
					1: 0,
				},
				transitions: map[stateId][]transition{
					0: {
						{
							eventRange: eventRange{start: 99, end: 99},
							nextState:  1,
						},
					},
				},
			},
		},
		{
			name: "StarExpr",
			cmdExprs: []CmdExpr{
				{
					CmdId: 0,
					Expr: StarExpr{
						Child: EventExpr{Event: 99},
					},
				},
			},
			expected: &StateMachine{
				numStates: 1,
				acceptCmd: map[stateId]CmdId{
					0: 0,
				},
				transitions: map[stateId][]transition{
					0: {
						{
							eventRange: eventRange{start: 99, end: 99},
							nextState:  0,
						},
					},
				},
			},
		},
		{
			name: "CaptureExpr",
			cmdExprs: []CmdExpr{
				{
					CmdId: 0,
					Expr: CaptureExpr{
						Child: EventExpr{Event: 99},
					},
				},
			},
			expected: &StateMachine{
				numStates: 2,
				acceptCmd: map[stateId]CmdId{
					1: 0,
				},
				transitions: map[stateId][]transition{
					0: {
						{
							eventRange: eventRange{start: 99, end: 99},
							nextState:  1,
							captures: map[CmdId]CaptureId{
								0: 0,
							},
						},
					},
				},
			},
		},
		{
			name: "Multiple commands with overlapping transitions",
			cmdExprs: []CmdExpr{
				{
					CmdId: 0,
					Expr: EventRangeExpr{
						StartEvent: 2,
						EndEvent:   10,
					},
				},
				{
					CmdId: 1,
					Expr: EventRangeExpr{
						StartEvent: 0,
						EndEvent:   3,
					},
				},
				{
					CmdId: 2,
					Expr: EventRangeExpr{
						StartEvent: 8,
						EndEvent:   13,
					},
				},
			},
			expected: &StateMachine{
				numStates: 4,
				acceptCmd: map[stateId]CmdId{
					1: 1,
					2: 0,
					3: 2,
				},
				transitions: map[stateId][]transition{
					0: {
						{
							eventRange: eventRange{start: 0, end: 1},
							nextState:  1,
						},
						{
							eventRange: eventRange{start: 2, end: 3},
							nextState:  2,
						},
						{
							eventRange: eventRange{start: 4, end: 7},
							nextState:  2,
						},
						{
							eventRange: eventRange{start: 8, end: 10},
							nextState:  2,
						},
						{
							eventRange: eventRange{start: 11, end: 13},
							nextState:  3,
						},
					},
				},
			},
		},
		{
			name: "Multiple commands with same suffix",
			cmdExprs: []CmdExpr{
				{
					CmdId: 0,
					Expr: ConcatExpr{
						Children: []Expr{
							EventExpr{Event: 1},
							EventExpr{Event: 2},
							EventExpr{Event: 3},
						},
					},
				},
				{
					CmdId: 1,
					Expr: ConcatExpr{
						Children: []Expr{
							EventExpr{Event: 4},
							EventExpr{Event: 5},
							EventExpr{Event: 3},
						},
					},
				},
			},
			expected: &StateMachine{
				numStates: 7,
				acceptCmd: map[stateId]CmdId{
					1: 1,
					2: 0,
				},
				transitions: map[stateId][]transition{
					0: {
						{
							eventRange: eventRange{start: 1, end: 1},
							nextState:  3,
						},
						{
							eventRange: eventRange{start: 4, end: 4},
							nextState:  4,
						},
					},
					3: {
						{
							eventRange: eventRange{start: 2, end: 2},
							nextState:  6,
						},
					},
					4: {
						{
							eventRange: eventRange{start: 5, end: 5},
							nextState:  5,
						},
					},
					5: {
						{
							eventRange: eventRange{start: 3, end: 3},
							nextState:  1,
						},
					},
					6: {
						{
							eventRange: eventRange{start: 3, end: 3},
							nextState:  2,
						},
					},
				},
			},
		},
		{
			name: "Sequential overlapping transitions",
			cmdExprs: []CmdExpr{
				{
					CmdId: 0,
					Expr: EventRangeExpr{
						StartEvent: 0,
						EndEvent:   1,
					},
				},
				{
					CmdId: 1,
					Expr: EventRangeExpr{
						StartEvent: 1,
						EndEvent:   2,
					},
				},
				{
					CmdId: 2,
					Expr: EventRangeExpr{
						StartEvent: 2,
						EndEvent:   3,
					},
				},
			},
			expected: &StateMachine{
				numStates: 4,
				acceptCmd: map[stateId]CmdId{
					1: 0,
					2: 1,
					3: 2,
				},
				transitions: map[stateId][]transition{
					0: {
						{
							eventRange: eventRange{start: 0, end: 0},
							nextState:  1,
						},
						{
							eventRange: eventRange{start: 1, end: 1},
							nextState:  1,
						},
						{
							eventRange: eventRange{start: 2, end: 2},
							nextState:  2,
						},
						{
							eventRange: eventRange{start: 3, end: 3},
							nextState:  3,
						},
					},
				},
			},
		},
		{
			name: "Overlapping transitions same start, first is longer",
			cmdExprs: []CmdExpr{
				{
					CmdId: 0,
					Expr: EventRangeExpr{
						StartEvent: 0,
						EndEvent:   7,
					},
				},
				{
					CmdId: 1,
					Expr: EventRangeExpr{
						StartEvent: 0,
						EndEvent:   2,
					},
				},
			},
			expected: &StateMachine{
				numStates: 2,
				acceptCmd: map[stateId]CmdId{
					1: 0,
				},
				transitions: map[stateId][]transition{
					0: {
						{
							eventRange: eventRange{start: 0, end: 2},
							nextState:  1,
						},
						{
							eventRange: eventRange{start: 3, end: 7},
							nextState:  1,
						},
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			sm, err := Compile(tc.cmdExprs)
			require.NoError(t, err)
			assert.Equal(t, tc.expected, sm)
		})
	}
}

func TestCompileValidation(t *testing.T) {
	testCases := []struct {
		name     string
		cmdExprs []CmdExpr
		errMsg   string
	}{
		{
			name: "valid program",
			cmdExprs: []CmdExpr{
				{
					CmdId: 0,
					Expr:  EventExpr{Event: 1},
				},
			},
			errMsg: "",
		},
		{
			name: "nil expression",
			cmdExprs: []CmdExpr{
				{
					CmdId: 0,
					Expr:  nil,
				},
			},
			errMsg: "Invalid expression for cmd 0: Invalid expression type <nil>",
		},
		{
			name: "invalid event range",
			cmdExprs: []CmdExpr{
				{
					CmdId: 0,
					Expr: EventRangeExpr{
						StartEvent: 5,
						EndEvent:   4,
					},
				},
			},
			errMsg: "Invalid expression for cmd 0: Invalid event range [5, 4]",
		},
		{
			name: "duplicate command IDs",
			cmdExprs: []CmdExpr{
				{
					CmdId: 0,
					Expr:  EventExpr{Event: 1},
				},
				{
					CmdId: 0,
					Expr:  EventExpr{Event: 2},
				},
			},
			errMsg: "Duplicate command ID detected: 0",
		},
		{
			name: "two commands with the same capture ID",
			cmdExprs: []CmdExpr{
				{
					CmdId: 0,
					Expr: CaptureExpr{
						CaptureId: 1,
						Child:     EventExpr{Event: 1},
					},
				},
				{
					CmdId: 1,
					Expr: CaptureExpr{
						CaptureId: 1,
						Child:     EventExpr{Event: 1},
					},
				},
			},
			errMsg: "",
		},
		{
			name: "nested capture ID",
			cmdExprs: []CmdExpr{
				{
					CmdId: 0,
					Expr: CaptureExpr{
						CaptureId: 1,
						Child: AltExpr{
							Children: []Expr{
								CaptureExpr{
									CaptureId: 2,
									Child:     EventExpr{Event: 1},
								},
								EventExpr{Event: 2},
							},
						},
					},
				},
			},
			errMsg: "Invalid expression for cmd 0: Nested capture detected: 2",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := Compile(tc.cmdExprs)
			if tc.errMsg == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.errMsg)
			}
		})
	}
}
