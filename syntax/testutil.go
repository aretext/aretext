package syntax

import (
	"strings"
	"unicode/utf8"

	"github.com/aretext/aretext/syntax/parser"
	"github.com/pkg/errors"
)

// TokenWithText is a token that includes its text value.
type TokenWithText struct {
	Role parser.TokenRole
	Text string
}

// ParseTokensWithText tokenizes the input string using the specified language.
// This is useful for testing tokenizer rules.
func ParseTokensWithText(language Language, s string) ([]TokenWithText, error) {
	tokenizer := TokenizerForLanguage(language)
	if tokenizer == nil {
		return nil, errors.New("No tokenizer for language")
	}

	r := &parser.ReadSeekerInput{R: strings.NewReader(s)}
	textLen := uint64(utf8.RuneCountInString(s))
	tokenTree, err := tokenizer.TokenizeAll(r, textLen)
	if err != nil {
		return nil, errors.Wrapf(err, "TokenizeAll")
	}

	tokens := tokenTree.IterFromPosition(0, parser.IterDirectionForward).Collect()
	tokensWithText := make([]TokenWithText, 0, len(tokens))
	for _, tok := range tokens {
		if tok.Role == parser.TokenRoleNone {
			continue
		}

		tokensWithText = append(tokensWithText, TokenWithText{
			Role: tok.Role,
			Text: runeSlice(s, tok.StartPos, tok.EndPos),
		})
	}

	return tokensWithText, nil
}

func runeSlice(s string, startPos, endPos uint64) string {
	var pos uint64
	var runes []rune
	for _, r := range s {
		if pos < startPos {
			pos++
			continue
		}
		if pos >= endPos {
			break
		}
		runes = append(runes, r)
		pos++
	}
	return string(runes)
}
