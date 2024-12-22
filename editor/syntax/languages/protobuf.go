package languages

import (
	"github.com/aretext/aretext/editor/syntax/parser"
)

// Protocol Buffers Version 3
// See "Protocol Buffers Version 3 Language Specification"
// https://developers.google.com/protocol-buffers/docs/reference/proto3-spec
func ProtobufParseFunc() parser.Func {
	return initialState(protobufParseState{depth: 0},
		protobufCommentParseFunc().
			Or(protobufStringLiteralParseFunc()).
			Or(protobufFloatLiteralParseFunc()).
			Or(protobufIntegerLiteralParseFunc()).
			Or(protobufOperatorParseFunc()).
			Or(protobufBraceParseFunc()).
			Or(protobufKeywordParseFunc()),
	)
}

func protobufCommentParseFunc() parser.Func {
	consumeLineComment := consumeString("//").
		ThenMaybe(consumeToNextLineFeed)

	consumeBlockComment := consumeString("/*").
		Then(consumeToString("*/"))

	return consumeLineComment.
		Or(consumeBlockComment).
		Map(recognizeToken(parser.TokenRoleComment))
}

func protobufStringLiteralParseFunc() parser.Func {
	return parseCStyleString('\'', false).Or(parseCStyleString('"', false))
}

func protobufIntegerLiteralParseFunc() parser.Func {
	consumeHexLiteral := consumeString("0").
		Then(consumeSingleRuneLike(func(r rune) bool {
			return r == 'x' || r == 'X'
		})).
		Then(consumeRunesLike(func(r rune) bool {
			return (r >= '0' && r <= '9') || (r >= 'A' && r <= 'F') || (r >= 'a' && r <= 'f')
		}))

	// Slightly more permissive than the spec, since it allows digits "8" and "9" after a leading "0".
	consumeDecimalOrOctalLiteral := consumeRunesLike(func(r rune) bool {
		return r >= '0' && r <= '9'
	})

	consumePlusOrMinus := consumeSingleRuneLike(func(r rune) bool {
		return r == '+' || r == '-'
	})
	consumeInt := consumeHexLiteral.Or(consumeDecimalOrOctalLiteral)

	return (consumePlusOrMinus.Then(consumeInt)).
		Or(consumeInt).
		Map(recognizeToken(parser.TokenRoleNumber))
}

func protobufFloatLiteralParseFunc() parser.Func {
	consumeDecimals := consumeRunesLike(func(r rune) bool {
		return r >= '0' && r <= '9'
	})

	consumeExponent := consumeSingleRuneLike(func(r rune) bool {
		return r == 'e' || r == 'E'
	}).
		ThenMaybe(consumeSingleRuneLike(func(r rune) bool {
			return r == '+' || r == '-'
		})).
		Then(consumeDecimals)

	consumeFloatFormA := consumeDecimals.
		Then(consumeString(".")).
		ThenMaybe(consumeDecimals).
		ThenMaybe(consumeExponent)

	consumeFloatFormB := consumeDecimals.Then(consumeExponent)

	consumeFloatFormC := consumeString(".").
		Then(consumeDecimals).
		ThenMaybe(consumeExponent)

	consumeFloatFormD := consumeString("inf").Or(consumeString("nan"))

	consumePlusOrMinus := consumeSingleRuneLike(func(r rune) bool {
		return r == '+' || r == '-'
	})
	consumeFloat := consumeFloatFormA.Or(consumeFloatFormB).Or(consumeFloatFormC).Or(consumeFloatFormD)

	return (consumePlusOrMinus.Then(consumeFloat)).
		Or(consumeFloat).
		Map(recognizeToken(parser.TokenRoleNumber))
}

func protobufOperatorParseFunc() parser.Func {
	return consumeString("=").
		Map(recognizeToken(parser.TokenRoleOperator))
}

func protobufBraceParseFunc() parser.Func {
	// Open brace increases the depth by one.
	openBraceParseFunc := consumeString("{").
		Map(func(result parser.Result) parser.Result {
			depth := result.NextState.(protobufParseState).depth
			result.NextState = protobufParseState{depth: depth + 1}
			return result
		})

	// Close brace decreases the depth by one, with a minimum of zero.
	closeBraceParseFunc := consumeString("}").
		Map(func(result parser.Result) parser.Result {
			depth := result.NextState.(protobufParseState).depth
			if depth > 0 {
				result.NextState = protobufParseState{depth: depth - 1}
			}
			return result
		})

	return openBraceParseFunc.Or(closeBraceParseFunc)
}

func protobufKeywordParseFunc() parser.Func {
	isLetter := func(r rune) bool {
		return (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z')
	}

	isLetterDigitPeriodOrUnderscore := func(r rune) bool {
		return isLetter(r) || (r >= '0' && r <= '9') || r == '.' || r == '_'
	}

	allLevelKeywords := []string{"true", "false", "message", "enum", "option"}

	topLevelKeywords := append(
		[]string{"syntax", "import", "weak", "public", "package", "service"},
		allLevelKeywords...,
	)

	nestedLevelKeywords := append(
		[]string{
			"double", "float", "int32", "int64",
			"uint32", "uint64", "sint32", "sint64", "fixed32",
			"fixed64", "sfixed32", "sfixed64",
			"bool", "string", "bytes", "repeated", "oneof",
			"map", "reserved", "rpc", "returns", "to",
			"required", "optional",
		},
		allLevelKeywords...,
	)

	recognizeTopLevelKeywordOrConsume := recognizeKeywordOrConsume(topLevelKeywords)
	recognizeNestedLevelKeywordOrConsume := recognizeKeywordOrConsume(nestedLevelKeywords)

	// Consume an identifier, then check whether it's a keyword.
	// The parser recognizes different keywords at the top-level than within a block (nested in open/close parens).
	return consumeSingleRuneLike(isLetter).
		ThenMaybe(consumeRunesLike(isLetterDigitPeriodOrUnderscore)).
		MapWithInput(func(result parser.Result, iter parser.TrackingRuneIter, state parser.State) parser.Result {
			depth := result.NextState.(protobufParseState).depth
			if depth == 0 {
				return recognizeTopLevelKeywordOrConsume(result, iter, state)
			} else {
				return recognizeNestedLevelKeywordOrConsume(result, iter, state)
			}
		})
}

// protobufParseState tracks the nesting depth (open/closed braces).
type protobufParseState struct {
	depth int
}

func (s protobufParseState) Equals(other parser.State) bool {
	otherState, ok := other.(protobufParseState)
	return ok && s == otherState
}
