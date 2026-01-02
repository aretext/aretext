package locate

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/aretext/aretext/syntax"
	"github.com/aretext/aretext/syntax/parser"
	"github.com/aretext/aretext/text"
)

func textTreeAndSyntaxParser(t *testing.T, s string, syntaxLanguage syntax.Language) (*text.Tree, *parser.P) {
	textTree, err := text.NewTreeFromString(s)
	require.NoError(t, err)

	syntaxParser := syntax.ParserForLanguage(syntaxLanguage)
	if syntaxParser != nil {
		syntaxParser.ParseAll(textTree)
	}

	return textTree, syntaxParser
}
