package input

import (
	"github.com/gdamore/tcell/v2"

	"github.com/aretext/aretext/clipboard"
	"github.com/aretext/aretext/state"
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
	Name          string
	Pattern       []EventMatcher
	ActionBuilder ActionBuilder
}

// These commands control cursor movement in normal and visual mode.
func cursorCommands() []Command {
	decorate := func(action Action) Action {
		return func(s *state.EditorState) {
			wrappedAction := func(s *state.EditorState) {
				action(s)
				state.ScrollViewToCursor(s)
				state.SetStatusMsg(s, state.StatusMsg{})
			}
			state.CheckpointUndoLog(s)
			wrappedAction(s)
			state.AddToRecordingUserMacro(s, state.MacroAction(wrappedAction))
		}
	}

	return []Command{
		{
			Name: "cursor left (arrow)",
			Pattern: []EventMatcher{
				{Key: tcell.KeyLeft},
			},
			ActionBuilder: func(p ActionBuilderParams) Action {
				return decorate(CursorLeft)
			},
		},
		{
			Name: "cursor left (h)",
			Pattern: []EventMatcher{
				{Key: tcell.KeyRune, Rune: 'h'},
			},
			ActionBuilder: func(p ActionBuilderParams) Action {
				return decorate(CursorLeft)
			},
		},
		{
			Name: "cursor right (arrow)",
			Pattern: []EventMatcher{
				{Key: tcell.KeyRight},
			},
			ActionBuilder: func(p ActionBuilderParams) Action {
				return decorate(CursorRight)
			},
		},
		{
			Name: "cursor right (l)",
			Pattern: []EventMatcher{
				{Key: tcell.KeyRune, Rune: 'l'},
			},
			ActionBuilder: func(p ActionBuilderParams) Action {
				return decorate(CursorRight)
			},
		},
		{
			Name: "cursor up (arrow)",
			Pattern: []EventMatcher{
				{Key: tcell.KeyUp},
			},
			ActionBuilder: func(p ActionBuilderParams) Action {
				return decorate(CursorUp)
			},
		},
		{
			Name: "cursor up (k)",
			Pattern: []EventMatcher{
				{Key: tcell.KeyRune, Rune: 'k'},
			},
			ActionBuilder: func(p ActionBuilderParams) Action {
				return decorate(CursorUp)
			},
		},
		{
			Name: "cursor down (arrow)",
			Pattern: []EventMatcher{
				{Key: tcell.KeyDown},
			},
			ActionBuilder: func(p ActionBuilderParams) Action {
				return decorate(CursorDown)
			},
		},
		{
			Name: "cursor down (j)",
			Pattern: []EventMatcher{
				{Key: tcell.KeyRune, Rune: 'j'},
			},
			ActionBuilder: func(p ActionBuilderParams) Action {
				return decorate(CursorDown)
			},
		},
		{
			Name: "cursor back (backspace)",
			Pattern: []EventMatcher{
				{Key: tcell.KeyBackspace},
			},
			ActionBuilder: func(p ActionBuilderParams) Action {
				return decorate(CursorBack)
			},
		},
		{
			Name: "cursor back (backspace2)",
			Pattern: []EventMatcher{
				{Key: tcell.KeyBackspace2},
			},
			ActionBuilder: func(p ActionBuilderParams) Action {
				return decorate(CursorBack)
			},
		},
		{
			Name: "cursor next word start (w)",
			Pattern: []EventMatcher{
				{Key: tcell.KeyRune, Rune: 'w'},
			},
			ActionBuilder: func(p ActionBuilderParams) Action {
				return decorate(CursorNextWordStart)
			},
		},
		{
			Name: "cursor prev word start (b)",
			Pattern: []EventMatcher{
				{Key: tcell.KeyRune, Rune: 'b'},
			},
			ActionBuilder: func(p ActionBuilderParams) Action {
				return decorate(CursorPrevWordStart)
			},
		},
		{
			Name: "cursor next word end (e)",
			Pattern: []EventMatcher{
				{Key: tcell.KeyRune, Rune: 'e'},
			},
			ActionBuilder: func(p ActionBuilderParams) Action {
				return decorate(CursorNextWordEnd)
			},
		},
		{
			Name: "cursor prev paragraph ({)",
			Pattern: []EventMatcher{
				{Key: tcell.KeyRune, Rune: '{'},
			},
			ActionBuilder: func(p ActionBuilderParams) Action {
				return decorate(CursorPrevParagraph)
			},
		},
		{
			Name: "cursor next paragraph (})",
			Pattern: []EventMatcher{
				{Key: tcell.KeyRune, Rune: '}'},
			},
			ActionBuilder: func(p ActionBuilderParams) Action {
				return decorate(CursorNextParagraph)
			},
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
				return decorate(CursorToNextMatchingChar(char, p.CountOrDefault(), true))
			},
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
				return decorate(CursorToPrevMatchingChar(char, p.CountOrDefault(), true))
			},
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
				return decorate(CursorToNextMatchingChar(char, p.CountOrDefault(), false))
			},
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
				return decorate(CursorToPrevMatchingChar(char, p.CountOrDefault(), false))
			},
		},
		{
			Name: "cursor line start (0)",
			Pattern: []EventMatcher{
				{Key: tcell.KeyRune, Rune: '0'},
			},
			ActionBuilder: func(p ActionBuilderParams) Action {
				return decorate(CursorLineStart)
			},
		},
		{
			Name: "cursor line start non-whitespace (^)",
			Pattern: []EventMatcher{
				{Key: tcell.KeyRune, Rune: '^'},
			},
			ActionBuilder: func(p ActionBuilderParams) Action {
				return decorate(CursorLineStartNonWhitespace)
			},
		},
		{
			Name: "cursor line end ($)",
			Pattern: []EventMatcher{
				{Key: tcell.KeyRune, Rune: '$'},
			},
			ActionBuilder: func(p ActionBuilderParams) Action {
				return decorate(CursorLineEnd)
			},
		},
		{
			Name: "cursor start of line num (gg)",
			Pattern: []EventMatcher{
				{Key: tcell.KeyRune, Rune: 'g'},
				{Key: tcell.KeyRune, Rune: 'g'},
			},
			ActionBuilder: func(p ActionBuilderParams) Action {
				return decorate(CursorStartOfLineNum(p.CountOrDefault()))
			},
		},
		{
			Name: "cursor start of last line (G)",
			Pattern: []EventMatcher{
				{Key: tcell.KeyRune, Rune: 'G'},
			},
			ActionBuilder: func(p ActionBuilderParams) Action {
				return decorate(CursorStartOfLastLine)
			},
		},
		{
			Name: "scroll up (ctrl-u)",
			Pattern: []EventMatcher{
				{Key: tcell.KeyCtrlU},
			},
			ActionBuilder: func(p ActionBuilderParams) Action {
				return decorate(ScrollUp(p.Config))
			},
		},
		{
			Name: "scroll down (ctrl-d)",
			Pattern: []EventMatcher{
				{Key: tcell.KeyCtrlD},
			},
			ActionBuilder: func(p ActionBuilderParams) Action {
				return decorate(ScrollDown(p.Config))
			},
		},
	}
}

type addToMacro struct {
	lastAction bool
	user       bool
}

func decorateNormalOrVisual(action Action, addToMacro addToMacro) Action {
	return func(s *state.EditorState) {
		wrappedAction := func(s *state.EditorState) {
			action(s)
			state.ScrollViewToCursor(s)
			state.SetStatusMsg(s, state.StatusMsg{})
		}

		state.CheckpointUndoLog(s)
		wrappedAction(s)

		if addToMacro.lastAction {
			state.ClearLastActionMacro(s)
			state.AddToLastActionMacro(s, state.MacroAction(wrappedAction))
		}

		if addToMacro.user {
			state.AddToRecordingUserMacro(s, state.MacroAction(wrappedAction))
		}
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
				return decorateNormalOrVisual(
					DeleteNextCharInLine(p.CountOrDefault(), p.ClipboardPageOrDefault()),
					addToMacro{lastAction: true, user: true})
			},
		},
		{
			Name: "enter insert mode (i)",
			Pattern: []EventMatcher{
				{Key: tcell.KeyRune, Rune: 'i'},
			},
			ActionBuilder: func(p ActionBuilderParams) Action {
				return decorateNormalOrVisual(
					EnterInsertMode,
					addToMacro{lastAction: true, user: true})
			},
		},
		{
			Name: "enter insert mode at start of line (I)",
			Pattern: []EventMatcher{
				{Key: tcell.KeyRune, Rune: 'I'},
			},
			ActionBuilder: func(p ActionBuilderParams) Action {
				return decorateNormalOrVisual(
					EnterInsertModeAtStartOfLine,
					addToMacro{lastAction: true, user: true})
			},
		},
		{
			Name: "enter insert mode at next pos (a)",
			Pattern: []EventMatcher{
				{Key: tcell.KeyRune, Rune: 'a'},
			},
			ActionBuilder: func(p ActionBuilderParams) Action {
				return decorateNormalOrVisual(
					EnterInsertModeAtNextPos,
					addToMacro{lastAction: true, user: true})
			},
		},
		{
			Name: "enter insert mode at end of line (A)",
			Pattern: []EventMatcher{
				{Key: tcell.KeyRune, Rune: 'A'},
			},
			ActionBuilder: func(p ActionBuilderParams) Action {
				return decorateNormalOrVisual(
					EnterInsertModeAtEndOfLine,
					addToMacro{lastAction: true, user: true})
			},
		},
		{
			Name: "begin new line below (o)",
			Pattern: []EventMatcher{
				{Key: tcell.KeyRune, Rune: 'o'},
			},
			ActionBuilder: func(p ActionBuilderParams) Action {
				return decorateNormalOrVisual(
					BeginNewLineBelow,
					addToMacro{lastAction: true, user: true})
			},
		},
		{
			Name: "begin new line above (O)",
			Pattern: []EventMatcher{
				{Key: tcell.KeyRune, Rune: 'O'},
			},
			ActionBuilder: func(p ActionBuilderParams) Action {
				return decorateNormalOrVisual(
					BeginNewLineAbove,
					addToMacro{lastAction: true, user: true})
			},
		},
		{
			Name: "join lines (J)",
			Pattern: []EventMatcher{
				{Key: tcell.KeyRune, Rune: 'J'},
			},
			ActionBuilder: func(p ActionBuilderParams) Action {
				return decorateNormalOrVisual(
					JoinLines,
					addToMacro{lastAction: true, user: true})
			},
		},
		{
			Name: "delete line (dd)",
			Pattern: []EventMatcher{
				{Key: tcell.KeyRune, Rune: 'd'},
				{Key: tcell.KeyRune, Rune: 'd'},
			},
			ActionBuilder: func(p ActionBuilderParams) Action {
				return decorateNormalOrVisual(
					DeleteLines(p.CountOrDefault(), p.ClipboardPageOrDefault()),
					addToMacro{lastAction: true, user: true})
			},
		},
		{
			Name: "delete prev char in line (dh)",
			Pattern: []EventMatcher{
				{Key: tcell.KeyRune, Rune: 'd'},
				{Key: tcell.KeyRune, Rune: 'h'},
			},
			ActionBuilder: func(p ActionBuilderParams) Action {
				return decorateNormalOrVisual(
					DeletePrevCharInLine(p.ClipboardPageOrDefault()),
					addToMacro{lastAction: true, user: true})
			},
		},
		{
			Name: "delete down (dj)",
			Pattern: []EventMatcher{
				{Key: tcell.KeyRune, Rune: 'd'},
				{Key: tcell.KeyRune, Rune: 'j'},
			},
			ActionBuilder: func(p ActionBuilderParams) Action {
				return decorateNormalOrVisual(
					DeleteDown(p.ClipboardPageOrDefault()),
					addToMacro{lastAction: true, user: true})
			},
		},
		{
			Name: "delete up (dk)",
			Pattern: []EventMatcher{
				{Key: tcell.KeyRune, Rune: 'd'},
				{Key: tcell.KeyRune, Rune: 'k'},
			},
			ActionBuilder: func(p ActionBuilderParams) Action {
				return decorateNormalOrVisual(
					DeleteUp(p.ClipboardPageOrDefault()),
					addToMacro{lastAction: true, user: true})
			},
		},
		{
			Name: "delete next char in line (dl)",
			Pattern: []EventMatcher{
				{Key: tcell.KeyRune, Rune: 'd'},
				{Key: tcell.KeyRune, Rune: 'l'},
			},
			ActionBuilder: func(p ActionBuilderParams) Action {
				return decorateNormalOrVisual(
					DeleteNextCharInLine(p.CountOrDefault(), p.ClipboardPageOrDefault()),
					addToMacro{lastAction: true, user: true})
			},
		},
		{
			Name: "delete to end of line (d$)",
			Pattern: []EventMatcher{
				{Key: tcell.KeyRune, Rune: 'd'},
				{Key: tcell.KeyRune, Rune: '$'},
			},
			ActionBuilder: func(p ActionBuilderParams) Action {
				return decorateNormalOrVisual(
					DeleteToEndOfLine(p.ClipboardPageOrDefault()),
					addToMacro{lastAction: true, user: true})
			},
		},
		{
			Name: "delete to start of line (d0)",
			Pattern: []EventMatcher{
				{Key: tcell.KeyRune, Rune: 'd'},
				{Key: tcell.KeyRune, Rune: '0'},
			},
			ActionBuilder: func(p ActionBuilderParams) Action {
				return decorateNormalOrVisual(
					DeleteToStartOfLine(p.ClipboardPageOrDefault()),
					addToMacro{lastAction: true, user: true})
			},
		},
		{
			Name: "delete to start of line non-whitespace (d^)",
			Pattern: []EventMatcher{
				{Key: tcell.KeyRune, Rune: 'd'},
				{Key: tcell.KeyRune, Rune: '^'},
			},
			ActionBuilder: func(p ActionBuilderParams) Action {
				return decorateNormalOrVisual(
					DeleteToStartOfLineNonWhitespace(p.ClipboardPageOrDefault()),
					addToMacro{lastAction: true, user: true})
			},
		},
		{
			Name: "delete to end of line (D)",
			Pattern: []EventMatcher{
				{Key: tcell.KeyRune, Rune: 'D'},
			},
			ActionBuilder: func(p ActionBuilderParams) Action {
				return decorateNormalOrVisual(
					DeleteToEndOfLine(p.ClipboardPageOrDefault()),
					addToMacro{lastAction: true, user: true})
			},
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
				return decorateNormalOrVisual(
					DeleteToNextMatchingChar(char, p.CountOrDefault(), p.ClipboardPageOrDefault(), true),
					addToMacro{lastAction: true, user: true})
			},
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
				return decorateNormalOrVisual(
					DeleteToPrevMatchingChar(char, p.CountOrDefault(), p.ClipboardPageOrDefault(), true),
					addToMacro{lastAction: true, user: true})
			},
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
				return decorateNormalOrVisual(
					DeleteToNextMatchingChar(char, p.CountOrDefault(), p.ClipboardPageOrDefault(), false),
					addToMacro{lastAction: true, user: true})
			},
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
				return decorateNormalOrVisual(
					DeleteToPrevMatchingChar(char, p.CountOrDefault(), p.ClipboardPageOrDefault(), false),
					addToMacro{lastAction: true, user: true})
			},
		},
		{
			Name: "delete to start of next word (dw)",
			Pattern: []EventMatcher{
				{Key: tcell.KeyRune, Rune: 'd'},
				{Key: tcell.KeyRune, Rune: 'w'},
			},
			ActionBuilder: func(p ActionBuilderParams) Action {
				return decorateNormalOrVisual(
					DeleteToStartOfNextWord(p.ClipboardPageOrDefault()),
					addToMacro{lastAction: true, user: true})
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
				return decorateNormalOrVisual(
					DeleteAWord(p.ClipboardPageOrDefault()),
					addToMacro{lastAction: true, user: true})
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
				return decorateNormalOrVisual(
					DeleteInnerWord(p.ClipboardPageOrDefault()),
					addToMacro{lastAction: true, user: true})
			},
		},
		{
			Name: "change to start of next word (cw)",
			Pattern: []EventMatcher{
				{Key: tcell.KeyRune, Rune: 'c'},
				{Key: tcell.KeyRune, Rune: 'w'},
			},
			ActionBuilder: func(p ActionBuilderParams) Action {
				return decorateNormalOrVisual(
					ChangeToStartOfNextWord(p.ClipboardPageOrDefault()),
					addToMacro{lastAction: true, user: true})
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
				return decorateNormalOrVisual(
					ChangeAWord(p.ClipboardPageOrDefault()),
					addToMacro{lastAction: true, user: true})
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
				return decorateNormalOrVisual(
					ChangeInnerWord(p.ClipboardPageOrDefault()),
					addToMacro{lastAction: true, user: true})
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
				char := p.LastChar()
				if char == '\x00' {
					return EmptyAction
				}
				return decorateNormalOrVisual(
					ChangeToNextMatchingChar(char, p.CountOrDefault(), p.ClipboardPageOrDefault(), true),
					addToMacro{lastAction: true, user: true})
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
				char := p.LastChar()
				if char == '\x00' {
					return EmptyAction
				}
				return decorateNormalOrVisual(
					ChangeToPrevMatchingChar(char, p.CountOrDefault(), p.ClipboardPageOrDefault(), true),
					addToMacro{lastAction: true, user: true})
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
				char := p.LastChar()
				if char == '\x00' {
					return EmptyAction
				}
				return decorateNormalOrVisual(
					ChangeToNextMatchingChar(char, p.CountOrDefault(), p.ClipboardPageOrDefault(), false),
					addToMacro{lastAction: true, user: true})
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
				char := p.LastChar()
				if char == '\x00' {
					return EmptyAction
				}
				return decorateNormalOrVisual(
					ChangeToPrevMatchingChar(char, p.CountOrDefault(), p.ClipboardPageOrDefault(), false),
					addToMacro{lastAction: true, user: true})
			},
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

				return decorateNormalOrVisual(
					ReplaceCharacter(newChar),
					addToMacro{lastAction: true, user: true})
			},
		},
		{
			Name: "toggle case (~)",
			Pattern: []EventMatcher{
				{Key: tcell.KeyRune, Rune: '~'},
			},
			ActionBuilder: func(p ActionBuilderParams) Action {
				return decorateNormalOrVisual(
					ToggleCaseAtCursor,
					addToMacro{lastAction: true, user: true})
			},
		},
		{
			Name: "indent (>>)",
			Pattern: []EventMatcher{
				{Key: tcell.KeyRune, Rune: '>'},
				{Key: tcell.KeyRune, Rune: '>'},
			},
			ActionBuilder: func(p ActionBuilderParams) Action {
				return decorateNormalOrVisual(
					IndentLine,
					addToMacro{lastAction: true, user: true})
			},
		},
		{
			Name: "outdent (<<)",
			Pattern: []EventMatcher{
				{Key: tcell.KeyRune, Rune: '<'},
				{Key: tcell.KeyRune, Rune: '<'},
			},
			ActionBuilder: func(p ActionBuilderParams) Action {
				return decorateNormalOrVisual(
					OutdentLine,
					addToMacro{lastAction: true, user: true})
			},
		},
		{
			Name: "yank to start of next word (yw)",
			Pattern: []EventMatcher{
				{Key: tcell.KeyRune, Rune: 'y'},
				{Key: tcell.KeyRune, Rune: 'w'},
			},
			ActionBuilder: func(p ActionBuilderParams) Action {
				return decorateNormalOrVisual(
					CopyToStartOfNextWord(p.ClipboardPageOrDefault()),
					addToMacro{lastAction: true, user: true})
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
				return decorateNormalOrVisual(
					CopyAWord(p.ClipboardPageOrDefault()),
					addToMacro{lastAction: true, user: true})
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
				return decorateNormalOrVisual(
					CopyInnerWord(p.ClipboardPageOrDefault()),
					addToMacro{lastAction: true, user: true})
			},
		},
		{
			Name: "yank line (yy)",
			Pattern: []EventMatcher{
				{Key: tcell.KeyRune, Rune: 'y'},
				{Key: tcell.KeyRune, Rune: 'y'},
			},
			ActionBuilder: func(p ActionBuilderParams) Action {
				return decorateNormalOrVisual(
					CopyLines(p.ClipboardPageOrDefault()),
					addToMacro{lastAction: true, user: true})
			},
		},
		{
			Name: "put after cursor (p)",
			Pattern: []EventMatcher{
				{Key: tcell.KeyRune, Rune: 'p'},
			},
			ActionBuilder: func(p ActionBuilderParams) Action {
				return decorateNormalOrVisual(
					PasteAfterCursor(p.ClipboardPageOrDefault()),
					addToMacro{lastAction: true, user: true})
			},
		},
		{
			Name: "put before cursor (P)",
			Pattern: []EventMatcher{
				{Key: tcell.KeyRune, Rune: 'P'},
			},
			ActionBuilder: func(p ActionBuilderParams) Action {
				return decorateNormalOrVisual(
					PasteBeforeCursor(p.ClipboardPageOrDefault()),
					addToMacro{lastAction: true, user: true})
			},
		},
		{
			Name: "show command menu",
			Pattern: []EventMatcher{
				{Key: tcell.KeyRune, Rune: ':'},
			},
			ActionBuilder: func(p ActionBuilderParams) Action {
				return decorateNormalOrVisual(
					ShowCommandMenu(p.Config),
					addToMacro{})
			},
		},
		{
			Name: "start forward search",
			Pattern: []EventMatcher{
				{Key: tcell.KeyRune, Rune: '/'},
			},
			ActionBuilder: func(p ActionBuilderParams) Action {
				return decorateNormalOrVisual(
					StartSearchForward,
					addToMacro{user: true})
			},
		},
		{
			Name: "start backward search",
			Pattern: []EventMatcher{
				{Key: tcell.KeyRune, Rune: '?'},
			},
			ActionBuilder: func(p ActionBuilderParams) Action {
				return decorateNormalOrVisual(
					StartSearchBackward,
					addToMacro{user: true})
			},
		},
		{
			Name: "find next match",
			Pattern: []EventMatcher{
				{Key: tcell.KeyRune, Rune: 'n'},
			},
			ActionBuilder: func(p ActionBuilderParams) Action {
				return decorateNormalOrVisual(
					FindNextMatch,
					addToMacro{user: true})
			},
		},
		{
			Name: "find previous match",
			Pattern: []EventMatcher{
				{Key: tcell.KeyRune, Rune: 'N'},
			},
			ActionBuilder: func(p ActionBuilderParams) Action {
				return decorateNormalOrVisual(
					FindPrevMatch,
					addToMacro{user: true})
			},
		},
		{
			Name: "undo (u)",
			Pattern: []EventMatcher{
				{Key: tcell.KeyRune, Rune: 'u'},
			},
			ActionBuilder: func(p ActionBuilderParams) Action {
				return decorateNormalOrVisual(
					Undo,
					addToMacro{user: true})
			},
		},
		{
			Name: "redo (ctrl-r)",
			Pattern: []EventMatcher{
				{Key: tcell.KeyCtrlR},
			},
			ActionBuilder: func(p ActionBuilderParams) Action {
				return decorateNormalOrVisual(
					Redo,
					addToMacro{user: true})
			},
		},
		{
			Name: "enter visual mode charwise (v)",
			Pattern: []EventMatcher{
				{Key: tcell.KeyRune, Rune: 'v'},
			},
			ActionBuilder: func(p ActionBuilderParams) Action {
				return decorateNormalOrVisual(
					ToggleVisualModeCharwise,
					addToMacro{user: true})
			},
		},
		{
			Name: "enter visual mode linewise (V)",
			Pattern: []EventMatcher{
				{Key: tcell.KeyRune, Rune: 'V'},
			},
			ActionBuilder: func(p ActionBuilderParams) Action {
				return decorateNormalOrVisual(
					ToggleVisualModeLinewise,
					addToMacro{user: true})
			},
		},
		{
			Name: "repeat last action (.)",
			Pattern: []EventMatcher{
				{Key: tcell.KeyRune, Rune: '.'},
			},
			ActionBuilder: func(p ActionBuilderParams) Action {
				return decorateNormalOrVisual(
					ReplayLastActionMacro(p.CountOrDefault()),
					addToMacro{user: true})
			},
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
				return decorateNormalOrVisual(
					ToggleVisualModeCharwise,
					addToMacro{user: true})
			},
		},
		{
			Name: "toggle visual mode linewise (V)",
			Pattern: []EventMatcher{
				{Key: tcell.KeyRune, Rune: 'V'},
			},
			ActionBuilder: func(p ActionBuilderParams) Action {
				return decorateNormalOrVisual(
					ToggleVisualModeLinewise,
					addToMacro{user: true})
			},
		},
		{
			Name: "return to normal mode (esc)",
			Pattern: []EventMatcher{
				{Key: tcell.KeyEscape},
			},
			ActionBuilder: func(p ActionBuilderParams) Action {
				return decorateNormalOrVisual(
					ReturnToNormalMode,
					addToMacro{user: true})
			},
		},
		{
			Name: "show command menu",
			Pattern: []EventMatcher{
				{Key: tcell.KeyRune, Rune: ':'},
			},
			ActionBuilder: func(p ActionBuilderParams) Action {
				return decorateNormalOrVisual(
					ShowCommandMenu(p.Config),
					addToMacro{})
			},
		},
		{
			Name: "delete selection (x)",
			Pattern: []EventMatcher{
				{Key: tcell.KeyRune, Rune: 'x'},
			},
			ActionBuilder: func(p ActionBuilderParams) Action {
				return decorateNormalOrVisual(
					DeleteSelectionAndReturnToNormalMode(
						p.ClipboardPageOrDefault(),
						p.Config.SelectionMode,
						p.Config.SelectionEndLocator,
					), addToMacro{lastAction: true, user: true})
			},
		},
		{
			Name: "delete selection (d)",
			Pattern: []EventMatcher{
				{Key: tcell.KeyRune, Rune: 'd'},
			},
			ActionBuilder: func(p ActionBuilderParams) Action {
				return decorateNormalOrVisual(
					DeleteSelectionAndReturnToNormalMode(
						p.ClipboardPageOrDefault(),
						p.Config.SelectionMode,
						p.Config.SelectionEndLocator,
					), addToMacro{lastAction: true, user: true})
			},
		},
		{
			Name: "change selection (c)",
			Pattern: []EventMatcher{
				{Key: tcell.KeyRune, Rune: 'c'},
			},
			ActionBuilder: func(p ActionBuilderParams) Action {
				return decorateNormalOrVisual(
					ChangeSelection(
						p.ClipboardPageOrDefault(),
						p.Config.SelectionMode,
						p.Config.SelectionEndLocator,
					), addToMacro{lastAction: true, user: true})
			},
		},
		{
			Name: "toggle case for selection (~)",
			Pattern: []EventMatcher{
				{Key: tcell.KeyRune, Rune: '~'},
			},
			ActionBuilder: func(p ActionBuilderParams) Action {
				return decorateNormalOrVisual(
					ToggleCaseInSelectionAndReturnToNormalMode(p.Config.SelectionEndLocator),
					addToMacro{lastAction: true, user: true})
			},
		},
		{
			Name: "indent selection (>)",
			Pattern: []EventMatcher{
				{Key: tcell.KeyRune, Rune: '>'},
			},
			ActionBuilder: func(p ActionBuilderParams) Action {
				return decorateNormalOrVisual(
					IndentSelectionAndReturnToNormalMode(p.Config.SelectionEndLocator),
					addToMacro{lastAction: true, user: true})
			},
		},
		{
			Name: "outdent selection (<)",
			Pattern: []EventMatcher{
				{Key: tcell.KeyRune, Rune: '<'},
			},
			ActionBuilder: func(p ActionBuilderParams) Action {
				return decorateNormalOrVisual(
					OutdentSelectionAndReturnToNormalMode(p.Config.SelectionEndLocator),
					addToMacro{lastAction: true, user: true})
			},
		},
		{
			Name: "yank selection (y)",
			Pattern: []EventMatcher{
				{Key: tcell.KeyRune, Rune: 'y'},
			},
			ActionBuilder: func(p ActionBuilderParams) Action {
				return decorateNormalOrVisual(
					CopySelectionAndReturnToNormalMode(p.ClipboardPageOrDefault()),
					addToMacro{user: true})
			},
		},
	}...)
}
