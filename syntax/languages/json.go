package languages

import (
	"unicode"

	"github.com/aretext/aretext/syntax/parser"
)

// JsonParseFunc returns a parse func for JSON.
func JsonParseFunc() parser.Func {
	return jsonNumberParseFunc().
		Or(jsonStringOrKeyParseFunc()).
		Or(jsonKeywordParseFunc())
}

func jsonIdentifierRune(r rune) bool {
	return unicode.IsLetter(r) || (r >= '0' && r <= '9') || r == '.' || r == '_' || r == '-'
}

func jsonNumberParseFunc() parser.Func {
	consumeDigits := consumeRunesLike(func(r rune) bool { return r >= '0' && r <= '9' })
	consumeExponentIndicator := consumeString("e").Or(consumeString("E"))
	consumePositiveNumber := consumeDigits.
		ThenMaybe(consumeString(".").Then(consumeDigits)).
		ThenMaybe(
			consumeExponentIndicator.
				ThenMaybe(consumeString("-")).
				Then(consumeDigits))
	consumeNegativeNumber := consumeString("-").Then(consumePositiveNumber)

	return consumeNegativeNumber.Or(consumePositiveNumber).
		ThenNot(consumeSingleRuneLike(jsonIdentifierRune)).
		Map(recognizeToken(parser.TokenRoleNumber))
}

func jsonConsumeToKeyEndParseFunc() parser.Func {
	// Match pattern /[ \t]*:/
	return func(iter parser.TrackingRuneIter, state parser.State) parser.Result {
		var n uint64
		for {
			r, err := iter.NextRune()
			n++
			if err == nil && r == ':' {
				return parser.Result{
					NumConsumed: n,
					NextState:   state,
				}
			}

			if err != nil || !(r == ' ' || r == '\t') {
				return parser.FailedResult
			}
		}
	}
}

func jsonStringOrKeyParseFunc() parser.Func {
	const tokenRoleKey = parser.TokenRoleCustom1
	recognizeKeyToken := recognizeToken(tokenRoleKey)
	return parseCStyleString('"').
		ThenMaybe(jsonConsumeToKeyEndParseFunc()).
		Map(func(r parser.Result) parser.Result {
			if len(r.ComputedTokens) == 1 && r.NumConsumed > r.ComputedTokens[0].Length {
				// Must have parsed additional characters after the end of the string,
				// so this is a key.
				return recognizeKeyToken(r)
			} else {
				return r
			}
		})
}

func jsonKeywordParseFunc() parser.Func {
	keywords := []string{"true", "false", "null"}
	return consumeRunesLike(jsonIdentifierRune).
		MapWithInput(recognizeKeywordOrConsume(keywords))
}
