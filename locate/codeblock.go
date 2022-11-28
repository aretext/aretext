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
	if err != nil || !(ParenPair.MatchRune(r) || BracketPair.MatchRune(r) || BracePair.MatchRune(r) || AnglePair.MatchRune(r)) {
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
	case AnglePair.OpenRune:
		return searchForwardMatch(AnglePair, textTree, syntaxParser, startToken, pos)
	case AnglePair.CloseRune:
		return searchBackwardMatch(AnglePair, textTree, syntaxParser, startToken, pos)
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
	matchToken := stringOrCommentTokenAtPos(syntaxParser, pos)
	startPos, endPos := delimitedBlockMatchSyntaxToken(delimiterPair, textTree, syntaxParser, includeDelimiters, pos, matchToken)
	if startPos == endPos && matchToken.Role != parser.TokenRoleNone {
		// If we can't find the delimiter in a comment/string, retry looking outside the comment/string.
		matchToken = parser.Token{}
		startPos, endPos = delimitedBlockMatchSyntaxToken(delimiterPair, textTree, syntaxParser, includeDelimiters, pos, matchToken)
	}
	return startPos, endPos
}

func delimitedBlockMatchSyntaxToken(delimiterPair DelimiterPair, textTree *text.Tree, syntaxParser *parser.P, includeDelimiters bool, pos uint64, matchSyntaxToken parser.Token) (uint64, uint64) {
	reader := textTree.ReaderAtPosition(pos)
	r, _, err := reader.ReadRune()
	if err != nil {
		return pos, pos
	}

	var ok bool
	startPos := pos // If we start on an open delimiter, use it.
	if r != delimiterPair.OpenRune {
		// Otherwise, search backwards to find the start delimiter.
		startPos, ok = searchBackwardMatch(delimiterPair, textTree, syntaxParser, matchSyntaxToken, pos)
		if !ok {
			return pos, pos
		}
	}

	// Search forward from the start delimiter to find the matching end delimiter.
	endPos, ok := searchForwardMatch(delimiterPair, textTree, syntaxParser, matchSyntaxToken, startPos)
	if !ok {
		return pos, pos
	}

	if includeDelimiters {
		endPos++
	} else {
		startPos++
	}

	return startPos, endPos
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

		if matchSyntaxToken.Role != parser.TokenRoleNone && pos > matchSyntaxToken.EndPos {
			// If we're searching for a specific token, and we're past the end of that token, exit early.
			return 0, false
		}
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

		if matchSyntaxToken.Role != parser.TokenRoleNone && pos < matchSyntaxToken.StartPos {
			// If we're searching for a specific token, and we're before the beginning of that token, exit early.
			return 0, false
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
