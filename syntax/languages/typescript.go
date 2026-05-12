package languages

import (
	"unicode"

	"github.com/aretext/aretext/syntax/parser"
)

// TypescriptParseFunc parses TypeScript.
func TypescriptParseFunc() parser.Func {
	return typescriptCommentParseFunc().
		Or(typescriptStringParseFunc()).
		Or(typescriptNumberParseFunc()).
		Or(typescriptIdentifierOrKeywordParseFunc()).
		Or(typescriptOperatorParseFunc()).
		Or(consumeRunesLike(unicode.IsSpace))
}

func typescriptCommentParseFunc() parser.Func {
	consumeLineComment := consumeString("//").
		ThenMaybe(consumeToNextLineFeed)

	consumeBlockComment := consumeString("/*").
		Then(consumeToString("*/"))

	return consumeLineComment.
		Or(consumeBlockComment).
		Map(recognizeToken(parser.TokenRoleComment))
}

func typescriptStringParseFunc() parser.Func {
	return consumeCStyleString('\'', false).
		Or(consumeCStyleString('"', false)).
		Or(consumeCStyleString('`', true)).
		Map(recognizeToken(parser.TokenRoleString))
}

func typescriptNumberParseFunc() parser.Func {
	consumeDecimalDigits := consumeSingleRuneLike(func(r rune) bool {
		return r >= '0' && r <= '9'
	}).ThenMaybe(consumeDigitsAndSeparators(true, func(r rune) bool {
		return r >= '0' && r <= '9'
	}))

	consumeDecimalExponent := consumeSingleRuneLike(func(r rune) bool {
		return r == 'e' || r == 'E'
	}).ThenMaybe(consumeSingleRuneLike(func(r rune) bool {
		return r == '+' || r == '-'
	})).Then(consumeDecimalDigits)

	consumeDecimalFloat := consumeDecimalDigits.
		Then(consumeString(".")).
		ThenMaybe(consumeDigitsAndSeparators(false, func(r rune) bool { return r >= '0' && r <= '9' })).
		ThenMaybe(consumeDecimalExponent).
		Or(consumeDecimalDigits.Then(consumeDecimalExponent)).
		Or(consumeString(".").Then(consumeDecimalDigits).ThenMaybe(consumeDecimalExponent))

	consumeIntegerSuffix := consumeString("n")

	consumeDecimalInteger := consumeDecimalDigits.ThenMaybe(consumeIntegerSuffix)

	consumeBinaryInteger := consumeString("0").
		Then(consumeSingleRuneLike(func(r rune) bool { return r == 'b' || r == 'B' })).
		Then(consumeDigitsAndSeparators(false, func(r rune) bool { return r == '0' || r == '1' })).
		ThenMaybe(consumeIntegerSuffix)

	consumeOctalInteger := consumeString("0").
		Then(consumeSingleRuneLike(func(r rune) bool { return r == 'o' || r == 'O' })).
		Then(consumeDigitsAndSeparators(false, func(r rune) bool { return r >= '0' && r <= '7' })).
		ThenMaybe(consumeIntegerSuffix)

	consumeHexInteger := consumeString("0").
		Then(consumeSingleRuneLike(func(r rune) bool { return r == 'x' || r == 'X' })).
		Then(consumeDigitsAndSeparators(false, func(r rune) bool {
			return (r >= '0' && r <= '9') || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F')
		})).
		ThenMaybe(consumeIntegerSuffix)

	return consumeDecimalFloat.
		Or(consumeBinaryInteger).
		Or(consumeOctalInteger).
		Or(consumeHexInteger).
		Or(consumeDecimalInteger).
		Map(recognizeToken(parser.TokenRoleNumber))
}

func typescriptIdentifierOrKeywordParseFunc() parser.Func {
	isIdStart := func(r rune) bool {
		return r == '_' || r == '$' || unicode.IsLetter(r)
	}

	isIdContinue := func(r rune) bool {
		return isIdStart(r) || unicode.IsDigit(r)
	}

	keywords := []string{
		"abstract", "any", "as", "asserts", "async", "await", "bigint", "boolean",
		"break", "case", "catch", "class", "const", "constructor", "continue",
		"debugger", "declare", "default", "delete", "do", "else", "enum", "export",
		"extends", "false", "finally", "for", "from", "function", "if",
		"implements", "import", "in", "infer", "instanceof", "interface", "is",
		"keyof", "let", "module", "namespace", "never", "new", "null", "number",
		"object", "of", "private", "protected", "public", "readonly", "require",
		"return", "static", "string", "super", "switch", "symbol", "this",
		"throw", "true", "try", "type", "typeof", "undefined", "unique", "unknown",
		"var", "void", "while", "with", "yield",
	}

	return consumeSingleRuneLike(isIdStart).
		ThenMaybe(consumeRunesLike(isIdContinue)).
		MapWithInput(recognizeKeywordOrConsume(keywords, true))
}

func typescriptOperatorParseFunc() parser.Func {
	return consumeLongestMatchingOption([]string{
		">>>=", "===", "!==", ">>>", "<<=", ">>=", "**=", "&&=", "||=", "??=",
		"=>", "++", "--", "**", "&&", "||", "??", "?.", "==", "!=", "<=", ">=",
		"+=", "-=", "*=", "/=", "%=", "&=", "|=", "^=", "<<", ">>", "+?", "...",
		"=", "+", "-", "*", "/", "%", "!", "~", "&", "|", "^", "<", ">", "?", ":",
	}).Map(recognizeToken(parser.TokenRoleOperator))
}
