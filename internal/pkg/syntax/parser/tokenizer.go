package parser

import (
	"io"

	"github.com/pkg/errors"
	"github.com/wedaly/aretext/internal/pkg/text/utf8"
)

// TokenizerRule represents a rule for parsing a particular token.
type TokenizerRule struct {
	Regexp    string
	TokenRole TokenRole
}

// Tokenizer parses a text into tokens based on a set of rules.
type Tokenizer struct {
	StateMachine *Dfa
	Rules        []TokenizerRule
}

// GenerateTokenizer compiles a tokenizer from a set of rules.
func GenerateTokenizer(rules []TokenizerRule) (*Tokenizer, error) {
	nfa := EmptyLanguageNfa()
	for i, r := range rules {
		regexp, err := ParseRegexp(r.Regexp)
		if err != nil {
			return nil, errors.Wrapf(err, "parse rule regexp")
		}
		ruleNfa := regexp.CompileNfa().SetAcceptAction(i)
		nfa = nfa.Union(ruleNfa)
	}

	tokenizer := &Tokenizer{
		StateMachine: nfa.CompileDfa(),
		Rules:        rules,
	}

	return tokenizer, nil
}

// TokenizeAll splits the entire input text into tokens.
// The input text MUST be valid UTF-8.
func (t *Tokenizer) TokenizeAll(r io.ReadSeeker, textLen uint64) (*TokenTree, error) {
	var tokens []Token
	pos := uint64(0)
	for pos < textLen {
		nextPos, nextTok, err := t.nextToken(r, textLen, pos)
		if err != nil {
			return nil, err
		}
		tokens = append(tokens, nextTok)
		pos = nextPos
	}

	// By construction, tokens are ordered ascending by start position and non-overlapping.
	return NewTokenTree(tokens), nil
}

// ReaderAtPosFunc returns a reader at the requested position.
type ReaderAtPosFunc func(pos uint64) io.ReadSeeker

// RetokenizeAfterEdit updates tokens based on an edit to the text.
// The algorithm is based on Wagner (1998) Practical Algorithms for Incremental Software Development Environments, Chapter 5.
func (t *Tokenizer) RetokenizeAfterEdit(tree *TokenTree, edit Edit, textLen uint64, readerAtPos ReaderAtPosFunc) (*TokenTree, error) {
	// The edit (insert/delete) changes the start/end/lookahead positions of tokens after the edit.
	// We need to update the token positions to match the edited text.
	// We do this operation first so that we can compare the positions of tokens in the tree with
	// the positions of new tokens parsed from the edited text.
	tree.ShiftPositionsAfterEdit(edit)

	// Find the start position of the first affected token in the tree.
	// We will start retokenizing from this position.
	// This assumes that the intervals (StartPos, LookaheadPos] of tokens in the tree cover the edit position,
	// which should be true because every position is covered by either an token from a DFA match
	// or an empty token (TokenRoleNone) from error recovery.
	var pos uint64
	var tok Token
	iter := tree.IterFromFirstAffected(edit.Pos)
	if iter.Get(&tok) {
		pos = tok.StartPos
	} else {
		iter = tree.IterFromPosition(0)
	}

	// Retrieve a reader from the start of the first affected token in the tree.
	r := readerAtPos(pos)

	// Run the tokenizer from the start position.
	// This will generate new tokens that replace existing tokens in the tree.
	// We continue until we encounter a token after the edit that matches the next generated token.
	// At this point, we know all subsequent tokens in the tree match what the tokenizer would output,
	// so we can stop retokenizing.
	var insertedTokens []Token
	for pos < textLen {
		nextPos, nextTok, err := t.nextToken(r, textLen, pos)
		if err != nil {
			return nil, errors.Wrapf(err, "next token")
		}

		// Delete all tokens up to the new token, so that the next token we check
		// is on or after the next token's start position.
		iter.DeleteToPos(nextTok.StartPos)

		var oldTok Token
		if nextTok.StartPos > edit.Pos && iter.Get(&oldTok) && nextTok == oldTok {
			// The new token exactly matches an existing token after the edit position,
			// so all subsequent tokens would match as well.  This means we're done!
			break
		}

		// Delete all tokens that overlap with the new token we're going to insert.
		iter.DeleteToPos(nextPos)

		// Defer insertions until later to avoid messing up the iterator.
		insertedTokens = append(insertedTokens, nextTok)
		pos = nextPos
	}

	if pos == textLen {
		iter.DeleteRemaining()
	}

	// Insert new tokens into the tree.
	// These tokens won't overlap with any existing tokens in the tree because
	// we would have deleted any overlapping tokens in the loop above.
	for _, tok := range insertedTokens {
		tree.InsertToken(tok)
	}

	return tree, nil
}

func (t *Tokenizer) nextToken(r io.ReadSeeker, textLen uint64, pos uint64) (uint64, Token, error) {
	emptyToken := Token{
		StartPos:     pos,
		EndPos:       pos,
		LookaheadPos: pos + 1,
		Role:         TokenRoleNone,
	}

	for pos < textLen {
		accepted, endPos, lookaheadPos, actions, rewindReader, err := t.StateMachine.MatchLongest(r, pos, textLen)
		if err != nil {
			return 0, Token{}, errors.Wrapf(err, "tokenizing input")
		}

		// Identify a token if the DFA accepts AND the accepted prefix is non-empty.
		// (An empty prefix can occur when "a*" matches "" at the beginning of "bcd")
		if accepted && endPos > pos {
			// We already skipped some characters, so we need to output an empty token.
			if emptyToken.StartPos < emptyToken.EndPos {
				if err := rewindReader(); err != nil {
					return 0, Token{}, errors.Wrapf(err, "rewind reader")
				}
				return pos, emptyToken, nil
			}

			rule := t.actionsToRule(actions) // Choose matched rule with the lowest index.
			acceptedToken := Token{
				StartPos:     pos,
				EndPos:       endPos,
				LookaheadPos: lookaheadPos,
				Role:         rule.TokenRole,
			}
			return endPos, acceptedToken, nil
		}

		// We couldn't find a match, so advance to the next position and try again.
		pos++
		if err := advanceReaderOneRune(r); err != nil {
			return 0, Token{}, errors.Wrapf(err, "advance reader")
		}

		// Cover the skipped position with an empty token.
		emptyToken.EndPos++
		emptyToken.LookaheadPos++
	}

	if emptyToken.LookaheadPos > textLen {
		emptyToken.LookaheadPos = textLen
	}

	return pos, emptyToken, nil
}

func (t *Tokenizer) actionsToRule(actions []int) TokenizerRule {
	ruleIdx := actions[0]
	for _, a := range actions[1:] {
		if a < ruleIdx {
			ruleIdx = a
		}
	}
	return t.Rules[ruleIdx]
}

func advanceReaderOneRune(r io.ReadSeeker) error {
	var buf [1]byte

	if _, err := r.Read(buf[:1]); err != nil && err != io.EOF {
		return errors.Wrapf(err, "read first byte")
	}

	w := int64(utf8.CharWidth[buf[0]])
	if w > 1 {
		if _, err := r.Seek(w-1, io.SeekCurrent); err != nil {
			return errors.Wrapf(err, "seek forward")
		}
	}

	return nil
}
