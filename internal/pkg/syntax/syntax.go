package syntax

import (
	"log"

	"github.com/wedaly/aretext/internal/pkg/syntax/parser"
)

//go:generate go run gen_tokenizers.go

// Language is an enum of available languages that we can parse.
type Language int

const (
	LanguageUndefined = Language(iota)
	LanguageJson
)

func (language Language) String() string {
	switch language {
	case LanguageUndefined:
		return "undefined"
	case LanguageJson:
		return "json"
	default:
		return ""
	}
}

func LanguageFromString(s string) Language {
	switch s {
	case "undefined":
		return LanguageUndefined
	case "json":
		return LanguageJson
	default:
		log.Printf("Unrecognized syntax language '%s'\n", s)
		return LanguageUndefined
	}
}

// TokenizerForLanguage returns a tokenizer for the specified language.
// If no tokenizer is available (e.g. for LanguageUndefined), this returns nil.
func TokenizerForLanguage(language Language) *parser.Tokenizer {
	switch language {
	case LanguageJson:
		return JsonTokenizer
	default:
		return nil
	}
}
