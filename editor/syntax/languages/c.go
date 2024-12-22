package languages

import (
	"unicode"

	"github.com/aretext/aretext/editor/syntax/parser"
)

const cTokenRolePreprocessorDirective = parser.TokenRoleCustom1

// CParseFunc returns a parse func for C99 with GNU extensions.
// See "The GNU C Reference Manual"
// http://www.gnu.org/software/gnu-c-manual/gnu-c-manual.html#Lexical-Elements
// and "The C Preprocessor"
// http://gcc.gnu.org/onlinedocs/cpp/
func CParseFunc() parser.Func {
	return cCommentParseFunc().
		Or(cPreprocessorDirective()).
		Or(cIdentifierOrKeywordParseFunc()).
		Or(cOperatorParseFunc()).
		Or(cStringParseFunc()).
		Or(cNumberParseFunc()).
		Or(consumeRunesLike(unicode.IsSpace))
}

func cCommentParseFunc() parser.Func {
	consumeLineComment := consumeString("//").
		ThenMaybe(consumeToNextLineFeed)

	consumeBlockComment := consumeString("/*").
		Then(consumeToString("*/"))

	return consumeLineComment.
		Or(consumeBlockComment).
		Map(recognizeToken(parser.TokenRoleComment))
}

func cPreprocessorDirective() parser.Func {
	directives := []string{
		"include", "pragma", "ifndef", "define", "error", "undef",
		"endif", "ifdef", "elif", "else", "if",
	}
	return consumeCStylePreprocessorDirective(directives).
		Map(recognizeToken(cTokenRolePreprocessorDirective))
}

func cIdentifierOrKeywordParseFunc() parser.Func {
	isIdStart := func(r rune) bool {
		return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || r == '_' || r == '$'
	}

	isIdContinue := func(r rune) bool {
		return isIdStart(r) || (r >= '0' && r <= '9')
	}

	keywords := []string{
		"auto", "break", "case", "char", "const", "continue",
		"default", "do", "double", "else", "enum", "extern",
		"float", "for", "goto", "if", "int", "long", "register",
		"return", "short", "signed", "sizeof", "static",
		"struct", "switch", "typedef", "union", "unsigned",
		"void", "volatile", "while",
		"inline", "_Bool", "_Complex", "_Imaginary",
		"noreturn", "_Noreturn", "NULL", "bool", "true", "false",
		"__FUNCTION__", "__PRETTY_FUNCTION__", "__alignof", "__alignof__", "__asm",
		"__asm__", "__attribute", "__attribute__", "__builtin_offsetof",
		"__builtin_va_arg", "__complex", "__complex__", "__const",
		"__extension__", "__func__", "__imag", "__imag__",
		"__inline", "__inline__", "__label__", "__null", "__real", "__real__",
		"__restrict", "__restrict__", "__signed", "__signed__", "__thread", "__typeof",
		"__volatile", "__volatile__",
		"restrict",
	}

	return consumeSingleRuneLike(isIdStart).
		ThenMaybe(consumeRunesLike(isIdContinue)).
		MapWithInput(recognizeKeywordOrConsume(keywords))
}

func cOperatorParseFunc() parser.Func {
	return consumeLongestMatchingOption([]string{
		"=", "==", "+", "++", "+=", "-", "--", "-=",
		"*", "*=", "/", "/=", "%", "%=",
		"<", "<=", ">", ">=", "<<", "<<=", ">>", ">>=",
		"^", "^=", "|", "|=", "||", "~",
		"!", "!=", "&", "&=", "&&", "sizeof",
	}).Map(recognizeToken(parser.TokenRoleOperator))
}

func cStringParseFunc() parser.Func {
	return consumeCStyleString('\'', false).
		Or(consumeCStyleString('"', false)).
		Map(recognizeToken(parser.TokenRoleString))
}

func cNumberParseFunc() parser.Func {
	isDigit := func(r rune) bool { return r >= '0' && r <= '9' }
	isHex := func(r rune) bool {
		return isDigit(r) || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F')
	}
	consumeIntTypeSuffix := consumeLongestMatchingOption([]string{"ULL", "LL", "u", "U", "l", "L"})
	consumeHex := consumeString("0x").
		Then(consumeRunesLike(isHex)).
		ThenMaybe(consumeIntTypeSuffix)
	consumeDecimal := consumeRunesLike(isDigit) // Implicitly handles octal (0 at start)
	consumeInteger := consumeDecimal.ThenMaybe(consumeIntTypeSuffix)

	consumeExponent := (consumeString("e").Or(consumeString("E"))).
		ThenMaybe(consumeString("-")).
		Then(consumeDecimal)
	consumeRealTypeSuffix := consumeLongestMatchingOption([]string{"l", "L", "f", "F"})
	consumeRealWithDecimal := (consumeString(".").Then(consumeDecimal)).
		Or(consumeDecimal.Then(consumeString(".")).ThenMaybe(consumeDecimal)).
		ThenMaybe(consumeExponent)
	consumeRealJustExponent := consumeDecimal.Then(consumeExponent)
	consumeReal := consumeRealWithDecimal.
		Or(consumeRealJustExponent).
		ThenMaybe(consumeRealTypeSuffix)

	return consumeHex.Or(consumeReal).Or(consumeInteger).
		Map(recognizeToken(parser.TokenRoleNumber))
}
