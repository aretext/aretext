package languages

import (
	"io"
	"strings"
	"unicode"

	"github.com/aretext/aretext/syntax/parser"
)

type dockerfileParseState struct {
	kind         dockerfileParseStateKind
	escape       rune
	continuation bool
}

type dockerfileParseStateKind uint8

const (
	dockerfileParseStateStart dockerfileParseStateKind = iota
	dockerfileParseStateMaybeJSON
	dockerfileParseStateShellArgs
	dockerfileParseStateJSONArgs
	dockerfileParseStateKVArgs
	dockerfileParseStateFromArgs
	dockerfileParseStateOnbuild
	dockerfileParseStateHealthcheck
)

func (s dockerfileParseState) Equals(other parser.State) bool {
	otherState, ok := other.(dockerfileParseState)
	return ok && s == otherState
}

func DockerfileParseFunc() parser.Func {
	parseState := func(state parser.State) dockerfileParseState {
		if s, ok := state.(dockerfileParseState); ok {
			return s
		}
		return dockerfileParseState{kind: dockerfileParseStateStart, escape: '\\'}
	}

	withKind := func(kind dockerfileParseStateKind) parser.MapFn {
		return func(result parser.Result) parser.Result {
			s := parseState(result.NextState)
			s.kind = kind
			s.continuation = false
			result.NextState = s
			return result
		}
	}

	parseWhitespace := consumeRunesLike(func(r rune) bool { return r == ' ' || r == '\t' })

	parseComment := func(iter parser.TrackingRuneIter, state parser.State) parser.Result {
		s := parseState(state)
		if s.kind != dockerfileParseStateStart && !s.continuation {
			return parser.FailedResult
		}

		result := dockerfileCommentParseFunc()(iter, s)
		if result.IsFailure() {
			return parser.FailedResult
		}

		text := readInputString(iter, result.NumConsumed)
		if strings.EqualFold(strings.TrimSpace(text), "# escape=`") {
			s.escape = '`'
		}
		result.ComputedTokens = []parser.ComputedToken{{
			Length: result.NumConsumed,
			Role:   parser.TokenRoleComment,
		}}
		result.NextState = s
		return result
	}

	parseLineContinuation := func(iter parser.TrackingRuneIter, state parser.State) parser.Result {
		s := parseState(state)
		r, err := iter.NextRune()
		if err != nil || r != s.escape {
			return parser.FailedResult
		}

		var n uint64 = 1
		for {
			r, err = iter.NextRune()
			if err != nil {
				return parser.FailedResult
			}
			n++
			if r == '\n' {
				s.continuation = true
				return parser.Result{NumConsumed: n, NextState: s}
			}
			if r != ' ' && r != '\t' {
				return parser.FailedResult
			}
		}
	}

	parseNewline := func(iter parser.TrackingRuneIter, state parser.State) parser.Result {
		s := parseState(state)
		result := consumeString("\n")(iter, s)
		if result.IsFailure() {
			return parser.FailedResult
		}
		if s.kind != dockerfileParseStateStart && !s.continuation {
			s.kind = dockerfileParseStateStart
		}
		s.continuation = false
		result.NextState = s
		return result
	}

	parseInstruction := dockerfileInstructionParseFunc(withKind)
	parseOnbuildInstruction := dockerfileOnbuildInstructionParseFunc(withKind)
	parseHealthcheckCmd := dockerfileKeywordOnlyParseFunc(
		[]string{"CMD"},
		func(string) dockerfileParseStateKind { return dockerfileParseStateMaybeJSON },
		withKind)
	parseFromAs := dockerfileKeywordOnlyParseFunc(
		[]string{"AS"},
		func(string) dockerfileParseStateKind { return dockerfileParseStateFromArgs },
		withKind)

	parseJSONStart := matchStateKind(
		dockerfileParseStateMaybeJSON,
		consumeString("[").Map(withKind(dockerfileParseStateJSONArgs)))

	parseString := matchStatesKind(
		[]dockerfileParseStateKind{dockerfileParseStateJSONArgs, dockerfileParseStateKVArgs},
		parseCStyleString('"', false))

	parseHeredoc := matchStatesKind(
		[]dockerfileParseStateKind{dockerfileParseStateShellArgs, dockerfileParseStateMaybeJSON},
		dockerfileHeredocParseFunc().Map(withKind(dockerfileParseStateStart)))

	parseBody := matchStatesKind(
		[]dockerfileParseStateKind{
			dockerfileParseStateMaybeJSON,
			dockerfileParseStateShellArgs,
			dockerfileParseStateJSONArgs,
			dockerfileParseStateKVArgs,
			dockerfileParseStateFromArgs,
			dockerfileParseStateOnbuild,
			dockerfileParseStateHealthcheck,
		},
		consumeSingleRuneLike(func(r rune) bool { return r != '\n' }).
			Map(func(result parser.Result) parser.Result {
				s := parseState(result.NextState)
				if s.kind == dockerfileParseStateMaybeJSON {
					s.kind = dockerfileParseStateShellArgs
				}
				s.continuation = false
				result.NextState = s
				return result
			}))

	return initialState(
		dockerfileParseState{kind: dockerfileParseStateStart, escape: '\\'},
		parser.Func(parseLineContinuation).
			Or(parseNewline).
			Or(parseComment).
			Or(matchStateKind(dockerfileParseStateStart, parseWhitespace)).
			Or(matchStateKind(dockerfileParseStateStart, parseInstruction)).
			Or(matchStateKind(dockerfileParseStateOnbuild, parseWhitespace)).
			Or(matchStateKind(dockerfileParseStateOnbuild, parseOnbuildInstruction)).
			Or(matchStateKind(dockerfileParseStateHealthcheck, parseWhitespace)).
			Or(matchStateKind(dockerfileParseStateHealthcheck, parseHealthcheckCmd)).
			Or(matchStatesKind(
				[]dockerfileParseStateKind{
					dockerfileParseStateMaybeJSON,
					dockerfileParseStateShellArgs,
					dockerfileParseStateJSONArgs,
					dockerfileParseStateKVArgs,
					dockerfileParseStateFromArgs,
				},
				parseWhitespace)).
			Or(matchStateKind(dockerfileParseStateFromArgs, parseFromAs)).
			Or(parseJSONStart).
			Or(parseString).
			Or(parseHeredoc).
			Or(parseBody))
}

func matchStateKind(kind dockerfileParseStateKind, f parser.Func) parser.Func {
	return func(iter parser.TrackingRuneIter, state parser.State) parser.Result {
		s, ok := state.(dockerfileParseState)
		if !ok || s.kind != kind {
			return parser.FailedResult
		}
		return f(iter, state)
	}
}

func matchStatesKind(kinds []dockerfileParseStateKind, f parser.Func) parser.Func {
	return func(iter parser.TrackingRuneIter, state parser.State) parser.Result {
		s, ok := state.(dockerfileParseState)
		if !ok {
			return parser.FailedResult
		}
		for _, kind := range kinds {
			if s.kind == kind {
				return f(iter, state)
			}
		}
		return parser.FailedResult
	}
}

func dockerfileInstructionParseFunc(withKind func(dockerfileParseStateKind) parser.MapFn) parser.Func {
	return dockerfileKeywordParseFunc(
		[]string{
			"ADD", "ARG", "CMD", "COPY", "ENTRYPOINT", "ENV", "EXPOSE", "FROM",
			"HEALTHCHECK", "LABEL", "MAINTAINER", "ONBUILD", "RUN", "SHELL",
			"STOPSIGNAL", "USER", "VOLUME", "WORKDIR",
		},
		dockerfileStateAfterInstruction,
		withKind)
}

func dockerfileOnbuildInstructionParseFunc(withKind func(dockerfileParseStateKind) parser.MapFn) parser.Func {
	return dockerfileKeywordParseFunc(
		[]string{
			"ADD", "ARG", "CMD", "COPY", "ENTRYPOINT", "ENV", "EXPOSE", "FROM",
			"HEALTHCHECK", "LABEL", "MAINTAINER", "RUN", "SHELL", "STOPSIGNAL",
			"USER", "VOLUME", "WORKDIR",
		},
		dockerfileStateAfterInstruction,
		withKind)
}

func dockerfileKeywordParseFunc(
	keywords []string,
	nextKind func(string) dockerfileParseStateKind,
	withKind func(dockerfileParseStateKind) parser.MapFn,
) parser.Func {
	isWordStart := func(r rune) bool { return unicode.IsLetter(r) }
	isWordContinue := func(r rune) bool { return unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' }

	return consumeSingleRuneLike(isWordStart).
		ThenMaybe(consumeRunesLike(isWordContinue)).
		MapWithInput(func(result parser.Result, iter parser.TrackingRuneIter, state parser.State) parser.Result {
			text := readInputString(iter, result.NumConsumed)
			for _, keyword := range keywords {
				if strings.EqualFold(text, keyword) {
					result.ComputedTokens = []parser.ComputedToken{{
						Length: result.NumConsumed,
						Role:   parser.TokenRoleKeyword,
					}}
					return withKind(nextKind(keyword))(result)
				}
			}
			return withKind(dockerfileParseStateShellArgs)(result)
		})
}

func dockerfileKeywordOnlyParseFunc(
	keywords []string,
	nextKind func(string) dockerfileParseStateKind,
	withKind func(dockerfileParseStateKind) parser.MapFn,
) parser.Func {
	isWordStart := func(r rune) bool { return unicode.IsLetter(r) }
	isWordContinue := func(r rune) bool { return unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' }

	return consumeSingleRuneLike(isWordStart).
		ThenMaybe(consumeRunesLike(isWordContinue)).
		MapWithInput(func(result parser.Result, iter parser.TrackingRuneIter, state parser.State) parser.Result {
			text := readInputString(iter, result.NumConsumed)
			for _, keyword := range keywords {
				if strings.EqualFold(text, keyword) {
					result.ComputedTokens = []parser.ComputedToken{{
						Length: result.NumConsumed,
						Role:   parser.TokenRoleKeyword,
					}}
					return withKind(nextKind(keyword))(result)
				}
			}
			return parser.FailedResult
		})
}

func dockerfileStateAfterInstruction(keyword string) dockerfileParseStateKind {
	switch keyword {
	case "FROM":
		return dockerfileParseStateFromArgs
	case "ONBUILD":
		return dockerfileParseStateOnbuild
	case "HEALTHCHECK":
		return dockerfileParseStateHealthcheck
	case "LABEL", "ENV":
		return dockerfileParseStateKVArgs
	case "CMD", "ENTRYPOINT", "RUN":
		return dockerfileParseStateMaybeJSON
	case "SHELL", "VOLUME":
		return dockerfileParseStateJSONArgs
	default:
		return dockerfileParseStateShellArgs
	}
}

func dockerfileHeredocParseFunc() parser.Func {
	return func(iter parser.TrackingRuneIter, state parser.State) parser.Result {
		result := consumeString("<<").Then(consumeRunesLike(func(r rune) bool {
			return unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' || r == '-'
		}))(iter, state)
		if result.IsFailure() {
			return parser.FailedResult
		}

		word := strings.TrimPrefix(readInputString(iter, result.NumConsumed), "<<")
		iter.Skip(result.NumConsumed)
		var n = result.NumConsumed
		var line strings.Builder
		for {
			r, err := iter.NextRune()
			if err == io.EOF {
				return parser.Result{NumConsumed: n, NextState: state}
			} else if err != nil {
				return parser.FailedResult
			}
			n++

			if r == '\n' {
				if strings.TrimSuffix(line.String(), "\r") == word {
					return parser.Result{NumConsumed: n, NextState: state}
				}
				line.Reset()
				continue
			}
			line.WriteRune(r)
		}
	}
}

func dockerfileCommentParseFunc() parser.Func {
	return func(iter parser.TrackingRuneIter, state parser.State) parser.Result {
		r, err := iter.NextRune()
		if err != nil || r != '#' {
			return parser.FailedResult
		}

		var n uint64 = 1
		for {
			r, err = iter.NextRune()
			if err == io.EOF {
				break
			} else if err != nil {
				return parser.FailedResult
			}
			if r == '\n' {
				break
			}
			n++
		}

		return parser.Result{NumConsumed: n, NextState: state}
	}
}
