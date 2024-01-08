package languages

import "github.com/aretext/aretext/syntax/parser"

const (
	p4TokenRolePreprocessorDirective = parser.TokenRoleCustom1
	p4TokenRoleAnnotation            = parser.TokenRoleCustom2
)

// P4ParseFunc returns a parse func for P4-16.
// See https://p4.org/p4-spec/docs/P4-16-v1.0.0-spec.html for the spec.
// See also p4.json for syntax highlighting rules:
// https://github.com/p4lang/p4-spec/blob/c84896fcd87f940983648b185ef9acf2b6f14838/p4-16/spec/p4.json
func P4ParseFunc() parser.Func {
	return p4CommentParseFunc().
		Or(p4PreprocessorDirectiveParseFunc()).
		Or(p4AnnotationParseFunc()).
		Or(p4IdentifierOrKeywordParseFunc()).
		Or(p4OperatorParseFunc()).
		Or(p4StringParseFunc()).
		Or(p4NumberParseFunc())
}

func p4CommentParseFunc() parser.Func {
	consumeLineComment := consumeString("//").
		ThenMaybe(consumeToNextLineFeed)

	consumeBlockComment := consumeString("/*").
		Then(consumeToString("*/"))

	return consumeLineComment.
		Or(consumeBlockComment).
		Map(recognizeToken(parser.TokenRoleComment))
}

func p4PreprocessorDirectiveParseFunc() parser.Func {
	directives := []string{
		"include", "if", "endif", "ifdef",
		"define", "ifndef", "undef", "line",
	}
	return consumeCStylePreprocessorDirective(directives).
		Map(recognizeToken(p4TokenRolePreprocessorDirective))
}

func p4AnnotationParseFunc() parser.Func {
	annotations := []string{
		"atomic", "defaultonly", "deprecated", "name", "noSideEffects", "noWarn",
		"optional", "priority", "pure", "tableonly", "hidden", "globalname",
	}
	return consumeString("@").
		Then(consumeLongestMatchingOption(annotations)).
		Map(recognizeToken(p4TokenRoleAnnotation))
}

func p4IdentifierOrKeywordParseFunc() parser.Func {
	isIdStart := func(r rune) bool {
		return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || r == '_' || r == '$'
	}

	isIdContinue := func(r rune) bool {
		return isIdStart(r) || (r >= '0' && r <= '9')
	}

	keywords := []string{
		"abstract", "action", "apply", "control", "default", "else",
		"extern", "exit", "false", "if",
		"package", "parser", "return", "select", "state", "switch",
		"table", "this", "transition", "true", "type", "typedef", "value_set", "verify",
		"bool", "bit", "const", "enum", "entries", "error", "header", "header_union",
		"in", "inout", "int", "list", "match_kind", "out", "string", "tuple", "struct", "varbit", "void",
	}

	return consumeSingleRuneLike(isIdStart).
		ThenMaybe(consumeRunesLike(isIdContinue)).
		MapWithInput(recognizeKeywordOrConsume(keywords))
}

func p4OperatorParseFunc() parser.Func {
	return consumeLongestMatchingOption([]string{
		"=", ">", "<", "!", "~", "?", ":",
		"==", "<=", ">=", "!=", "&&", "||", "++",
		"+", "-", "*", "/", "&", "|", "^", "%", "<<",
		">>", "&&&", "..",
	}).Map(recognizeToken(parser.TokenRoleOperator))
}

func p4StringParseFunc() parser.Func {
	return parseCStyleString('"', false)
}

func p4NumberParseFunc() parser.Func {
	// NOTE: the number regex patterns in the spec's syntax highlighting definition (p4.json)
	// differs from the spec itself (P4-16-spec.mdk). Follow the latter here.
	consumeDigitsWithUnderscores := func(isDigit func(r rune) bool) parser.Func {
		return func(iter parser.TrackingRuneIter, state parser.State) parser.Result {
			var numUnderscores, numDigits uint64
			for {
				r, err := iter.NextRune()
				if err != nil {
					break
				} else if r == '_' {
					numUnderscores++
				} else if isDigit(r) {
					numDigits++
				} else {
					break
				}
			}

			if numDigits == 0 {
				return parser.FailedResult
			}

			return parser.Result{
				NumConsumed: numUnderscores + numDigits,
				NextState:   state,
			}
		}
	}

	isDecimalDigit := func(r rune) bool { return r >= '0' && r <= '9' }
	isHexDigit := func(r rune) bool {
		return (r >= '0' && r <= '9') || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F')
	}
	isOctalDigit := func(r rune) bool {
		return r >= '0' && r <= '7'
	}
	isBinaryDigit := func(r rune) bool {
		return r == '0' || r == '1'
	}

	consumeWidthPrefix := consumeRunesLike(isDecimalDigit).
		Then(consumeSingleRuneLike(func(r rune) bool { return r == 'w' || r == 's' }))

	consumeHex := consumeString("0").
		Then(consumeSingleRuneLike(func(r rune) bool { return r == 'x' || r == 'X' })).
		Then(consumeDigitsWithUnderscores(isHexDigit))

	consumeOctal := consumeString("0").
		Then(consumeSingleRuneLike(func(r rune) bool { return r == 'o' || r == 'O' })).
		Then(consumeDigitsWithUnderscores(isOctalDigit))

	consumeBinary := consumeString("0").
		Then(consumeSingleRuneLike(func(r rune) bool { return r == 'b' || r == 'B' })).
		Then(consumeDigitsWithUnderscores(isBinaryDigit))

	consumeDecimalWithPrefix := consumeString("0").
		Then(consumeSingleRuneLike(func(r rune) bool { return r == 'd' || r == 'D' })).
		Then(consumeDigitsWithUnderscores(isDecimalDigit))

	// Ensure first digit is not an underscore.
	consumeDecimalWithoutPrefix := consumeSingleRuneLike(isDecimalDigit).
		ThenMaybe(consumeDigitsWithUnderscores(isDecimalDigit))

	return consumeWidthPrefix.
		MaybeBefore(
			consumeHex.
				Or(consumeOctal).
				Or(consumeBinary).
				Or(consumeDecimalWithPrefix).
				Or(consumeDecimalWithoutPrefix)).
		Map(recognizeToken(parser.TokenRoleNumber))
}
