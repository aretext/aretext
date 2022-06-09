package languages

import (
	"unicode"

	"github.com/aretext/aretext/syntax/parser"
)

// PythonParseFunc returns a parse func for Python.
// See "The Python Language Reference"
// https://docs.python.org/3/reference/
func PythonParseFunc() parser.Func {
	return pythonCommentParseFunc().
		Or(pythonStringLiteralParseFunc()).
		Or(pythonNumberLiteralParseFunc()).
		Or(pythonIdentifierOrKeywordParseFunc()).
		Or(pythonOperatorParseFunc())
}

func pythonCommentParseFunc() parser.Func {
	return consumeString("#").
		ThenMaybe(consumeToNextLineFeed).
		Map(recognizeToken(parser.TokenRoleComment))
}

func pythonStringLiteralParseFunc() parser.Func {
	consumeStringPrefix := (consumeString("r").
		ThenMaybe(consumeString("f").
			Or(consumeString("F")).
			Or(consumeString("b")).
			Or(consumeString("B")))).
		Or(consumeString("u")).
		Or(consumeString("U")).
		Or(consumeString("R").ThenMaybe(consumeString("f").Or(consumeString("F")))).
		Or(consumeString("f").ThenMaybe(consumeString("r").Or(consumeString("R")))).
		Or(consumeString("F").ThenMaybe(consumeString("r").Or(consumeString("R")))).
		Or(consumeString("b").ThenMaybe(consumeString("r").Or(consumeString("R")))).
		Or(consumeString("B").ThenMaybe(consumeString("r").Or(consumeString("R"))))

	// Technically byte strings (prefix "b") should include only ASCII characters,
	// but we accept non-ASCII.
	consumeShortString := parseCStyleString('\'', false).Or(parseCStyleString('"', false))
	consumeLongString := (consumeString(`"""`).Then(consumeToString(`"""`))).
		Or(consumeString(`'''`).Then(consumeToString(`'''`)))
	consumeLongOrShortString := consumeLongString.Or(consumeShortString)

	return (consumeStringPrefix.Then(consumeLongOrShortString)).
		Or(consumeLongOrShortString).
		Map(recognizeToken(parser.TokenRoleString))
}

func pythonNumberLiteralParseFunc() parser.Func {
	consumeDecimalZero := consumeString("0").
		ThenMaybe(consumeDigitsAndSeparators(true, func(r rune) bool { return r == '0' }))

	consumeDecimalNonZero := consumeSingleRuneLike(func(r rune) bool {
		return r >= '1' && r <= '9'
	}).ThenMaybe(consumeDigitsAndSeparators(true, func(r rune) bool {
		return r >= '0' && r <= '9'
	}))

	consumeDecimalLiteral := consumeDecimalZero.Or(consumeDecimalNonZero)

	consumeBinaryLiteral := (consumeString("0b").Or(consumeString("0B"))).
		Then(consumeDigitsAndSeparators(true, func(r rune) bool {
			return r == '0' || r == '1'
		}))

	consumeOctalLiteral := (consumeString("0o").Or(consumeString("0O"))).
		Then(consumeDigitsAndSeparators(true, func(r rune) bool {
			return r >= '0' && r <= '7'
		}))

	consumeHexLiteral := (consumeString("0x").Or(consumeString("0X"))).
		Then(consumeDigitsAndSeparators(true, func(r rune) bool {
			return (r >= '0' && r <= '9') || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F')
		}))

	consumeIntLiteral := consumeBinaryLiteral.
		Or(consumeOctalLiteral).
		Or(consumeHexLiteral).
		Or(consumeDecimalLiteral)

	consumeDigitPart := consumeDigitsAndSeparators(false, func(r rune) bool {
		return r >= '0' && r <= '9'
	})
	consumePointFloat := (consumeDigitPart.Then(consumeString(".")).ThenMaybe(consumeDigitPart)).
		Or(consumeString(".").Then(consumeDigitPart))

	consumeExponentFloat := ((consumePointFloat).Or(consumeDigitPart)).
		Then((consumeString("e").Or(consumeString("E")))).
		ThenMaybe((consumeString("+").Or(consumeString("-")))).
		Then(consumeDigitPart)

	consumeFloatLiteral := consumeExponentFloat.Or(consumePointFloat)

	consumeImaginaryLiteral := (consumeFloatLiteral.Or(consumeDigitPart)).
		Then(consumeString("j").Or(consumeString("J")))

	return consumeImaginaryLiteral.
		Or(consumeFloatLiteral).
		Or(consumeIntLiteral).
		Map(recognizeToken(parser.TokenRoleNumber))
}

func pythonIdentifierOrKeywordParseFunc() parser.Func {
	// We are not handling NFKC normalization.
	isIdentifierStart := func(r rune) bool {
		return r == '_' || unicode.In(r, unicode.Lu, unicode.Ll, unicode.Lt, unicode.Lm, unicode.Lo, unicode.Nl, unicode.Other_ID_Start)
	}

	isIdentifierContinue := func(r rune) bool {
		return isIdentifierStart(r) || unicode.In(r, unicode.Mn, unicode.Mc, unicode.Nd, unicode.Pc, unicode.Other_ID_Continue)
	}

	// We are not handling soft keywords ("match", "case", "_").
	keywords := []string{
		"False", "await", "else", "import", "pass",
		"None", "break", "except", "in", "raise",
		"True", "class", "finally", "is", "return",
		"and", "continue", "for", "lambda", "try",
		"as", "def", "from", "nonlocal", "while",
		"assert", "del", "global", "not", "with",
		"async", "elif", "if", "or", "yield",
	}

	return consumeSingleRuneLike(isIdentifierStart).
		ThenMaybe(consumeRunesLike(isIdentifierContinue)).
		MapWithInput(recognizeKeywordOrConsume(keywords))
}

func pythonOperatorParseFunc() parser.Func {
	return consumeLongestMatchingOption([]string{
		"+", "+=", "-", "->", "-=",
		"*", "*=", "**", "**=", "/", "/=",
		"//", "//=", "%", "%=",
		"@", "@=", "<", "<=", "<<", "<<=",
		">", ">=", ">>", ">>=", "&", "&=",
		"|", "|=", "^", "^=", "=", "==",
		"~", ":=", "!=",
	}).Map(recognizeToken(parser.TokenRoleOperator))
}
