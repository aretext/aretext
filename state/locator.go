package state

import (
	"github.com/aretext/aretext/syntax/parser"
	"github.com/aretext/aretext/text"
)

// LocatorParams are inputs to a function that locates a position in the document.
type LocatorParams struct {
	TextTree          *text.Tree
	TokenTree         *parser.TokenTree
	CursorPos         uint64
	AutoIndentEnabled bool
	TabSize           uint64
}

func locatorParamsForBuffer(buffer *BufferState) LocatorParams {
	return LocatorParams{
		TextTree:          buffer.textTree,
		TokenTree:         buffer.tokenTree,
		CursorPos:         buffer.cursor.position,
		AutoIndentEnabled: buffer.autoIndent,
		TabSize:           buffer.tabSize,
	}
}

// Locator is a function that locates a position in the document.
type Locator func(LocatorParams) uint64
