package state

import (
	"github.com/aretext/aretext/editor/syntax"
	"github.com/aretext/aretext/editor/syntax/parser"
)

// SetSyntax sets the syntax language for the current document.
func SetSyntax(state *EditorState, language syntax.Language) {
	setSyntaxAndRetokenize(state.documentBuffer, language)
}

// setSyntaxAndRetokenize changes the syntax language of the buffer and updates the tokens.
func setSyntaxAndRetokenize(buffer *BufferState, language syntax.Language) {
	buffer.syntaxLanguage = language
	buffer.syntaxParser = syntax.ParserForLanguage(language)

	if buffer.syntaxParser == nil {
		buffer.syntaxLanguage = syntax.LanguagePlaintext
		return
	}

	buffer.syntaxParser.ParseAll(buffer.textTree)
}

// retokenizeAfterEdit updates syntax tokens after an edit to the text (insert or delete).
func retokenizeAfterEdit(buffer *BufferState, edit parser.Edit) {
	if buffer.syntaxParser == nil {
		return
	}

	buffer.syntaxParser.ReparseAfterEdit(buffer.textTree, edit)
}
