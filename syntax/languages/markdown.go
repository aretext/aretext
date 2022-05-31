package languages

import (
	"unicode"

	"github.com/aretext/aretext/syntax/parser"
)

const (
	markdownCodeBlockRole      = parser.TokenRoleString
	markdownCodeSpanRole       = parser.TokenRoleString
	markdownListNumberRole     = parser.TokenRoleNumber
	markdownListBulletRole     = parser.TokenRoleOperator
	markdownThematicBreakRole  = parser.TokenRoleOperator
	markdownHeadingRole        = parser.TokenRoleCustom1
	markdownEmphasisRole       = parser.TokenRoleCustom2
	markdownStrongEmphasisRole = parser.TokenRoleCustom3
	markdownLinkRole           = parser.TokenRoleCustom4
)

type markdownParseState uint8

const (
	markdownParseStateNormal = markdownParseState(iota)
	markdownParseStateInListItem
)

func (s markdownParseState) Equals(other parser.State) bool {
	otherState, ok := other.(markdownParseState)
	return ok && s == otherState
}

// MarkdownParseFunc returns a parse func for Markdown.
// This attempts to follow the CommonMark 0.30 spec,
// but deviates in some cases to simplify the implementation.
//
// Known limitations include:
// * Incorrect handling of nested emphasis in some cases.
// * No support for inline HTML.
// * No support for autolinks.
// * No support for indented code blocks.
// * No support for block quotes.
// * No support for entity and numeric character references.
// * Some differences in handling of nested lists.
// * Some differences in handling link and code span precedence.
// * No restriction on the number of digits in a list item.
// * No restriction on indentation for lists, code fences, headings, etc.
//
// See https://spec.commonmark.org/0.30/ for details.
//
func MarkdownParseFunc() parser.Func {
	// Incrementally parse one block at a time (headings, paragraphs, list items, etc.).
	// This ensures that each parse func invocation starts at the beginning of a line.
	parseListItem := markdownNumberListItemParseFunc().
		Or(markdownBulletListItemParseFunc()).
		Map(setState(markdownParseStateInListItem))

	parseThematicBreak := matchState(
		markdownParseStateNormal,
		markdownThematicBreakParseFunc())

	parseHeadings := matchState(
		markdownParseStateNormal,
		markdownAtxHeadingParseFunc().
			Or(markdownSetextHeadingParseFunc()),
	)

	parseOtherBlocks := markdownFencedCodeBlockParseFunc().
		Or(markdownParagraphParseFunc()).
		Or(consumeToNextLineFeed).
		Map(setState(markdownParseStateNormal))

	return initialState(
		markdownParseStateNormal,
		parseThematicBreak.
			Or(parseListItem).
			Or(parseHeadings).
			Or(parseOtherBlocks),
	)
}

func markdownSkipLeadingIndentation(iter *parser.TrackingRuneIter) uint64 {
	lookaheadIter := *iter
	var n uint64
	for {
		r, err := lookaheadIter.NextRune()
		if err != nil || !(r == ' ' || r == '\t') {
			break
		}
		n++
	}
	iter.Skip(n)
	return n
}

func markdownThematicBreakParseFunc() parser.Func {
	// A thematic break consists of three or more matching '-', '_', or '*''s,
	// optionally preceded and/or followed by whitespace.
	return func(iter parser.TrackingRuneIter, state parser.State) parser.Result {
		var n uint64
		var breakRune rune
		var breakCount int
		for {
			r, err := iter.NextRune()
			if err != nil {
				// End of text.
				break
			}

			n++

			if r == '\n' {
				// End of line (include the newline).
				break
			} else if breakRune == '\x00' && (r == '-' || r == '_' || r == '*') {
				// Start matching thematic break characters.
				breakRune = r
				breakCount = 1
				continue
			} else if breakRune == r {
				// Continue matching thematic break character.
				breakCount++
				continue
			} else if r == ' ' || r == '\t' {
				// Allow whitespace between thematic break characters.
				continue
			} else {
				return parser.FailedResult
			}
		}

		if breakCount < 3 {
			// Need at least three thematic break characters.
			return parser.FailedResult
		}

		return parser.Result{
			NumConsumed: n,
			ComputedTokens: []parser.ComputedToken{
				{
					Offset: 0,
					Length: n,
					Role:   markdownThematicBreakRole,
				},
			},
			NextState: state,
		}
	}
}

func markdownAtxHeadingParseFunc() parser.Func {
	// An ATX heading consists of a sequence of between 1 and 6 "#", optionally preceded by
	// indentation, followed by a space then the rest of the line.
	// The CommonMark spec has some additional requirements for sequences of closing "#"
	// at the end of the line, but we don't enforce those.
	consumeOpener := func(iter parser.TrackingRuneIter, state parser.State) parser.Result {
		// Leading indentation.
		indentCount := markdownSkipLeadingIndentation(&iter)

		// Allow between 1-6 "#"'s.
		var n uint64
		for {
			r, err := iter.NextRune()
			if err != nil || r != '#' {
				break
			}
			n++
		}

		if n < 1 || n > 6 {
			return parser.FailedResult
		}

		return parser.Result{
			NumConsumed: indentCount + n,
			NextState:   state,
		}
	}

	return parser.Func(consumeOpener).
		ThenNot(consumeSingleRuneLike(func(r rune) bool { return !unicode.IsSpace(r) })).
		Then(consumeToNextLineFeed).
		Map(recognizeToken(markdownHeadingRole))
}

func markdownSetextHeadingParseFunc() parser.Func {
	// A setext heading consists of one or more non-blank lines, followed by
	// a setext underline (sequence of one or more "-" or "=").
	// The setext underline may have leading indentation and/or trailing whitespace.
	consumeFirstLine := func(iter parser.TrackingRuneIter, state parser.State) parser.Result {
		// Leading indentation.
		indentCount := markdownSkipLeadingIndentation(&iter)

		// Consume the rest of the line.
		var n uint64
		var r rune
		var err error
		for {
			r, err = iter.NextRune()
			if err != nil {
				break
			}

			n++

			if r == '\n' {
				break
			}
		}

		// Heading cannot be empty.
		if (r == '\n' && n < 2) || (r != '\n' && n < 1) {
			return parser.FailedResult
		}

		return parser.Result{
			NumConsumed: indentCount + n,
			NextState:   state,
		}
	}

	checkSubsequentLine := func(iter parser.TrackingRuneIter) (uint64, bool) {
		var n uint64
		for {
			r, err := iter.NextRune()
			if err != nil {
				if n == 0 {
					// Empty line ending in EOF.
					return 0, false
				} else {
					// Non-empty line ending in EOF.
					return n, true
				}
			}

			n++

			if r == '\n' {
				if n == 1 {
					// Empty line ending in a newline.
					return 0, false
				} else {
					// Non-empty line ending in a newline.
					return n, true
				}
			}
		}
	}

	checkUnderline := func(iter parser.TrackingRuneIter) (uint64, bool) {
		var n uint64
		indentCount := markdownSkipLeadingIndentation(&iter)
		n += indentCount

		// Check if this is an '-' or '=' underline.
		underlineRune, err := iter.NextRune()
		if err != nil || !(underlineRune == '-' || underlineRune == '=') {
			return 0, false
		}
		n++

		// Consume repeats of the underline rune.
		for {
			r, err := iter.NextRune()
			if err != nil {
				// Found EOF (without trailing whitespace),
				// so this is a valid setext underline.
				return n, true
			}

			n++

			if r == '\n' {
				// Found end of line (without trailing whitespace),
				// so this is a valid setext underline.
				return n, true
			} else if r == ' ' || r == '\t' {
				break
			} else if r != underlineRune {
				return 0, false
			}
		}

		// Consume trailing whitespace.
		for {
			r, err := iter.NextRune()
			if err != nil {
				break
			}

			n++

			if r == '\n' {
				break
			} else if r == ' ' || r == '\t' {
				continue
			} else {
				return 0, false
			}
		}

		// Found a valid setext underline.
		return n, true
	}

	consumeToUnderline := func(iter parser.TrackingRuneIter, state parser.State) parser.Result {
		var n uint64
		for {
			// Check if the current line is a setext underline.
			underlineLen, found := checkUnderline(iter)
			if found {
				// Found a setext underline, so consume it.
				n += underlineLen
				return parser.Result{
					NumConsumed: n,
					NextState:   state,
				}
			}

			// Check if the current line is non-empty.
			lineLen, found := checkSubsequentLine(iter)
			if found {
				// Found a non-empty line. Consume it and keep looking for the setext underline.
				n += lineLen
				iter.Skip(lineLen)
				continue
			}

			// Otherwise, we found an empty line, so this isn't a setext heading.
			return parser.FailedResult
		}
	}

	return parser.Func(consumeFirstLine).
		Then(consumeToUnderline).
		Map(recognizeToken(markdownHeadingRole))
}

func markdownFencedCodeBlockParseFunc() parser.Func {
	// A fenced code block consists of a fence ("```" or "~~~" of length >= 3)
	// until a closing fence of at least the same length or EOF.
	// The fences may have leading indentation.
	// Commonmark allows the opening fence to be followed by
	// an optional "info" string (e.g. specifying the code language), which we include
	// within the coe block token (no special treatment).
	checkFenceLen := func(fenceRune rune, iter parser.TrackingRuneIter) (uint64, bool) {
		var n uint64
		for {
			r, err := iter.NextRune()
			if err != nil || r != fenceRune {
				break
			}
			n++
		}

		if n < 3 {
			return 0, false
		}
		return n, true
	}

	checkClosingCodeFence := func(fenceRune rune, openFenceLen uint64, iter parser.TrackingRuneIter) (uint64, bool) {
		var n uint64
		for {
			maybeFence := true

			// Leading indentation.
			indentCount := markdownSkipLeadingIndentation(&iter)
			n += indentCount

			closeFenceLen, found := checkFenceLen(fenceRune, iter)
			if found && closeFenceLen >= openFenceLen {
				iter.Skip(closeFenceLen)
				n += closeFenceLen
			} else {
				maybeFence = false
			}

			// Consume to the end of the line or file.
			for {
				r, err := iter.NextRune()
				if err != nil {
					// If we hit the EOF, then close the code block.
					return n, true
				}
				n++
				if r == '\n' {
					break
				} else if maybeFence && !(r == ' ' || r == '\t') {
					// Only trailing whitespace allowed after code fence.
					maybeFence = false
				}
			}

			if maybeFence {
				return n, true
			}
		}
	}

	return func(iter parser.TrackingRuneIter, state parser.State) parser.Result {
		var n uint64

		// Leading indentation.
		indentCount := markdownSkipLeadingIndentation(&iter)
		n += indentCount

		// Read the opening fence (first check '`', then fallback to '~')
		fenceRune := '`'
		openFenceLen, found := checkFenceLen(fenceRune, iter)
		if !found {
			fenceRune = '~'
			openFenceLen, found = checkFenceLen(fenceRune, iter)
		}

		if !found || openFenceLen < 3 {
			return parser.FailedResult
		}

		iter.Skip(openFenceLen)
		n += openFenceLen

		// Consume to the end of the first line.
		for {
			r, err := iter.NextRune()
			if err != nil {
				break
			}
			n++
			if r == '\n' {
				break
			}
		}

		// Read subsequent lines until we find a closing code fence or EOF.
		for {
			lineLen, found := checkClosingCodeFence(fenceRune, openFenceLen, iter)
			n += lineLen
			iter.Skip(lineLen)
			if found {
				break
			}
		}

		// Found the end of the code fence, so return the token.
		return parser.Result{
			NumConsumed: n,
			ComputedTokens: []parser.ComputedToken{
				{
					Offset: 0,
					Length: n,
					Role:   markdownCodeBlockRole,
				},
			},
			NextState: state,
		}
	}
}

func markdownNumberListItemParseFunc() parser.Func {
	// A numbered list item is a sequence of digits followed by '.' or ')' and a space,
	// optionally preceded by indentation.
	// Commonmark requires no more than nine digits, but we allow more.
	return consumeRunesLike(func(r rune) bool { return r == ' ' || r == '\t' }).
		MaybeBefore(
			consumeRunesLike(func(r rune) bool { return r >= '0' && r <= '9' }).
				Then(consumeSingleRuneLike(func(r rune) bool { return r == '.' || r == ')' })).
				Map(recognizeToken(markdownListNumberRole))).
		Then(consumeSingleRuneLike(func(r rune) bool { return r == ' ' || r == '\t' }))
}

func markdownBulletListItemParseFunc() parser.Func {
	// A bullet list item is a '-', '+', or '*' character followed by a space,
	// optionally preceded by indentation.
	return consumeRunesLike(func(r rune) bool { return r == ' ' || r == '\t' }).
		MaybeBefore(
			consumeSingleRuneLike(func(r rune) bool { return r == '-' || r == '+' || r == '*' }).
				Map(recognizeToken(markdownListBulletRole))).
		Then(consumeSingleRuneLike(func(r rune) bool { return r == ' ' || r == '\t' }))
}

func markdownParagraphParseFunc() parser.Func {
	// A paragraph consists of a sequence of non-empty lines that cannot be interpreted
	// as another kind of block.
	// We parse paragraphs in two passes: first find the paragraph contents,
	// then tokenize the paragraph's inlines (emphasis, links, etc.).
	isEmptyLineOrEof := func(iter parser.TrackingRuneIter) bool {
		r, err := iter.NextRune()
		return err != nil || r == '\n'
	}

	parseNumberList := markdownNumberListItemParseFunc()
	parseBulletList := markdownBulletListItemParseFunc()
	parseAtxHeading := markdownAtxHeadingParseFunc()
	parseThematicBreak := markdownThematicBreakParseFunc()
	parseStartOfCodeBlock := consumeRunesLike(func(r rune) bool { return r == ' ' || r == '\t' }).
		MaybeBefore(consumeString("```").Or(consumeString("~~~")))

	isStartOfAnotherBlock := func(iter parser.TrackingRuneIter, state parser.State) bool {
		// Setext headings are already handled by an earlier parse func.
		return (parseNumberList(iter, state).IsSuccess() ||
			parseBulletList(iter, state).IsSuccess() ||
			parseAtxHeading(iter, state).IsSuccess() ||
			parseThematicBreak(iter, state).IsSuccess() ||
			parseStartOfCodeBlock(iter, state).IsSuccess())
	}

	consumeParagraphLines := func(iter parser.TrackingRuneIter, state parser.State) parser.Result {
		var n uint64

		for {
			if isEmptyLineOrEof(iter) || isStartOfAnotherBlock(iter, state) {
				// Found end of paragraph.
				break
			}

			// Leading indentation.
			indentCount := markdownSkipLeadingIndentation(&iter)
			n += indentCount

			// Consume to end of line.
			for {
				r, err := iter.NextRune()
				if err != nil {
					break
				}
				n++
				if r == '\n' {
					break
				}
			}
		}

		if n == 0 {
			// Didn't consume any lines, so this isn't a paragraph.
			return parser.FailedResult
		}

		return parser.Result{
			NumConsumed: n,
			NextState:   state,
		}
	}

	parseInlineCodeSpan := func(iter parser.TrackingRuneIter, state parser.State) parser.Result {
		const (
			codeSpanStateNone = iota
			codeSpanStateStartDelim
			codeSpanStateContent
			codeSpanStateEndDelim
		)

		var done bool
		var tokenStart uint64
		var startDelimLen, endDelimLen int
		var codeSpanState int
		var n uint64
		for !done {
			r, err := iter.NextRune()
			if err != nil {
				// One more iteration for tokens ending at EOF.
				done = true
				r = '\x00'
			} else {
				n++
			}

			switch codeSpanState {
			case codeSpanStateNone:
				if r == '`' {
					tokenStart = n - 1
					startDelimLen = 1
					codeSpanState = codeSpanStateStartDelim
				} else {
					// Code span must start with a backtick.
					return parser.FailedResult
				}

			case codeSpanStateStartDelim:
				if r == '`' {
					startDelimLen++
				} else {
					codeSpanState = codeSpanStateContent
				}

			case codeSpanStateContent:
				if r == '`' {
					endDelimLen = 1
					codeSpanState = codeSpanStateEndDelim
				}

			case codeSpanStateEndDelim:
				if r == '`' {
					endDelimLen++
				} else if startDelimLen != endDelimLen {
					endDelimLen = 0
					codeSpanState = codeSpanStateContent
				} else {
					tokenEnd := n
					if !done {
						// Compensate for lookahead character (unless we're at EOF)
						tokenEnd--
					}
					return parser.Result{
						NumConsumed: tokenEnd,
						ComputedTokens: []parser.ComputedToken{
							{
								Offset: tokenStart,
								Length: tokenEnd - tokenStart,
								Role:   markdownCodeSpanRole,
							},
						},
						NextState: state,
					}
				}
			}
		}

		if startDelimLen > 0 {
			// Skip an unmatched start delimiter to avoid partial matches later.
			return parser.Result{
				NumConsumed: tokenStart + uint64(startDelimLen),
				NextState:   state,
			}
		}

		return parser.FailedResult
	}

	// Parse inline emphasis and strong emphasis, delimited by '*' or '_'.
	// This implementation doesn't handle all the edge cases in the CommonMark spec
	// involving nested emphasis, but it should handle the most common cases reasonably.
	parseInlineEmphasis := func(delimRune rune, allowWithinWord bool) parser.Func {
		return func(iter parser.TrackingRuneIter, state parser.State) parser.Result {
			const (
				emphStateNone = iota
				emphStateStartDelim
				emphStateContent
				emphStateEndDelim
			)

			var emphState int
			var done bool
			var tokenStart uint64
			var startDelimLen, endDelimLen int
			var lastWasSpace bool
			var lastWasDelim bool
			var n uint64
			for !done {
				r, err := iter.NextRune()
				if err != nil {
					done = true
					r = '\x00'
				} else {
					n++
				}

				switch emphState {
				case emphStateNone:
					if r == delimRune {
						tokenStart = n - 1
						startDelimLen = 1
						emphState = emphStateStartDelim
					} else {
						return parser.FailedResult
					}

				case emphStateStartDelim:
					if r == delimRune {
						startDelimLen++
					} else if unicode.IsSpace(r) {
						return parser.FailedResult
					} else {
						endDelimLen = 0
						emphState = emphStateContent
					}
				case emphStateContent:
					if r == delimRune && !lastWasSpace && !lastWasDelim {
						endDelimLen++
						emphState = emphStateEndDelim
					} else if r == '`' {
						// Code span takes precedence.
						return parser.Result{
							NumConsumed: n - 1,
							NextState:   state,
						}
					}
				case emphStateEndDelim:
					if r == delimRune {
						endDelimLen++
					} else if !allowWithinWord && !(unicode.IsSpace(r) || unicode.IsPunct(r) || r == '\x00') {
						// Disallow delimiters within a word for "_".
						// For example "foo_bar_baz" should not emphasis "bar".
						endDelimLen = 0
						emphState = emphStateContent
					} else if endDelimLen < startDelimLen {
						emphState = emphStateContent
					} else {
						var role parser.TokenRole
						if startDelimLen < 2 {
							role = markdownEmphasisRole
						} else {
							role = markdownStrongEmphasisRole
						}

						tokenEnd := n
						if !done {
							// Compensate for lookahead character (unless we're at EOF)
							tokenEnd--
						}

						return parser.Result{
							NumConsumed: tokenEnd,
							ComputedTokens: []parser.ComputedToken{
								{
									Offset: tokenStart,
									Length: tokenEnd - tokenStart,
									Role:   role,
								},
							},
							NextState: state,
						}
					}
				}

				lastWasSpace = unicode.IsSpace(r)
				lastWasDelim = r == delimRune
			}

			if startDelimLen > 0 {
				// Skip an unmatched start delimiter to avoid partial matches later.
				return parser.Result{
					NumConsumed: tokenStart + uint64(startDelimLen),
					NextState:   state,
				}
			}

			return parser.FailedResult
		}
	}

	consumeLinkPart := func(startDelim, endDelim rune) parser.Func {
		return func(iter parser.TrackingRuneIter, state parser.State) parser.Result {
			var n uint64
			var depth int
			for {
				r, err := iter.NextRune()
				if err != nil {
					break
				}
				n++

				if n == 1 && r != startDelim {
					return parser.FailedResult
				}

				if r == '\\' {
					// Backslash escape.
					n++
					iter.Skip(1)
					continue
				}

				if r == '\n' {
					// Links cannot contain newlines.
					return parser.FailedResult
				}

				if r == startDelim {
					depth++
				} else if r == endDelim {
					depth--
				}

				if depth == 0 {
					return parser.Result{
						NumConsumed: n,
						NextState:   state,
					}
				}
			}

			return parser.FailedResult
		}
	}

	parseInlineLink := consumeString("!").
		MaybeBefore(consumeLinkPart('[', ']')).
		ThenMaybe(consumeLinkPart('(', ')')).
		Map(recognizeToken(markdownLinkRole))

	consumeToNextPossibleStartDelim := func(iter parser.TrackingRuneIter, state parser.State) parser.Result {
		allowUnderscore := true
		var n uint64
		for {
			r, err := iter.NextRune()
			if err != nil || r == '\\' || r == '`' || r == '*' || r == '!' || r == '[' || (allowUnderscore && r == '_') {
				break
			}
			n++

			// Don't allow an underscore within a word or following another underscore.
			allowUnderscore = r != '_' && (unicode.IsSpace(r) || unicode.IsPunct(r))
		}
		return parser.Result{
			NumConsumed: n,
			NextState:   state,
		}
	}

	consumeBackslashEscape := consumeString("\\").
		Then(consumeSingleRuneLike(func(r rune) bool {
			// ASCII punctuation
			return (r >= '!' && r <= '/') || (r >= ':' && r <= '@') || (r >= '[' && r <= '`') || (r >= '{' && r <= '~')
		}))

	parseInlineToken := consumeBackslashEscape.
		Or(parseInlineCodeSpan).
		Or(parseInlineEmphasis('*', true)).
		Or(parseInlineLink).
		Or(parseInlineEmphasis('_', false)).
		Or(consumeToNextPossibleStartDelim)

	recognizeInlineTokens := func(result parser.Result, iter parser.TrackingRuneIter, state parser.State) parser.Result {
		var n uint64
		for n < result.NumConsumed {
			inlineResult := parseInlineToken(iter, state)
			if inlineResult.IsSuccess() {
				for _, tok := range inlineResult.ComputedTokens {
					tok.Offset += n
					result.ComputedTokens = append(result.ComputedTokens, tok)
				}
				iter.Skip(inlineResult.NumConsumed)
				n += inlineResult.NumConsumed
			} else {
				iter.Skip(1)
				n++
			}
		}
		return result
	}

	return parser.Func(consumeParagraphLines).
		MapWithInput(recognizeInlineTokens)
}
