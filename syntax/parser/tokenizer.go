package parser

import (
	"io"
	"log"

	"github.com/pkg/errors"

	"github.com/aretext/aretext/text/utf8"
)

// Edit represents a change to a document.
type Edit struct {
	Pos         uint64 // Position of the first character inserted/deleted.
	NumInserted uint64
	NumDeleted  uint64
}

// TokenizerRule represents a rule for parsing a particular token.
type TokenizerRule struct {
	Regexp    string
	TokenRole TokenRole
	SubRules  []TokenizerRule
}

// Tokenizer parses a text into tokens based on a set of rules.
type Tokenizer struct {
	StateMachine  *Dfa
	SubTokenizers []*Tokenizer
	Rules         []TokenizerRule
	buf           [4]byte
}

// GenerateTokenizer compiles a tokenizer from a set of rules.
func GenerateTokenizer(rules []TokenizerRule) (*Tokenizer, error) {
	subTokenizers := make([]*Tokenizer, len(rules))
	nfa := EmptyLanguageNfa()
	for i, r := range rules {
		regexp, err := ParseRegexp(r.Regexp)
		if err != nil {
			return nil, errors.Wrapf(err, "parse rule regexp")
		}
		ruleNfa := regexp.CompileNfa().SetAcceptAction(i)
		nfa = nfa.Union(ruleNfa)

		if len(r.SubRules) > 0 {
			subTokenizers[i], err = GenerateTokenizer(r.SubRules)
			if err != nil {
				return nil, err
			}
		}
	}

	tokenizer := &Tokenizer{
		StateMachine:  nfa.CompileDfa(),
		SubTokenizers: subTokenizers,
		Rules:         rules,
	}

	return tokenizer, nil
}

// TokenizeAll splits the entire input text into tokens.
// The input text MUST be valid UTF-8.
func (t *Tokenizer) TokenizeAll(r InputReader, textLen uint64) (*TokenTree, error) {
	var tokens []Token
	pos := uint64(0)
	for pos < textLen {
		nextPos, nextTokens, err := t.nextTokens(r, textLen, pos)
		if err != nil {
			return nil, err
		}
		tokens = append(tokens, nextTokens...)
		pos = nextPos
	}

	// By construction, tokens are ordered ascending by start position and non-overlapping.
	return NewTokenTree(tokens), nil
}

// ReaderAtPosFunc returns a reader at the requested position.
type ReaderAtPosFunc func(pos uint64) InputReader

// RetokenizeAfterEdit updates tokens based on an edit to the text.
// The algorithm is based on Wagner (1998) Practical Algorithms for Incremental Software Development Environments, Chapter 5.
// This method assumes that the token tree is up-to-date with the text before the most recent edit; if not, it may panic.
func (t *Tokenizer) RetokenizeAfterEdit(tree *TokenTree, edit Edit, textLen uint64, readerAtPos ReaderAtPosFunc) (*TokenTree, error) {
	// The edit (insert/delete) changes the start/end/lookahead positions of tokens after the edit.
	// We need to update the token positions to match the edited text.
	// We do this operation first so that we can compare the positions of tokens in the tree with
	// the positions of new tokens parsed from the edited text.
	t.shiftPositionsAfterEdit(tree, edit)

	// Find the start position of the first affected token in the tree.
	// We will start retokenizing from this position.
	// This assumes that the intervals (StartPos, LookaheadPos] of tokens in the tree cover the edit position,
	// which should be true because every position is covered by either an token from a DFA match
	// or an empty token (TokenRoleNone) from error recovery.
	var pos uint64
	var tok Token
	iter := tree.iterFromFirstAffected(edit.Pos, IterDirectionForward)
	if iter.Get(&tok) {
		pos = tok.StartPos
	} else {
		iter = tree.IterFromPosition(0, IterDirectionForward)
	}

	// Retrieve a reader from the start of the first affected token in the tree.
	r := readerAtPos(pos)

	// Run the tokenizer from the start position.
	// This will generate new tokens that replace existing tokens in the tree.
	// We continue until we encounter a set of tokens after the edit that match the next generated set of tokens.
	// At this point, we know all subsequent tokens in the tree match what the tokenizer would output,
	// so we can stop retokenizing.
	startPos := pos
	var newTokens []Token
	for pos < textLen {
		nextPos, nextTokens, err := t.nextTokens(r, textLen, pos)
		if err != nil {
			return nil, errors.Wrapf(err, "next token")
		}

		if len(nextTokens) > 0 && nextTokens[0].StartPos > edit.Pos+edit.NumInserted {
			var mismatchDetected bool
			for _, nextTok := range nextTokens {
				// Advance to the next existing token on or after the next new token's start position.
				advanceIterToPos(iter, nextTok.StartPos)

				var oldTok Token
				if iter.Get(&oldTok) && oldTok == nextTok {
					continue
				}

				mismatchDetected = true
				break
			}

			// The new tokens exactly match the existing tokens after the edit,
			// so all subsequent tokens would match as well.  This means we're done!
			if !mismatchDetected {
				break
			}
		}

		// Advance past all tokens that overlap with the new token.
		advanceIterToPos(iter, nextPos)

		newTokens = append(newTokens, nextTokens...)
		pos = nextPos
	}

	// Rewrite existing tokens in the tree with the new tokens.
	t.rewriteTokens(tree, startPos, pos, newTokens)
	return tree, nil
}

func (t *Tokenizer) shiftPositionsAfterEdit(tree *TokenTree, edit Edit) {
	treeTextLen := tree.textLen()
	if edit.Pos > treeTextLen {
		panic("Edit cannot occur past end of text in token tree")
	}

	if edit.NumInserted > 0 {
		if edit.Pos == treeTextLen {
			// If the edit occurred at the end of the text, insert a placeholder token.
			tree.insertToken(Token{
				StartPos:     treeTextLen,
				EndPos:       treeTextLen + edit.NumInserted,
				LookaheadPos: treeTextLen + edit.NumInserted,
				Role:         TokenRoleNone,
			})
		} else {
			// If the edit occurred within the text, there must be some token in the tree that overlaps the edit.
			// Extend the length of that token so subsequent token positions match the updated text.
			tree.extendTokenIntersectingPos(edit.Pos, edit.NumInserted)
		}
	}

	if edit.NumDeleted > 0 {
		if treeTextLen-edit.Pos < edit.NumDeleted {
			panic("Not enough characters to delete")
		}
		tree.deleteRange(edit.Pos, edit.NumDeleted)
	}
}

// nextTokens runs the top-level DFA to produce a single "next" token (which may be empty).
// If the rule that produced the token has sub-rules, then the token is further tokenized to potentially
// produce more tokens, and so on recursively until all relevant sub-tokenizers have executed.
func (t *Tokenizer) nextTokens(r InputReader, textLen uint64, pos uint64) (uint64, []Token, error) {
	emptyToken := Token{
		StartPos:     pos,
		EndPos:       pos,
		LookaheadPos: pos + 1,
		Role:         TokenRoleNone,
	}

	for pos < textLen {
		accepted, endPos, lookaheadPos, actions, numBytesReadAtLastAccept, err := t.StateMachine.MatchLongest(r, pos, textLen)
		if err != nil {
			return 0, nil, errors.Wrapf(err, "tokenizing input")
		}

		// Identify a token if the DFA accepts AND the accepted prefix is non-empty.
		// (An empty prefix can occur when "a*" matches "" at the beginning of "bcd")
		if accepted && endPos > pos {
			// We already skipped some characters, so we need to output an empty token.
			if emptyToken.StartPos < emptyToken.EndPos {
				if err := r.SeekBackward(uint64(numBytesReadAtLastAccept)); err != nil {
					return 0, nil, errors.Wrapf(err, "rewind reader")
				}
				return pos, []Token{emptyToken}, nil
			}

			ruleIdx, rule := t.actionsToRule(actions) // Choose matched rule with the lowest index.
			acceptedToken := Token{
				StartPos:     pos,
				EndPos:       endPos,
				LookaheadPos: lookaheadPos,
				Role:         rule.TokenRole,
			}

			subTokenizer := t.SubTokenizers[ruleIdx]
			if subTokenizer != nil {
				// Rewind the reader to the start of the token.
				if err := r.SeekBackward(uint64(numBytesReadAtLastAccept)); err != nil {
					return 0, nil, errors.Wrapf(err, "rewind reader")
				}
				// Run the relevant sub-tokenizer on the contents of the token.
				return t.runSubTokenizer(r, subTokenizer, acceptedToken)
			}

			return endPos, []Token{acceptedToken}, nil
		}

		// We couldn't find a match, so advance to the next position and try again.
		pos++
		if err := t.advanceReaderOneRune(r); err != nil {
			return 0, nil, errors.Wrapf(err, "advance reader")
		}

		// Cover the skipped position with an empty token.
		emptyToken.EndPos = pos
		if lookaheadPos > emptyToken.LookaheadPos {
			emptyToken.LookaheadPos = lookaheadPos
		}
	}

	return pos, []Token{emptyToken}, nil
}

func (t *Tokenizer) runSubTokenizer(r InputReader, subTokenizer *Tokenizer, token Token) (uint64, []Token, error) {
	// The sub tokenizer treats the start/end of token as the start/end of the text.
	// This affects the behavior of the start-of-text (^) and end-of-text ($) patterns.
	// For this reason, we "translate" the text positions relative to the start of the token
	// and set textLen to the end of the token.
	var newTokens []Token
	var posInToken uint64
	tokenLen := token.EndPos - token.StartPos
	for posInToken < tokenLen {
		nextPos, nextTokens, err := subTokenizer.nextTokens(r, tokenLen, posInToken)
		if err != nil {
			return 0, nil, err
		}

		newTokens = append(newTokens, nextTokens...)
		posInToken = nextPos
	}

	for i := 0; i < len(newTokens); i++ {
		newTok := &(newTokens[i])

		// Translate the positions back to the original text.
		newTok.StartPos += token.StartPos
		newTok.EndPos += token.StartPos

		// Set the lookahead position to the same as the parent token so that
		// modifications to any of the tokens trigger retokenization from the
		// start of the first token (if not earlier).
		newTok.LookaheadPos = token.LookaheadPos

		// Any tokens unrecognized by the sub-tokenizer default to the role of the parent token.
		if newTok.Role == TokenRoleNone {
			newTok.Role = token.Role
		}
	}

	return token.EndPos, newTokens, nil
}

func (t *Tokenizer) rewriteTokens(tree *TokenTree, startPos uint64, endPos uint64, newTokens []Token) {
	rangeLen := endPos - startPos
	log.Printf("Rewriting tokens from %d to %d (length = %d)\n", startPos, endPos, rangeLen)
	tree.deleteRange(startPos, rangeLen)
	for _, tok := range newTokens {
		tree.insertToken(tok)
	}
}

func (t *Tokenizer) actionsToRule(actions []int) (int, TokenizerRule) {
	ruleIdx := actions[0]
	for _, a := range actions[1:] {
		if a < ruleIdx {
			ruleIdx = a
		}
	}
	return ruleIdx, t.Rules[ruleIdx]
}

func (t *Tokenizer) advanceReaderOneRune(r InputReader) error {
	if _, err := r.Read(t.buf[:1]); err != nil && err != io.EOF {
		return errors.Wrapf(err, "read first byte")
	}

	w := int64(utf8.CharWidth[t.buf[0]])
	if w > 1 {
		if _, err := r.Read(t.buf[:w-1]); err != nil {
			return errors.Wrapf(err, "read next bytes")
		}
	}

	return nil
}

func advanceIterToPos(iter *TokenIter, pos uint64) {
	var tok Token
	for iter.Get(&tok) {
		if tok.StartPos >= pos {
			return
		}
		iter.Advance()
	}
}
