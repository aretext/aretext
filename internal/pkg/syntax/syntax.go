package syntax

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"github.com/wedaly/aretext/internal/pkg/syntax/parser"
)

//go:generate go run gen_tokenizers.go

// Language is an enum of available languages that we can parse.
type Language int

const (
	UndefinedLanguage = Language(iota)
	JsonLanguage
)

func (language Language) String() string {
	switch language {
	case UndefinedLanguage:
		return "undefined"
	case JsonLanguage:
		return "json"
	default:
		return ""
	}
}

func LanguageFromString(s string) (Language, error) {
	switch s {
	case "undefined":
		return UndefinedLanguage, nil
	case "json":
		return JsonLanguage, nil
	default:
		availableLanguages := strings.Join([]string{
			JsonLanguage.String(),
		}, ", ")
		err := errors.New(fmt.Sprintf("Unrecognized language, please choose one of [%s]", availableLanguages))
		return UndefinedLanguage, err
	}
}

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
