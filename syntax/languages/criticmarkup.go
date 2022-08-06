package languages

import (
	"sort"

	"github.com/aretext/aretext/syntax/parser"
)

const (
	criticMarkupCommentRole = parser.TokenRoleComment

	// Use higher-numbered custom roles to avoid conflict
	// with custom roles used for markdown.
	criticMarkupAddRole       = parser.TokenRoleCustom9
	criticMarkupDelRole       = parser.TokenRoleCustom10
	criticMarkupSubRole       = parser.TokenRoleCustom11
	criticMarkupHighlightRole = parser.TokenRoleCustom12
)

// CriticMarkupParseFunc returns a parse func for CriticMarkup.
// https://github.com/CriticMarkup/CriticMarkup-toolkit/blob/master/README.md
func CriticMarkupParseFunc() parser.Func {
	/*
		This is a bit of a hack.

		We first run the markdown parser, then run the CriticMarkup parser on whatever
		the markdown parser consumed (if we see the start of a CriticMarkup tag, we
		may continue past where the markdown parser stopped).

		We then delete/truncate/split markdown tokens to make space for
		the CriticMarkup tokens.

		This works, but notice that the text within a CriticMarkup tag is still
		processed by the Markdown parser! So, for example, an asterisk "*" inside
		a CriticMarkup tag can terminate a Markdown emphasis tag.
		Fortunately, CriticMarkup explicitly forbids nesting Markdown tags,
		so if the user is doing this, it's a mistake and we can interpret
		it however we want.
	*/

	parseMarkdown := MarkdownParseFunc()
	parseCriticMarkup := criticMarkupParseFunc()
	return func(iter parser.TrackingRuneIter, state parser.State) parser.Result {
		result := parseMarkdown(iter, state)

		// Lookahead as far as the markdown parser consumed.
		lookaheadLimit := result.NumConsumed

		// If the markdown parser failed, lookahead to the one rune
		// that would be consumed by error recovery.
		// This shouldn't ever happen because the markdown parser always
		// tries to consume something, but it's safer to check.
		if lookaheadLimit == 0 {
			lookaheadLimit = 1
		}

		// Attempt to parse this part of the document as CriticMarkup.
		var criticMarkupTokens []parser.ComputedToken
		var n uint64
		for n < lookaheadLimit {
			cmResult := parseCriticMarkup(iter, state)
			if cmResult.IsSuccess() {
				for _, tok := range cmResult.ComputedTokens {
					tok.Offset += n
					criticMarkupTokens = append(criticMarkupTokens, tok)
				}
				iter.Skip(cmResult.NumConsumed)
				n += cmResult.NumConsumed
			} else {
				iter.Skip(1)
				n++
			}
		}

		// CriticMarkup tokens may overlap the markdown tokens.
		// Delete/truncate/split the markdown tokens as necessary to make space.
		result.ComputedTokens = criticMarkupConsolidateTokens(result.ComputedTokens, criticMarkupTokens)

		// There may be CriticMarkup tokens that started within this computation
		// but extend past the end of the computation. If so, update NumConsumed.
		if len(result.ComputedTokens) > 0 {
			lastToken := result.ComputedTokens[len(result.ComputedTokens)-1]
			lastTokenEnd := lastToken.Offset + lastToken.Length
			if lastTokenEnd > result.NumConsumed {
				result.NumConsumed = lastTokenEnd
			}
		}

		return result
	}
}

func criticMarkupParseFunc() parser.Func {
	parseAdd := consumeString("{++").
		Then(consumeToString("++}")).
		Map(recognizeToken(criticMarkupAddRole))

	// Examples in the CriticMarkup README use U+2010 hyphens, so allow those as well.
	parseDel := (consumeString("{--").Then(consumeToString("--}"))).
		Or(consumeString("{\u2010\u2010").Then(consumeToString("\u2010\u2010}"))).
		Map(recognizeToken(criticMarkupDelRole))

	parseSub := consumeString("{~~").
		Then(consumeToString("~~}")).
		Map(recognizeToken(criticMarkupSubRole))

	parseComment := consumeString("{>>").
		Then(consumeToString("<<}")).
		Map(recognizeToken(criticMarkupCommentRole))

	parseHighlight := consumeString("{==").
		Then(consumeToString("==}")).
		Map(recognizeToken(criticMarkupHighlightRole))

	return parseAdd.
		Or(parseDel).
		Or(parseSub).
		Or(parseComment).
		Or(parseHighlight)
}

func criticMarkupConsolidateTokens(mdTokens, cmTokens []parser.ComputedToken) []parser.ComputedToken {
	// Fast path if we have only Markdown or only CriticMarkup.
	if len(cmTokens) == 0 {
		return mdTokens
	} else if len(mdTokens) == 0 {
		return cmTokens
	}

	// Assume that mdTokens and cmTokens are each sorted ascending and non-overlapping.
	tokens := make([]parser.ComputedToken, 0, len(mdTokens)+len(cmTokens))
	tokens = append(tokens, mdTokens...)

	for _, cmTok := range cmTokens {
		// Each iteration of this loop eliminates an overlap by deleting, truncating, or splitting
		// one token. Once there are no overlaps, it inserts cmTok and exits the loop.
		for {
			i := sort.Search(len(tokens), func(i int) bool {
				return tokens[i].Offset >= cmTok.Offset
			})

			if i > 0 {
				tokBefore := tokens[i-1]
				if tokBefore.Offset+tokBefore.Length > cmTok.Offset+cmTok.Length {
					// tokBefore contains cmTok, so split tokBefore to make space.
					tokens = append(tokens, parser.ComputedToken{})
					copy(tokens[i+1:], tokens[i:])
					tokens[i-1].Length = cmTok.Offset - tokBefore.Offset
					tokens[i] = parser.ComputedToken{
						Offset: cmTok.Offset + cmTok.Length,
						Length: (tokBefore.Offset + tokBefore.Length) - (cmTok.Offset + cmTok.Length),
						Role:   tokBefore.Role,
					}
					continue
				} else if tokBefore.Offset+tokBefore.Length > cmTok.Offset {
					// Truncate end of prev token
					tokens[i-1].Length = cmTok.Offset - tokBefore.Offset
					continue
				}
			}

			if i < len(tokens) {
				tokAfter := tokens[i]
				if cmTok.Offset+cmTok.Length >= tokAfter.Offset+tokAfter.Length {
					// cmTok contains the following token, so delete it to make space.
					copy(tokens[i:], tokens[i+1:])
					tokens = tokens[0 : len(tokens)-1]
					continue
				} else if cmTok.Offset+cmTok.Length > tokAfter.Offset {
					// Truncate start of next token.
					tokens[i].Offset = cmTok.Offset + cmTok.Length
					tokens[i].Length -= (cmTok.Offset + cmTok.Length) - tokAfter.Offset
					continue
				}
			}

			// No overlap, so insert the token and exit the loop.
			tokens = append(tokens, parser.ComputedToken{})
			copy(tokens[i+1:], tokens[i:])
			tokens[i] = cmTok
			break
		}
	}

	return tokens
}
