package languages

import (
	"unicode"

	"github.com/aretext/aretext/syntax/parser"
)

type gotemplateParseState uint8

const (
	gotemplateParseStateInText = gotemplateParseState(iota)
	gotemplateParseStateInAction
)

func (s gotemplateParseState) Equals(other parser.State) bool {
	otherState, ok := other.(gotemplateParseState)
	return ok && s == otherState
}

// GoTemplateParseFunc returns a parse func for Go templates.
// See https://pkg.go.dev/text/template
func GoTemplateParseFunc() parser.Func {
	parseText := matchState(
		gotemplateParseStateInText,
		goTemplateTextParseFunc())

	parseActionStartDelim := matchState(
		gotemplateParseStateInText,
		goTemplateActionStartDelimParseFunc().
			Map(setState(gotemplateParseStateInAction)))

	parseActionEndDelim := matchState(
		gotemplateParseStateInAction,
		goTemplateActionEndDelimParseFunc().
			Map(setState(gotemplateParseStateInText)))

	parseActionContents := matchState(
		gotemplateParseStateInAction,
		goTemplateActionContentsParseFunc())

	return initialState(
		gotemplateParseStateInText,
		parseText.
			Or(parseActionStartDelim).
			Or(parseActionContents).
			Or(parseActionEndDelim))
}

func goTemplateActionStartDelimParseFunc() parser.Func {
	return consumeString("{{").
		ThenMaybe(consumeString("-")).
		Map(recognizeToken(parser.TokenRoleOperator))
}

func goTemplateActionEndDelimParseFunc() parser.Func {
	return consumeString("-").
		MaybeBefore(consumeString("}}")).
		Map(recognizeToken(parser.TokenRoleOperator))
}

func goTemplateActionContentsParseFunc() parser.Func {
	// Comments are supposed to start/end only at the beginning/end of an action,
	// but we don't enforce that.
	parseComment := golangGeneralCommentParseFunc()

	parseString := golangRuneLiteralParseFunc().
		Or(golangRawStringLiteralParseFunc()).
		Or(golangInterpretedStringLiteralParseFunc())

	parseOperator := consumeLongestMatchingOption([]string{"|", "$", ":="}).
		Map(recognizeToken(parser.TokenRoleOperator))

	isLetterOrPunct := func(r rune) bool { return unicode.IsLetter(r) || r == '_' || r == '.' }
	isLetterPunctOrDigit := func(r rune) bool { return isLetterOrPunct(r) || unicode.IsDigit(r) }
	keywords := []string{
		"if", "else", "end", "range", "break", "continue", "template", "block", "with",
		"and", "call", "html", "index", "slice", "js", "len", "not", "or",
		"print", "printf", "println", "urlquery", "define",
		"eq", "ne", "lt", "le", "gt", "ge",
	}
	parseKeywordOrIdentifier := consumeSingleRuneLike(isLetterOrPunct).
		ThenMaybe(consumeRunesLike(isLetterPunctOrDigit)).
		MapWithInput(recognizeKeywordOrConsume(keywords, true))

	return parseString.
		Or(parseComment).
		Or(parseOperator).
		Or(parseKeywordOrIdentifier)
}

func goTemplateTextParseFunc() parser.Func {
	// Consume up to, but not including, the next '{' if it exists (may start an action).
	// Otherwise, consume the rest of the line.
	return func(iter parser.TrackingRuneIter, state parser.State) parser.Result {
		var numConsumed uint64
		for {
			r, err := iter.NextRune()
			if err != nil || r == '{' {
				break
			}

			numConsumed++

			if r == '\n' {
				break
			}
		}
		return parser.Result{
			NumConsumed: numConsumed,
			NextState:   state,
		}
	}
}
