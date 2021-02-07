package syntax

import (
	"log"

	"github.com/aretext/aretext/syntax/parser"
)

//go:generate go run gen_tokenizers.go

// Language is an enum of available languages that we can parse.
type Language int

const (
	LanguageUndefined = Language(iota)
	LanguageJson
	LanguageGo
)

func (language Language) String() string {
	switch language {
	case LanguageUndefined:
		return "undefined"
	case LanguageJson:
		return "json"
	case LanguageGo:
		return "go"
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
	case "go":
		return LanguageGo
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
	case LanguageGo:
		return GolangTokenizer
	default:
		return nil
	}
}
