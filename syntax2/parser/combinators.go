package parser

import "github.com/aretext/aretext/text"

// recoverFromFailure consumes runes until the first successful parse.
func (f Func) recoverFromFailure() Func {
	return func(iter text.CloneableRuneIter, state State) Result {
		var numSkipped uint64
		for {
			result := f(iter.Clone(), state)
			if result.IsSuccess() {
				if numSkipped > 0 {
					// Shift the result to consume skipped runes.
					result.NumConsumed += numSkipped
					for i := 0; i < len(result.ComputedTokens); i++ {
						result.ComputedTokens[i].Offset += numSkipped
					}
				}
				return result
			}

			// Recover by skipping one rune ahead.
			n := advanceIter(iter, 1)
			numSkipped += n
			if n == 0 {
				return Result{
					NumConsumed: numSkipped,
					NextState:   state,
				}
			}
		}
	}
}

func advanceIter(iter text.CloneableRuneIter, n uint64) uint64 {
	for i := uint64(0); i < n; i++ {
		_, err := iter.NextRune()
		if err != nil {
			return i
		}
	}
	return n
}
