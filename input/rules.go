package input

import (
	"github.com/gdamore/tcell/v2"
)

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
		ActionBuilder: CursorLeft,
	},
	{
		Name: "cursor left (h)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: 'h'},
		},
		ActionBuilder: CursorLeft,
	},
	{
		Name: "cursor right (arrow)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRight},
		},
		ActionBuilder: CursorRight,
	},
	{
		Name: "cursor right (l)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: 'l'},
		},
		ActionBuilder: CursorRight,
	},
	{
		Name: "cursor up (arrow)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyUp},
		},
		ActionBuilder: CursorUp,
	},
	{
		Name: "cursor up (k)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: 'k'},
		},
		ActionBuilder: CursorUp,
	},
	{
		Name: "cursor down (arrow)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyDown},
		},
		ActionBuilder: CursorDown,
	},
	{
		Name: "cursor down (j)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: 'j'},
		},
		ActionBuilder: CursorDown,
	},
	{
		Name: "cursor back (backspace)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyBackspace},
		},
		ActionBuilder: CursorBack,
	},
	{
		Name: "cursor back (backspace2)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyBackspace2},
		},
		ActionBuilder: CursorBack,
	},
	{
		Name: "cursor next word start (w)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: 'w'},
		},
		ActionBuilder: CursorNextWordStart,
	},
	{
		Name: "cursor prev word start (b)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: 'b'},
		},
		ActionBuilder: CursorPrevWordStart,
	},
	{
		Name: "cursor next word end (e)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: 'e'},
		},
		ActionBuilder: CursorNextWordEnd,
	},
	{
		Name: "cursor prev paragraph ({)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: '{'},
		},
		ActionBuilder: CursorPrevParagraph,
	},
	{
		Name: "cursor next paragraph (})",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: '}'},
		},
		ActionBuilder: CursorNextParagraph,
	},
	{
		Name: "cursor line start (0)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: '0'},
		},
		ActionBuilder: CursorLineStart,
	},
	{
		Name: "cursor line start non-whitespace (^)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: '^'},
		},
		ActionBuilder: CursorLineStartNonWhitespace,
	},
	{
		Name: "cursor line end ($)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: '$'},
		},
		ActionBuilder: CursorLineEnd,
	},
	{
		Name: "cursor start of line num (gg)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: 'g'},
			{Key: tcell.KeyRune, Rune: 'g'},
		},
		ActionBuilder: CursorStartOfLineNum,
	},
	{
		Name: "cursor start of last line (G)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: 'G'},
		},
		ActionBuilder: CursorStartOfLastLine,
	},
	{
		Name: "delete next char in line (x)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: 'x'},
		},
		ActionBuilder: DeleteNextCharInLine,
	},
	{
		Name: "enter insert mode (i)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: 'i'},
		},
		ActionBuilder: EnterInsertMode,
	},
	{
		Name: "enter insert mode at start of line (I)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: 'I'},
		},
		ActionBuilder: EnterInsertModeAtStartOfLine,
	},
	{
		Name: "enter insert mode at next pos (a)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: 'a'},
		},
		ActionBuilder: EnterInsertModeAtNextPos,
	},
	{
		Name: "enter insert mode at end of line (A)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: 'A'},
		},
		ActionBuilder: EnterInsertModeAtEndOfLine,
	},
	{
		Name: "begin new line below (o)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: 'o'},
		},
		ActionBuilder: BeginNewLineBelow,
	},
	{
		Name: "begin new line above (O)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: 'O'},
		},
		ActionBuilder: BeginNewLineAbove,
	},
	{
		Name: "delete line (dd)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: 'd'},
			{Key: tcell.KeyRune, Rune: 'd'},
		},
		ActionBuilder: DeleteLine,
	},
	{
		Name: "delete prev char in line (dh)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: 'd'},
			{Key: tcell.KeyRune, Rune: 'h'},
		},
		ActionBuilder: DeletePrevCharInLine,
	},
	{
		Name: "delete down (dj)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: 'd'},
			{Key: tcell.KeyRune, Rune: 'j'},
		},
		ActionBuilder: DeleteDown,
	},
	{
		Name: "delete up (dk)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: 'd'},
			{Key: tcell.KeyRune, Rune: 'k'},
		},
		ActionBuilder: DeleteUp,
	},
	{
		Name: "delete next char in line (dl)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: 'd'},
			{Key: tcell.KeyRune, Rune: 'l'},
		},
		ActionBuilder: DeleteNextCharInLine,
	},
	{
		Name: "delete to end of line (d$)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: 'd'},
			{Key: tcell.KeyRune, Rune: '$'},
		},
		ActionBuilder: DeleteToEndOfLine,
	},
	{
		Name: "delete to start of line (d0)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: 'd'},
			{Key: tcell.KeyRune, Rune: '0'},
		},
		ActionBuilder: DeleteToStartOfLine,
	},
	{
		Name: "delete to start of line non-whitespace (d^)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: 'd'},
			{Key: tcell.KeyRune, Rune: '^'},
		},
		ActionBuilder: DeleteToStartOfLineNonWhitespace,
	},
	{
		Name: "delete to end of line (D)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: 'D'},
		},
		ActionBuilder: DeleteToEndOfLine,
	},
	{
		Name: "delete inner word (diw)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: 'd'},
			{Key: tcell.KeyRune, Rune: 'i'},
			{Key: tcell.KeyRune, Rune: 'w'},
		},
		ActionBuilder: DeleteInnerWord,
	},
	{
		Name: "change inner word (ciw)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: 'c'},
			{Key: tcell.KeyRune, Rune: 'i'},
			{Key: tcell.KeyRune, Rune: 'w'},
		},
		ActionBuilder: ChangeInnerWord,
	},
	{
		Name: "replace character (r)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: 'r'},
			{Wildcard: true},
		},
		ActionBuilder: ReplaceCharacter,
	},
	{
		Name: "scroll up",
		Pattern: []EventMatcher{
			{Key: tcell.KeyCtrlU},
		},
		ActionBuilder: ScrollUp,
	},
	{
		Name: "scroll down",
		Pattern: []EventMatcher{
			{Key: tcell.KeyCtrlD},
		},
		ActionBuilder: ScrollDown,
	},
	{
		Name: "show command menu",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: ':'},
		},
		ActionBuilder: ShowCommandMenu,
	},
	{
		Name: "start forward search",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: '/'},
		},
		ActionBuilder: StartSearchForward,
	},
	{
		Name: "start backward search",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: '?'},
		},
		ActionBuilder: StartSearchBackward,
	},
	{
		Name: "find next match",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: 'n'},
		},
		ActionBuilder: FindNextMatch,
	},
	{
		Name: "find previous match",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: 'N'},
		},
		ActionBuilder: FindPrevMatch,
	},
}
