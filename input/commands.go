package input

import (
	"github.com/gdamore/tcell/v2"

	"github.com/aretext/aretext/clipboard"
)

// ActionBuilder is invoked when the input parser accepts a sequence of keypresses matching a command.
type ActionBuilder func(p ActionBuilderParams) Action

type ActionBuilderParams struct {
	InputEvents          []*tcell.EventKey
	CountArg             *uint64
	ClipboardPageNameArg *rune
	Config               Config
}

func (p ActionBuilderParams) CountOrDefault() uint64 {
	if p.CountArg == nil {
		return 1
	}
	return *p.CountArg
}

func (p ActionBuilderParams) ClipboardPageOrDefault() clipboard.PageId {
	if p.ClipboardPageNameArg == nil {
		return clipboard.PageDefault
	}
	return clipboard.PageIdForLetter(*p.ClipboardPageNameArg)
}

func (p ActionBuilderParams) LastChar() rune {
	if len(p.InputEvents) == 0 {
		return '\x00'
	}
	lastEvent := p.InputEvents[len(p.InputEvents)-1]
	if lastEvent.Key() != tcell.KeyRune {
		return '\x00'
	}
	return lastEvent.Rune()
}

// Command defines a command that the input parser can recognize.
// The pattern is a sequence of keypresses that trigger the command.
type Command struct {
	Name                 string
	Pattern              []EventMatcher
	ActionBuilder        ActionBuilder
	AddToLastActionMacro bool
	AddToUserMacro       bool
}

// These commands control cursor movement in normal and visual mode.
func cursorCommands() []Command {
	return []Command{
		{
			Name: "cursor left (arrow)",
			Pattern: []EventMatcher{
				{Key: tcell.KeyLeft},
			},
			ActionBuilder: func(p ActionBuilderParams) Action {
				return CursorLeft
			},
			AddToUserMacro: true,
		},
		{
			Name: "cursor left (h)",
			Pattern: []EventMatcher{
				{Key: tcell.KeyRune, Rune: 'h'},
			},
			ActionBuilder: func(p ActionBuilderParams) Action {
				return CursorLeft
			},
			AddToUserMacro: true,
		},
		{
			Name: "cursor right (arrow)",
			Pattern: []EventMatcher{
				{Key: tcell.KeyRight},
			},
			ActionBuilder: func(p ActionBuilderParams) Action {
				return CursorRight
			},
			AddToUserMacro: true,
		},
		{
			Name: "cursor right (l)",
			Pattern: []EventMatcher{
				{Key: tcell.KeyRune, Rune: 'l'},
			},
			ActionBuilder: func(p ActionBuilderParams) Action {
				return CursorRight
			},
			AddToUserMacro: true,
		},
		{
			Name: "cursor up (arrow)",
			Pattern: []EventMatcher{
				{Key: tcell.KeyUp},
			},
			ActionBuilder: func(p ActionBuilderParams) Action {
				return CursorUp
			},
			AddToUserMacro: true,
		},
		{
			Name: "cursor up (k)",
			Pattern: []EventMatcher{
				{Key: tcell.KeyRune, Rune: 'k'},
			},
			ActionBuilder: func(p ActionBuilderParams) Action {
				return CursorUp
			},
			AddToUserMacro: true,
		},
		{
			Name: "cursor down (arrow)",
			Pattern: []EventMatcher{
				{Key: tcell.KeyDown},
			},
			ActionBuilder: func(p ActionBuilderParams) Action {
				return CursorDown
			},
			AddToUserMacro: true,
		},
		{
			Name: "cursor down (j)",
			Pattern: []EventMatcher{
				{Key: tcell.KeyRune, Rune: 'j'},
			},
			ActionBuilder: func(p ActionBuilderParams) Action {
				return CursorDown
			},
			AddToUserMacro: true,
		},
		{
			Name: "cursor back (backspace)",
			Pattern: []EventMatcher{
				{Key: tcell.KeyBackspace},
			},
			ActionBuilder: func(p ActionBuilderParams) Action {
				return CursorBack
			},
			AddToUserMacro: true,
		},
		{
			Name: "cursor back (backspace2)",
			Pattern: []EventMatcher{
				{Key: tcell.KeyBackspace2},
			},
			ActionBuilder: func(p ActionBuilderParams) Action {
				return CursorBack
			},
			AddToUserMacro: true,
		},
		{
			Name: "cursor next word start (w)",
			Pattern: []EventMatcher{
				{Key: tcell.KeyRune, Rune: 'w'},
			},
			ActionBuilder: func(p ActionBuilderParams) Action {
				return CursorNextWordStart
			},
			AddToUserMacro: true,
		},
		{
			Name: "cursor prev word start (b)",
			Pattern: []EventMatcher{
				{Key: tcell.KeyRune, Rune: 'b'},
			},
			ActionBuilder: func(p ActionBuilderParams) Action {
				return CursorPrevWordStart
			},
			AddToUserMacro: true,
		},
		{
			Name: "cursor next word end (e)",
			Pattern: []EventMatcher{
				{Key: tcell.KeyRune, Rune: 'e'},
			},
			ActionBuilder: func(p ActionBuilderParams) Action {
				return CursorNextWordEnd
			},
			AddToUserMacro: true,
		},
		{
			Name: "cursor prev paragraph ({)",
			Pattern: []EventMatcher{
				{Key: tcell.KeyRune, Rune: '{'},
			},
			ActionBuilder: func(p ActionBuilderParams) Action {
				return CursorPrevParagraph
			},
			AddToUserMacro: true,
		},
		{
			Name: "cursor next paragraph (})",
			Pattern: []EventMatcher{
				{Key: tcell.KeyRune, Rune: '}'},
			},
			ActionBuilder: func(p ActionBuilderParams) Action {
				return CursorNextParagraph
			},
			AddToUserMacro: true,
		},
		{
			Name: "cursor to next matching char (f{char})",
			Pattern: []EventMatcher{
				{Key: tcell.KeyRune, Rune: 'f'},
				{Wildcard: true},
			},
			ActionBuilder: func(p ActionBuilderParams) Action {
				char := p.LastChar()
				if char == '\x00' {
					return EmptyAction
				}
				return CursorToNextMatchingChar(char, p.CountOrDefault(), true)
			},
			AddToUserMacro: true,
		},
		{
			Name: "cursor to prev matching char (F{char})",
			Pattern: []EventMatcher{
				{Key: tcell.KeyRune, Rune: 'F'},
				{Wildcard: true},
			},
			ActionBuilder: func(p ActionBuilderParams) Action {
				char := p.LastChar()
				if char == '\x00' {
					return EmptyAction
				}
				return CursorToPrevMatchingChar(char, p.CountOrDefault(), true)
			},
			AddToUserMacro: true,
		},
		{
			Name: "cursor till next matching char (t{char})",
			Pattern: []EventMatcher{
				{Key: tcell.KeyRune, Rune: 't'},
				{Wildcard: true},
			},
			ActionBuilder: func(p ActionBuilderParams) Action {
				char := p.LastChar()
				if char == '\x00' {
					return EmptyAction
				}
				return CursorToNextMatchingChar(char, p.CountOrDefault(), false)
			},
			AddToUserMacro: true,
		},
		{
			Name: "cursor to prev matching char (T{char})",
			Pattern: []EventMatcher{
				{Key: tcell.KeyRune, Rune: 'T'},
				{Wildcard: true},
			},
			ActionBuilder: func(p ActionBuilderParams) Action {
				char := p.LastChar()
				if char == '\x00' {
					return EmptyAction
				}
				return CursorToPrevMatchingChar(char, p.CountOrDefault(), false)
			},
			AddToUserMacro: true,
		},
		{
			Name: "cursor line start (0)",
			Pattern: []EventMatcher{
				{Key: tcell.KeyRune, Rune: '0'},
			},
			ActionBuilder: func(p ActionBuilderParams) Action {
				return CursorLineStart
			},
			AddToUserMacro: true,
		},
		{
			Name: "cursor line start non-whitespace (^)",
			Pattern: []EventMatcher{
				{Key: tcell.KeyRune, Rune: '^'},
			},
			ActionBuilder: func(p ActionBuilderParams) Action {
				return CursorLineStartNonWhitespace
			},
			AddToUserMacro: true,
		},
		{
			Name: "cursor line end ($)",
			Pattern: []EventMatcher{
				{Key: tcell.KeyRune, Rune: '$'},
			},
			ActionBuilder: func(p ActionBuilderParams) Action {
				return CursorLineEnd
			},
			AddToUserMacro: true,
		},
		{
			Name: "cursor start of line num (gg)",
			Pattern: []EventMatcher{
				{Key: tcell.KeyRune, Rune: 'g'},
				{Key: tcell.KeyRune, Rune: 'g'},
			},
			ActionBuilder: func(p ActionBuilderParams) Action {
				return CursorStartOfLineNum(p.CountOrDefault())
			},
			AddToUserMacro: true,
		},
		{
			Name: "cursor start of last line (G)",
			Pattern: []EventMatcher{
				{Key: tcell.KeyRune, Rune: 'G'},
			},
			ActionBuilder: func(p ActionBuilderParams) Action {
				return CursorStartOfLastLine
			},
			AddToUserMacro: true,
		},
		{
			Name: "scroll up (ctrl-u)",
			Pattern: []EventMatcher{
				{Key: tcell.KeyCtrlU},
			},
			ActionBuilder: func(p ActionBuilderParams) Action {
				return ScrollUp(p.Config)
			},
		},
		{
			Name: "scroll down (ctrl-d)",
			Pattern: []EventMatcher{
				{Key: tcell.KeyCtrlD},
			},
			ActionBuilder: func(p ActionBuilderParams) Action {
				return ScrollDown(p.Config)
			},
		},
	}
}

// These commands are used when the editor is in normal mode.
func normalModeCommands() []Command {
	return append(cursorCommands(), []Command{
		{
			Name: "delete next char in line (x)",
			Pattern: []EventMatcher{
				{Key: tcell.KeyRune, Rune: 'x'},
			},
			ActionBuilder: func(p ActionBuilderParams) Action {
				return DeleteNextCharInLine(p.CountOrDefault(), p.ClipboardPageOrDefault())
			},
			AddToLastActionMacro: true,
			AddToUserMacro:       true,
		},
		{
			Name: "enter insert mode (i)",
			Pattern: []EventMatcher{
				{Key: tcell.KeyRune, Rune: 'i'},
			},
			ActionBuilder: func(p ActionBuilderParams) Action {
				return EnterInsertMode
			},
			AddToLastActionMacro: true,
			AddToUserMacro:       true,
		},
		{
			Name: "enter insert mode at start of line (I)",
			Pattern: []EventMatcher{
				{Key: tcell.KeyRune, Rune: 'I'},
			},
			ActionBuilder: func(p ActionBuilderParams) Action {
				return EnterInsertModeAtStartOfLine
			},
			AddToLastActionMacro: true,
			AddToUserMacro:       true,
		},
		{
			Name: "enter insert mode at next pos (a)",
			Pattern: []EventMatcher{
				{Key: tcell.KeyRune, Rune: 'a'},
			},
			ActionBuilder: func(p ActionBuilderParams) Action {
				return EnterInsertModeAtNextPos
			},
			AddToLastActionMacro: true,
			AddToUserMacro:       true,
		},
		{
			Name: "enter insert mode at end of line (A)",
			Pattern: []EventMatcher{
				{Key: tcell.KeyRune, Rune: 'A'},
			},
			ActionBuilder: func(p ActionBuilderParams) Action {
				return EnterInsertModeAtEndOfLine
			},
			AddToLastActionMacro: true,
			AddToUserMacro:       true,
		},
		{
			Name: "begin new line below (o)",
			Pattern: []EventMatcher{
				{Key: tcell.KeyRune, Rune: 'o'},
			},
			ActionBuilder: func(p ActionBuilderParams) Action {
				return BeginNewLineBelow
			},
			AddToLastActionMacro: true,
			AddToUserMacro:       true,
		},
		{
			Name: "begin new line above (O)",
			Pattern: []EventMatcher{
				{Key: tcell.KeyRune, Rune: 'O'},
			},
			ActionBuilder: func(p ActionBuilderParams) Action {
				return BeginNewLineAbove
			},
			AddToLastActionMacro: true,
			AddToUserMacro:       true,
		},
		{
			Name: "join lines (J)",
			Pattern: []EventMatcher{
				{Key: tcell.KeyRune, Rune: 'J'},
			},
			ActionBuilder: func(p ActionBuilderParams) Action {
				return JoinLines
			},
			AddToLastActionMacro: true,
			AddToUserMacro:       true,
		},
		{
			Name: "delete line (dd)",
			Pattern: []EventMatcher{
				{Key: tcell.KeyRune, Rune: 'd'},
				{Key: tcell.KeyRune, Rune: 'd'},
			},
			ActionBuilder: func(p ActionBuilderParams) Action {
				return DeleteLines(p.CountOrDefault(), p.ClipboardPageOrDefault())
			},
			AddToLastActionMacro: true,
			AddToUserMacro:       true,
		},
		{
			Name: "delete prev char in line (dh)",
			Pattern: []EventMatcher{
				{Key: tcell.KeyRune, Rune: 'd'},
				{Key: tcell.KeyRune, Rune: 'h'},
			},
			ActionBuilder: func(p ActionBuilderParams) Action {
				return DeletePrevCharInLine(p.ClipboardPageOrDefault())
			},
			AddToLastActionMacro: true,
			AddToUserMacro:       true,
		},
		{
			Name: "delete down (dj)",
			Pattern: []EventMatcher{
				{Key: tcell.KeyRune, Rune: 'd'},
				{Key: tcell.KeyRune, Rune: 'j'},
			},
			ActionBuilder: func(p ActionBuilderParams) Action {
				return DeleteDown(p.ClipboardPageOrDefault())
			},
			AddToLastActionMacro: true,
			AddToUserMacro:       true,
		},
		{
			Name: "delete up (dk)",
			Pattern: []EventMatcher{
				{Key: tcell.KeyRune, Rune: 'd'},
				{Key: tcell.KeyRune, Rune: 'k'},
			},
			ActionBuilder: func(p ActionBuilderParams) Action {
				return DeleteUp(p.ClipboardPageOrDefault())
			},
			AddToLastActionMacro: true,
			AddToUserMacro:       true,
		},
		{
			Name: "delete next char in line (dl)",
			Pattern: []EventMatcher{
				{Key: tcell.KeyRune, Rune: 'd'},
				{Key: tcell.KeyRune, Rune: 'l'},
			},
			ActionBuilder: func(p ActionBuilderParams) Action {
				return DeleteNextCharInLine(p.CountOrDefault(), p.ClipboardPageOrDefault())
			},
			AddToLastActionMacro: true,
			AddToUserMacro:       true,
		},
		{
			Name: "delete to end of line (d$)",
			Pattern: []EventMatcher{
				{Key: tcell.KeyRune, Rune: 'd'},
				{Key: tcell.KeyRune, Rune: '$'},
			},
			ActionBuilder: func(p ActionBuilderParams) Action {
				return DeleteToEndOfLine(p.ClipboardPageOrDefault())
			},
			AddToLastActionMacro: true,
			AddToUserMacro:       true,
		},
		{
			Name: "delete to start of line (d0)",
			Pattern: []EventMatcher{
				{Key: tcell.KeyRune, Rune: 'd'},
				{Key: tcell.KeyRune, Rune: '0'},
			},
			ActionBuilder: func(p ActionBuilderParams) Action {
				return DeleteToStartOfLine(p.ClipboardPageOrDefault())
			},
			AddToLastActionMacro: true,
			AddToUserMacro:       true,
		},
		{
			Name: "delete to start of line non-whitespace (d^)",
			Pattern: []EventMatcher{
				{Key: tcell.KeyRune, Rune: 'd'},
				{Key: tcell.KeyRune, Rune: '^'},
			},
			ActionBuilder: func(p ActionBuilderParams) Action {
				return DeleteToStartOfLineNonWhitespace(p.ClipboardPageOrDefault())
			},
			AddToLastActionMacro: true,
			AddToUserMacro:       true,
		},
		{
			Name: "delete to end of line (D)",
			Pattern: []EventMatcher{
				{Key: tcell.KeyRune, Rune: 'D'},
			},
			ActionBuilder: func(p ActionBuilderParams) Action {
				return DeleteToEndOfLine(p.ClipboardPageOrDefault())
			},
			AddToLastActionMacro: true,
			AddToUserMacro:       true,
		},
		{
			Name: "delete to next matching char (df{char})",
			Pattern: []EventMatcher{
				{Key: tcell.KeyRune, Rune: 'd'},
				{Key: tcell.KeyRune, Rune: 'f'},
				{Wildcard: true},
			},
			ActionBuilder: func(p ActionBuilderParams) Action {
				char := p.LastChar()
				if char == '\x00' {
					return EmptyAction
				}
				return DeleteToNextMatchingChar(char, p.CountOrDefault(), p.ClipboardPageOrDefault(), true)
			},
			AddToLastActionMacro: true,
			AddToUserMacro:       true,
		},
		{
			Name: "delete to prev matching char (dF{char})",
			Pattern: []EventMatcher{
				{Key: tcell.KeyRune, Rune: 'd'},
				{Key: tcell.KeyRune, Rune: 'F'},
				{Wildcard: true},
			},
			ActionBuilder: func(p ActionBuilderParams) Action {
				char := p.LastChar()
				if char == '\x00' {
					return EmptyAction
				}
				return DeleteToPrevMatchingChar(char, p.CountOrDefault(), p.ClipboardPageOrDefault(), true)
			},
			AddToLastActionMacro: true,
			AddToUserMacro:       true,
		},
		{
			Name: "delete till next matching char (dt{char})",
			Pattern: []EventMatcher{
				{Key: tcell.KeyRune, Rune: 'd'},
				{Key: tcell.KeyRune, Rune: 't'},
				{Wildcard: true},
			},
			ActionBuilder: func(p ActionBuilderParams) Action {
				char := p.LastChar()
				if char == '\x00' {
					return EmptyAction
				}
				return DeleteToNextMatchingChar(char, p.CountOrDefault(), p.ClipboardPageOrDefault(), false)
			},
			AddToLastActionMacro: true,
			AddToUserMacro:       true,
		},
		{
			Name: "delete till prev matching char (dT{char})",
			Pattern: []EventMatcher{
				{Key: tcell.KeyRune, Rune: 'd'},
				{Key: tcell.KeyRune, Rune: 'T'},
				{Wildcard: true},
			},
			ActionBuilder: func(p ActionBuilderParams) Action {
				char := p.LastChar()
				if char == '\x00' {
					return EmptyAction
				}
				return DeleteToPrevMatchingChar(char, p.CountOrDefault(), p.ClipboardPageOrDefault(), false)
			},
			AddToLastActionMacro: true,
			AddToUserMacro:       true,
		},
		{
			Name: "delete to start of next word (dw)",
			Pattern: []EventMatcher{
				{Key: tcell.KeyRune, Rune: 'd'},
				{Key: tcell.KeyRune, Rune: 'w'},
			},
			ActionBuilder: func(p ActionBuilderParams) Action {
				return DeleteToStartOfNextWord(p.ClipboardPageOrDefault())
			},
			AddToLastActionMacro: true,
			AddToUserMacro:       true,
		},
		{
			Name: "delete a word (daw)",
			Pattern: []EventMatcher{
				{Key: tcell.KeyRune, Rune: 'd'},
				{Key: tcell.KeyRune, Rune: 'a'},
				{Key: tcell.KeyRune, Rune: 'w'},
			},
			ActionBuilder: func(p ActionBuilderParams) Action {
				return DeleteAWord(p.ClipboardPageOrDefault())
			},
			AddToLastActionMacro: true,
			AddToUserMacro:       true,
		},
		{
			Name: "delete inner word (diw)",
			Pattern: []EventMatcher{
				{Key: tcell.KeyRune, Rune: 'd'},
				{Key: tcell.KeyRune, Rune: 'i'},
				{Key: tcell.KeyRune, Rune: 'w'},
			},
			ActionBuilder: func(p ActionBuilderParams) Action {
				return DeleteInnerWord(p.ClipboardPageOrDefault())
			},
			AddToLastActionMacro: true,
			AddToUserMacro:       true,
		},
		{
			Name: "change to start of next word (cw)",
			Pattern: []EventMatcher{
				{Key: tcell.KeyRune, Rune: 'c'},
				{Key: tcell.KeyRune, Rune: 'w'},
			},
			ActionBuilder: func(p ActionBuilderParams) Action {
				return ChangeToStartOfNextWord(p.ClipboardPageOrDefault())
			},
			AddToLastActionMacro: true,
			AddToUserMacro:       true,
		},
		{
			Name: "change a word (caw)",
			Pattern: []EventMatcher{
				{Key: tcell.KeyRune, Rune: 'c'},
				{Key: tcell.KeyRune, Rune: 'a'},
				{Key: tcell.KeyRune, Rune: 'w'},
			},
			ActionBuilder: func(p ActionBuilderParams) Action {
				return ChangeAWord(p.ClipboardPageOrDefault())
			},
			AddToLastActionMacro: true,
			AddToUserMacro:       true,
		},
		{
			Name: "change inner word (ciw)",
			Pattern: []EventMatcher{
				{Key: tcell.KeyRune, Rune: 'c'},
				{Key: tcell.KeyRune, Rune: 'i'},
				{Key: tcell.KeyRune, Rune: 'w'},
			},
			ActionBuilder: func(p ActionBuilderParams) Action {
				return ChangeInnerWord(p.ClipboardPageOrDefault())
			},
			AddToLastActionMacro: true,
			AddToUserMacro:       true,
		},
		{
			Name: "change to next matching char (cf{char})",
			Pattern: []EventMatcher{
				{Key: tcell.KeyRune, Rune: 'c'},
				{Key: tcell.KeyRune, Rune: 'f'},
				{Wildcard: true},
			},
			ActionBuilder: func(p ActionBuilderParams) Action {
				char := p.LastChar()
				if char == '\x00' {
					return EmptyAction
				}
				return ChangeToNextMatchingChar(char, p.CountOrDefault(), p.ClipboardPageOrDefault(), true)
			},
			AddToLastActionMacro: true,
			AddToUserMacro:       true,
		},
		{
			Name: "change to prev matching char (cF{char})",
			Pattern: []EventMatcher{
				{Key: tcell.KeyRune, Rune: 'c'},
				{Key: tcell.KeyRune, Rune: 'F'},
				{Wildcard: true},
			},
			ActionBuilder: func(p ActionBuilderParams) Action {
				char := p.LastChar()
				if char == '\x00' {
					return EmptyAction
				}
				return ChangeToPrevMatchingChar(char, p.CountOrDefault(), p.ClipboardPageOrDefault(), true)
			},
			AddToLastActionMacro: true,
			AddToUserMacro:       true,
		},
		{
			Name: "change till next matching char (ct{char})",
			Pattern: []EventMatcher{
				{Key: tcell.KeyRune, Rune: 'c'},
				{Key: tcell.KeyRune, Rune: 't'},
				{Wildcard: true},
			},
			ActionBuilder: func(p ActionBuilderParams) Action {
				char := p.LastChar()
				if char == '\x00' {
					return EmptyAction
				}
				return ChangeToNextMatchingChar(char, p.CountOrDefault(), p.ClipboardPageOrDefault(), false)
			},
			AddToLastActionMacro: true,
			AddToUserMacro:       true,
		},
		{
			Name: "change till prev matching char (cT{char})",
			Pattern: []EventMatcher{
				{Key: tcell.KeyRune, Rune: 'c'},
				{Key: tcell.KeyRune, Rune: 'T'},
				{Wildcard: true},
			},
			ActionBuilder: func(p ActionBuilderParams) Action {
				char := p.LastChar()
				if char == '\x00' {
					return EmptyAction
				}
				return ChangeToPrevMatchingChar(char, p.CountOrDefault(), p.ClipboardPageOrDefault(), false)
			},
			AddToLastActionMacro: true,
			AddToUserMacro:       true,
		},
		{
			Name: "replace character (r)",
			Pattern: []EventMatcher{
				{Key: tcell.KeyRune, Rune: 'r'},
				{Wildcard: true},
			},
			ActionBuilder: func(p ActionBuilderParams) Action {
				if len(p.InputEvents) == 0 {
					return EmptyAction
				}

				lastEvent := p.InputEvents[len(p.InputEvents)-1]
				var newChar rune
				switch lastEvent.Key() {
				case tcell.KeyEnter:
					newChar = '\n'
				case tcell.KeyTab:
					newChar = '\t'
				case tcell.KeyRune:
					newChar = lastEvent.Rune()
				default:
					return EmptyAction
				}

				return ReplaceCharacter(newChar)
			},
			AddToLastActionMacro: true,
			AddToUserMacro:       true,
		},
		{
			Name: "toggle case (~)",
			Pattern: []EventMatcher{
				{Key: tcell.KeyRune, Rune: '~'},
			},
			ActionBuilder: func(p ActionBuilderParams) Action {
				return ToggleCaseAtCursor
			},
			AddToLastActionMacro: true,
			AddToUserMacro:       true,
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
			AddToLastActionMacro: true,
			AddToUserMacro:       true,
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
			AddToLastActionMacro: true,
			AddToUserMacro:       true,
		},
		{
			Name: "yank to start of next word (yw)",
			Pattern: []EventMatcher{
				{Key: tcell.KeyRune, Rune: 'y'},
				{Key: tcell.KeyRune, Rune: 'w'},
			},
			ActionBuilder: func(p ActionBuilderParams) Action {
				return CopyToStartOfNextWord(p.ClipboardPageOrDefault())
			},
			AddToLastActionMacro: true,
			AddToUserMacro:       true,
		},
		{
			Name: "yank a word (yaw)",
			Pattern: []EventMatcher{
				{Key: tcell.KeyRune, Rune: 'y'},
				{Key: tcell.KeyRune, Rune: 'a'},
				{Key: tcell.KeyRune, Rune: 'w'},
			},
			ActionBuilder: func(p ActionBuilderParams) Action {
				return CopyAWord(p.ClipboardPageOrDefault())
			},
			AddToLastActionMacro: true,
			AddToUserMacro:       true,
		},
		{
			Name: "yank inner word (yiw)",
			Pattern: []EventMatcher{
				{Key: tcell.KeyRune, Rune: 'y'},
				{Key: tcell.KeyRune, Rune: 'i'},
				{Key: tcell.KeyRune, Rune: 'w'},
			},
			ActionBuilder: func(p ActionBuilderParams) Action {
				return CopyInnerWord(p.ClipboardPageOrDefault())
			},
			AddToLastActionMacro: true,
			AddToUserMacro:       true,
		},
		{
			Name: "yank line (yy)",
			Pattern: []EventMatcher{
				{Key: tcell.KeyRune, Rune: 'y'},
				{Key: tcell.KeyRune, Rune: 'y'},
			},
			ActionBuilder: func(p ActionBuilderParams) Action {
				return CopyLines(p.ClipboardPageOrDefault())
			},
			AddToLastActionMacro: true,
			AddToUserMacro:       true,
		},
		{
			Name: "put after cursor (p)",
			Pattern: []EventMatcher{
				{Key: tcell.KeyRune, Rune: 'p'},
			},
			ActionBuilder: func(p ActionBuilderParams) Action {
				return PasteAfterCursor(p.ClipboardPageOrDefault())
			},
			AddToLastActionMacro: true,
			AddToUserMacro:       true,
		},
		{
			Name: "put before cursor (P)",
			Pattern: []EventMatcher{
				{Key: tcell.KeyRune, Rune: 'P'},
			},
			ActionBuilder: func(p ActionBuilderParams) Action {
				return PasteBeforeCursor(p.ClipboardPageOrDefault())
			},
			AddToLastActionMacro: true,
			AddToUserMacro:       true,
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
			AddToUserMacro: true,
		},
		{
			Name: "start backward search",
			Pattern: []EventMatcher{
				{Key: tcell.KeyRune, Rune: '?'},
			},
			ActionBuilder: func(p ActionBuilderParams) Action {
				return StartSearchBackward
			},
			AddToUserMacro: true,
		},
		{
			Name: "find next match",
			Pattern: []EventMatcher{
				{Key: tcell.KeyRune, Rune: 'n'},
			},
			ActionBuilder: func(p ActionBuilderParams) Action {
				return FindNextMatch
			},
			AddToUserMacro: true,
		},
		{
			Name: "find previous match",
			Pattern: []EventMatcher{
				{Key: tcell.KeyRune, Rune: 'N'},
			},
			ActionBuilder: func(p ActionBuilderParams) Action {
				return FindPrevMatch
			},
			AddToUserMacro: true,
		},
		{
			Name: "undo (u)",
			Pattern: []EventMatcher{
				{Key: tcell.KeyRune, Rune: 'u'},
			},
			ActionBuilder: func(p ActionBuilderParams) Action {
				return Undo
			},
			AddToUserMacro: true,
		},
		{
			Name: "redo (ctrl-r)",
			Pattern: []EventMatcher{
				{Key: tcell.KeyCtrlR},
			},
			ActionBuilder: func(p ActionBuilderParams) Action {
				return Redo
			},
			AddToUserMacro: true,
		},
		{
			Name: "enter visual mode charwise (v)",
			Pattern: []EventMatcher{
				{Key: tcell.KeyRune, Rune: 'v'},
			},
			ActionBuilder: func(p ActionBuilderParams) Action {
				return ToggleVisualModeCharwise
			},
			AddToUserMacro: true,
		},
		{
			Name: "enter visual mode linewise (V)",
			Pattern: []EventMatcher{
				{Key: tcell.KeyRune, Rune: 'V'},
			},
			ActionBuilder: func(p ActionBuilderParams) Action {
				return ToggleVisualModeLinewise
			},
			AddToUserMacro: true,
		},
		{
			Name: "repeat last action (.)",
			Pattern: []EventMatcher{
				{Key: tcell.KeyRune, Rune: '.'},
			},
			ActionBuilder: func(p ActionBuilderParams) Action {
				return ReplayLastActionMacro(p.CountOrDefault())
			},
			AddToUserMacro: true,
		},
	}...)
}

// These commands are used when the editor is in visual mode.
func visualModeCommands() []Command {
	return append(cursorCommands(), []Command{
		{
			Name: "toggle visual mode charwise (v)",
			Pattern: []EventMatcher{
				{Key: tcell.KeyRune, Rune: 'v'},
			},
			ActionBuilder: func(p ActionBuilderParams) Action {
				return ToggleVisualModeCharwise
			},
			AddToUserMacro: true,
		},
		{
			Name: "toggle visual mode linewise (V)",
			Pattern: []EventMatcher{
				{Key: tcell.KeyRune, Rune: 'V'},
			},
			ActionBuilder: func(p ActionBuilderParams) Action {
				return ToggleVisualModeLinewise
			},
			AddToUserMacro: true,
		},
		{
			Name: "return to normal mode (esc)",
			Pattern: []EventMatcher{
				{Key: tcell.KeyEscape},
			},
			ActionBuilder: func(p ActionBuilderParams) Action {
				return ReturnToNormalMode
			},
			AddToUserMacro: true,
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
				return DeleteSelectionAndReturnToNormalMode(
					p.ClipboardPageOrDefault(),
					p.Config.SelectionMode,
					p.Config.SelectionEndLocator,
				)
			},
			AddToLastActionMacro: true,
			AddToUserMacro:       true,
		},
		{
			Name: "delete selection (d)",
			Pattern: []EventMatcher{
				{Key: tcell.KeyRune, Rune: 'd'},
			},
			ActionBuilder: func(p ActionBuilderParams) Action {
				return DeleteSelectionAndReturnToNormalMode(
					p.ClipboardPageOrDefault(),
					p.Config.SelectionMode,
					p.Config.SelectionEndLocator,
				)
			},
			AddToLastActionMacro: true,
			AddToUserMacro:       true,
		},
		{
			Name: "change selection (c)",
			Pattern: []EventMatcher{
				{Key: tcell.KeyRune, Rune: 'c'},
			},
			ActionBuilder: func(p ActionBuilderParams) Action {
				return ChangeSelection(
					p.ClipboardPageOrDefault(),
					p.Config.SelectionMode,
					p.Config.SelectionEndLocator,
				)
			},
			AddToLastActionMacro: true,
			AddToUserMacro:       true,
		},
		{
			Name: "toggle case for selection (~)",
			Pattern: []EventMatcher{
				{Key: tcell.KeyRune, Rune: '~'},
			},
			ActionBuilder: func(p ActionBuilderParams) Action {
				return ToggleCaseInSelectionAndReturnToNormalMode(p.Config.SelectionEndLocator)
			},
			AddToLastActionMacro: true,
			AddToUserMacro:       true,
		},
		{
			Name: "indent selection (>)",
			Pattern: []EventMatcher{
				{Key: tcell.KeyRune, Rune: '>'},
			},
			ActionBuilder: func(p ActionBuilderParams) Action {
				return IndentSelectionAndReturnToNormalMode(p.Config.SelectionEndLocator)
			},
			AddToLastActionMacro: true,
			AddToUserMacro:       true,
		},
		{
			Name: "outdent selection (<)",
			Pattern: []EventMatcher{
				{Key: tcell.KeyRune, Rune: '<'},
			},
			ActionBuilder: func(p ActionBuilderParams) Action {
				return OutdentSelectionAndReturnToNormalMode(p.Config.SelectionEndLocator)
			},
			AddToLastActionMacro: true,
			AddToUserMacro:       true,
		},
		{
			Name: "yank selection (y)",
			Pattern: []EventMatcher{
				{Key: tcell.KeyRune, Rune: 'y'},
			},
			ActionBuilder: func(p ActionBuilderParams) Action {
				return CopySelectionAndReturnToNormalMode(p.ClipboardPageOrDefault())
			},
			AddToUserMacro: true,
		},
	}...)
}
