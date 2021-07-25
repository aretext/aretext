package parser

import (
	"math"

	"github.com/aretext/aretext/text"
)

// ParseFunc incrementally parses a document into tokens.
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
type ParseFunc func(text.CloneableRuneIter, State) (uint64, []ComputedToken, State)

// Parser parses a document into tokens.
// It caches the results from the last parse so it can efficiently
// reparse a document after an edit (insertion/deletion).
type Parser struct {
	parseFunc       ParseFunc
	prevComputation *Computation
}

// NewParser constructs a new parser for the language recognized by parseFunc.
func NewParser(parseFunc ParseFunc) *Parser {
	return &Parser{parseFunc: parseFunc}
}

// ParseAll parses the entire document.
func (p *Parser) ParseAll(tree *text.Tree) *Computation {
	var pos uint64
	state := State(EmptyState{})
	leafComputations := make([]*Computation, 0)
	n := tree.NumChars()
	for pos < n {
		c := p.runParseFunc(tree, pos, state)
		pos += c.ConsumedLength()
		state = c.EndState()
		leafComputations = append(leafComputations, c)
	}
	c := ConcatLeafComputations(leafComputations)
	p.prevComputation = c
	return c
}

// ReparseAfterEdit parses a document after an edit (insertion/deletion),
// re-using cached results from previous computations when possible.
// This should be called *after* at least one invocation of ParseAll().
// It must be called for *every* edit to the document, otherwise the
// tokens may not match the current state of the document.
func (p *Parser) ReparseAfterEdit(tree *text.Tree, edit Edit) *Computation {
	var pos uint64
	var c *Computation
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
	p.prevComputation = c
	return c
}

func (p *Parser) runParseFunc(tree *text.Tree, pos uint64, state State) *Computation {
	reader := tree.ReaderAtPosition(pos, text.ReadDirectionForward)
	runeIter := text.NewCloneableForwardRuneIter(reader)
	trackingIter := NewTrackingRuneIter(runeIter)
	numConsumed, computedTokens, nextState := p.parseFunc(trackingIter, state)
	return NewComputation(
		trackingIter.MaxRead(),
		numConsumed,
		state,
		nextState,
		computedTokens,
	)
}

func (p *Parser) findReusableComputation(pos uint64, edit Edit, state State) *Computation {
	if pos < edit.pos {
		// If the parser is starting before the edit, look for a subcomputation
		// from that position up to the start of the edit.
		return p.prevComputation.LargestSubComputationInRange(
			pos,
			edit.pos,
			state,
		)
	}

	if edit.numInserted > 0 && pos >= edit.pos+edit.numInserted {
		// If the parser is past the last character inserted,
		// translate the position to the previous document by subtracting
		// the number of inserted characters.
		return p.prevComputation.LargestSubComputationInRange(
			pos-edit.numInserted,
			math.MaxUint64,
			state,
		)
	}

	if edit.numDeleted > 0 && pos >= edit.pos {
		// If the parser is past a deletion,
		// translate the position to the previous document by adding
		// the number of deleted characters.
		return p.prevComputation.LargestSubComputationInRange(
			pos+edit.numDeleted,
			math.MaxUint64,
			state,
		)
	}

	// The parser is starting within the edit range, so we can can't re-use
	// any of the previous computation.
	return nil
}
