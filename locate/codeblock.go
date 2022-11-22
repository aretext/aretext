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
	AnglePair   = DelimiterPair{OpenRune: '<', CloseRune: '>'}
)

// MatchingCodeBlockDelimiter locates the matching paren, brace, or bracket at a position, if it exists.
func MatchingCodeBlockDelimiter(textTree *text.Tree, syntaxParser *parser.P, pos uint64) (uint64, bool) {
	startToken := stringOrCommentTokenAtPos(syntaxParser, pos)
	reader := textTree.ReaderAtPosition(pos)
	r, _, err := reader.ReadRune()
	if err != nil || !(ParenPair.MatchRune(r) || BracketPair.MatchRune(r) || BracePair.MatchRune(r)) {
		return 0, false
	}

	switch r {
	case ParenPair.OpenRune:
		return searchForwardMatch(ParenPair, textTree, syntaxParser, startToken, pos)
	case ParenPair.CloseRune:
		return searchBackwardMatch(ParenPair, textTree, syntaxParser, startToken, pos)
	case BracePair.OpenRune:
		return searchForwardMatch(BracePair, textTree, syntaxParser, startToken, pos)
	case BracePair.CloseRune:
		return searchBackwardMatch(BracePair, textTree, syntaxParser, startToken, pos)
	case BracketPair.OpenRune:
		return searchForwardMatch(BracketPair, textTree, syntaxParser, startToken, pos)
	case BracketPair.CloseRune:
		return searchBackwardMatch(BracketPair, textTree, syntaxParser, startToken, pos)
	default:
		return 0, false
	}
}

// PrevUnmatchedOpenDelimiter locates the previous unmatched open delimiter before a position.
func PrevUnmatchedOpenDelimiter(delimiterPair DelimiterPair, textTree *text.Tree, syntaxParser *parser.P, pos uint64) (uint64, bool) {
	startToken := stringOrCommentTokenAtPos(syntaxParser, pos)
	matchPos, ok := searchBackwardMatch(delimiterPair, textTree, syntaxParser, startToken, pos)
	if !ok && startToken.Role != parser.TokenRoleNone {
		// If we can't find the delimiter in a comment/string, retry looking outside the comment/string.
		matchPos, ok = searchBackwardMatch(delimiterPair, textTree, syntaxParser, parser.Token{}, pos)
	}
	return matchPos, ok
}

// NextUnmatchedCloseDelimiter locates the next unmatched close delimiter after a position.
func NextUnmatchedCloseDelimiter(delimiterPair DelimiterPair, textTree *text.Tree, syntaxParser *parser.P, pos uint64) (uint64, bool) {
	startToken := stringOrCommentTokenAtPos(syntaxParser, pos)
	matchPos, ok := searchForwardMatch(delimiterPair, textTree, syntaxParser, startToken, pos)
	if !ok && startToken.Role != parser.TokenRoleNone {
		// If we can't find the delimiter in a comment/string, retry looking outside the comment/string.
		matchPos, ok = searchForwardMatch(delimiterPair, textTree, syntaxParser, parser.Token{}, pos)
	}
	return matchPos, ok
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
			startToken := stringOrCommentTokenAtPos(syntaxParser, startPos)
			endPos, ok := searchForwardMatch(delimiterPair, textTree, syntaxParser, startToken, startPos)
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

func searchForwardMatch(delimiterPair DelimiterPair, textTree *text.Tree, syntaxParser *parser.P, matchSyntaxToken parser.Token, pos uint64) (uint64, bool) {
	pos++
	depth := 1
	reader := textTree.ReaderAtPosition(pos)
	for {
		r, _, err := reader.ReadRune()
		if err != nil {
			return 0, false
		}

		if delimiterPair.MatchRune(r) {
			if stringOrCommentTokenAtPos(syntaxParser, pos) == matchSyntaxToken {
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

func searchBackwardMatch(delimiterPair DelimiterPair, textTree *text.Tree, syntaxParser *parser.P, matchSyntaxToken parser.Token, pos uint64) (uint64, bool) {
	depth := 1
	reader := textTree.ReverseReaderAtPosition(pos)
	for {
		r, _, err := reader.ReadRune()
		if err != nil {
			return 0, false
		}

		pos--

		if delimiterPair.MatchRune(r) {
			if stringOrCommentTokenAtPos(syntaxParser, pos) == matchSyntaxToken {
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
