package locate

import (
	"github.com/aretext/aretext/syntax/parser"
	"github.com/aretext/aretext/text"
	"github.com/aretext/aretext/text/segment"
)

// NextWordStart locates the start of the next word after the cursor.
// Word boundaries occur:
//  1) at the first non-whitespace after a whitespace
//  2) at the start of an empty line
//  3) at the start of a non-empty syntax token
func NextWordStart(textTree *text.Tree, tokenTree *parser.TokenTree, pos uint64) uint64 {
	newPos := findNextWordStartFromWhitespace(textTree, pos)
	if tokenTree != nil {
		nextTokenPos := findNextWordStartFromSyntaxTokens(tokenTree, pos)
		if nextTokenPos < newPos && !isWhitespaceAtPos(textTree, nextTokenPos) {
			newPos = nextTokenPos
		}
	}
	return newPos
}

func findNextWordStartFromWhitespace(textTree *text.Tree, pos uint64) uint64 {
	segmentIter := segment.NewGraphemeClusterIterForTree(textTree, pos, text.ReadDirectionForward)
	seg := segment.Empty()
	var whitespaceFlag, newlineFlag bool
	var offset, prevOffset uint64
	for {
		eof := segment.NextSegmentOrEof(segmentIter, seg)
		if eof {
			// Stop on (not after) the last character in the document.
			return pos + prevOffset
		}
		if seg.HasNewline() {
			if newlineFlag {
				// An empty line is a word boundary.
				break
			}
			newlineFlag = true
		}
		if seg.IsWhitespace() {
			whitespaceFlag = true
		} else if whitespaceFlag {
			// Non-whitespace after whitespace is a word boundary.
			break
		}
		prevOffset = offset
		offset += seg.NumRunes()
	}
	return pos + offset
}

func findNextWordStartFromSyntaxTokens(tokenTree *parser.TokenTree, pos uint64) uint64 {
	startPos := pos
	iter := tokenTree.IterFromPosition(pos, parser.IterDirectionForward)
	var tok parser.Token
	for iter.Get(&tok) {
		pos = tok.StartPos
		if tok.Role != parser.TokenRoleNone && pos > startPos {
			break
		}
		pos = tok.EndPos
		iter.Advance()
	}
	return pos
}

// PrevWordStart locates the start of the word before the cursor.
// It uses the same word break rules as NextWordStart.
func PrevWordStart(textTree *text.Tree, tokenTree *parser.TokenTree, pos uint64) uint64 {
	newPos := findPrevWordStartFromWhitespace(textTree, pos)
	if tokenTree != nil {
		prevTokenPos := findPrevWordStartFromSyntaxTokens(tokenTree, pos)
		if prevTokenPos > newPos && !isWhitespaceAtPos(textTree, prevTokenPos) {
			newPos = prevTokenPos
		}
	}
	return newPos
}

func findPrevWordStartFromWhitespace(textTree *text.Tree, pos uint64) uint64 {
	segmentIter := segment.NewGraphemeClusterIterForTree(textTree, pos, text.ReadDirectionBackward)
	seg := segment.Empty()
	var nonwhitespaceFlag, newlineFlag bool
	var offset uint64
	for {
		eof := segment.NextSegmentOrEof(segmentIter, seg)
		if eof {
			// Start of the document.
			return 0
		}
		if seg.HasNewline() {
			if newlineFlag {
				// An empty line is a word boundary.
				return pos - offset
			}
			newlineFlag = true
		}
		if seg.IsWhitespace() {
			if nonwhitespaceFlag {
				// A whitespace after a nonwhitespace is a word boundary.
				return pos - offset
			}
		} else {
			nonwhitespaceFlag = true
		}
		offset += seg.NumRunes()
	}
}

func findPrevWordStartFromSyntaxTokens(tokenTree *parser.TokenTree, pos uint64) uint64 {
	startPos := pos
	iter := tokenTree.IterFromPosition(startPos, parser.IterDirectionBackward)
	var tok parser.Token
	for iter.Get(&tok) {
		pos = tok.StartPos
		if tok.Role != parser.TokenRoleNone && pos < startPos {
			break
		}
		iter.Advance()
	}
	return pos
}

// NextWordEnd locates the next word-end boundary after the cursor.
// The word break rules are the same as for NextWordStart, except
// that empty lines are NOT treated as word boundaries.
func NextWordEnd(textTree *text.Tree, tokenTree *parser.TokenTree, pos uint64) uint64 {
	newPos := findNextWordEndFromWhitespace(textTree, pos)
	if tokenTree != nil {
		nextTokenPos := findNextWordEndFromSyntaxTokens(tokenTree, pos)
		if nextTokenPos < newPos && !isWhitespaceAtPos(textTree, nextTokenPos) {
			newPos = nextTokenPos
		}
	}
	return newPos
}

func findNextWordEndFromWhitespace(textTree *text.Tree, pos uint64) uint64 {
	segmentIter := segment.NewGraphemeClusterIterForTree(textTree, pos, text.ReadDirectionForward)
	seg := segment.Empty()
	var prevWasNonwhitespace bool
	var offset, prevOffset uint64
	for {
		eof := segment.NextSegmentOrEof(segmentIter, seg)
		if eof {
			// Stop on (not after) the last character in the document.
			return pos + prevOffset
		}

		if seg.IsWhitespace() {
			if prevWasNonwhitespace && offset > 1 {
				// Nonwhitespace followed by whitespace should stop at the nonwhitespace.
				return pos + prevOffset
			}
			prevWasNonwhitespace = false
		} else {
			prevWasNonwhitespace = true
		}

		prevOffset = offset
		offset += seg.NumRunes()
	}
}

func findNextWordEndFromSyntaxTokens(tokenTree *parser.TokenTree, pos uint64) uint64 {
	startPos := pos
	iter := tokenTree.IterFromPosition(startPos, parser.IterDirectionForward)
	var tok parser.Token
	for iter.Get(&tok) {
		// tok.EndPos should always be greater than zero,
		// but check anyway to protect against underflow.
		if tok.EndPos > 0 {
			pos = tok.EndPos - 1
		}
		if tok.Role != parser.TokenRoleNone && pos > startPos {
			break
		}
		iter.Advance()
	}
	return pos
}

// CurrentWordStart locates the start of the word or whitespace under the cursor.
// Word boundaries are determined by both whitespace and syntax tokens.
func CurrentWordStart(textTree *text.Tree, tokenTree *parser.TokenTree, pos uint64) uint64 {
	if isWhitespaceAtPos(textTree, pos) {
		return findEndOfWordBeforeWhitespace(textTree, tokenTree, pos)
	} else {
		return findStartOfCurrentWord(textTree, tokenTree, pos)
	}
}

func findEndOfWordBeforeWhitespace(textTree *text.Tree, tokenTree *parser.TokenTree, pos uint64) uint64 {
	newPos := findStartOfWhitespace(textTree, pos)
	if tokenTree != nil {
		tokenPos := findEndOfPrevNonEmptyToken(tokenTree, pos)
		if tokenPos > pos {
			newPos = tokenPos
		}
	}
	return newPos
}

func findStartOfWhitespace(textTree *text.Tree, pos uint64) uint64 {
	segmentIter := segment.NewGraphemeClusterIterForTree(textTree, pos, text.ReadDirectionBackward)
	seg := segment.Empty()
	var offset uint64
	for {
		eof := segment.NextSegmentOrEof(segmentIter, seg)
		if eof {
			return 0
		} else if seg.HasNewline() || !seg.IsWhitespace() {
			return pos - offset
		}
		offset += seg.NumRunes()
	}
}

func findEndOfPrevNonEmptyToken(tokenTree *parser.TokenTree, pos uint64) uint64 {
	iter := tokenTree.IterFromPosition(pos, parser.IterDirectionBackward)
	var tok parser.Token
	for iter.Get(&tok) {
		if tok.Role != parser.TokenRoleNone && tok.EndPos < pos {
			return tok.EndPos
		}
		iter.Advance()
	}
	return 0
}

func findStartOfCurrentWord(textTree *text.Tree, tokenTree *parser.TokenTree, pos uint64) uint64 {
	newPos := findLastNonWhitspaceBeforePos(textTree, pos)
	if tokenTree != nil {
		tokenPos := findStartOfCurrentToken(tokenTree, pos)
		if tokenPos > newPos {
			newPos = tokenPos
		}
	}
	return newPos
}

func findLastNonWhitspaceBeforePos(textTree *text.Tree, pos uint64) uint64 {
	segmentIter := segment.NewGraphemeClusterIterForTree(textTree, pos, text.ReadDirectionBackward)
	seg := segment.Empty()
	var offset uint64
	for {
		eof := segment.NextSegmentOrEof(segmentIter, seg)
		if eof {
			return 0
		} else if seg.HasNewline() || seg.IsWhitespace() {
			return pos - offset
		}
		offset += seg.NumRunes()
	}
}

func findStartOfCurrentToken(tokenTree *parser.TokenTree, pos uint64) uint64 {
	iter := tokenTree.IterFromPosition(pos, parser.IterDirectionBackward)
	var tok parser.Token
	if iter.Get(&tok) && tok.Role != parser.TokenRoleNone {
		return tok.StartPos
	}
	return 0
}

// CurrentWordEnd locates the end of the word or whitespace under the cursor.
// The returned position is one past the last character in the word or whitespace,
// so this can be used in a DeleteMutator to delete all the characters in the word.
// Word boundaries are determined by whitespace and syntax tokens.
func CurrentWordEnd(textTree *text.Tree, tokenTree *parser.TokenTree, pos uint64) uint64 {
	if isWhitespaceAtPos(textTree, pos) {
		return findStartOfWordAfterWhitespace(textTree, tokenTree, pos)
	} else {
		return findEndOfCurrentWord(textTree, tokenTree, pos)
	}
}

func findStartOfWordAfterWhitespace(textTree *text.Tree, tokenTree *parser.TokenTree, pos uint64) uint64 {
	nextWordStartPos := NextWordStart(textTree, tokenTree, pos)
	nextLineBoundaryPos := NextLineBoundary(textTree, true, pos)
	if nextWordStartPos < nextLineBoundaryPos {
		return nextWordStartPos
	} else {
		return nextLineBoundaryPos
	}
}

func findEndOfCurrentWord(textTree *text.Tree, tokenTree *parser.TokenTree, pos uint64) uint64 {
	newPos := findFirstWhitespaceAfterPos(textTree, pos)
	if tokenTree != nil {
		tokenPos := findEndOfCurrentToken(tokenTree, pos)
		if tokenPos < newPos {
			newPos = tokenPos
		}
	}
	return newPos
}

func findFirstWhitespaceAfterPos(textTree *text.Tree, pos uint64) uint64 {
	segmentIter := segment.NewGraphemeClusterIterForTree(textTree, pos, text.ReadDirectionForward)
	seg := segment.Empty()
	var offset uint64
	for {
		eof := segment.NextSegmentOrEof(segmentIter, seg)
		if eof {
			break
		} else if seg.HasNewline() || seg.IsWhitespace() {
			break
		}
		offset += seg.NumRunes()
	}
	return pos + offset
}

func findEndOfCurrentToken(tokenTree *parser.TokenTree, pos uint64) uint64 {
	iter := tokenTree.IterFromPosition(pos, parser.IterDirectionBackward)
	var tok parser.Token
	if iter.Get(&tok) && tok.Role != parser.TokenRoleNone {
		return tok.EndPos
	}
	return pos
}

// isWhitespace returns whether the character at the position is whitespace.
func isWhitespaceAtPos(tree *text.Tree, pos uint64) bool {
	segmentIter := segment.NewGraphemeClusterIterForTree(tree, pos, text.ReadDirectionForward)
	seg := segment.Empty()
	eof := segment.NextSegmentOrEof(segmentIter, seg)
	return !eof && seg.IsWhitespace()
}
