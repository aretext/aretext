package vm

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
