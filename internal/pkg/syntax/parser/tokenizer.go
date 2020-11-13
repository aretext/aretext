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
		accepted, endPos, lookaheadPos, actions, err := t.StateMachine.MatchLongest(r, pos, textLen)
		if err != nil {
			return nil, errors.Wrapf(err, "tokenizing input")
		}

		// Identify a token if the DFA accepts AND the accepted prefix is non-empty.
		// (An empty prefix can occur when "a*" matches "" at the beginning of "bcd")
		if accepted && endPos > pos {
			rule := t.actionsToRule(actions) // Choose matched rule with the lowest index.
			tokens = append(tokens, Token{
				StartPos:     pos,
				EndPos:       endPos,
				LookaheadPos: lookaheadPos,
				Role:         rule.TokenRole,
			})
			pos = endPos
		} else {
			// If we couldn't find a match, advance to the next position and try again.
			pos++
			if err := advanceReaderOneRune(r); err != nil {
				return nil, errors.Wrapf(err, "advance reader")
			}
		}
	}

	// By construction, tokens are ordered ascending by start position and non-overlapping.
	return NewTokenTree(tokens), nil
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
