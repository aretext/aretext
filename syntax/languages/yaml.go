package languages

import (
	"unicode"

	"github.com/aretext/aretext/syntax/parser"
)

const (
	yamlTokenRoleKey           = parser.TokenRoleCustom1
	yamlTokenRoleAliasOrAnchor = parser.TokenRoleCustom2
)

// YamlParseFunc returns a parse func for YAML.
func YamlParseFunc() parser.Func {
	parseBlockStyle := yamlTransitionToFlowParseFunc().
		Or(yamlKeyParseFunc()).
		Or(yamlOverrideParseFunc()).
		Or(yamlListParseFunc()).
		Or(yamlAnchorAliasParseFunc()).
		Or(yamlCommentParseFunc()).
		Or(yamlBlockScalarParseFunc())

	parseFlowStyle := yamlFlowStyleBracketsAndBracesParseFunc().
		Or(yamlKeyParseFunc()).
		Or(yamlCommentParseFunc()).
		Or(yamlFlowScalarParseFunc())

	return initialState(
		yamlParseState{},
		func(iter parser.TrackingRuneIter, state parser.State) parser.Result {
			yamlState := state.(yamlParseState)
			if len(yamlState.flowStyleStack) > 0 {
				return parseFlowStyle(iter, state)
			} else {
				return parseBlockStyle(iter, state)
			}
		})
}

func yamlSkipIndentation(f parser.Func) parser.Func {
	consumeIndentation := consumeRunesLike(func(r rune) bool {
		return r == ' ' || r == '\t'
	})
	return consumeIndentation.MaybeBefore(f)
}

func yamlTransitionToFlowParseFunc() parser.Func {
	parseFlowMapStart := consumeString("{").
		Map(setState(yamlParseState{flowStyleStack: []yamlFlowStyle{yamlFlowStyleMap}}))

	parseFlowArrayStart := consumeString("[").
		Map(setState(yamlParseState{flowStyleStack: []yamlFlowStyle{yamlFlowStyleArray}}))

	return yamlSkipIndentation(parseFlowMapStart.Or(parseFlowArrayStart))
}

func yamlFlowStyleBracketsAndBracesParseFunc() parser.Func {
	pushFlowStack := func(flowState yamlFlowStyle) parser.MapFn {
		return func(result parser.Result) parser.Result {
			yamlState := result.NextState.(yamlParseState)
			yamlState.flowStyleStack = append(yamlState.flowStyleStack, flowState)
			return parser.Result{
				NumConsumed:    result.NumConsumed,
				ComputedTokens: result.ComputedTokens,
				NextState:      parser.State(yamlState),
			}
		}
	}

	popFlowStack := func(flowState yamlFlowStyle) parser.MapFn {
		return func(result parser.Result) parser.Result {
			yamlState := result.NextState.(yamlParseState)
			n := len(yamlState.flowStyleStack)
			if n == 0 || yamlState.flowStyleStack[n-1] != flowState {
				return result
			}
			yamlState.flowStyleStack = yamlState.flowStyleStack[0 : n-1]
			return parser.Result{
				NumConsumed:    result.NumConsumed,
				ComputedTokens: result.ComputedTokens,
				NextState:      parser.State(yamlState),
			}
		}
	}

	parseOpenBrace := consumeString("{").
		Map(pushFlowStack(yamlFlowStyleMap))

	parseCloseBrace := consumeString("}").
		Map(popFlowStack(yamlFlowStyleMap))

	parseOpenBracket := consumeString("[").
		Map(pushFlowStack(yamlFlowStyleArray))

	parseCloseBracket := consumeString("]").
		Map(popFlowStack(yamlFlowStyleArray))

	return parseOpenBrace.
		Or(parseCloseBrace).
		Or(parseOpenBracket).
		Or(parseCloseBracket)
}

func yamlKeyParseFunc() parser.Func {
	consumeToKeyEnd := jsonConsumeToKeyEndParseFunc()

	parseUnquotedKey := consumeRunesLike(jsonIdentifierRune).
		Then(consumeToKeyEnd)

	parseQuotedKey := yamlStringParseFunc().Then(consumeToKeyEnd)

	return yamlSkipIndentation(
		parseUnquotedKey.
			Or(parseQuotedKey).
			ThenNot(consumeSingleRuneLike(func(r rune) bool {
				return !unicode.IsSpace(r) // must be followed by space, newline, or EOF
			})).
			Map(recognizeToken(yamlTokenRoleKey)))
}

func yamlOverrideParseFunc() parser.Func {
	return yamlSkipIndentation(
		consumeString("<<:").
			Map(recognizeToken(parser.TokenRoleOperator)))
}

func yamlAnchorAliasParseFunc() parser.Func {
	return yamlSkipIndentation(
		consumeString("*").
			Or(consumeString("&")).
			Then(consumeRunesLike(jsonIdentifierRune)).
			Map(recognizeToken(yamlTokenRoleAliasOrAnchor)))
}

func yamlListParseFunc() parser.Func {
	return yamlSkipIndentation(
		consumeString("-").
			ThenNot(consumeSingleRuneLike(func(r rune) bool {
				return !unicode.IsSpace(r) // must be followed by space, newline, or EOF
			})).
			Map(recognizeToken(parser.TokenRoleOperator)))
}

func yamlStringParseFunc() parser.Func {
	// Consume string wrapped in single-quotes, but treat '' as an escaped quote.
	consumeSingleQuotedString := parser.Func(
		func(iter parser.TrackingRuneIter, state parser.State) parser.Result {
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
				if err != nil {
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
							NextState:   state,
						}
					}

					// The next rune is also a quote, so this is an escaped quote.
					// Consume it and keep going.
					n++
				}
			}
		})

	return consumeSingleQuotedString.
		Or(consumeCStyleString('"', true)).
		Map(recognizeToken(parser.TokenRoleString))
}

// Match at least two words separated by a space, continuing to the end of the line.
// Example: "123 abc" matches, but "123" does not.
func yamlMultiWordUnquotedScalarParseFunc(endRunePredicate func(r rune) bool) parser.Func {
	return func(iter parser.TrackingRuneIter, state parser.State) parser.Result {
		var searchState int
		var n uint64
		for {
			r, err := iter.NextRune()
			if err != nil || endRunePredicate(r) {
				break
			}
			n++

			switch searchState {
			case 0:
				// Initially looking for non-space.
				if !(r == ' ' || r == '\t') {
					searchState = 1
				}
			case 1:
				// Then look for a space.
				if r == ' ' || r == '\t' {
					searchState = 2
				}

			case 2:
				// Then look for a non-space again.
				if !(r == ' ' || r == '\t') {
					searchState = 3
				}
			}
		}

		if searchState == 3 {
			return parser.Result{
				NumConsumed: n,
				NextState:   state,
			}
		} else {
			return parser.FailedResult
		}
	}
}

func yamlSingleWordUnquotedScalarParseFunc() parser.Func {
	keywords := []string{"true", "false", "null"}

	yamlScalarRune := func(r rune) bool {
		return !(unicode.IsSpace(r) || r == '#')
	}

	return consumeRunesLike(yamlScalarRune).
		MapWithInput(recognizeKeywordOrConsume(keywords))
}

func yamlBlockScalarParseFunc() parser.Func {
	endOfBlockScalarRune := func(r rune) bool {
		return r == '#' || r == '\n'
	}
	parseMultiWordUnquotedScalar := yamlMultiWordUnquotedScalarParseFunc(endOfBlockScalarRune)

	parseBlockStyleIndicator := consumeString("|").Or(consumeString(">")).
		Map(recognizeToken(parser.TokenRoleOperator))

	parseBlockChompingIndicator := consumeString("-").Or(consumeString("+")).
		Map(recognizeToken(parser.TokenRoleOperator))

	// Count indentation from the current position.
	// The YAML spec forbids mixing tabs and spaces in indentation, so we don't differentiate
	// between them (both increase the indentation level by one).
	countIndentation := func(iter parser.TrackingRuneIter) uint64 {
		var n uint64
		for {
			r, err := iter.NextRune()
			if err != nil || !(r == ' ' || r == '\t' || r == '\n') {
				break
			}
			n++
		}
		return n
	}

	skipToEndOfLineOrFile := func(iter *parser.TrackingRuneIter) uint64 {
		var n uint64
		for {
			r, err := iter.NextRune()
			if err != nil {
				break
			}
			n++
			if r == '\n' {
				break
			}
		}
		return n
	}

	// This is a bit of a hack that most YAML syntax highlighters use.
	// Suppose we have a block scalar like this:
	//
	//    key1:
	//      first line
	//      second line
	//      third line
	//    key2: 123
	//
	// How do we know that the scalar ends before "key2"?
	// We count the spaces at the start of "first line", then consume every
	// subsequent line indented by at least that many spaces.
	consumeBlockLines := parser.Func(
		func(iter parser.TrackingRuneIter, state parser.State) parser.Result {
			firstLineIndent := countIndentation(iter)
			n := skipToEndOfLineOrFile(&iter)
			for {
				nextLineIndent := countIndentation(iter)
				if nextLineIndent < firstLineIndent {
					break
				} else {
					m := skipToEndOfLineOrFile(&iter)
					if m == 0 {
						// End of file.
						break
					}
					n += m
				}
			}

			return parser.Result{
				NumConsumed: n,
				NextState:   state,
			}
		})

	parseBlockLines := consumeToNextLineFeed.
		ThenMaybe(consumeBlockLines).
		Map(recognizeToken(parser.TokenRoleString))

	parseBlockScalar := parseBlockStyleIndicator.
		ThenMaybe(parseBlockChompingIndicator).
		ThenMaybe(parseBlockLines)

	return yamlSkipIndentation(
		parseBlockScalar.
			Or(yamlStringParseFunc()).
			Or(parseMultiWordUnquotedScalar).
			Or(jsonNumberParseFunc()).
			Or(yamlSingleWordUnquotedScalarParseFunc()))
}

func yamlFlowScalarParseFunc() parser.Func {
	endOfFlowItemRune := func(r rune) bool {
		return r == ',' || r == '{' || r == '}' || r == '[' || r == ']' || r == '#' || r == '\n'
	}

	return yamlSkipIndentation(
		yamlStringParseFunc().
			Or(yamlMultiWordUnquotedScalarParseFunc(endOfFlowItemRune)).
			Or(jsonNumberParseFunc()).
			Or(yamlSingleWordUnquotedScalarParseFunc()))
}

func yamlCommentParseFunc() parser.Func {
	return consumeString("#").
		ThenMaybe(consumeToNextLineFeed).
		Map(recognizeToken(parser.TokenRoleComment))
}

// yamlFlowStyle represents the current flow style (map or array)
// "Flow" style in YAML uses brackets and braces to deliminate maps and arrays
// (similar to JSON). This contrasts with the more common "block" style.
type yamlFlowStyle byte

const (
	yamlFlowStyleMap = yamlFlowStyle(iota)
	yamlFlowStyleArray
)

// yamlParseState represents the state of the YAML parser.
// We use this to determine whether we're currently parsing a flow style or block style.
// To do this, maintain a stack of flow style states. Starting a map or array ("{" or "[")
// pushes to the stack; ending a map or array ("}" or "]") pops from the stack.
// If the stack is empty, we're in block style; otherwise, we're in flow style.
type yamlParseState struct {
	flowStyleStack []yamlFlowStyle
}

func (s yamlParseState) Equals(other parser.State) bool {
	otherState, ok := other.(yamlParseState)
	if !ok {
		return false
	}

	if len(s.flowStyleStack) != len(otherState.flowStyleStack) {
		return false
	}

	for i := 0; i < len(s.flowStyleStack); i++ {
		if s.flowStyleStack[i] != otherState.flowStyleStack[i] {
			return false
		}
	}

	return true
}
