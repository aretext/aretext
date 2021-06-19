package input

import (
	"github.com/gdamore/tcell/v2"
)

// ActionBuilder is invoked when the input parser accepts a sequence of keypresses matching a rule.
type ActionBuilder func(p ActionBuilderParams) Action

type ActionBuilderParams struct {
	InputEvents          []*tcell.EventKey
	CountArg             *uint64
	ClipboardPageNameArg *rune
	MacroRecorder        *MacroRecorder
	Config               Config
}

// Rule defines a command that the input parser can recognize.
// The pattern is a sequence of keypresses that trigger the rule.
type Rule struct {
	Name          string
	Pattern       []EventMatcher
	ActionBuilder ActionBuilder
	SkipMacro     bool
}

// These rules control cursor movement in normal and visual mode.
var cursorRules = []Rule{
	{
		Name: "cursor left (arrow)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyLeft},
		},
		ActionBuilder: func(p ActionBuilderParams) Action {
			return CursorLeft
		},
		SkipMacro: true,
	},
	{
		Name: "cursor left (h)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: 'h'},
		},
		ActionBuilder: func(p ActionBuilderParams) Action {
			return CursorLeft
		},
		SkipMacro: true,
	},
	{
		Name: "cursor right (arrow)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRight},
		},
		ActionBuilder: func(p ActionBuilderParams) Action {
			return CursorRight
		},
		SkipMacro: true,
	},
	{
		Name: "cursor right (l)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: 'l'},
		},
		ActionBuilder: func(p ActionBuilderParams) Action {
			return CursorRight
		},
		SkipMacro: true,
	},
	{
		Name: "cursor up (arrow)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyUp},
		},
		ActionBuilder: func(p ActionBuilderParams) Action {
			return CursorUp
		},
		SkipMacro: true,
	},
	{
		Name: "cursor up (k)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: 'k'},
		},
		ActionBuilder: func(p ActionBuilderParams) Action {
			return CursorUp
		},
		SkipMacro: true,
	},
	{
		Name: "cursor down (arrow)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyDown},
		},
		ActionBuilder: func(p ActionBuilderParams) Action {
			return CursorDown
		},
		SkipMacro: true,
	},
	{
		Name: "cursor down (j)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: 'j'},
		},
		ActionBuilder: func(p ActionBuilderParams) Action {
			return CursorDown
		},
		SkipMacro: true,
	},
	{
		Name: "cursor back (backspace)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyBackspace},
		},
		ActionBuilder: func(p ActionBuilderParams) Action {
			return CursorBack
		},
		SkipMacro: true,
	},
	{
		Name: "cursor back (backspace2)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyBackspace2},
		},
		ActionBuilder: func(p ActionBuilderParams) Action {
			return CursorBack
		},
		SkipMacro: true,
	},
	{
		Name: "cursor next word start (w)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: 'w'},
		},
		ActionBuilder: func(p ActionBuilderParams) Action {
			return CursorNextWordStart
		},
		SkipMacro: true,
	},
	{
		Name: "cursor prev word start (b)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: 'b'},
		},
		ActionBuilder: func(p ActionBuilderParams) Action {
			return CursorPrevWordStart
		},
		SkipMacro: true,
	},
	{
		Name: "cursor next word end (e)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: 'e'},
		},
		ActionBuilder: func(p ActionBuilderParams) Action {
			return CursorNextWordEnd
		},
		SkipMacro: true,
	},
	{
		Name: "cursor prev paragraph ({)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: '{'},
		},
		ActionBuilder: func(p ActionBuilderParams) Action {
			return CursorPrevParagraph
		},
		SkipMacro: true,
	},
	{
		Name: "cursor next paragraph (})",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: '}'},
		},
		ActionBuilder: func(p ActionBuilderParams) Action {
			return CursorNextParagraph
		},
		SkipMacro: true,
	},
	{
		Name: "cursor to next matching char (f{char})",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: 'f'},
			{Wildcard: true},
		},
		ActionBuilder: func(p ActionBuilderParams) Action {
			return CursorToNextMatchingChar(p.InputEvents, p.CountArg, true)
		},
		SkipMacro: true,
	},
	{
		Name: "cursor to prev matching char (F{char})",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: 'F'},
			{Wildcard: true},
		},
		ActionBuilder: func(p ActionBuilderParams) Action {
			return CursorToPrevMatchingChar(p.InputEvents, p.CountArg, true)
		},
		SkipMacro: true,
	},
	{
		Name: "cursor till next matching char (t{char})",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: 't'},
			{Wildcard: true},
		},
		ActionBuilder: func(p ActionBuilderParams) Action {
			return CursorToNextMatchingChar(p.InputEvents, p.CountArg, false)
		},
		SkipMacro: true,
	},
	{
		Name: "cursor to prev matching char (T{char})",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: 'T'},
			{Wildcard: true},
		},
		ActionBuilder: func(p ActionBuilderParams) Action {
			return CursorToPrevMatchingChar(p.InputEvents, p.CountArg, false)
		},
		SkipMacro: true,
	},
	{
		Name: "cursor line start (0)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: '0'},
		},
		ActionBuilder: func(p ActionBuilderParams) Action {
			return CursorLineStart
		},
		SkipMacro: true,
	},
	{
		Name: "cursor line start non-whitespace (^)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: '^'},
		},
		ActionBuilder: func(p ActionBuilderParams) Action {
			return CursorLineStartNonWhitespace
		},
		SkipMacro: true,
	},
	{
		Name: "cursor line end ($)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: '$'},
		},
		ActionBuilder: func(p ActionBuilderParams) Action {
			return CursorLineEnd
		},
		SkipMacro: true,
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
		SkipMacro: true,
	},
	{
		Name: "cursor start of last line (G)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: 'G'},
		},
		ActionBuilder: func(p ActionBuilderParams) Action {
			return CursorStartOfLastLine
		},
		SkipMacro: true,
	},
	{
		Name: "scroll up (ctrl-u)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyCtrlU},
		},
		ActionBuilder: func(p ActionBuilderParams) Action {
			return ScrollUp(p.Config)
		},
		SkipMacro: true,
	},
	{
		Name: "scroll down (ctrl-d)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyCtrlD},
		},
		ActionBuilder: func(p ActionBuilderParams) Action {
			return ScrollDown(p.Config)
		},
		SkipMacro: true,
	},
}

// These rules are used when the editor is in normal mode.
var normalModeRules = append(cursorRules, []Rule{
	{
		Name: "delete next char in line (x)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: 'x'},
		},
		ActionBuilder: func(p ActionBuilderParams) Action {
			return DeleteNextCharInLine(p.CountArg, p.ClipboardPageNameArg)
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
		Name: "join lines (J)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: 'J'},
		},
		ActionBuilder: func(p ActionBuilderParams) Action {
			return JoinLines
		},
	},
	{
		Name: "delete line (dd)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: 'd'},
			{Key: tcell.KeyRune, Rune: 'd'},
		},
		ActionBuilder: func(p ActionBuilderParams) Action {
			return DeleteLines(p.CountArg, p.ClipboardPageNameArg)
		},
	},
	{
		Name: "delete prev char in line (dh)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: 'd'},
			{Key: tcell.KeyRune, Rune: 'h'},
		},
		ActionBuilder: func(p ActionBuilderParams) Action {
			return DeletePrevCharInLine(p.ClipboardPageNameArg)
		},
	},
	{
		Name: "delete down (dj)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: 'd'},
			{Key: tcell.KeyRune, Rune: 'j'},
		},
		ActionBuilder: func(p ActionBuilderParams) Action {
			return DeleteDown(p.ClipboardPageNameArg)
		},
	},
	{
		Name: "delete up (dk)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: 'd'},
			{Key: tcell.KeyRune, Rune: 'k'},
		},
		ActionBuilder: func(p ActionBuilderParams) Action {
			return DeleteUp(p.ClipboardPageNameArg)
		},
	},
	{
		Name: "delete next char in line (dl)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: 'd'},
			{Key: tcell.KeyRune, Rune: 'l'},
		},
		ActionBuilder: func(p ActionBuilderParams) Action {
			return DeleteNextCharInLine(p.CountArg, p.ClipboardPageNameArg)
		},
	},
	{
		Name: "delete to end of line (d$)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: 'd'},
			{Key: tcell.KeyRune, Rune: '$'},
		},
		ActionBuilder: func(p ActionBuilderParams) Action {
			return DeleteToEndOfLine(p.ClipboardPageNameArg)
		},
	},
	{
		Name: "delete to start of line (d0)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: 'd'},
			{Key: tcell.KeyRune, Rune: '0'},
		},
		ActionBuilder: func(p ActionBuilderParams) Action {
			return DeleteToStartOfLine(p.ClipboardPageNameArg)
		},
	},
	{
		Name: "delete to start of line non-whitespace (d^)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: 'd'},
			{Key: tcell.KeyRune, Rune: '^'},
		},
		ActionBuilder: func(p ActionBuilderParams) Action {
			return DeleteToStartOfLineNonWhitespace(p.ClipboardPageNameArg)
		},
	},
	{
		Name: "delete to end of line (D)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: 'D'},
		},
		ActionBuilder: func(p ActionBuilderParams) Action {
			return DeleteToEndOfLine(p.ClipboardPageNameArg)
		},
	},
	{
		Name: "delete to next matching char (df{char}",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: 'd'},
			{Key: tcell.KeyRune, Rune: 'f'},
			{Wildcard: true},
		},
		ActionBuilder: func(p ActionBuilderParams) Action {
			return DeleteToNextMatchingChar(p.InputEvents, p.CountArg, p.ClipboardPageNameArg, true)
		},
	},
	{
		Name: "delete to prev matching char (dF{char}",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: 'd'},
			{Key: tcell.KeyRune, Rune: 'F'},
			{Wildcard: true},
		},
		ActionBuilder: func(p ActionBuilderParams) Action {
			return DeleteToPrevMatchingChar(p.InputEvents, p.CountArg, p.ClipboardPageNameArg, true)
		},
	},
	{
		Name: "delete till next matching char (dt{char}",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: 'd'},
			{Key: tcell.KeyRune, Rune: 't'},
			{Wildcard: true},
		},
		ActionBuilder: func(p ActionBuilderParams) Action {
			return DeleteToNextMatchingChar(p.InputEvents, p.CountArg, p.ClipboardPageNameArg, false)
		},
	},
	{
		Name: "delete till prev matching char (dT{char}",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: 'd'},
			{Key: tcell.KeyRune, Rune: 'T'},
			{Wildcard: true},
		},
		ActionBuilder: func(p ActionBuilderParams) Action {
			return DeleteToPrevMatchingChar(p.InputEvents, p.CountArg, p.ClipboardPageNameArg, false)
		},
	},
	{
		Name: "delete to start of next word (dw)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: 'd'},
			{Key: tcell.KeyRune, Rune: 'w'},
		},
		ActionBuilder: func(p ActionBuilderParams) Action {
			return DeleteToStartOfNextWord(p.ClipboardPageNameArg)
		},
	},
	{
		Name: "delete a word (daw)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: 'd'},
			{Key: tcell.KeyRune, Rune: 'a'},
			{Key: tcell.KeyRune, Rune: 'w'},
		},
		ActionBuilder: func(p ActionBuilderParams) Action {
			return DeleteAWord(p.ClipboardPageNameArg)
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
			return DeleteInnerWord(p.ClipboardPageNameArg)
		},
	},
	{
		Name: "change to start of next word (cw)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: 'c'},
			{Key: tcell.KeyRune, Rune: 'w'},
		},
		ActionBuilder: func(p ActionBuilderParams) Action {
			return ChangeToStartOfNextWord(p.ClipboardPageNameArg)
		},
	},
	{
		Name: "change a word (caw)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: 'c'},
			{Key: tcell.KeyRune, Rune: 'a'},
			{Key: tcell.KeyRune, Rune: 'w'},
		},
		ActionBuilder: func(p ActionBuilderParams) Action {
			return ChangeAWord(p.ClipboardPageNameArg)
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
			return ChangeInnerWord(p.ClipboardPageNameArg)
		},
	},
	{
		Name: "change to next matching char (cf{char})",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: 'c'},
			{Key: tcell.KeyRune, Rune: 'f'},
			{Wildcard: true},
		},
		ActionBuilder: func(p ActionBuilderParams) Action {
			return ChangeToNextMatchingChar(p.InputEvents, p.CountArg, p.ClipboardPageNameArg, true)
		},
	},
	{
		Name: "change to prev matching char (cF{char})",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: 'c'},
			{Key: tcell.KeyRune, Rune: 'F'},
			{Wildcard: true},
		},
		ActionBuilder: func(p ActionBuilderParams) Action {
			return ChangeToPrevMatchingChar(p.InputEvents, p.CountArg, p.ClipboardPageNameArg, true)
		},
	},
	{
		Name: "change till next matching char (ct{char})",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: 'c'},
			{Key: tcell.KeyRune, Rune: 't'},
			{Wildcard: true},
		},
		ActionBuilder: func(p ActionBuilderParams) Action {
			return ChangeToNextMatchingChar(p.InputEvents, p.CountArg, p.ClipboardPageNameArg, false)
		},
	},
	{
		Name: "change till prev matching char (cT{char})",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: 'c'},
			{Key: tcell.KeyRune, Rune: 'T'},
			{Wildcard: true},
		},
		ActionBuilder: func(p ActionBuilderParams) Action {
			return ChangeToPrevMatchingChar(p.InputEvents, p.CountArg, p.ClipboardPageNameArg, false)
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
		Name: "toggle case (~)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: '~'},
		},
		ActionBuilder: func(p ActionBuilderParams) Action {
			return ToggleCaseAtCursor
		},
	},
	{
		Name: "indent (>>)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: '>'},
			{Key: tcell.KeyRune, Rune: '>'},
		},
		ActionBuilder: func(p ActionBuilderParams) Action {
			return IndentLine
		},
	},
	{
		Name: "outdent (<<)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: '<'},
			{Key: tcell.KeyRune, Rune: '<'},
		},
		ActionBuilder: func(p ActionBuilderParams) Action {
			return OutdentLine
		},
	},
	{
		Name: "yank to start of next word (yw)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: 'y'},
			{Key: tcell.KeyRune, Rune: 'w'},
		},
		ActionBuilder: func(p ActionBuilderParams) Action {
			return CopyToStartOfNextWord(p.ClipboardPageNameArg)
		},
	},
	{
		Name: "yank a word (yaw)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: 'y'},
			{Key: tcell.KeyRune, Rune: 'a'},
			{Key: tcell.KeyRune, Rune: 'w'},
		},
		ActionBuilder: func(p ActionBuilderParams) Action {
			return CopyAWord(p.ClipboardPageNameArg)
		},
	},
	{
		Name: "yank inner word (yiw)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: 'y'},
			{Key: tcell.KeyRune, Rune: 'i'},
			{Key: tcell.KeyRune, Rune: 'w'},
		},
		ActionBuilder: func(p ActionBuilderParams) Action {
			return CopyInnerWord(p.ClipboardPageNameArg)
		},
	},
	{
		Name: "yank line (yy)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: 'y'},
			{Key: tcell.KeyRune, Rune: 'y'},
		},
		ActionBuilder: func(p ActionBuilderParams) Action {
			return CopyLines(p.ClipboardPageNameArg)
		},
	},
	{
		Name: "put after cursor (p)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: 'p'},
		},
		ActionBuilder: func(p ActionBuilderParams) Action {
			return PasteAfterCursor(p.ClipboardPageNameArg)
		},
	},
	{
		Name: "put before cursor (P)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: 'P'},
		},
		ActionBuilder: func(p ActionBuilderParams) Action {
			return PasteBeforeCursor(p.ClipboardPageNameArg)
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
		SkipMacro: true,
	},
	{
		Name: "start forward search",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: '/'},
		},
		ActionBuilder: func(p ActionBuilderParams) Action {
			return StartSearchForward
		},
		SkipMacro: true,
	},
	{
		Name: "start backward search",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: '?'},
		},
		ActionBuilder: func(p ActionBuilderParams) Action {
			return StartSearchBackward
		},
		SkipMacro: true,
	},
	{
		Name: "find next match",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: 'n'},
		},
		ActionBuilder: func(p ActionBuilderParams) Action {
			return FindNextMatch
		},
		SkipMacro: true,
	},
	{
		Name: "find previous match",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: 'N'},
		},
		ActionBuilder: func(p ActionBuilderParams) Action {
			return FindPrevMatch
		},
		SkipMacro: true,
	},
	{
		Name: "undo (u)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: 'u'},
		},
		ActionBuilder: func(p ActionBuilderParams) Action {
			return Undo
		},
		SkipMacro: true,
	},
	{
		Name: "redo (ctrl-r)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyCtrlR},
		},
		ActionBuilder: func(p ActionBuilderParams) Action {
			return Redo
		},
		SkipMacro: true,
	},
	{
		Name: "enter visual mode charwise (v)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: 'v'},
		},
		ActionBuilder: func(p ActionBuilderParams) Action {
			return ToggleVisualModeCharwise
		},
	},
	{
		Name: "enter visual mode linewise (V)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: 'V'},
		},
		ActionBuilder: func(p ActionBuilderParams) Action {
			return ToggleVisualModeLinewise
		},
	},
	{
		Name: "repeat last action (.)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: '.'},
		},
		ActionBuilder: func(p ActionBuilderParams) Action {
			return p.MacroRecorder.LastAction()
		},
	},
}...)

// These rules are used when the editor is in visual mode.
var visualModeRules = append(cursorRules, []Rule{
	{
		Name: "toggle visual mode charwise (v)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: 'v'},
		},
		ActionBuilder: func(p ActionBuilderParams) Action {
			return ToggleVisualModeCharwise
		},
		SkipMacro: true,
	},
	{
		Name: "toggle visual mode linewise (V)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: 'V'},
		},
		ActionBuilder: func(p ActionBuilderParams) Action {
			return ToggleVisualModeLinewise
		},
		SkipMacro: true,
	},
	{
		Name: "return to normal mode (esc)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyEscape},
		},
		ActionBuilder: func(p ActionBuilderParams) Action {
			return ReturnToNormalMode
		},
		SkipMacro: true,
	},
	{
		Name: "show command menu",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: ':'},
		},
		ActionBuilder: func(p ActionBuilderParams) Action {
			return ShowCommandMenu(p.Config)
		},
		SkipMacro: true,
	},
	{
		Name: "delete selection (x)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: 'x'},
		},
		ActionBuilder: func(p ActionBuilderParams) Action {
			return DeleteSelectionAndReturnToNormalMode(
				p.ClipboardPageNameArg,
				p.Config.SelectionMode,
				p.Config.SelectionEndLocator,
			)
		},
	},
	{
		Name: "delete selection (d)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: 'd'},
		},
		ActionBuilder: func(p ActionBuilderParams) Action {
			return DeleteSelectionAndReturnToNormalMode(
				p.ClipboardPageNameArg,
				p.Config.SelectionMode,
				p.Config.SelectionEndLocator,
			)
		},
	},
	{
		Name: "change selection (c)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: 'c'},
		},
		ActionBuilder: func(p ActionBuilderParams) Action {
			return ChangeSelection(
				p.ClipboardPageNameArg,
				p.Config.SelectionMode,
				p.Config.SelectionEndLocator,
			)
		},
	},
	{
		Name: "toggle case for selection (~)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: '~'},
		},
		ActionBuilder: func(p ActionBuilderParams) Action {
			return ToggleCaseInSelectionAndReturnToNormalMode(p.Config.SelectionEndLocator)
		},
	},
	{
		Name: "indent selection (>)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: '>'},
		},
		ActionBuilder: func(p ActionBuilderParams) Action {
			return IndentSelectionAndReturnToNormalMode(p.Config.SelectionEndLocator)
		},
	},
	{
		Name: "outdent selection (<)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: '<'},
		},
		ActionBuilder: func(p ActionBuilderParams) Action {
			return OutdentSelectionAndReturnToNormalMode(p.Config.SelectionEndLocator)
		},
	},
	{
		Name: "yank selection (y)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: 'y'},
		},
		ActionBuilder: func(p ActionBuilderParams) Action {
			return CopySelectionAndReturnToNormalMode(p.ClipboardPageNameArg)
		},
		SkipMacro: true,
	},
}...)
