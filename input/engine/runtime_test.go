package engine

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type step struct {
	event  Event
	result Result
}

func TestCompileAndRun(t *testing.T) {
	testCases := []struct {
		name     string
		cmdExprs []CmdExpr
		steps    []step
	}{
		{
			name: "EventExpr",
			cmdExprs: []CmdExpr{
				{
					CmdId: 0,
					Expr: EventExpr{
						Event: 123,
					},
				},
			},
			steps: []step{
				{
					event:  123,
					result: Result{Decision: DecisionAccept},
				},
				{
					event:  9,
					result: Result{Decision: DecisionReject},
				},
				{
					event:  123,
					result: Result{Decision: DecisionAccept},
				},
			},
		},
		{
			name: "EventRangeExpr",
			cmdExprs: []CmdExpr{
				{
					CmdId: 0,
					Expr: EventRangeExpr{
						StartEvent: 3,
						EndEvent:   5,
					},
				},
			},
			steps: []step{
				{
					event:  2,
					result: Result{Decision: DecisionReject},
				},
				{
					event:  3,
					result: Result{Decision: DecisionAccept},
				},
				{
					event:  4,
					result: Result{Decision: DecisionAccept},
				},
				{
					event:  5,
					result: Result{Decision: DecisionAccept},
				},
				{
					event:  6,
					result: Result{Decision: DecisionReject},
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
							EventExpr{Event: 2},
							EventExpr{Event: 3},
							EventExpr{Event: 4},
						},
					},
				},
			},
			steps: []step{
				{
					event:  1,
					result: Result{Decision: DecisionReject},
				},
				{
					event:  2,
					result: Result{Decision: DecisionWait},
				},
				{
					event:  3,
					result: Result{Decision: DecisionWait},
				},
				{
					event:  4,
					result: Result{Decision: DecisionAccept},
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
							EventExpr{Event: 2},
							EventExpr{Event: 3},
						},
					},
				},
			},
			steps: []step{
				{
					event:  1,
					result: Result{Decision: DecisionReject},
				},
				{
					event:  2,
					result: Result{Decision: DecisionAccept},
				},
				{
					event:  3,
					result: Result{Decision: DecisionAccept},
				},
				{
					event:  4,
					result: Result{Decision: DecisionReject},
				},
			},
		},
		{
			name: "OptionExpr",
			cmdExprs: []CmdExpr{
				{
					CmdId: 0,
					Expr: ConcatExpr{
						Children: []Expr{
							OptionExpr{
								Child: EventExpr{Event: 3},
							},
							EventExpr{Event: 4},
						},
					},
				},
			},
			steps: []step{
				{
					event:  4,
					result: Result{Decision: DecisionAccept},
				},
				{
					event:  3,
					result: Result{Decision: DecisionWait},
				},
				{
					event:  4,
					result: Result{Decision: DecisionAccept},
				},
				{
					event:  4,
					result: Result{Decision: DecisionAccept},
				},
				{
					event:  5,
					result: Result{Decision: DecisionReject},
				},
			},
		},
		{
			name: "StarExpr",
			cmdExprs: []CmdExpr{
				{
					CmdId: 0,
					Expr: ConcatExpr{
						Children: []Expr{
							StarExpr{Child: EventExpr{Event: 3}},
							EventExpr{Event: 4},
						},
					},
				},
			},
			steps: []step{
				{
					event:  3,
					result: Result{Decision: DecisionWait},
				},
				{
					event:  3,
					result: Result{Decision: DecisionWait},
				},
				{
					event:  3,
					result: Result{Decision: DecisionWait},
				},
				{
					event:  4,
					result: Result{Decision: DecisionAccept},
				},
				{
					event:  4,
					result: Result{Decision: DecisionAccept},
				},
			},
		},
		{
			name: "CaptureExpr",
			cmdExprs: []CmdExpr{
				{
					CmdId: 0,
					Expr: ConcatExpr{
						Children: []Expr{
							CaptureExpr{
								CaptureId: 1,
								Child: StarExpr{
									Child: EventExpr{Event: 1},
								},
							},
							CaptureExpr{
								CaptureId: 2,
								Child:     EventExpr{Event: 2},
							},
						},
					},
				},
			},
			steps: []step{
				{
					event:  1,
					result: Result{Decision: DecisionWait},
				},
				{
					event:  1,
					result: Result{Decision: DecisionWait},
				},
				{
					event: 2,
					result: Result{
						Decision: DecisionAccept,
						Captures: map[CaptureId][]Event{
							1: {1, 1},
							2: {2},
						},
					},
				},
			},
		},
		{
			name: "multiple commands",
			cmdExprs: []CmdExpr{
				{
					CmdId: 0,
					Expr: EventExpr{
						Event: 11,
					},
				},
				{
					CmdId: 1,
					Expr: EventExpr{
						Event: 22,
					},
				},
				{
					CmdId: 2,
					Expr: EventExpr{
						Event: 33,
					},
				},
			},
			steps: []step{
				{
					event:  22,
					result: Result{Decision: DecisionAccept, CmdId: 1},
				},
				{
					event:  11,
					result: Result{Decision: DecisionAccept, CmdId: 0},
				},
				{
					event:  33,
					result: Result{Decision: DecisionAccept, CmdId: 2},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			sm, err := Compile(tc.cmdExprs)
			require.NoError(t, err)
			runtime := NewRuntime(sm, 1024)
			for i, step := range tc.steps {
				result := runtime.ProcessEvent(step.event)
				assert.Equal(
					t, step.result, result,
					"After processing event %v from step %d",
					step.event, i,
				)
			}
		})
	}
}

func TestRuntimeMaxInputLen(t *testing.T) {
	cmdExprs := []CmdExpr{
		{
			CmdId: 0,
			Expr: ConcatExpr{
				Children: []Expr{
					StarExpr{Child: EventExpr{Event: 1}},
					EventExpr{Event: 2},
				},
			},
		},
	}
	sm, err := Compile(cmdExprs)
	require.NoError(t, err)

	const maxInputLen = 8
	runtime := NewRuntime(sm, maxInputLen)

	// Events before max input length should continue.
	for i := 0; i < maxInputLen; i++ {
		result := runtime.ProcessEvent(1)
		assert.Equal(t, DecisionWait, result.Decision, "Should wait at iteration %d", i)
	}

	// First event past max input length should reject and reset.
	result := runtime.ProcessEvent(1)
	assert.Equal(t, DecisionReject, result.Decision)
}
