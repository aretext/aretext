package languages

import (
	"unicode"

	"github.com/aretext/aretext/editor/syntax/parser"
)

const rustTokenRoleLifetime = parser.TokenRoleCustom1

// RustParseFunc returns a parse func for Rust.
// See "The Rust Reference"
// https://doc.rust-lang.org/stable/reference/
func RustParseFunc() parser.Func {
	return rustCommentParseFunc().
		Or(rustOperatorParseFunc()).
		Or(rustLifetimeParseFunc()).
		Or(rustStringLiteralParseFunc()).
		Or(rustNumberLiteralParseFunc()).
		Or(rustIdentifierOrKeywordParseFunc())
}

func rustCommentParseFunc() parser.Func {
	// These rules implicitly covers the doc forms ("//!", "/*!", ...)
	consumeLineComment := consumeString("//").
		ThenMaybe(consumeToNextLineFeed)
	consumeBlockComment := consumeString("/*").
		Then(consumeToString("*/"))
	return consumeLineComment.
		Or(consumeBlockComment).
		Map(recognizeToken(parser.TokenRoleComment))
}

func rustOperatorParseFunc() parser.Func {
	return consumeLongestMatchingOption([]string{
		"@", "#", "$", "?", ":", "::",
		"+", "+=", "-", "-=", "->",
		"*", "*=", "/", "/=", "%", "%=",
		"^", "^=", "!", "!=", "&", "&&", "&=",
		"|", "||", "|=", "=", "==", "=>",
		"<<", "<<=", ">>", ">>=", ">", ">=", "<", "<=",
	}).Map(recognizeToken(parser.TokenRoleOperator))
}

func rustLifetimeParseFunc() parser.Func {
	return consumeString("'_").
		Or(consumeString("'").Then(rustConsumeIdentifierOrKeyword())).
		ThenNot(consumeString("'")).
		Map(recognizeToken(rustTokenRoleLifetime))
}

func rustConsumeRawString(iter parser.TrackingRuneIter, state parser.State) parser.Result {
	// Count number of "#" at start of string.
	var n uint64
	var numHashMarks int
	for {
		r, err := iter.NextRune()
		if err != nil {
			return parser.FailedResult
		}

		n++
		if r == '#' {
			numHashMarks++
		} else if numHashMarks > 0 && r == '"' {
			break
		} else {
			return parser.FailedResult
		}
	}

	// Consume everything until we find a quote followed by the same number of hash marks.
	hashMarkRun := -1
	for {
		r, err := iter.NextRune()
		if err != nil {
			return parser.FailedResult
		}

		n++
		if hashMarkRun < 0 && r == '"' {
			// Start a run.
			hashMarkRun = 0
		} else if hashMarkRun >= 0 && r == '#' {
			hashMarkRun++
			if hashMarkRun == numHashMarks {
				// Success
				return parser.Result{
					NumConsumed: n,
					ComputedTokens: []parser.ComputedToken{
						{Length: n},
					},
					NextState: state,
				}
			}
		} else {
			// Not in a run.
			hashMarkRun = -1
		}
	}
}

func rustStringLiteralParseFunc() parser.Func {
	consumeCharacter := consumeCStyleString('\'', false)
	consumeQuoteString := consumeCStyleString('"', true)
	consumeRawString := consumeString("r").Then(rustConsumeRawString)

	// Rust restricts byte strings to ASCII, but we don't enforce that.
	consumeByteString := consumeString("b").
		Then(consumeCharacter.Or(consumeQuoteString))
	consumeRawByteString := consumeString("br").Then(rustConsumeRawString)

	return consumeCharacter.
		Or(consumeQuoteString).
		Or(consumeRawString).
		Or(consumeByteString).
		Or(consumeRawByteString).
		Map(recognizeToken(parser.TokenRoleString))
}

func rustConsumeDigitsAndSeparators(isDigit func(r rune) bool) parser.Func {
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

func rustNumberLiteralParseFunc() parser.Func {
	consumeDecimalLiteral := rustConsumeDigitsAndSeparators(func(r rune) bool {
		return r >= '0' && r <= '9'
	})

	consumeBinaryLiteral := consumeString("0b").
		Then(rustConsumeDigitsAndSeparators(func(r rune) bool {
			return r == '0' || r == '1'
		}))

	consumeOctalLiteral := consumeString("0o").
		Then(rustConsumeDigitsAndSeparators(func(r rune) bool {
			return r >= '0' && r <= '7'
		}))

	consumeHexLiteral := consumeString("0x").
		Then(rustConsumeDigitsAndSeparators(func(r rune) bool {
			return (r >= '0' && r <= '9') || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F')
		}))

	consumeIntegerSuffix := consumeLongestMatchingOption([]string{
		"u8", "u16", "u32", "u64", "u128", "usize",
		"i8", "i16", "i32", "i64", "i128", "isize",
	})

	consumeIntegerLiteral := (consumeBinaryLiteral.
		Or(consumeOctalLiteral).
		Or(consumeHexLiteral).
		Or(consumeDecimalLiteral)).
		ThenMaybe(consumeIntegerSuffix)

	consumeFloatExponent := consumeSingleRuneLike(func(r rune) bool {
		return r == 'e' || r == 'E'
	}).ThenMaybe(consumeSingleRuneLike(func(r rune) bool {
		return r == '+' || r == '-'
	})).ThenMaybe(consumeDecimalLiteral)

	consumeFloatSuffix := consumeLongestMatchingOption([]string{"f32", "f64"})

	consumeFloatFormA := consumeDecimalLiteral.
		Then(consumeString(".")).
		ThenNot(consumeSingleRuneLike(func(r rune) bool {
			return r == '.' || r == '_' || unicode.Is(unicode.Other_ID_Start, r)
		}))

	consumeFloatFormB := consumeDecimalLiteral.
		Then(consumeFloatExponent)

	consumeFloatFormC := consumeDecimalLiteral.
		Then(consumeString(".")).
		Then(consumeDecimalLiteral).
		ThenMaybe(consumeFloatExponent)

	consumeFloatFormD := consumeDecimalLiteral.
		ThenMaybe(consumeString(".").Then(consumeDecimalLiteral)).
		ThenMaybe(consumeFloatExponent).
		Then(consumeFloatSuffix)

	consumeFloatLiteral := consumeFloatFormD.
		Or(consumeFloatFormC).
		Or(consumeFloatFormB).
		Or(consumeFloatFormA)

	return (consumeFloatLiteral.Or(consumeIntegerLiteral)).
		Map(recognizeToken(parser.TokenRoleNumber))
}

func rustConsumeIdentifierOrKeyword() parser.Func {
	isIdStart := func(r rune) bool {
		return unicode.In(r, unicode.L, unicode.Nl, unicode.Other_ID_Start) && !unicode.In(r, unicode.Pattern_Syntax, unicode.Pattern_White_Space)
	}

	isIdContinue := func(r rune) bool {
		return isIdStart(r) || unicode.In(r, unicode.Mn, unicode.Mc, unicode.Nd, unicode.Pc, unicode.Other_ID_Continue) && !unicode.In(r, unicode.Pattern_Syntax, unicode.Pattern_White_Space)
	}

	return (consumeString("_").Then(consumeRunesLike(isIdContinue))).
		Or(consumeSingleRuneLike(isIdStart).ThenMaybe(consumeRunesLike(isIdContinue)))
}

func rustIdentifierOrKeywordParseFunc() parser.Func {
	// Highlight strict and reserved keywords.
	// Ignore weak keywords.
	keywords := []string{
		"true", "false", "as", "break", "const", "continue", "crate", "else", "enum",
		"extern", "false", "fn", "for", "if", "impl", "in", "let", "loop", "match",
		"mod", "move", "mut", "pub", "ref", "return", "self", "Self", "static", "struct",
		"super", "trait", "true", "type", "unsafe", "use", "where", "while", "async",
		"await", "dyn", "abstract", "become", "box", "do", "final", "macro", "override",
		"priv", "typeof", "unsized", "virtual", "yield", "try",
	}
	consumeIdentifierOrKeyword := rustConsumeIdentifierOrKeyword()

	consumeRawIdentifier := consumeString("r#").Then(consumeIdentifierOrKeyword).
		MapWithInput(failIfMatchTerm([]string{"r#crate", "r#self", "r#super", "r#Self"}))

	return consumeRawIdentifier.Or(consumeIdentifierOrKeyword).
		MapWithInput(recognizeKeywordOrConsume(keywords))
}
