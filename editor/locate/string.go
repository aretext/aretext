package locate

import (
	"github.com/aretext/aretext/editor/syntax/parser"
	"github.com/aretext/aretext/editor/text"
)

// StringObject locates the start and end positions for a single- or double-quoted string.
func StringObject(quoteRune rune, textTree *text.Tree, syntaxParser *parser.P, includeQuotes bool, pos uint64) (uint64, uint64) {
	// If the cursor is inside a string syntax token starting with the quote rune, use that.
	startPos, endPos, ok := stringObjectFromSyntaxToken(quoteRune, textTree, syntaxParser, includeQuotes, pos)
	if ok {
		return startPos, endPos
	}

	// Otherwise, find the string object from opening/closing quotes.
	return stringObjectFromOpenAndCloseQuotes(quoteRune, textTree, includeQuotes, pos)
}

func stringObjectFromSyntaxToken(quoteRune rune, textTree *text.Tree, syntaxParser *parser.P, includeQuotes bool, pos uint64) (uint64, uint64, bool) {
	if syntaxParser == nil {
		return 0, 0, false
	}

	token := syntaxParser.TokenAtPosition(pos)
	if token.Role == parser.TokenRoleString {
		reader := textTree.ReaderAtPosition(token.StartPos)
		r, _, err := reader.ReadRune()
		if err == nil && r == quoteRune {
			startPos, endPos := adjustStringObjectForIncludeQuotes(token.StartPos, token.EndPos, includeQuotes)
			return startPos, endPos, true
		}
	}

	return 0, 0, false
}

func stringObjectFromOpenAndCloseQuotes(quoteRune rune, textTree *text.Tree, includeQuotes bool, pos uint64) (uint64, uint64) {
	if isCursorOnQuote(quoteRune, textTree, pos) {
		// The cursor is on a quote, but we don't know if it's an opening or closing quote.
		// First check if there's a closing quote after this quote on the same line.
		endPos, ok := findNextQuoteInLine(quoteRune, textTree, pos+1)
		if ok {
			return adjustStringObjectForIncludeQuotes(pos, endPos, includeQuotes)
		}

		// Otherwise check if there's an opening quote before this quote on the same line.
		startPos, ok := findPrevQuoteInLine(quoteRune, textTree, pos)
		if ok {
			return adjustStringObjectForIncludeQuotes(startPos, pos+1, includeQuotes)
		}
	} else {
		// The cursor isn't on a quote, so look backwards for the start quote, then forwards for the end quote.
		startPos, ok := findPrevQuoteInLine(quoteRune, textTree, pos)
		if ok {
			endPos, ok := findNextQuoteInLine(quoteRune, textTree, pos)
			if ok {
				return adjustStringObjectForIncludeQuotes(startPos, endPos, includeQuotes)
			}
		}
	}

	// Could not find string object.
	return pos, pos
}

func adjustStringObjectForIncludeQuotes(startPos uint64, endPos uint64, includeQuotes bool) (uint64, uint64) {
	if !includeQuotes {
		startPos++
		if endPos > startPos {
			endPos--
		} else {
			endPos = startPos
		}
	}

	return startPos, endPos
}

func isCursorOnQuote(quoteRune rune, textTree *text.Tree, pos uint64) bool {
	reader := textTree.ReaderAtPosition(pos)
	r, _, err := reader.ReadRune()
	return err == nil && r == quoteRune
}

func findNextQuoteInLine(quoteRune rune, textTree *text.Tree, pos uint64) (uint64, bool) {
	reader := textTree.ReaderAtPosition(pos)
	for {
		r, _, err := reader.ReadRune()
		if err != nil || r == '\n' {
			return 0, false
		}

		pos++

		if r == quoteRune {
			return pos, true
		}
	}
}

func findPrevQuoteInLine(quoteRune rune, textTree *text.Tree, pos uint64) (uint64, bool) {
	reader := textTree.ReverseReaderAtPosition(pos)
	for {
		r, _, err := reader.ReadRune()
		if err != nil || r == '\n' {
			return 0, false
		}

		pos--

		if r == quoteRune {
			return pos, true
		}
	}
}
