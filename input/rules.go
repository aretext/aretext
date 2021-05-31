package input

import (
	"github.com/gdamore/tcell/v2"
)

// ActionBuilder is invoked when the input parser accepts a sequence of keypresses matching a rule.
type ActionBuilder func(p ActionBuilderParams) Action

type ActionBuilderParams struct {
	InputEvents   []*tcell.EventKey
	CountArg      *int64
	MacroRecorder *MacroRecorder
	Config        Config
}

// Rule defines a command that the input parser can recognize.
// The pattern is a sequence of keypresses that trigger the rule.
type Rule struct {
	Name                  string
	Pattern               []EventMatcher
	ActionBuilder         ActionBuilder
	SkipMacroInNormalMode bool
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
		SkipMacroInNormalMode: true,
	},
	{
		Name: "cursor left (h)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: 'h'},
		},
		ActionBuilder: func(p ActionBuilderParams) Action {
			return CursorLeft
		},
		SkipMacroInNormalMode: true,
	},
	{
		Name: "cursor right (arrow)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRight},
		},
		ActionBuilder: func(p ActionBuilderParams) Action {
			return CursorRight
		},
		SkipMacroInNormalMode: true,
	},
	{
		Name: "cursor right (l)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: 'l'},
		},
		ActionBuilder: func(p ActionBuilderParams) Action {
			return CursorRight
		},
		SkipMacroInNormalMode: true,
	},
	{
		Name: "cursor up (arrow)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyUp},
		},
		ActionBuilder: func(p ActionBuilderParams) Action {
			return CursorUp
		},
		SkipMacroInNormalMode: true,
	},
	{
		Name: "cursor up (k)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: 'k'},
		},
		ActionBuilder: func(p ActionBuilderParams) Action {
			return CursorUp
		},
		SkipMacroInNormalMode: true,
	},
	{
		Name: "cursor down (arrow)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyDown},
		},
		ActionBuilder: func(p ActionBuilderParams) Action {
			return CursorDown
		},
		SkipMacroInNormalMode: true,
	},
	{
		Name: "cursor down (j)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: 'j'},
		},
		ActionBuilder: func(p ActionBuilderParams) Action {
			return CursorDown
		},
		SkipMacroInNormalMode: true,
	},
	{
		Name: "cursor back (backspace)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyBackspace},
		},
		ActionBuilder: func(p ActionBuilderParams) Action {
			return CursorBack
		},
		SkipMacroInNormalMode: true,
	},
	{
		Name: "cursor back (backspace2)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyBackspace2},
		},
		ActionBuilder: func(p ActionBuilderParams) Action {
			return CursorBack
		},
		SkipMacroInNormalMode: true,
	},
	{
		Name: "cursor next word start (w)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: 'w'},
		},
		ActionBuilder: func(p ActionBuilderParams) Action {
			return CursorNextWordStart
		},
		SkipMacroInNormalMode: true,
	},
	{
		Name: "cursor prev word start (b)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: 'b'},
		},
		ActionBuilder: func(p ActionBuilderParams) Action {
			return CursorPrevWordStart
		},
		SkipMacroInNormalMode: true,
	},
	{
		Name: "cursor next word end (e)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: 'e'},
		},
		ActionBuilder: func(p ActionBuilderParams) Action {
			return CursorNextWordEnd
		},
		SkipMacroInNormalMode: true,
	},
	{
		Name: "cursor prev paragraph ({)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: '{'},
		},
		ActionBuilder: func(p ActionBuilderParams) Action {
			return CursorPrevParagraph
		},
		SkipMacroInNormalMode: true,
	},
	{
		Name: "cursor next paragraph (})",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: '}'},
		},
		ActionBuilder: func(p ActionBuilderParams) Action {
			return CursorNextParagraph
		},
		SkipMacroInNormalMode: true,
	},
	{
		Name: "cursor line start (0)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: '0'},
		},
		ActionBuilder: func(p ActionBuilderParams) Action {
			return CursorLineStart
		},
		SkipMacroInNormalMode: true,
	},
	{
		Name: "cursor line start non-whitespace (^)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: '^'},
		},
		ActionBuilder: func(p ActionBuilderParams) Action {
			return CursorLineStartNonWhitespace
		},
		SkipMacroInNormalMode: true,
	},
	{
		Name: "cursor line end ($)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: '$'},
		},
		ActionBuilder: func(p ActionBuilderParams) Action {
			return CursorLineEnd
		},
		SkipMacroInNormalMode: true,
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
		SkipMacroInNormalMode: true,
	},
	{
		Name: "cursor start of last line (G)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: 'G'},
		},
		ActionBuilder: func(p ActionBuilderParams) Action {
			return CursorStartOfLastLine
		},
		SkipMacroInNormalMode: true,
	},
	{
		Name: "scroll up (ctrl-u)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyCtrlU},
		},
		ActionBuilder: func(p ActionBuilderParams) Action {
			return ScrollUp(p.Config)
		},
		SkipMacroInNormalMode: true,
	},
	{
		Name: "scroll down (ctrl-d)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyCtrlD},
		},
		ActionBuilder: func(p ActionBuilderParams) Action {
			return ScrollDown(p.Config)
		},
		SkipMacroInNormalMode: true,
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
		Name: "delete to start of next word (dw)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: 'd'},
			{Key: tcell.KeyRune, Rune: 'w'},
		},
		ActionBuilder: func(p ActionBuilderParams) Action {
			return DeleteToStartOfNextWord
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
			return DeleteAWord
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
		Name: "change to start of next word (cw)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: 'c'},
			{Key: tcell.KeyRune, Rune: 'w'},
		},
		ActionBuilder: func(p ActionBuilderParams) Action {
			return ChangeToStartOfNextWord
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
			return ChangeAWord
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
			return CopyToStartOfNextWord
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
			return CopyAWord
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
			return CopyInnerWord
		},
	},
	{
		Name: "yank line (yy)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: 'y'},
			{Key: tcell.KeyRune, Rune: 'y'},
		},
		ActionBuilder: func(p ActionBuilderParams) Action {
			return CopyLines
		},
	},
	{
		Name: "put after cursor (p)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: 'p'},
		},
		ActionBuilder: func(p ActionBuilderParams) Action {
			return PasteAfterCursor
		},
	},
	{
		Name: "put before cursor (P)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: 'P'},
		},
		ActionBuilder: func(p ActionBuilderParams) Action {
			return PasteBeforeCursor
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
		SkipMacroInNormalMode: true,
	},
	{
		Name: "start forward search",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: '/'},
		},
		ActionBuilder: func(p ActionBuilderParams) Action {
			return StartSearchForward
		},
		SkipMacroInNormalMode: true,
	},
	{
		Name: "start backward search",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: '?'},
		},
		ActionBuilder: func(p ActionBuilderParams) Action {
			return StartSearchBackward
		},
		SkipMacroInNormalMode: true,
	},
	{
		Name: "find next match",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: 'n'},
		},
		ActionBuilder: func(p ActionBuilderParams) Action {
			return FindNextMatch
		},
		SkipMacroInNormalMode: true,
	},
	{
		Name: "find previous match",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: 'N'},
		},
		ActionBuilder: func(p ActionBuilderParams) Action {
			return FindPrevMatch
		},
		SkipMacroInNormalMode: true,
	},
	{
		Name: "undo (u)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: 'u'},
		},
		ActionBuilder: func(p ActionBuilderParams) Action {
			return Undo
		},
		SkipMacroInNormalMode: true,
	},
	{
		Name: "redo (ctrl-r)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyCtrlR},
		},
		ActionBuilder: func(p ActionBuilderParams) Action {
			return Redo
		},
		SkipMacroInNormalMode: true,
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
	},
	{
		Name: "toggle visual mode linewise (V)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: 'V'},
		},
		ActionBuilder: func(p ActionBuilderParams) Action {
			return ToggleVisualModeLinewise
		},
	},
	{
		Name: "return to normal mode (esc)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyEscape},
		},
		ActionBuilder: func(p ActionBuilderParams) Action {
			return ReturnToNormalMode
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
		Name: "delete selection (x)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: 'x'},
		},
		ActionBuilder: func(p ActionBuilderParams) Action {
			return DeleteSelectionAndReturnToNormalMode
		},
	},
	{
		Name: "delete selection (d)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: 'd'},
		},
		ActionBuilder: func(p ActionBuilderParams) Action {
			return DeleteSelectionAndReturnToNormalMode
		},
	},
	{
		Name: "change selection (c)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: 'c'},
		},
		ActionBuilder: func(p ActionBuilderParams) Action {
			return ChangeSelection
		},
	},
	{
		Name: "toggle case for selection (~)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: '~'},
		},
		ActionBuilder: func(p ActionBuilderParams) Action {
			return ToggleCaseInSelectionAndReturnToNormalMode
		},
	},
	{
		Name: "indent selection (>)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: '>'},
		},
		ActionBuilder: func(p ActionBuilderParams) Action {
			return IndentSelectionAndReturnToNormalMode
		},
	},
	{
		Name: "outdent selection (<)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: '<'},
		},
		ActionBuilder: func(p ActionBuilderParams) Action {
			return OutdentSelectionAndReturnToNormalMode
		},
	},
	{
		Name: "yank selection (y)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: 'y'},
		},
		ActionBuilder: func(p ActionBuilderParams) Action {
			return CopySelectionAndReturnToNormalMode
		},
	},
}...)
