package languages

import (
	"github.com/aretext/aretext/syntax/parser"
)

// YamlParseFunc returns a parse func for YAML.
func YamlParseFunc() parser.Func {
	return yamlSingleQuoteStringOrKeyParseFunc().
		Or(yamlCommentParseFunc()).
		Or(yamlKeyParseFunc()).
		Or(JsonParseFunc()) // YAML is a superset of JSON.
}

const yamlTokenRoleKey = parser.TokenRoleCustom1

func yamlSingleQuoteStringParseFunc() parser.Func {
	// Parse string wrapped in single-quotes, but treat '' as an escaped quote.
	return func(iter parser.TrackingRuneIter, state parser.State) parser.Result {
		// Consume first single-quote.
		var n uint64
		r, err := iter.NextRune()
		if err != nil || r != '\'' {
			return parser.FailedResult
		}
		n++

		// Consume the rest of the string.
		for {
			r, err = iter.NextRune()
			if err != nil || r == '\n' {
				// Couldn't find a closing quote.
				return parser.FailedResult
			}
			n++

			if r == '\'' {
				// Found a single quote.
				r, err = iter.NextRune()
				if err != nil || r != '\'' {
					// Found the end of the string.
					return parser.Result{
						NumConsumed: n,
						ComputedTokens: []parser.ComputedToken{
							{
								Length: n,
								Role:   parser.TokenRoleString,
							},
						},
						NextState: state,
					}
				}

				// The next rune is also a quote, so this is an escaped quote.
				// Consume it and keep going.
				n++
			}
		}
	}
}

func yamlSingleQuoteStringOrKeyParseFunc() parser.Func {
	return yamlSingleQuoteStringParseFunc().
		ThenMaybe(jsonConsumeToKeyEndParseFunc()).
		Map(func(r parser.Result) parser.Result {
			if len(r.ComputedTokens) == 1 && r.NumConsumed > r.ComputedTokens[0].Length {
				// Must have parsed additional characters after the end of the string,
				// so this is a key.
				return recognizeToken(yamlTokenRoleKey)(r)
			} else {
				return r
			}
		})
}

func yamlCommentParseFunc() parser.Func {
	return consumeString("#").
		ThenMaybe(consumeToNextLineFeed).
		Map(recognizeToken(parser.TokenRoleComment))
}

func yamlKeyParseFunc() parser.Func {
	return consumeRunesLike(jsonIdentifierRune).
		Then(jsonConsumeToKeyEndParseFunc()).
		Map(recognizeToken(yamlTokenRoleKey))
}
