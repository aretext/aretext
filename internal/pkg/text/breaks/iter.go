package breaks

import (
	"io"
)

// BreakIter iterates through valid breakpoints in a string.
type BreakIter interface {
	// NextBreak returns the next valid breakpoint.
	// The breakpoint is an offset from the beginning of the string the iterator has processed.
	// If there are no more breakpoints, NextBreak returns an io.EOF error.
	NextBreak() (uint64, error)
}

// SkipBreak advances past the next breakpoint.
// If there are no more breakpoints, this is a no-op.
func SkipBreak(iter BreakIter) error {
	if _, err := iter.NextBreak(); err != nil && err != io.EOF {
		return err
	}
	return nil
}
