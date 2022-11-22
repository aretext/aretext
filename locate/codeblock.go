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

// NextUnmatchedCloseBrace locates the next unmatched close brace after a position.
func NextUnmatchedCloseBrace(textTree *text.Tree, syntaxParser *parser.P, pos uint64) (uint64, bool) {
	return searchForwardMatch(BracePair, textTree, syntaxParser, pos)
}

// PrevUnmatchedOpenBrace locates the previous unmatched open brace before a position.
func PrevUnmatchedOpenBrace(textTree *text.Tree, syntaxParser *parser.P, pos uint64) (uint64, bool) {
	return searchBackwardMatch(BracePair, textTree, syntaxParser, pos)
}

// NextUnmatchedCloseParen locates the next unmatched close paren after a position.
func NextUnmatchedCloseParen(textTree *text.Tree, syntaxParser *parser.P, pos uint64) (uint64, bool) {
	return searchForwardMatch(ParenPair, textTree, syntaxParser, pos)
}

// PrevUnmatchedOpenParen locates the previous unmatched open paren before a position.
func PrevUnmatchedOpenParen(textTree *text.Tree, syntaxParser *parser.P, pos uint64) (uint64, bool) {
	return searchBackwardMatch(ParenPair, textTree, syntaxParser, pos)
}

// DelimitedBlock locates the start and end positions for matched open/close delimiters.
func DelimitedBlock(delimiterPair DelimiterPair, textTree *text.Tree, includeDelimiters bool, pos uint64) (uint64, uint64) {
	reader := textTree.ReaderAtPosition(pos)
	r, _, err := reader.ReadRune()
	if err != nil {
		return pos, pos
	} else if r == delimiterPair.OpenRune {
		// On an open delimiter, search forward for matching close delimiters.
		endPos, ok := searchForwardMatch(delimiterPair, textTree, nil, pos)
		if !ok {
			return pos, pos
		}
		return delimitedBlockRange(includeDelimiters, pos, endPos)
	} else if r == delimiterPair.CloseRune {
		// On a close delimiter, search backward for matching open delimiters.
		startPos, ok := searchBackwardMatch(delimiterPair, textTree, nil, pos)
		if !ok {
			return pos, pos
		}
		return delimitedBlockRange(includeDelimiters, startPos, pos)
	} else {
		// Search backwards/forwards for open/close delimiters.
		startPos, ok := searchBackwardMatch(delimiterPair, textTree, nil, pos)
		if !ok {
			return pos, pos
		}
		endPos, ok := searchForwardMatch(delimiterPair, textTree, nil, pos)
		if !ok {
			return pos, pos
		}
		return delimitedBlockRange(includeDelimiters, startPos, endPos)
	}
}

func delimitedBlockRange(includeDelimiters bool, startPos, endPos uint64) (uint64, uint64) {
	if includeDelimiters {
		endPos++
	} else {
		startPos++
	}
	if startPos > endPos {
		endPos = startPos
	}
	return startPos, endPos
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
