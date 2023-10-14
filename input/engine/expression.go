package engine

// Event is an input event.
// This usually represents a keypress, but the compiled state machine doesn't assume
// that the events have any particular meaning.
type Event int64

// CaptureId is an identifier for a subsequence of events.
type CaptureId uint64

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
// Captures must NOT be nested.
type CaptureExpr struct {
	CaptureId CaptureId
	Child     Expr
}
