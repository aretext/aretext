package input

import (
	"github.com/gdamore/tcell/v2"
)

// ActionBuilder is invoked when the input parser accepts a sequence of keypresses matching a rule.
type ActionBuilder func(p ActionBuilderParams) Action

type ActionBuilderParams struct {
	InputEvents []*tcell.EventKey
	CountArg    *int64
	Config      Config
}

// Rule defines a command that the input parser can recognize.
// The pattern is a sequence of keypresses that trigger the rule.
type Rule struct {
	Name          string
	Pattern       []EventMatcher
	ActionBuilder ActionBuilder
}

// These rules are used when the editor is in normal mode.
var normalModeRules = []Rule{
	{
		Name: "cursor left (arrow)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyLeft},
		},
		ActionBuilder: func(p ActionBuilderParams) Action {
			return CursorLeft
		},
	},
	{
		Name: "cursor left (h)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: 'h'},
		},
		ActionBuilder: func(p ActionBuilderParams) Action {
			return CursorLeft
		},
	},
	{
		Name: "cursor right (arrow)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRight},
		},
		ActionBuilder: func(p ActionBuilderParams) Action {
			return CursorRight
		},
	},
	{
		Name: "cursor right (l)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: 'l'},
		},
		ActionBuilder: func(p ActionBuilderParams) Action {
			return CursorRight
		},
	},
	{
		Name: "cursor up (arrow)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyUp},
		},
		ActionBuilder: func(p ActionBuilderParams) Action {
			return CursorUp
		},
	},
	{
		Name: "cursor up (k)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: 'k'},
		},
		ActionBuilder: func(p ActionBuilderParams) Action {
			return CursorUp
		},
	},
	{
		Name: "cursor down (arrow)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyDown},
		},
		ActionBuilder: func(p ActionBuilderParams) Action {
			return CursorDown
		},
	},
	{
		Name: "cursor down (j)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: 'j'},
		},
		ActionBuilder: func(p ActionBuilderParams) Action {
			return CursorDown
		},
	},
	{
		Name: "cursor back (backspace)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyBackspace},
		},
		ActionBuilder: func(p ActionBuilderParams) Action {
			return CursorBack
		},
	},
	{
		Name: "cursor back (backspace2)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyBackspace2},
		},
		ActionBuilder: func(p ActionBuilderParams) Action {
			return CursorBack
		},
	},
	{
		Name: "cursor next word start (w)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: 'w'},
		},
		ActionBuilder: func(p ActionBuilderParams) Action {
			return CursorNextWordStart
		},
	},
	{
		Name: "cursor prev word start (b)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: 'b'},
		},
		ActionBuilder: func(p ActionBuilderParams) Action {
			return CursorPrevWordStart
		},
	},
	{
		Name: "cursor next word end (e)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: 'e'},
		},
		ActionBuilder: func(p ActionBuilderParams) Action {
			return CursorNextWordEnd
		},
	},
	{
		Name: "cursor prev paragraph ({)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: '{'},
		},
		ActionBuilder: func(p ActionBuilderParams) Action {
			return CursorPrevParagraph
		},
	},
	{
		Name: "cursor next paragraph (})",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: '}'},
		},
		ActionBuilder: func(p ActionBuilderParams) Action {
			return CursorNextParagraph
		},
	},
	{
		Name: "cursor line start (0)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: '0'},
		},
		ActionBuilder: func(p ActionBuilderParams) Action {
			return CursorLineStart
		},
	},
	{
		Name: "cursor line start non-whitespace (^)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: '^'},
		},
		ActionBuilder: func(p ActionBuilderParams) Action {
			return CursorLineStartNonWhitespace
		},
	},
	{
		Name: "cursor line end ($)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: '$'},
		},
		ActionBuilder: func(p ActionBuilderParams) Action {
			return CursorLineEnd
		},
	},
	{
		Name: "cursor start of line num (gg)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: 'g'},
			{Key: tcell.KeyRune, Rune: 'g'},
		},
		ActionBuilder: func(p ActionBuilderParams) Action {
			return CursorStartOfLineNum(p.CountArg)
		},
	},
	{
		Name: "cursor start of last line (G)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: 'G'},
		},
		ActionBuilder: func(p ActionBuilderParams) Action {
			return CursorStartOfLastLine
		},
	},
	{
		Name: "delete next char in line (x)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: 'x'},
		},
		ActionBuilder: func(p ActionBuilderParams) Action {
			return DeleteNextCharInLine
		},
	},
	{
		Name: "enter insert mode (i)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: 'i'},
		},
		ActionBuilder: func(p ActionBuilderParams) Action {
			return EnterInsertMode
		},
	},
	{
		Name: "enter insert mode at start of line (I)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: 'I'},
		},
		ActionBuilder: func(p ActionBuilderParams) Action {
			return EnterInsertModeAtStartOfLine
		},
	},
	{
		Name: "enter insert mode at next pos (a)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: 'a'},
		},
		ActionBuilder: func(p ActionBuilderParams) Action {
			return EnterInsertModeAtNextPos
		},
	},
	{
		Name: "enter insert mode at end of line (A)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: 'A'},
		},
		ActionBuilder: func(p ActionBuilderParams) Action {
			return EnterInsertModeAtEndOfLine
		},
	},
	{
		Name: "begin new line below (o)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: 'o'},
		},
		ActionBuilder: func(p ActionBuilderParams) Action {
			return BeginNewLineBelow
		},
	},
	{
		Name: "begin new line above (O)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: 'O'},
		},
		ActionBuilder: func(p ActionBuilderParams) Action {
			return BeginNewLineAbove
		},
	},
	{
		Name: "delete line (dd)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: 'd'},
			{Key: tcell.KeyRune, Rune: 'd'},
		},
		ActionBuilder: func(p ActionBuilderParams) Action {
			return DeleteLine
		},
	},
	{
		Name: "delete prev char in line (dh)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: 'd'},
			{Key: tcell.KeyRune, Rune: 'h'},
		},
		ActionBuilder: func(p ActionBuilderParams) Action {
			return DeletePrevCharInLine
		},
	},
	{
		Name: "delete down (dj)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: 'd'},
			{Key: tcell.KeyRune, Rune: 'j'},
		},
		ActionBuilder: func(p ActionBuilderParams) Action {
			return DeleteDown
		},
	},
	{
		Name: "delete up (dk)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: 'd'},
			{Key: tcell.KeyRune, Rune: 'k'},
		},
		ActionBuilder: func(p ActionBuilderParams) Action {
			return DeleteUp
		},
	},
	{
		Name: "delete next char in line (dl)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: 'd'},
			{Key: tcell.KeyRune, Rune: 'l'},
		},
		ActionBuilder: func(p ActionBuilderParams) Action {
			return DeleteNextCharInLine
		},
	},
	{
		Name: "delete to end of line (d$)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: 'd'},
			{Key: tcell.KeyRune, Rune: '$'},
		},
		ActionBuilder: func(p ActionBuilderParams) Action {
			return DeleteToEndOfLine
		},
	},
	{
		Name: "delete to start of line (d0)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: 'd'},
			{Key: tcell.KeyRune, Rune: '0'},
		},
		ActionBuilder: func(p ActionBuilderParams) Action {
			return DeleteToStartOfLine
		},
	},
	{
		Name: "delete to start of line non-whitespace (d^)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: 'd'},
			{Key: tcell.KeyRune, Rune: '^'},
		},
		ActionBuilder: func(p ActionBuilderParams) Action {
			return DeleteToStartOfLineNonWhitespace
		},
	},
	{
		Name: "delete to end of line (D)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: 'D'},
		},
		ActionBuilder: func(p ActionBuilderParams) Action {
			return DeleteToEndOfLine
		},
	},
	{
		Name: "delete inner word (diw)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: 'd'},
			{Key: tcell.KeyRune, Rune: 'i'},
			{Key: tcell.KeyRune, Rune: 'w'},
		},
		ActionBuilder: func(p ActionBuilderParams) Action {
			return DeleteInnerWord
		},
	},
	{
		Name: "change inner word (ciw)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: 'c'},
			{Key: tcell.KeyRune, Rune: 'i'},
			{Key: tcell.KeyRune, Rune: 'w'},
		},
		ActionBuilder: func(p ActionBuilderParams) Action {
			return ChangeInnerWord
		},
	},
	{
		Name: "replace character (r)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: 'r'},
			{Wildcard: true},
		},
		ActionBuilder: func(p ActionBuilderParams) Action {
			return ReplaceCharacter(p.InputEvents)
		},
	},
	{
		Name: "scroll up",
		Pattern: []EventMatcher{
			{Key: tcell.KeyCtrlU},
		},
		ActionBuilder: func(p ActionBuilderParams) Action {
			return ScrollUp(p.Config)
		},
	},
	{
		Name: "scroll down",
		Pattern: []EventMatcher{
			{Key: tcell.KeyCtrlD},
		},
		ActionBuilder: func(p ActionBuilderParams) Action {
			return ScrollDown(p.Config)
		},
	},
	{
		Name: "show command menu",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: ':'},
		},
		ActionBuilder: func(p ActionBuilderParams) Action {
			return ShowCommandMenu(p.Config)
		},
	},
	{
		Name: "start forward search",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: '/'},
		},
		ActionBuilder: func(p ActionBuilderParams) Action {
			return StartSearchForward
		},
	},
	{
		Name: "start backward search",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: '?'},
		},
		ActionBuilder: func(p ActionBuilderParams) Action {
			return StartSearchBackward
		},
	},
	{
		Name: "find next match",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: 'n'},
		},
		ActionBuilder: func(p ActionBuilderParams) Action {
			return FindNextMatch
		},
	},
	{
		Name: "find previous match",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: 'N'},
		},
		ActionBuilder: func(p ActionBuilderParams) Action {
			return FindPrevMatch
		},
	},
}
