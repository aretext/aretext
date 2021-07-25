package parser

import (
	"math"

	"github.com/aretext/aretext/text"
)

// ParseFunc incrementally parses a document into tokens.
//
// It returns the number of tokens consumed and a slice of tokens.
// The output MUST be deterministic based solely on the input text.
//
// Each invocation of the function is cached and may be reused
// when reparsing the document after an edit.
//
// The returned tokens must be sequential, non-overlapping,
// have non-zero length, and have positions within the range
// of consumed characters.
//
// Every successful parse must consume at least one rune.
type ParseFunc func(text.CloneableRuneIter) (uint64, []ComputedToken)

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
	leafComputations := make([]*Computation, 0)
	n := tree.NumChars()
	for pos < n {
		c := p.runParseFunc(tree, pos)
		pos += c.ConsumedLength()
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
	n := tree.NumChars()
	for pos < n {
		nextComputation := p.findReusableComputation(pos, edit)
		if nextComputation == nil {
			nextComputation = p.runParseFunc(tree, pos)
		}
		pos += nextComputation.ConsumedLength()
		c = c.Append(nextComputation)
	}
	p.prevComputation = c
	return c
}

func (p *Parser) runParseFunc(tree *text.Tree, pos uint64) *Computation {
	reader := tree.ReaderAtPosition(pos, text.ReadDirectionForward)
	runeIter := text.NewCloneableForwardRuneIter(reader)
	trackingIter := NewTrackingRuneIter(runeIter)
	numConsumed, computedTokens := p.parseFunc(trackingIter)
	return NewComputation(trackingIter.MaxRead(), numConsumed, computedTokens)
}

func (p *Parser) findReusableComputation(pos uint64, edit Edit) *Computation {
	if pos < edit.pos {
		// If the parser is starting before the edit, look for a subcomputation
		// from that position up to the start of the edit.
		return p.prevComputation.LargestSubComputationInRange(pos, edit.pos)
	}

	if edit.numInserted > 0 && pos >= edit.pos+edit.numInserted {
		// If the parser is past the last character inserted,
		// translate the position to the previous document by subtracting
		// the number of inserted characters.
		return p.prevComputation.LargestSubComputationInRange(pos-edit.numInserted, math.MaxUint64)
	}

	if edit.numDeleted > 0 && pos >= edit.pos {
		// If the parser is past a deletion,
		// translate the position to the previous document by adding
		// the number of deleted characters.
		return p.prevComputation.LargestSubComputationInRange(pos+edit.numDeleted, math.MaxUint64)
	}

	// The parser is starting within the edit range, so we can can't re-use
	// any of the previous computation.
	return nil
}
