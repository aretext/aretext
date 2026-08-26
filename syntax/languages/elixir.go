package languages

import (
	"unicode"

	"github.com/aretext/aretext/syntax/parser"
)

const (
	elixirTokenRoleAtom      = parser.TokenRoleCustom1
	elixirTokenRoleAttribute = parser.TokenRoleCustom2
)

// ElixirParseFunc returns a parse func for Elixir.
// See "Elixir Syntax Reference"
// https://hexdocs.pm/elixir/syntax-reference.html
func ElixirParseFunc() parser.Func {
	return elixirCommentParseFunc().
		Or(elixirCharParseFunc()).
		Or(elixirSigilParseFunc()).
		Or(elixirStringParseFunc()).
		Or(elixirAtomParseFunc()).
		Or(elixirAttributeParseFunc()).
		Or(elixirNumberParseFunc()).
		Or(elixirIdentifierOrKeywordParseFunc()).
		Or(elixirOperatorParseFunc())
}

func elixirCommentParseFunc() parser.Func {
	return consumeString("#").
		ThenMaybe(consumeToNextLineFeed).
		Map(recognizeToken(parser.TokenRoleComment))
}

// elixirCharParseFunc parses a character code like `?a` or `?\n`.
func elixirCharParseFunc() parser.Func {
	consumeEscaped := consumeString(`\`).
		Then(consumeSingleRuneLike(func(r rune) bool { return r != '\n' }))
	consumeUnescaped := consumeSingleRuneLike(func(r rune) bool {
		return !unicode.IsSpace(r)
	})
	return consumeString("?").
		Then(consumeEscaped.Or(consumeUnescaped)).
		Map(recognizeToken(parser.TokenRoleNumber))
}

// elixirSigilParseFunc parses a sigil like `~r/.../` or `~w[foo bar]a`.
func elixirSigilParseFunc() parser.Func {
	consumeName := consumeString("~").
		Then(consumeSingleRuneLike(unicode.IsLetter)).
		ThenMaybe(consumeRunesLike(unicode.IsLetter))

	// Paired delimiters do not track nesting, which is good enough for highlighting.
	consumeDelimited := consumeString("(").Then(consumeToString(")")).
		Or(consumeString("[").Then(consumeToString("]"))).
		Or(consumeString("{").Then(consumeToString("}"))).
		Or(consumeString("<").Then(consumeToString(">"))).
		Or(consumeString("/").Then(consumeToString("/"))).
		Or(consumeString("|").Then(consumeToString("|"))).
		Or(consumeString(`"`).Then(consumeToString(`"`))).
		Or(consumeString("'").Then(consumeToString("'")))

	consumeModifiers := consumeRunesLike(func(r rune) bool {
		return r >= 'a' && r <= 'z'
	})

	return consumeName.
		Then(consumeDelimited).
		ThenMaybe(consumeModifiers).
		Map(recognizeToken(parser.TokenRoleString))
}

func elixirStringParseFunc() parser.Func {
	consumeHeredoc := (consumeString(`"""`).Then(consumeToString(`"""`))).
		Or(consumeString("'''").Then(consumeToString("'''")))
	consumeSingleLine := parseCStyleString('"', false).
		Or(parseCStyleString('\'', false))
	return consumeHeredoc.
		Or(consumeSingleLine).
		Map(recognizeToken(parser.TokenRoleString))
}

// elixirAtomParseFunc parses an atom like `:ok`, `:"with spaces"`, or `:valid?`.
func elixirAtomParseFunc() parser.Func {
	isAtomStart := func(r rune) bool {
		return r == '_' || unicode.IsLetter(r)
	}
	isAtomContinue := func(r rune) bool {
		return r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r)
	}
	consumePlainAtom := consumeString(":").
		Then(consumeSingleRuneLike(isAtomStart)).
		ThenMaybe(consumeRunesLike(isAtomContinue)).
		ThenMaybe(consumeSingleRuneLike(func(r rune) bool {
			return r == '?' || r == '!'
		}))
	consumeQuotedAtom := consumeString(":").
		Then(parseCStyleString('"', false).Or(parseCStyleString('\'', false)))
	return consumePlainAtom.
		Or(consumeQuotedAtom).
		Map(recognizeToken(elixirTokenRoleAtom))
}

// elixirAttributeParseFunc parses a module attribute like `@moduledoc` or `@spec`.
func elixirAttributeParseFunc() parser.Func {
	isIdentStart := func(r rune) bool {
		return r == '_' || unicode.IsLetter(r)
	}
	isIdentContinue := func(r rune) bool {
		return r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r)
	}
	return consumeString("@").
		Then(consumeSingleRuneLike(isIdentStart)).
		ThenMaybe(consumeRunesLike(isIdentContinue)).
		Map(recognizeToken(elixirTokenRoleAttribute))
}

func elixirNumberParseFunc() parser.Func {
	consumeDecimalDigits := consumeSingleRuneLike(func(r rune) bool {
		return r >= '0' && r <= '9'
	}).ThenMaybe(consumeDigitsAndSeparators(true, func(r rune) bool {
		return r >= '0' && r <= '9'
	}))

	consumeHexLiteral := (consumeString("0x").Or(consumeString("0X"))).
		Then(consumeDigitsAndSeparators(true, func(r rune) bool {
			return (r >= '0' && r <= '9') || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F')
		}))

	consumeOctalLiteral := (consumeString("0o").Or(consumeString("0O"))).
		Then(consumeDigitsAndSeparators(true, func(r rune) bool {
			return r >= '0' && r <= '7'
		}))

	consumeBinaryLiteral := (consumeString("0b").Or(consumeString("0B"))).
		Then(consumeDigitsAndSeparators(true, func(r rune) bool {
			return r == '0' || r == '1'
		}))

	// Elixir floats require digits on both sides of the point, and an exponent
	// is only valid on a float.
	consumeExponent := (consumeString("e").Or(consumeString("E"))).
		ThenMaybe(consumeString("+").Or(consumeString("-"))).
		Then(consumeDecimalDigits)
	consumeFloatLiteral := consumeDecimalDigits.
		Then(consumeString(".")).
		Then(consumeDecimalDigits).
		ThenMaybe(consumeExponent)

	return consumeHexLiteral.
		Or(consumeOctalLiteral).
		Or(consumeBinaryLiteral).
		Or(consumeFloatLiteral).
		Or(consumeDecimalDigits).
		Map(recognizeToken(parser.TokenRoleNumber))
}

func elixirIdentifierOrKeywordParseFunc() parser.Func {
	isIdentStart := func(r rune) bool {
		return r == '_' || unicode.IsLetter(r)
	}
	isIdentContinue := func(r rune) bool {
		return r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r)
	}

	keywords := []string{
		"true", "false", "nil",
		"when", "and", "or", "not", "in",
		"fn", "do", "end", "catch", "rescue", "after", "else",
		"if", "unless", "case", "cond", "for", "with",
		"try", "receive", "raise", "throw",
		"def", "defp", "defmodule", "defmacro", "defmacrop",
		"defstruct", "defprotocol", "defimpl", "defdelegate",
		"defguard", "defguardp", "defexception", "defoverridable",
		"import", "alias", "require", "use",
		"quote", "unquote", "unquote_splicing", "super",
		"__MODULE__", "__DIR__", "__ENV__", "__CALLER__", "__STACKTRACE__",
	}

	return consumeSingleRuneLike(isIdentStart).
		ThenMaybe(consumeRunesLike(isIdentContinue)).
		ThenMaybe(consumeSingleRuneLike(func(r rune) bool {
			return r == '?' || r == '!'
		})).
		MapWithInput(recognizeKeywordOrConsume(keywords, true))
}

func elixirOperatorParseFunc() parser.Func {
	return consumeLongestMatchingOption([]string{
		".", "..", "...",
		"+", "-", "*", "/", "**",
		"++", "--",
		"==", "!=", "===", "!==", "=~",
		"<", ">", "<=", ">=",
		"&&", "||", "!",
		"&&&", "|||", "^^^", "~~~", "<<<", ">>>",
		"=", "=>", "->", "<-", "<>", "|>", "|", `\\`,
		"::", ":", "&", "^", "@", "%",
	}).Map(recognizeToken(parser.TokenRoleOperator))
}
