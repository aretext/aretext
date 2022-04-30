package vm

import "fmt"

// Expr is a regular expression that matches input events.
type Expr interface{}

// EventExpr matches a single input event.
type EventExpr struct {
	Event Event
}

// EventRangeExpr matches any event in the range [start, end].
// The range boundaries are inclusive.
type EventRangeExpr struct {
	StartEvent Event
	EndEvent   Event
}

// ConcatExpr matches a sequence of expressions.
type ConcatExpr struct {
	Children []Expr
}

// AltExpr matches any of a set of expressions.
type AltExpr struct {
	Children []Expr
}

// OptionExpr matches a child expression or an empty input sequence.
type OptionExpr struct {
	Child Expr
}

// StarExpr matches zero or more repetitions of a child expression.
type StarExpr struct {
	Child Expr
}

// CaptureExpr includes the matched child expression in a capture with the specified ID.
// Captures may be nested.
// Each CaptureId must appear at most once within an expression.
type CaptureExpr struct {
	CaptureId CaptureId
	Child     Expr
}

func validateExpr(expr Expr) error {
	captureIds := make(map[CaptureId]struct{}, 0)
	stack := []Expr{expr}
	var current Expr
	for len(stack) > 0 {
		current, stack = stack[len(stack)-1], stack[0:len(stack)-1]
		switch expr := current.(type) {
		case EventExpr:
			break
		case EventRangeExpr:
			if expr.StartEvent >= expr.EndEvent {
				return fmt.Errorf("Invalid event range [%d, %d]", expr.StartEvent, expr.EndEvent)
			}
		case ConcatExpr:
			stack = append(stack, expr.Children...)
		case AltExpr:
			stack = append(stack, expr.Children...)
		case OptionExpr:
			stack = append(stack, expr.Child)
		case StarExpr:
			stack = append(stack, expr.Child)
		case CaptureExpr:
			if _, ok := captureIds[expr.CaptureId]; ok {
				return fmt.Errorf("Duplicate capture ID %d", expr.CaptureId)
			}
			captureIds[expr.CaptureId] = struct{}{}
			stack = append(stack, expr.Child)
		default:
			return fmt.Errorf("Invalid expression type %T", expr)
		}
	}
	return nil
}
