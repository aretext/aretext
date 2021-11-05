package languages

import (
	"unicode"

	"github.com/aretext/aretext/syntax/parser"
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
		Or(cNumberParseFunc())
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
	return consumeString("#").
		Then(consumeString("include").
			Or(consumeString("pragma")).
			Or(consumeString("ifndef")).
			Or(consumeString("define")).
			Or(consumeString("error")).
			Or(consumeString("undef")).
			Or(consumeString("endif")).
			Or(consumeString("ifdef")).
			Or(consumeString("elif")).
			Or(consumeString("else")).
			Or(consumeString("if"))).
		ThenNot(consumeSingleRuneLike(func(r rune) bool {
			return !unicode.IsSpace(r) // must be followed by space, newline, or EOF
		})).
		ThenMaybe(consumeToNextLineFeed).
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

	recognizeIdentifier := recognizeToken(parser.TokenRoleIdentifier)
	return consumeSingleRuneLike(isIdStart).
		ThenMaybe(consumeRunesLike(isIdContinue)).
		MapWithInput(recognizeKeywordOrConsume(keywords)).
		Map(func(r parser.Result) parser.Result {
			if len(r.ComputedTokens) == 0 {
				return recognizeIdentifier(r)
			}
			return r
		})
}

func cOperatorParseFunc() parser.Func {
	return (consumeString("=").ThenMaybe(consumeString("="))).
		Or(consumeString("+").ThenMaybe(consumeString("=").Or(consumeString("+")))).
		Or(consumeString("-").ThenMaybe(consumeString("=").Or(consumeString("-")))).
		Or(consumeString("*").ThenMaybe(consumeString("="))).
		Or(consumeString("/").ThenMaybe(consumeString("="))).
		Or(consumeString("%").ThenMaybe(consumeString("="))).
		Or(consumeString("<").ThenMaybe(consumeString("="))).
		Or(consumeString(">").ThenMaybe(consumeString("="))).
		Or(consumeString("<<").ThenMaybe(consumeString("="))).
		Or(consumeString(">>").ThenMaybe(consumeString("="))).
		Or(consumeString("^").ThenMaybe(consumeString("="))).
		Or(consumeString("|").ThenMaybe(consumeString("=").Or(consumeString("|")))).
		Or(consumeString("~")).
		Or(consumeString("!").ThenMaybe(consumeString("="))).
		Or(consumeString("&").ThenMaybe(consumeString("=").Or(consumeString("&")))).
		Or(consumeString("sizeof")).
		Map(recognizeToken(parser.TokenRoleOperator))
}

func cStringParseFunc() parser.Func {
	return consumeCStyleString('\'').
		Or(consumeCStyleString('"')).
		Map(recognizeToken(parser.TokenRoleString))
}

func cNumberParseFunc() parser.Func {
	isDigit := func(r rune) bool { return r >= '0' && r <= '9' }
	isHex := func(r rune) bool {
		return isDigit(r) || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F')
	}
	consumeIntTypeSuffix := consumeString("ULL").
		Or(consumeString("LL")).
		Or(consumeString("u")).
		Or(consumeString("U")).
		Or(consumeString("l")).
		Or(consumeString("L"))

	consumeHex := consumeString("0x").
		Then(consumeRunesLike(isHex)).
		ThenMaybe(consumeIntTypeSuffix)

	consumeDecimal := consumeRunesLike(isDigit) // Implicitly handles octal (0 at start)
	consumeInteger := consumeDecimal.ThenMaybe(consumeIntTypeSuffix)

	consumeExponent := (consumeString("e").Or(consumeString("E"))).
		ThenMaybe(consumeString("-")).
		Then(consumeDecimal)
	consumeRealTypeSuffix := consumeString("l").
		Or(consumeString("L")).
		Or(consumeString("f")).
		Or(consumeString("F"))
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
