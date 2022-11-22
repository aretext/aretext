package locate

import (
	"github.com/aretext/aretext/syntax/parser"
	"github.com/aretext/aretext/text"
)

// DelimiterPair is a pair of matching open/close delimiters (parens, braces, etc.)
type DelimiterPair struct {
	OpenRune  rune
	CloseRune rune
}

func (p DelimiterPair) MatchRune(r rune) bool {
	return r == p.OpenRune || r == p.CloseRune
}

var (
	ParenPair   = DelimiterPair{OpenRune: '(', CloseRune: ')'}
	BracketPair = DelimiterPair{OpenRune: '[', CloseRune: ']'}
	BracePair   = DelimiterPair{OpenRune: '{', CloseRune: '}'}
)

// MatchingCodeBlockDelimiter locates the matching paren, brace, or bracket at a position, if it exists.
func MatchingCodeBlockDelimiter(textTree *text.Tree, syntaxParser *parser.P, pos uint64) (uint64, bool) {
	reader := textTree.ReaderAtPosition(pos)
	r, _, err := reader.ReadRune()
	if err != nil || !(ParenPair.MatchRune(r) || BracketPair.MatchRune(r) || BracePair.MatchRune(r)) {
		return 0, false
	}

	switch r {
	case ParenPair.OpenRune:
		return searchForwardMatch(ParenPair, textTree, syntaxParser, pos)
	case ParenPair.CloseRune:
		return searchBackwardMatch(ParenPair, textTree, syntaxParser, pos)
	case BracePair.OpenRune:
		return searchForwardMatch(BracePair, textTree, syntaxParser, pos)
	case BracePair.CloseRune:
		return searchBackwardMatch(BracePair, textTree, syntaxParser, pos)
	case BracketPair.OpenRune:
		return searchForwardMatch(BracketPair, textTree, syntaxParser, pos)
	case BracketPair.CloseRune:
		return searchBackwardMatch(BracketPair, textTree, syntaxParser, pos)
	default:
		return 0, false
	}
}

// PrevUnmatchedOpenDelimiter locates the previous unmatched open delimiter before a position.
func PrevUnmatchedOpenDelimiter(delimiterPair DelimiterPair, textTree *text.Tree, syntaxParser *parser.P, pos uint64) (uint64, bool) {
	return searchBackwardMatch(delimiterPair, textTree, syntaxParser, pos)
}

// NextUnmatchedCloseDelimiter locates the next unmatched close delimiter after a position.
func NextUnmatchedCloseDelimiter(delimiterPair DelimiterPair, textTree *text.Tree, syntaxParser *parser.P, pos uint64) (uint64, bool) {
	return searchForwardMatch(delimiterPair, textTree, syntaxParser, pos)
}

// DelimitedBlock locates the start and end positions for matched open/close delimiters.
func DelimitedBlock(delimiterPair DelimiterPair, textTree *text.Tree, syntaxParser *parser.P, includeDelimiters bool, pos uint64) (uint64, uint64) {
	startPos := pos + 1
	reader := textTree.ReverseReaderAtPosition(startPos)
	for {
		r, _, err := reader.ReadRune()
		if err != nil {
			return pos, pos
		}

		startPos--

		if r == delimiterPair.OpenRune {
			endPos, ok := searchForwardMatch(delimiterPair, textTree, syntaxParser, startPos)
			if ok {
				if includeDelimiters {
					endPos++
				} else {
					startPos++
				}
				return startPos, endPos
			}
		}
	}
}

func searchForwardMatch(delimiterPair DelimiterPair, textTree *text.Tree, syntaxParser *parser.P, pos uint64) (uint64, bool) {
	startToken := stringOrCommentTokenAtPos(syntaxParser, pos)
	pos++
	depth := 1
	reader := textTree.ReaderAtPosition(pos)
	for {
		r, _, err := reader.ReadRune()
		if err != nil {
			return 0, false
		}

		if delimiterPair.MatchRune(r) {
			if startToken == stringOrCommentTokenAtPos(syntaxParser, pos) {
				if r == delimiterPair.OpenRune {
					depth++
				} else {
					depth--
				}
			}
		}

		if depth == 0 {
			return pos, true
		}

		pos++
	}
}

func searchBackwardMatch(delimiterPair DelimiterPair, textTree *text.Tree, syntaxParser *parser.P, pos uint64) (uint64, bool) {
	startToken := stringOrCommentTokenAtPos(syntaxParser, pos)
	depth := 1
	reader := textTree.ReverseReaderAtPosition(pos)
	for {
		r, _, err := reader.ReadRune()
		if err != nil {
			return 0, false
		}

		pos--

		if delimiterPair.MatchRune(r) {
			if startToken == stringOrCommentTokenAtPos(syntaxParser, pos) {
				if r == delimiterPair.OpenRune {
					depth--
				} else if r == delimiterPair.CloseRune {
					depth++
				}
			}
		}

		if depth == 0 {
			return pos, true
		}
	}
}

func stringOrCommentTokenAtPos(syntaxParser *parser.P, pos uint64) parser.Token {
	if syntaxParser == nil {
		return parser.Token{}
	}
	token := syntaxParser.TokenAtPosition(pos)
	if token.Role != parser.TokenRoleComment && token.Role != parser.TokenRoleString {
		return parser.Token{}
	}
	return token
}
