package languages

import (
	"io"

	"github.com/aretext/aretext/editor/syntax/parser"
)

const (
	// This is for variable and function expansions.
	// Examples:
	//   $(VAR)
	//   ${VAR}
	//   $(func arg1 arg2)
	makefileTokenRoleVariable = parser.TokenRoleCustom1

	// This is for special patterns in targets or prereqs.
	// Example:
	//    %.o: %.c
	makefileTokenRolePattern = parser.TokenRoleCustom2
)

type makefileParseState uint8

const (
	// Top-level is the initial state.
	makefileTopLevelParseState = makefileParseState(iota)

	// Rule prereq when we're in the "prerequisites" section of a rule.
	//
	// Example:
	//   foo: bar baz # "bar" and "baz" are the prereqs.
	makefileRulePrereqParseState

	// Recipe is the section of a rule with shell commands.
	//
	// Example:
	//   foo: bar baz
	//     echo "hello"  # <-- this is the recipe.
	makefileRecipeCmdParseState

	// AssignmentVal is the value assigned to a variable.
	//
	// Example:
	//   X = abc <-- "abc" is the assignment val state.
	makefileAssignmentValParseState
)

func (s makefileParseState) Equals(other parser.State) bool {
	otherState, ok := other.(makefileParseState)
	return ok && s == otherState
}

// MakefileParseFunc returns a parse func for GNU Makefiles.
// See https://www.gnu.org/software/make/manual/make.html
// especially section 3.8 "How Makefiles are Parsed"
func MakefileParseFunc() parser.Func {
	// From top-level, if we see ":" then we must be in a rule,
	// so start parsing the prereqs and then parse the following tab-indented
	// lines as recipe commands.
	parseRuleSeparator := matchState(
		makefileTopLevelParseState,
		consumeString(":").
			ThenNot(consumeString("=")).
			Map(setState(makefileRulePrereqParseState)))

	// Recipe command are tab-indented, or following rule prereqs separated by a semicolon.
	// If the start of the recipe command has "@", treat that as an operator
	// meaning "do not echo this line".
	//
	// Treating every tab-indented line as part of a recipe isn't completely accurate,
	// since technically it's missing the target/prereqs, but I believe other editors
	// also make this assumption.
	parseAtOp := consumeString("@").
		Map(recognizeToken(parser.TokenRoleOperator))

	parseTabIndent := consumeString("\n\t").
		ThenMaybe(parseAtOp).
		Map(setState(makefileRecipeCmdParseState))

	parseSemicolonInRule := matchState(
		makefileRulePrereqParseState,
		consumeString(";").
			ThenMaybe(consumeRunesLike(func(r rune) bool { return r == ' ' || r == '\t' })).
			ThenMaybe(parseAtOp).
			Map(setState(makefileRecipeCmdParseState)))

	// Handle backslash line continuation everywhere except in recipe command.
	// This consumes the backslash and newline, then stays in the same state
	// rather than transitioning back to top-level.
	parseBackslashLineContinuation := matchStates(
		[]parser.State{makefileRulePrereqParseState, makefileRecipeCmdParseState, makefileAssignmentValParseState},
		consumeString(`\`).
			ThenMaybe(consumeRunesLike(func(r rune) bool { return r == ' ' || r == '\t' })).
			Then(consumeString("\n")))

	// A newline NOT followed by a tab transitions back to top-level state.
	parseBackToTopLevel := matchStates(
		[]parser.State{makefileRulePrereqParseState, makefileRecipeCmdParseState, makefileAssignmentValParseState},
		consumeString("\n").
			Map(setState(makefileTopLevelParseState)))

	// Parse some keywords only at top-level.
	parseTopLevelKeywords := matchState(
		makefileTopLevelParseState,
		consumeLongestMatchingOption([]string{
			"ifeq", "ifneq", "ifdef", "ifndef", "else", "endif",
			"export", "unexport", "override",
			"include", "-include", "define", "endef",
		}).Map(recognizeToken(parser.TokenRoleKeyword)))

	// Parse comments everywhere except in recipe commands.
	parseComment := matchStates(
		[]parser.State{makefileTopLevelParseState, makefileRulePrereqParseState},
		makefileCommentParseFunc())

	// Parse assign operators everywhere except in recipe commands.
	parseAssignOp := matchStates(
		[]parser.State{makefileTopLevelParseState, makefileRulePrereqParseState},
		consumeLongestMatchingOption([]string{"=", ":=", "::=", ":::=", "?=", "+=", "!="}).
			Map(recognizeToken(parser.TokenRoleOperator)).
			Map(setState(makefileAssignmentValParseState)))

	// Parse escaped dollar sign ($$).
	// This must come before any rule that parses "$" as a prefix
	// to ensure that "$$" isn't highlighted.
	parseEscapedDollarSign := consumeString("$$")

	// Parse automatic variables in all states.
	parseAutomaticVariables := consumeLongestMatchingOption([]string{
		"$@", "$%", "$<", "$?", "$^", "$+", "$|", "$*",
	}).Map(recognizeToken(makefileTokenRoleVariable))

	// Parse patterns only in rule target and prereqs (NOT in assignment).
	parseRulePattern := matchStates(
		[]parser.State{makefileTopLevelParseState, makefileRulePrereqParseState},
		consumeLongestMatchingOption([]string{
			"%", "%D", "%F", "+", "+D", "+F",
		}).Map(recognizeToken(makefileTokenRolePattern)))

	// Parse expansions (functions and variables) in all states.
	parseExpansion := consumeString("$").
		Then(makefileExpansionParseFunc()).
		Map(recognizeToken(makefileTokenRoleVariable))

	return initialState(
		makefileTopLevelParseState,
		parseRuleSeparator.
			Or(parseTabIndent).
			Or(parseSemicolonInRule).
			Or(parseBackslashLineContinuation).
			Or(parseBackToTopLevel).
			Or(parseComment).
			Or(parseAssignOp).
			Or(parseEscapedDollarSign).
			Or(parseAutomaticVariables).
			Or(parseRulePattern).
			Or(parseExpansion).
			Or(parseTopLevelKeywords))
}

// makefileCommentParseFunc parses a comment in a makefile.
// Comments are just "# ..." to end of line, but do NOT consume
// the line feed, since that determines state transitions.
func makefileCommentParseFunc() parser.Func {
	return func(iter parser.TrackingRuneIter, state parser.State) parser.Result {
		var numConsumed uint64

		startRune, err := iter.NextRune()
		if err != nil || startRune != '#' {
			return parser.FailedResult
		}
		numConsumed++

		for {
			r, err := iter.NextRune()
			if err == io.EOF {
				break
			} else if err != nil {
				return parser.FailedResult
			}

			if r == '\n' {
				// Do NOT consume the line feed.
				break
			}

			numConsumed++
		}

		return parser.Result{
			NumConsumed: numConsumed,
			ComputedTokens: []parser.ComputedToken{
				{
					Length: numConsumed,
					Role:   parser.TokenRoleComment,
				},
			},
			NextState: state,
		}
	}
}

// makefileExpansionParseFunc handles variable and function expansions.
//
// Examples:
//
//	${VAR}
//	$(VAR)
//	$(subst $(space),$(comma),$(foo))
func makefileExpansionParseFunc() parser.Func {
	return func(iter parser.TrackingRuneIter, state parser.State) parser.Result {
		var n uint64

		// Open delimiter.
		startRune, err := iter.NextRune()
		if err != nil || !(startRune == '{' || startRune == '(') {
			return parser.FailedResult
		}
		n++

		// Maintain a stack of open delimiters so we can check when they're closed.
		stack := []rune{startRune}

		// Consume runes until the stack is empty.
		for len(stack) > 0 {
			stackTop := stack[len(stack)-1]

			r, err := iter.NextRune()
			if err != nil {
				return parser.FailedResult
			}
			n++

			if r == '{' || r == '(' {
				// Push open delimiter to stack.
				stack = append(stack, r)
			} else if (stackTop == '{' && r == '}') || (stackTop == '(' && r == ')') {
				// Found close delimiter matching last open delimiter, so pop from stack.
				stack = stack[0 : len(stack)-1]
			}
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
