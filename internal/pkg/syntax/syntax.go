package syntax

import "github.com/wedaly/aretext/internal/pkg/syntax/parser"

//go:generate go run gen_tokenizers.go

// Language is an enum of available languages that we can parse.
type Language int

const (
	UndefinedLanguage = Language(iota)
	JsonLanguage
)

// TokenizerForLanguage returns a tokenizer for the specified language.
// If no tokenizer is available (e.g. for UndefinedLanguage), this returns nil.
func TokenizerForLanguage(language Language) *parser.Tokenizer {
	switch language {
	case JsonLanguage:
		return JsonTokenizer
	default:
		return nil
	}
}
