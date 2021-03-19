package input

import (
	"github.com/aretext/aretext/exec"
	"github.com/gdamore/tcell/v2"
)

// ActionFunc is invoked when the input parser accepts a sequence of keypresses matching a rule.
type ActionFunc func(inputEvents []*tcell.EventKey, count *int64, config Config) exec.Mutator

// Rule defines a command that the input parser can recognize.
// The pattern is a sequence of keypresses that trigger the rule.
type Rule struct {
	Name    string
	Pattern []EventMatcher
	Action  ActionFunc
}

// These rules are used when the editor is in normal mode.
var normalModeRules = []Rule{
	{
		Name: "cursor left (arrow)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyLeft},
		},
		Action: CursorLeft,
	},
	{
		Name: "cursor left (h)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: 'h'},
		},
		Action: CursorLeft,
	},
	{
		Name: "cursor right (arrow)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRight},
		},
		Action: CursorRight,
	},
	{
		Name: "cursor right (l)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: 'l'},
		},
		Action: CursorRight,
	},
	{
		Name: "cursor up (arrow)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyUp},
		},
		Action: CursorUp,
	},
	{
		Name: "cursor up (k)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: 'k'},
		},
		Action: CursorUp,
	},
	{
		Name: "cursor down (arrow)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyDown},
		},
		Action: CursorDown,
	},
	{
		Name: "cursor down (j)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: 'j'},
		},
		Action: CursorDown,
	},
	{
		Name: "cursor back (backspace)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyBackspace},
		},
		Action: CursorBack,
	},
	{
		Name: "cursor back (backspace2)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyBackspace2},
		},
		Action: CursorBack,
	},
	{
		Name: "cursor next word start (w)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: 'w'},
		},
		Action: CursorNextWordStart,
	},
	{
		Name: "cursor prev word start (b)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: 'b'},
		},
		Action: CursorPrevWordStart,
	},
	{
		Name: "cursor next word end (e)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: 'e'},
		},
		Action: CursorNextWordEnd,
	},
	{
		Name: "cursor prev paragraph ({)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: '{'},
		},
		Action: CursorPrevParagraph,
	},
	{
		Name: "cursor next paragraph (})",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: '}'},
		},
		Action: CursorNextParagraph,
	},
	{
		Name: "cursor line start (0)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: '0'},
		},
		Action: CursorLineStart,
	},
	{
		Name: "cursor line start non-whitespace (^)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: '^'},
		},
		Action: CursorLineStartNonWhitespace,
	},
	{
		Name: "cursor line end ($)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: '$'},
		},
		Action: CursorLineEnd,
	},
	{
		Name: "cursor start of line num (gg)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: 'g'},
			{Key: tcell.KeyRune, Rune: 'g'},
		},
		Action: CursorStartOfLineNum,
	},
	{
		Name: "cursor start of last line (G)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: 'G'},
		},
		Action: CursorStartOfLastLine,
	},
	{
		Name: "delete next char in line (x)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: 'x'},
		},
		Action: DeleteNextCharInLine,
	},
	{
		Name: "enter insert mode (i)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: 'i'},
		},
		Action: EnterInsertMode,
	},
	{
		Name: "enter insert mode at start of line (I)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: 'I'},
		},
		Action: EnterInsertModeAtStartOfLine,
	},
	{
		Name: "enter insert mode at next pos (a)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: 'a'},
		},
		Action: EnterInsertModeAtNextPos,
	},
	{
		Name: "enter insert mode at end of line (A)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: 'A'},
		},
		Action: EnterInsertModeAtEndOfLine,
	},
	{
		Name: "begin new line below (o)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: 'o'},
		},
		Action: BeginNewLineBelow,
	},
	{
		Name: "begin new line above (O)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: 'O'},
		},
		Action: BeginNewLineAbove,
	},
	{
		Name: "delete line (dd)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: 'd'},
			{Key: tcell.KeyRune, Rune: 'd'},
		},
		Action: DeleteLine,
	},
	{
		Name: "delete prev char in line (dh)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: 'd'},
			{Key: tcell.KeyRune, Rune: 'h'},
		},
		Action: DeletePrevCharInLine,
	},
	{
		Name: "delete down (dj)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: 'd'},
			{Key: tcell.KeyRune, Rune: 'j'},
		},
		Action: DeleteDown,
	},
	{
		Name: "delete up (dk)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: 'd'},
			{Key: tcell.KeyRune, Rune: 'k'},
		},
		Action: DeleteUp,
	},
	{
		Name: "delete next char in line (dl)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: 'd'},
			{Key: tcell.KeyRune, Rune: 'l'},
		},
		Action: DeleteNextCharInLine,
	},
	{
		Name: "delete to end of line (d$)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: 'd'},
			{Key: tcell.KeyRune, Rune: '$'},
		},
		Action: DeleteToEndOfLine,
	},
	{
		Name: "delete to start of line (d0)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: 'd'},
			{Key: tcell.KeyRune, Rune: '0'},
		},
		Action: DeleteToStartOfLine,
	},
	{
		Name: "delete to start of line non-whitespace (d^)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: 'd'},
			{Key: tcell.KeyRune, Rune: '^'},
		},
		Action: DeleteToStartOfLineNonWhitespace,
	},
	{
		Name: "delete to end of line (D)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: 'D'},
		},
		Action: DeleteToEndOfLine,
	},
	{
		Name: "delete inner word (diw)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: 'd'},
			{Key: tcell.KeyRune, Rune: 'i'},
			{Key: tcell.KeyRune, Rune: 'w'},
		},
		Action: DeleteInnerWord,
	},
	{
		Name: "change inner word (ciw)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: 'c'},
			{Key: tcell.KeyRune, Rune: 'i'},
			{Key: tcell.KeyRune, Rune: 'w'},
		},
		Action: ChangeInnerWord,
	},
	{
		Name: "replace character (r)",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: 'r'},
			{Wildcard: true},
		},
		Action: ReplaceCharacter,
	},
	{
		Name: "scroll up",
		Pattern: []EventMatcher{
			{Key: tcell.KeyCtrlU},
		},
		Action: ScrollUp,
	},
	{
		Name: "scroll down",
		Pattern: []EventMatcher{
			{Key: tcell.KeyCtrlD},
		},
		Action: ScrollDown,
	},
	{
		Name: "show command menu",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: ':'},
		},
		Action: ShowCommandMenu,
	},
	{
		Name: "start search",
		Pattern: []EventMatcher{
			{Key: tcell.KeyRune, Rune: '/'},
		},
		Action: StartSearchForward,
	},
}
