package vm

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCompilerValidation(t *testing.T) {
	testCases := []struct {
		name   string
		expr   Expr
		errMsg string
	}{
		{
			name:   "valid program",
			expr:   EventExpr{Event: 1},
			errMsg: "",
		},
		{
			name:   "nil expression",
			expr:   nil,
			errMsg: "Invalid expression type <nil>",
		},
		{
			name: "invalid event range",
			expr: EventRangeExpr{
				StartEvent: 5,
				EndEvent:   4,
			},
			errMsg: "Invalid event range [5, 4]",
		},
		{
			name: "duplicate capture ID, non-overlapping",
			expr: ConcatExpr{
				Children: []Expr{
					CaptureExpr{
						CaptureId: 1,
						Child:     EventExpr{Event: 1},
					},
					CaptureExpr{
						CaptureId: 1,
						Child:     EventExpr{Event: 1},
					},
				},
			},
			errMsg: "",
		},
		{
			name: "conflicting capture ID",
			expr: CaptureExpr{
				CaptureId: 1,
				Child: AltExpr{
					Children: []Expr{
						CaptureExpr{
							CaptureId: 1,
							Child:     EventExpr{Event: 1},
						},
						EventExpr{Event: 2},
					},
				},
			},
			errMsg: "Conflicting capture ID 1",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := Compile(tc.expr)
			if tc.errMsg == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.errMsg)
			}
		})
	}
}
