package parser

import (
	"math"

	"github.com/aretext/aretext/editor/text"
)

// Func incrementally parses a document into tokens.
//
// It returns the number of tokens consumed and a slice of tokens.
// The output MUST be deterministic based solely on the input args.
//
// Each invocation of the function is cached and may be reused
// when reparsing the document after an edit.
//
// The returned tokens must be sequential, non-overlapping,
// have non-zero length, and have positions within the range
// of consumed characters.
//
// Every successful parse must consume at least one rune.
//
// The state parameter allows the parse func to track state across invocations.
// The initial state is always EmptyState.  The parse func must return a non-nil
// state, which will be passed back to the parse func on the next invocation.
type Func func(TrackingRuneIter, State) Result

// Result represents the result of a single execution of a parse function.
type Result struct {
	NumConsumed    uint64
	ComputedTokens []ComputedToken
	NextState      State
}

// FailedResult represents a failed parse.
var FailedResult = Result{}

// IsSuccess returns whether the parse succeeded.
func (r Result) IsSuccess() bool {
	return r.NumConsumed > 0
}

// IsFailure returns whether the parse failed.
func (r Result) IsFailure() bool {
	return !r.IsSuccess()
}

// ShiftForward shifts the result offsets forward by the specified number of positions.
func (r Result) ShiftForward(n uint64) Result {
	if n > 0 {
		r.NumConsumed += n
		for i := 0; i < len(r.ComputedTokens); i++ {
			r.ComputedTokens[i].Offset += n
		}
	}
	return r
}

// P parses a document into tokens.
// It caches the results from the last parse so it can efficiently
// reparse a document after an edit (insertion/deletion).
type P struct {
	parseFunc       Func
	lastComputation *computation
}

// New constructs a new parser for the language recognized by parseFunc.
func New(f Func) *P {
	// This ensures that the parse func always makes progress.
	f = f.recoverFromFailure()
	return &P{parseFunc: f}
}

// TokenAtPosition returns the token containing a position.
// If no such token exists, it returns the Token zero value.
func (p *P) TokenAtPosition(pos uint64) Token {
	return p.lastComputation.TokenAtPosition(pos)
}

// TokensIntersectingRange returns tokens that overlap the interval [startPos, endPos)
func (p *P) TokensIntersectingRange(startPos, endPos uint64) []Token {
	return p.lastComputation.TokensIntersectingRange(startPos, endPos)
}

// Minimum consumed length for leaf computations on initial parse.
const minInitialConsumedLen = 1024

// ParseAll parses the entire document.
func (p *P) ParseAll(tree *text.Tree) {
	var pos uint64
	var prevComputation *computation
	state := State(EmptyState{})
	leafComputations := make([]*computation, 0)
	n := tree.NumChars()
	for pos < n {
		c := p.runParseFunc(tree, pos, state)
		pos += c.ConsumedLength()
		state = c.EndState()

		if prevComputation != nil && prevComputation.ConsumedLength() < minInitialConsumedLen {
			// For the initial parse, combine small leaves. This saves memory by reducing both
			// the number of leaves and parent nodes we need to allocate.
			combineLeaves(prevComputation, c)
		} else {
			leafComputations = append(leafComputations, c)
			prevComputation = c
		}
	}
	c := concatLeafComputations(leafComputations)
	p.lastComputation = c
}

func combineLeaves(prev, next *computation) {
	for _, tok := range next.tokens {
		tok.Offset += prev.consumedLength
		prev.tokens = append(prev.tokens, tok)
	}
	prev.consumedLength += next.consumedLength
	prev.readLength += next.readLength
	prev.endState = next.endState
}

// ReparseAfterEdit parses a document after an edit (insertion/deletion),
// re-using cached results from previous computations when possible.
// This should be called *after* at least one invocation of ParseAll().
// It must be called for *every* edit to the document, otherwise the
// tokens may not match the current state of the document.
func (p *P) ReparseAfterEdit(tree *text.Tree, edit Edit) {
	var pos uint64
	var c *computation
	state := State(EmptyState{})
	n := tree.NumChars()
	for pos < n {
		nextComputation := p.findReusableComputation(pos, edit, state)
		if nextComputation == nil {
			nextComputation = p.runParseFunc(tree, pos, state)
		}
		state = nextComputation.EndState()
		pos += nextComputation.ConsumedLength()
		c = c.Append(nextComputation)
	}
	p.lastComputation = c
}

func (p *P) runParseFunc(tree *text.Tree, pos uint64, state State) *computation {
	reader := tree.ReaderAtPosition(pos)
	trackingIter := NewTrackingRuneIter(reader)
	result := p.parseFunc(trackingIter, state)
	return newComputation(
		trackingIter.MaxRead(),
		result.NumConsumed,
		state,
		result.NextState,
		result.ComputedTokens,
	)
}

func (p *P) findReusableComputation(pos uint64, edit Edit, state State) *computation {
	if pos < edit.pos {
		// If the parser is starting before the edit, look for a subcomputation
		// from that position up to the start of the edit.
		return p.lastComputation.LargestMatchingSubComputation(
			pos,
			edit.pos,
			state,
		)
	}

	if edit.numInserted > 0 && pos >= edit.pos+edit.numInserted {
		// If the parser is past the last character inserted,
		// translate the position to the previous document by subtracting
		// the number of inserted characters.
		return p.lastComputation.LargestMatchingSubComputation(
			pos-edit.numInserted,
			math.MaxUint64,
			state,
		)
	}

	if edit.numDeleted > 0 && pos >= edit.pos {
		// If the parser is past a deletion,
		// translate the position to the previous document by adding
		// the number of deleted characters.
		return p.lastComputation.LargestMatchingSubComputation(
			pos+edit.numDeleted,
			math.MaxUint64,
			state,
		)
	}

	// The parser is starting within the edit range, so we can can't re-use
	// any of the last computation.
	return nil
}
