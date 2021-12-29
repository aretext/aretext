package languages

import (
	"io"
	"strings"

	"github.com/aretext/aretext/syntax/parser"
)

// initialState sets the initial parser state if it hasn't yet been set.
func initialState(initialState parser.State, f parser.Func) parser.Func {
	return func(iter parser.TrackingRuneIter, state parser.State) parser.Result {
		if state.Equals(parser.EmptyState{}) {
			state = initialState
		}
		return f(iter, state)
	}
}

// matchState executes `f` only if the parser state matches `targetState`.
func matchState(targetState parser.State, f parser.Func) parser.Func {
	return func(iter parser.TrackingRuneIter, state parser.State) parser.Result {
		if !state.Equals(targetState) {
			return parser.FailedResult
		}
		return f(iter, state)
	}
}

// setState sets the next parser state to `targetState`.
func setState(targetState parser.State) parser.MapFn {
	return func(result parser.Result) parser.Result {
		return parser.Result{
			NumConsumed:    result.NumConsumed,
			ComputedTokens: result.ComputedTokens,
			NextState:      targetState,
		}
	}
}

// consumeString consumes the characters in `s`.
func consumeString(s string) parser.Func {
	return func(iter parser.TrackingRuneIter, state parser.State) parser.Result {
		var numConsumed uint64
		for _, targetRune := range s {
			r, err := iter.NextRune()
			if err != nil || r != targetRune {
				return parser.FailedResult
			}
			numConsumed++
		}
		return parser.Result{
			NumConsumed: numConsumed,
			NextState:   state,
		}
	}
}

// consumeToString consumes all characters up to and including the string `s`.
func consumeToString(s string) parser.Func {
	f := consumeString(s)
	return func(iter parser.TrackingRuneIter, state parser.State) parser.Result {
		var numSkipped uint64
		for {
			r := f(iter, state)
			if r.IsSuccess() {
				return r.ShiftForward(numSkipped)
			}

			_, err := iter.NextRune()
			if err != nil {
				return parser.FailedResult
			}
			numSkipped++
		}
	}
}

// consumeSingleRuneLike consumes a single rune matching a predicate.
func consumeSingleRuneLike(predicateFn func(rune) bool) parser.Func {
	return func(iter parser.TrackingRuneIter, state parser.State) parser.Result {
		r, err := iter.NextRune()
		if err == nil && predicateFn(r) {
			return parser.Result{
				NumConsumed: 1,
				NextState:   state,
			}
		}
		return parser.FailedResult
	}
}

// consumeRunesLike consumes one or more runes matching a predicate.
func consumeRunesLike(predicateFn func(rune) bool) parser.Func {
	return func(iter parser.TrackingRuneIter, state parser.State) parser.Result {
		var numConsumed uint64
		for {
			r, err := iter.NextRune()
			if err != nil || !predicateFn(r) {
				return parser.Result{
					NumConsumed: numConsumed,
					NextState:   state,
				}
			}
			numConsumed++
		}
	}
}

// consumeUntilEofOrRuneLike consumes up to, but not including, a rune matching a predicate or EOF.
func consumeUntilEofOrRuneLike(predicate func(r rune) bool) parser.Func {
	return func(iter parser.TrackingRuneIter, state parser.State) parser.Result {
		var numConsumed uint64
		for {
			r, err := iter.NextRune()
			if err == io.EOF {
				break
			} else if err != nil {
				return parser.FailedResult
			} else if predicate(r) {
				break
			}
			numConsumed++
		}

		return parser.Result{
			NumConsumed: numConsumed,
			NextState:   state,
		}
	}
}

// consumeToEofOrRuneLike consumes up to and including a rune matching a predicate or EOF.
func consumeToEofOrRuneLike(predicate func(r rune) bool) parser.Func {
	return func(iter parser.TrackingRuneIter, state parser.State) parser.Result {
		var numConsumed uint64
		for {
			r, err := iter.NextRune()
			if err == io.EOF {
				break
			} else if err != nil {
				return parser.FailedResult
			}

			numConsumed++

			if predicate(r) {
				break
			}
		}
		return parser.Result{
			NumConsumed: numConsumed,
			NextState:   state,
		}
	}
}

// consumeToNextLineFeed consumes up to and including the next newline character or the last character in the document, whichever comes first.
var consumeToNextLineFeed = consumeToEofOrRuneLike(func(r rune) bool {
	return r == '\n'
})

func consumeDigitsAndSeparators(allowLeadingSeparator bool, isDigit func(r rune) bool) parser.Func {
	return func(iter parser.TrackingRuneIter, state parser.State) parser.Result {
		var numConsumed uint64
		var lastWasUnderscore bool
		for {
			r, err := iter.NextRune()
			if err != nil {
				break
			}

			if r == '_' && !lastWasUnderscore && (allowLeadingSeparator || numConsumed > 0) {
				lastWasUnderscore = true
				numConsumed++
				continue
			}

			if isDigit(r) {
				lastWasUnderscore = false
				numConsumed++
				continue
			}

			break
		}

		if lastWasUnderscore {
			numConsumed--
		}

		return parser.Result{
			NumConsumed: numConsumed,
			NextState:   state,
		}
	}

}

// recognizeToken recognizes the consumed characters in the result as a token.
func recognizeToken(tokenRole parser.TokenRole) parser.MapFn {
	return func(result parser.Result) parser.Result {
		token := parser.ComputedToken{
			Length: result.NumConsumed,
			Role:   tokenRole,
		}
		return parser.Result{
			NumConsumed:    result.NumConsumed,
			ComputedTokens: []parser.ComputedToken{token},
			NextState:      result.NextState,
		}
	}
}

func maxStrLen(ss []string) uint64 {
	maxLength := uint64(0)
	for _, s := range ss {
		length := uint64(len(s))
		if length > maxLength {
			maxLength = length
		}
	}
	return maxLength
}

// recognizeKeywordOrConsume recognizes a keyword from the list of `keywords`.
// If no keywords match, the result is returned unmodified.
func recognizeKeywordOrConsume(keywords []string) parser.MapWithInputFn {
	// Calculate the length of the longest keyword to limit how much
	// of the input needs to be reprocessed.
	maxLength := maxStrLen(keywords)
	return func(result parser.Result, iter parser.TrackingRuneIter, state parser.State) parser.Result {
		if result.NumConsumed > maxLength {
			return result
		}

		s := readInputString(iter, result.NumConsumed)
		for _, kw := range keywords {
			if kw == s {
				token := parser.ComputedToken{
					Role:   parser.TokenRoleKeyword,
					Length: result.NumConsumed,
				}
				return parser.Result{
					NumConsumed:    result.NumConsumed,
					ComputedTokens: []parser.ComputedToken{token},
					NextState:      state,
				}
			}
		}

		return result
	}
}

// failIfMatchTerm fails if the consumed string matches any of the excluded terms.
// Otherwise, it returns the result unmodified.
func failIfMatchTerm(terms []string) parser.MapWithInputFn {
	maxLength := maxStrLen(terms)
	return func(result parser.Result, iter parser.TrackingRuneIter, state parser.State) parser.Result {
		if result.NumConsumed > maxLength {
			return result
		}
		s := readInputString(iter, result.NumConsumed)
		for _, term := range terms {
			if term == s {
				return parser.FailedResult
			}
		}
		return result
	}
}

// readInputString reads a string from the text up to `n` characters long.
func readInputString(iter parser.TrackingRuneIter, n uint64) string {
	var sb strings.Builder
	for i := uint64(0); i < n; i++ {
		r, err := iter.NextRune()
		if err != nil {
			break
		}
		if _, err := sb.WriteRune(r); err != nil {
			panic(err)
		}
	}
	return sb.String()
}

// consumeCStyleString consumes a string with characters escaped by a backslash.
func consumeCStyleString(quoteRune rune, allowLineBreaks bool) parser.Func {
	return func(iter parser.TrackingRuneIter, state parser.State) parser.Result {
		var n uint64
		r, err := iter.NextRune()
		if err != nil || r != quoteRune {
			return parser.FailedResult
		}
		n++

		var inEscapeSeq bool
		for {
			r, err = iter.NextRune()
			if err != nil || (!allowLineBreaks && r == '\n') {
				return parser.FailedResult
			}
			n++

			if r == quoteRune && !inEscapeSeq {
				return parser.Result{
					NumConsumed: n,
					ComputedTokens: []parser.ComputedToken{
						{Length: n},
					},
					NextState: state,
				}
			}

			if r == '\\' && !inEscapeSeq {
				inEscapeSeq = true
				continue
			}

			if inEscapeSeq {
				inEscapeSeq = false
			}
		}
	}
}

// parseCStyleString parses a string with characters escaped by a backslash.
func parseCStyleString(quoteRune rune, allowLineBreaks bool) parser.Func {
	return consumeCStyleString(quoteRune, allowLineBreaks).
		Map(recognizeToken(parser.TokenRoleString))
}
