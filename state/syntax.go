package state

import (
	"log"

	"github.com/pkg/errors"

	"github.com/aretext/aretext/syntax"
	"github.com/aretext/aretext/syntax/parser"
)

// SetSyntax sets the syntax language for the current document.
func SetSyntax(state *EditorState, language syntax.Language) {
	buffer := state.documentBuffer
	if err := setSyntaxAndRetokenize(buffer, language); err != nil {
		log.Printf("Error setting syntax: %v\n", err)
	}
}

// setSyntaxAndRetokenize changes the syntax language of the buffer and updates the tokens.
func setSyntaxAndRetokenize(buffer *BufferState, language syntax.Language) error {
	buffer.syntaxLanguage = language
	buffer.tokenizer = syntax.TokenizerForLanguage(language)

	if buffer.tokenizer == nil {
		buffer.tokenTree = nil
		return nil
	}

	r := buffer.textTree.ReaderAtPosition(0)
	textLen := buffer.textTree.NumChars()
	tokenTree, err := buffer.tokenizer.TokenizeAll(&r, textLen)
	if err != nil {
		return errors.Wrap(err, "TokenizeAll")
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
		r := buffer.textTree.ReaderAtPosition(pos)
		return &r
	}
	updatedTokenTree, err := buffer.tokenizer.RetokenizeAfterEdit(buffer.tokenTree, edit, textLen, readerAtPos)
	if err != nil {
		return errors.Wrap(err, "RetokenizeAfterEdit")
	}

	buffer.tokenTree = updatedTokenTree
	return nil
}
