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

func LanguageFromString(s string) (Language, error) {
	switch s {
	case "undefined":
		return LanguageUndefined, nil
	case "json":
		return LanguageJson, nil
	default:
		availableLanguages := strings.Join([]string{
			LanguageJson.String(),
		}, ", ")
		err := errors.New(fmt.Sprintf("Unrecognized language, please choose one of [%s]", availableLanguages))
		return LanguageUndefined, err
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
