package languages

import (
	"io"
	"strings"
	"unicode"

	"github.com/aretext/aretext/editor/syntax/parser"
)

const (
	bashTokenRoleVariable           = parser.TokenRoleCustom1
	bashTokenRoleBackquoteExpansion = parser.TokenRoleCustom2
)

type bashParseState uint8

const (
	bashParseStateNormal = bashParseState(iota)
	bashParseStateInConditional
)

func (s bashParseState) Equals(other parser.State) bool {
	otherState, ok := other.(bashParseState)
	return ok && s == otherState
}

// BashParseFunc returns a parse func for bash.
// See https://www.gnu.org/software/bash/manual/bash.html
//
// Some known limitations with this implementation:
// * reserved keywords "in" and "do" recognized outside the context of a case/select/for.
// * no special parsing for numbers or arithmetic expressions $((...))
// * no special handling for file redirects after heredoc word.
func BashParseFunc() parser.Func {
	parseComment := consumeString("#").
		ThenMaybe(consumeToNextLineFeed).
		Map(recognizeToken(parser.TokenRoleComment))

	parseEscape := consumeString("\\").
		Then(consumeSingleRuneLike(func(r rune) bool { return r != '\n' }))

	// We are NOT handling context-sensitive rules for "in" an "do"
	// From the GNU bash manual:
	// "in is recognized as a reserved word if it is the third word of a case or select command.
	// in and do are recognized as reserved words if they are the third word in a for command."
	isWordStart := func(r rune) bool { return unicode.IsLetter(r) || r == '_' }
	isWordContinue := func(r rune) bool { return unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' || r == '-' }
	keywords := []string{
		"if", "then", "elif", "else", "fi", "time",
		"for", "in", "until", "while", "do", "done",
		"case", "esac", "coproc", "select", "function",
	}
	parseKeyword := consumeSingleRuneLike(isWordStart).
		ThenMaybe(consumeRunesLike(isWordContinue)).
		MapWithInput(recognizeKeywordOrConsume(keywords))

	isVariableNameRune := func(r rune) bool { return unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' }
	parseVariable := consumeString("$").
		Then(consumeRunesLike(isVariableNameRune).
			Or(consumeLongestMatchingOption([]string{"!", "#", "$", "*", "-", "0", "?", "@", "-"}))).
		Map(recognizeToken(bashTokenRoleVariable))

	parseVariableBrace := consumeString("$").
		Then(bashExpansionParseFunc('{')).
		Map(recognizeToken(bashTokenRoleVariable))

	consumeSingleQuoteString := consumeString("'").Then(consumeToString("'"))
	parseString := consumeSingleQuoteString.
		Or(bashExpansionParseFunc('"')).
		Map(recognizeToken(parser.TokenRoleString))

	parseBackquoteExpansion := bashExpansionParseFunc('`').
		Map(recognizeToken(bashTokenRoleBackquoteExpansion))

	parseHeredoc := bashHeredocParseFunc()

	parseOperator := consumeLongestMatchingOption([]string{
		"$", "!", "<", ">", "<<", ">>",
		"|", "&&", "||", "=", "+=", "&>", ">&", "&>>",
	}).Map(recognizeToken(parser.TokenRoleOperator))

	parseStartConditional := matchState(
		bashParseStateNormal,
		consumeString("[[").
			Map(setState(bashParseStateInConditional)))

	parseEndConditional := matchState(
		bashParseStateInConditional,
		consumeString("]]").
			Map(setState(bashParseStateNormal)))

	parseConditionalOperator := matchState(
		bashParseStateInConditional,
		consumeLongestMatchingOption([]string{
			"=~", "^", "==", "!=",
		}).Map(recognizeToken(parser.TokenRoleOperator)))

	parseConditional := parseStartConditional.
		Or(parseEndConditional).
		Or(parseConditionalOperator)

	return initialState(
		bashParseStateNormal,
		parseComment.
			Or(parseEscape).
			Or(parseString).
			Or(parseBackquoteExpansion).
			Or(parseConditional).
			Or(parseKeyword).
			Or(parseVariableBrace).
			Or(parseVariable).
			Or(parseHeredoc).
			Or(parseOperator))
}

// bashExpansionParseFunc handles expressions that can contain expansions.
//
// Examples:
//
//	"${echo "foo"}"
//	`echo $(echo "`")`
//	$(echo "$(pwd)")
//
// In these cases, we can't simply scan ahead for the next matching end delimiter,
// because the delimiter could be quoted in an expansion. We need to keep track
// of each expansion and string start/end delimiter.
func bashExpansionParseFunc(startRune rune) parser.Func {
	// States that can be pushed onto the stack.
	const (
		STRING    = iota // In a double-quoted string.
		SUBSHELL         // In a subshell expansion $(...)
		VARIABLE         // In a variable expansion ${...}
		BACKQUOTE        // In a backquote expansion `...`
		ESCAPE           // In a backslash escape like \$ or \"
	)

	return func(iter parser.TrackingRuneIter, state parser.State) parser.Result {
		var n uint64

		// Open delimiter.
		r, err := iter.NextRune()
		if err != nil || r != startRune {
			return parser.FailedResult
		}
		n++

		// Setup initial stack.
		var stack []int
		switch r {
		case '"':
			stack = append(stack, STRING)
		case '`':
			stack = append(stack, BACKQUOTE)
		case '{':
			stack = append(stack, VARIABLE)
		default:
			// Should never happen if this helper is used correctly.
			panic("Invalid start rune for bash expansion")
		}

		// Consume runes until the stack is empty.
		for len(stack) > 0 {
			r, err = iter.NextRune()
			if err != nil {
				return parser.FailedResult
			}
			n++

			switch stack[len(stack)-1] {

			case ESCAPE:
				// Escape this character.
				stack = stack[0 : len(stack)-1]
				goto loopend

			case STRING:
				if r == '"' {
					// Found end quote.
					stack = stack[0 : len(stack)-1]
					goto loopend
				}

			case SUBSHELL:
				if r == ')' {
					// Found end of $(...)
					stack = stack[0 : len(stack)-1]
					goto loopend
				}

			case VARIABLE:
				if r == '}' {
					// Found end of ${...}
					stack = stack[0 : len(stack)-1]
					goto loopend
				}

			case BACKQUOTE:
				if r == '`' {
					// Found end of `...`
					stack = stack[0 : len(stack)-1]
					goto loopend
				}

			default:
				panic("Unrecognized state for bash expansion")
			}

			// Fallthrough case: if we find a start delimiter,
			// push the corresponding state onto the stack.
			if r == '"' {
				stack = append(stack, STRING)
			} else if r == '$' {
				lookaheadIter := iter
				r, err = lookaheadIter.NextRune()
				if err == nil {
					if r == '(' {
						n++
						iter = lookaheadIter
						stack = append(stack, SUBSHELL)
					} else if r == '{' {
						n++
						iter = lookaheadIter
						stack = append(stack, VARIABLE)
					}
				}
			} else if r == '`' {
				stack = append(stack, BACKQUOTE)
			} else if r == '\\' {
				stack = append(stack, ESCAPE)
			}

		loopend:
			continue
		}

		return parser.Result{
			NumConsumed: n,
			ComputedTokens: []parser.ComputedToken{
				{Length: n},
			},
			NextState: state,
		}
	}
}

// bashHeredocParseFunc returns a parse func for a "here document".
//
// Example:
//
//	cat << EOF
//	  this is a heredoc
//	EOF
//
// We need to first identify the "end word" for the heredoc,
// then scan ahead until we find that word.
//
// Bash allows the heredoc word to be followed by a redirection, like this:
//
//	cat << EOF > out.txt
//	   this is a heredoc
//	EOF
//
// ... but we currently don't do any special handling for that, so "> out.txt"
// is considered part of the heredoc string.
func bashHeredocParseFunc() parser.Func {
	consumeOpen := consumeLongestMatchingOption([]string{"<<", "<<<", "<<-", "<<<-"})

	findHeredocWord := func(iter parser.TrackingRuneIter) (string, error) {
		var sb strings.Builder
		for {
			r, err := iter.NextRune()
			if err != nil {
				return "", err
			} else if r == '\n' || unicode.IsSpace(r) {
				word := sb.String()
				// If the word is quoted, search for the unquoted word.
				if len(word) >= 2 && ((word[0] == '\'' && word[len(word)-1] == '\'') || (word[0] == '"' && word[len(word)-1] == '"')) {
					word = word[1 : len(word)-1]
				} else if len(word) >= 1 && word[0] == '\\' {
					word = word[1:]
				}
				return word, nil
			} else {
				sb.WriteRune(r)
			}
		}
	}

	findEndOfHeredoc := func(iter parser.TrackingRuneIter, word string) (uint64, error) {
		var (
			n                 uint64
			sb                strings.Builder
			couldBeOnThisLine bool // Initially false since we're starting on the line that begins the heredoc.
		)
		for {
			r, err := iter.NextRune()
			if err != nil {
				return 0, err
			}

			n++

			if !couldBeOnThisLine {
				if r == '\n' {
					couldBeOnThisLine = true
					sb.Reset()
				}
				continue
			}

			if r == '\n' {
				if len(word) == 0 {
					// Special case: searching for quoted word "" and found an empty line.
					return n, nil
				}
				sb.Reset()
				continue
			}

			sb.WriteRune(r)

			if sb.Len() == len(word) {
				if sb.String() == word {
					// Found the word. Verify that the next rune is either a line feed or EOF.
					lookaheadIter := iter
					nextRune, err := lookaheadIter.NextRune()
					if nextRune == '\n' || err == io.EOF {
						// Found it!
						return n, nil
					}
				}

				// Not on this line, keep looking.
				couldBeOnThisLine = false
				continue
			}
		}
	}

	consumeHeredocWordAndContent := parser.Func(func(iter parser.TrackingRuneIter, state parser.State) parser.Result {
		// Find the word used to indicate the end of the here-document.
		heredocWord, err := findHeredocWord(iter)
		if err != nil {
			return parser.FailedResult
		}

		// Consume until we see a line containing only the word.
		heredocLength, err := findEndOfHeredoc(iter, heredocWord)
		if err != nil {
			return parser.FailedResult
		}

		return parser.Result{
			NumConsumed: heredocLength,
			ComputedTokens: []parser.ComputedToken{
				{Length: heredocLength},
			},
			NextState: state,
		}
	})

	return consumeOpen.Map(recognizeToken(parser.TokenRoleOperator)).
		ThenMaybe(consumeRunesLike(func(r rune) bool { return unicode.IsSpace(r) })).
		Then(consumeHeredocWordAndContent.Map(recognizeToken(parser.TokenRoleString)))
}
