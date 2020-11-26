package syntax

import (
	"strings"

	"github.com/pkg/errors"
	"github.com/wedaly/aretext/internal/pkg/syntax/parser"
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
	textLen := uint64(len(s))
	tokenTree, err := tokenizer.TokenizeAll(r, textLen)
	if err != nil {
		return nil, errors.Wrapf(err, "TokenizeAll()")
	}

	tokens := tokenTree.IterFromPosition(0).Collect()
	tokensWithText := make([]TokenWithText, 0, len(tokens))
	for _, tok := range tokens {
		if tok.Role == parser.TokenRoleNone {
			continue
		}

		tokensWithText = append(tokensWithText, TokenWithText{
			Role: tok.Role,
			Text: s[tok.StartPos:tok.EndPos],
		})
	}

	return tokensWithText, nil
}
