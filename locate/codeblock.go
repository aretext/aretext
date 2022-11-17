package locate

import (
	"github.com/aretext/aretext/syntax/parser"
	"github.com/aretext/aretext/text"
)

// MatchingCodeBlockDelimiter locates the matching paren, brace, or bracket at a position, if it exists.
func MatchingCodeBlockDelimiter(textTree *text.Tree, syntaxParser *parser.P, pos uint64) (uint64, bool) {
	reader := textTree.ReaderAtPosition(pos)
	r, _, err := reader.ReadRune()
	if err != nil || !(r == '(' || r == ')' || r == '{' || r == '}' || r == '[' || r == ']') {
		return 0, false
	}

	switch r {
	case '(':
		return searchForwardMatch(textTree, syntaxParser, pos, '(', ')')
	case ')':
		return searchBackwardMatch(textTree, syntaxParser, pos, '(', ')')
	case '[':
		return searchForwardMatch(textTree, syntaxParser, pos, '[', ']')
	case ']':
		return searchBackwardMatch(textTree, syntaxParser, pos, '[', ']')
	case '{':
		return searchForwardMatch(textTree, syntaxParser, pos, '{', '}')
	case '}':
		return searchBackwardMatch(textTree, syntaxParser, pos, '{', '}')
	default:
		return 0, false
	}
}

// NextUnmatchedCloseBrace locates the next unmatched close brace after a position.
func NextUnmatchedCloseBrace(textTree *text.Tree, syntaxParser *parser.P, pos uint64) (uint64, bool) {
	return searchForwardMatch(textTree, syntaxParser, pos, '{', '}')
}

// PrevUnmatchedOpenBrace locates the previous unmatched open brace before a position.
func PrevUnmatchedOpenBrace(textTree *text.Tree, syntaxParser *parser.P, pos uint64) (uint64, bool) {
	return searchBackwardMatch(textTree, syntaxParser, pos, '{', '}')
}

// NextUnmatchedCloseParen locates the next unmatched close paren after a position.
func NextUnmatchedCloseParen(textTree *text.Tree, syntaxParser *parser.P, pos uint64) (uint64, bool) {
	return searchForwardMatch(textTree, syntaxParser, pos, '(', ')')
}

// PrevUnmatchedOpenParen locates the previous unmatched open paren before a position.
func PrevUnmatchedOpenParen(textTree *text.Tree, syntaxParser *parser.P, pos uint64) (uint64, bool) {
	return searchBackwardMatch(textTree, syntaxParser, pos, '(', ')')
}

// InnerParenBlock locates the start and end positions inside matching parens.
func InnerParenBlock(textTree *text.Tree, syntaxParser *parser.P, pos uint64) (uint64, uint64) {
	reader := textTree.ReaderAtPosition(pos)
	r, _, err := reader.ReadRune()
	if err != nil {
		return pos, pos
	} else if r == '(' {
		// On an open paren, search forward for matching close paren.
		endPos, ok := searchForwardMatch(textTree, syntaxParser, pos, '(', ')')
		if !ok {
			return pos, pos
		}
		return innerCodeBlock(pos, endPos)
	} else if r == ')' {
		// On a close paren, search backward for matching open paren.
		startPos, ok := searchBackwardMatch(textTree, syntaxParser, pos, '(', ')')
		if !ok {
			return pos, pos
		}
		return innerCodeBlock(startPos, pos)
	} else {
		// Search backwards/forwards for open/close parens.
		startPos, ok := searchBackwardMatch(textTree, syntaxParser, pos, '(', ')')
		if !ok {
			return pos, pos
		}
		endPos, ok := searchForwardMatch(textTree, syntaxParser, pos, '(', ')')
		if !ok {
			return pos, pos
		}
		return innerCodeBlock(startPos, endPos)
	}
}

func searchForwardMatch(textTree *text.Tree, syntaxParser *parser.P, pos uint64, openRune rune, closeRune rune) (uint64, bool) {
	startToken := stringOrCommentTokenAtPos(syntaxParser, pos)
	pos++
	depth := 1
	reader := textTree.ReaderAtPosition(pos)
	for {
		r, _, err := reader.ReadRune()
		if err != nil {
			return 0, false
		}

		if r == openRune || r == closeRune {
			if startToken == stringOrCommentTokenAtPos(syntaxParser, pos) {
				if r == openRune {
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

func searchBackwardMatch(textTree *text.Tree, syntaxParser *parser.P, pos uint64, openRune rune, closeRune rune) (uint64, bool) {
	startToken := stringOrCommentTokenAtPos(syntaxParser, pos)
	depth := 1
	reader := textTree.ReverseReaderAtPosition(pos)
	for {
		r, _, err := reader.ReadRune()
		if err != nil {
			return 0, false
		}

		pos--

		if r == openRune || r == closeRune {
			if startToken == stringOrCommentTokenAtPos(syntaxParser, pos) {
				if r == openRune {
					depth--
				} else if r == closeRune {
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

func innerCodeBlock(startPos, endPos uint64) (uint64, uint64) {
	startPos++
	if startPos > endPos {
		endPos = startPos
	}
	return startPos, endPos
}
