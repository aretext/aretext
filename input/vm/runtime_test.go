package vm

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
		name  string
		expr  Expr
		steps []step
	}{
		{
			name: "EventExpr",
			expr: EventExpr{
				Event: 123,
			},
			steps: []step{
				{
					event:  123,
					result: Result{Accepted: true, Reset: true},
				},
				{
					event:  9,
					result: Result{Accepted: false, Reset: true},
				},
				{
					event:  123,
					result: Result{Accepted: true, Reset: true},
				},
			},
		},
		{
			name: "EventRangeExpr",
			expr: EventRangeExpr{
				StartEvent: 3,
				EndEvent:   5,
			},
			steps: []step{
				{
					event:  2,
					result: Result{Accepted: false, Reset: true},
				},
				{
					event:  3,
					result: Result{Accepted: true, Reset: true},
				},
				{
					event:  4,
					result: Result{Accepted: true, Reset: true},
				},
				{
					event:  5,
					result: Result{Accepted: true, Reset: true},
				},
				{
					event:  6,
					result: Result{Accepted: false, Reset: true},
				},
			},
		},
		{
			name: "ConcatExpr",
			expr: ConcatExpr{
				Children: []Expr{
					EventExpr{Event: 2},
					EventExpr{Event: 3},
					EventExpr{Event: 4},
				},
			},
			steps: []step{
				{
					event:  1,
					result: Result{Accepted: false, Reset: true},
				},
				{
					event:  2,
					result: Result{Accepted: false, Reset: false},
				},
				{
					event:  3,
					result: Result{Accepted: false, Reset: false},
				},
				{
					event:  4,
					result: Result{Accepted: true, Reset: true},
				},
			},
		},
		{
			name: "AltExpr",
			expr: AltExpr{
				Children: []Expr{
					EventExpr{Event: 2},
					EventExpr{Event: 3},
				},
			},
			steps: []step{
				{
					event:  1,
					result: Result{Accepted: false, Reset: true},
				},
				{
					event:  2,
					result: Result{Accepted: true, Reset: true},
				},
				{
					event:  3,
					result: Result{Accepted: true, Reset: true},
				},
				{
					event:  4,
					result: Result{Accepted: false, Reset: true},
				},
			},
		},
		{
			name: "OptionExpr",
			expr: ConcatExpr{
				Children: []Expr{
					OptionExpr{
						Child: EventExpr{Event: 3},
					},
					EventExpr{Event: 4},
				},
			},
			steps: []step{
				{
					event:  4,
					result: Result{Accepted: true, Reset: true},
				},
				{
					event:  3,
					result: Result{Accepted: false, Reset: false},
				},
				{
					event:  4,
					result: Result{Accepted: true, Reset: true},
				},
				{
					event:  4,
					result: Result{Accepted: true, Reset: true},
				},
				{
					event:  5,
					result: Result{Accepted: false, Reset: true},
				},
			},
		},
		{
			name: "StarExpr",
			expr: ConcatExpr{
				Children: []Expr{
					StarExpr{Child: EventExpr{Event: 3}},
					EventExpr{Event: 4},
				},
			},
			steps: []step{
				{
					event:  3,
					result: Result{Accepted: false, Reset: false},
				},
				{
					event:  3,
					result: Result{Accepted: false, Reset: false},
				},
				{
					event:  3,
					result: Result{Accepted: false, Reset: false},
				},
				{
					event:  4,
					result: Result{Accepted: true, Reset: true},
				},
				{
					event:  4,
					result: Result{Accepted: true, Reset: true},
				},
			},
		},
		{
			name: "CaptureExpr",
			expr: ConcatExpr{
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
			steps: []step{
				{
					event:  1,
					result: Result{Accepted: false, Reset: false},
				},
				{
					event:  1,
					result: Result{Accepted: false, Reset: false},
				},
				{
					event: 2,
					result: Result{
						Accepted: true,
						Reset:    true,
						Captures: []Capture{
							{Id: 1, StartIdx: 0, Length: 2},
							{Id: 2, StartIdx: 2, Length: 1},
						},
					},
				},
			},
		},
		{
			name: "nested CaptureExpr",
			expr: CaptureExpr{
				CaptureId: 1,
				Child: CaptureExpr{
					CaptureId: 2,
					Child:     EventExpr{Event: 1},
				},
			},
			steps: []step{
				{
					event: 1,
					result: Result{
						Accepted: true,
						Reset:    true,
						Captures: []Capture{
							{Id: 1, StartIdx: 0, Length: 1},
							{Id: 2, StartIdx: 0, Length: 1},
						},
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			program, err := Compile(tc.expr)
			require.NoError(t, err)
			runtime := NewRuntime(program)
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
