package syntax

import (
	"strings"
	"unicode/utf8"

	"github.com/aretext/aretext/syntax/parser"
)

//go:generate go run gen_tokenizers.go

// TokenizeString tokenizes a string based on the specified language.
func TokenizeString(language Language, s string) (*parser.TokenTree, error) {
	tokenizer := TokenizerForLanguage(language)
	if tokenizer == nil {
		return nil, nil
	}
	inputReader := &parser.ReadSeekerInput{
		R: strings.NewReader(s),
	}
	textLen := uint64(utf8.RuneCountInString(s))
	return tokenizer.TokenizeAll(inputReader, textLen)
}
