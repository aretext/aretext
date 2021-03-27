package exec

import (
	"unicode/utf8"

	"github.com/aretext/aretext/syntax"
	"github.com/aretext/aretext/syntax/parser"
	"github.com/aretext/aretext/text"
	"github.com/pkg/errors"
)

// insertRuneAtPosition inserts a rune into the document.
// It also updates the syntax tokens and unsaved changes flag.
// It does NOT move the cursor.
func insertRuneAtPosition(state *EditorState, r rune, pos uint64) error {
	buffer := state.documentBuffer
	if err := buffer.textTree.InsertAtPosition(pos, r); err != nil {
		return errors.Wrapf(err, "text.Tree.InsertAtPosition")
	}
	edit := parser.Edit{Pos: pos, NumInserted: 1}
	if err := retokenizeAfterEdit(buffer, edit); err != nil {
		return errors.Wrapf(err, "retokenizeAfterEdit")
	}
	state.hasUnsavedChanges = true
	return nil
}

func mustInsertRuneAtPosition(state *EditorState, r rune, pos uint64) {
	err := insertRuneAtPosition(state, r, pos)
	if err != nil {
		panic(err)
	}
}

// deleteRunes deletes text from the document.
// It also updates the syntax token and unsaved changes flag.
// It does NOT move the cursor.
func deleteRunes(state *EditorState, pos uint64, count uint64) error {
	buffer := state.documentBuffer
	for i := uint64(0); i < count; i++ {
		buffer.textTree.DeleteAtPosition(pos)
	}
	edit := parser.Edit{Pos: pos, NumDeleted: count}
	if err := retokenizeAfterEdit(buffer, edit); err != nil {
		return errors.Wrapf(err, "retokenizeAfterEdit")
	}
	state.hasUnsavedChanges = true
	return nil
}

// setSyntaxAndRetokenize changes the syntax language of the buffer and updates the tokens.
func setSyntaxAndRetokenize(buffer *BufferState, language syntax.Language) error {
	buffer.syntaxLanguage = language
	buffer.tokenizer = syntax.TokenizerForLanguage(language)

	if buffer.tokenizer == nil {
		buffer.tokenTree = nil
		return nil
	}

	r := buffer.textTree.ReaderAtPosition(0, text.ReadDirectionForward)
	textLen := buffer.textTree.NumChars()
	tokenTree, err := buffer.tokenizer.TokenizeAll(r, textLen)
	if err != nil {
		return errors.Wrapf(err, "TokenizeAll")
	}

	buffer.tokenTree = tokenTree
	return nil
}

// retokenizeAfterEdit updates syntax tokens after an edit to the text (insert or delete).
func retokenizeAfterEdit(buffer *BufferState, edit parser.Edit) error {
	if buffer.tokenizer == nil {
		return nil
	}

	textLen := buffer.textTree.NumChars()
	readerAtPos := func(pos uint64) parser.InputReader {
		return buffer.textTree.ReaderAtPosition(pos, text.ReadDirectionForward)
	}
	updatedTokenTree, err := buffer.tokenizer.RetokenizeAfterEdit(buffer.tokenTree, edit, textLen, readerAtPos)
	if err != nil {
		return errors.Wrapf(err, "RetokenizeAfterEdit")
	}

	buffer.tokenTree = updatedTokenTree
	return nil
}

// searchTextForward finds the position of the next occurrence of a query string on or after the start position.
func searchTextForward(startPos uint64, tree *text.Tree, query string) (bool, uint64) {
	r := tree.ReaderAtPosition(startPos, text.ReadDirectionForward)
	foundMatch, matchOffset, err := text.Search(query, r)
	if err != nil {
		panic(err) // should never happen because the tree reader shouldn't return an error.
	}

	if !foundMatch {
		return false, 0
	}

	return true, startPos + matchOffset
}

// searchTextBackward finds the beginning of the previous match before the start position.
func searchTextBackward(startPos uint64, tree *text.Tree, query string) (bool, uint64) {
	if len(query) == 0 {
		return false, 0
	}

	// Since we're searching backwards through the text, we need to find
	// the mirror image of the query string.  Note that we are reversing the bytes
	// of the query string, not runes or grapheme clusters.
	reversedQuery := make([]byte, len(query))
	for i := 0; i < len(query); i++ {
		reversedQuery[i] = query[len(query)-1-i]
	}

	// It is possible for the cursor to be in the middle of a search query,
	// in which case we want to match the beginning of the query.
	// Example: if the text is "...ab[c]d..." (where [] shows the cursor position)
	// and we're searching backwards for "abcd", the cursor should end up on "a".
	// To ensure that we find these matches, we need to start searching from the current
	// position plus one less than the length of the query (or the end of text if that comes sooner).
	numRunesInQuery := uint64(utf8.RuneCountInString(query))
	pos := startPos + numRunesInQuery - 1
	if n := tree.NumChars(); pos >= n {
		if n > 0 {
			pos = n - 1
		} else {
			pos = 0
		}
	}

	r := tree.ReaderAtPosition(pos, text.ReadDirectionBackward)
	foundMatch, matchOffset, err := text.Search(string(reversedQuery), r)
	if err != nil {
		panic(err) // should never happen because the tree reader shouldn't return an error.
	}

	if !foundMatch {
		return false, 0
	}

	matchStartPos := pos - matchOffset - numRunesInQuery
	return true, matchStartPos
}

// stringSliceToMap converts a slice of strings to a map with string keys.
func stringSliceToMap(ss []string) map[string]struct{} {
	m := make(map[string]struct{}, len(ss))
	for _, s := range ss {
		m[s] = struct{}{}
	}
	return m
}
