package breaks

// BreakIterator iterates through valid breakpoints in a string.
type BreakIterator interface {
	// NextBreak returns the next valid breakpoint.
	// The breakpoint is an offset from the beginning of the string the iterator has processed.
	NextBreak() (uint64, error)
}
