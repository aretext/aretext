package input

import (
	"github.com/gdamore/tcell/v2"

	"github.com/aretext/aretext/clipboard"
	"github.com/aretext/aretext/input/vm"
	"github.com/aretext/aretext/state"
)

// CommandParams are parameters parsed from user input.
type CommandParams struct {
	Count         uint64
	ClipboardPage clipboard.PageId
	MatchChar     rune
	ReplaceChar   rune
	InsertChar    rune
}

// Command defines a command that the input parser can recognize.
// The VM expression defines how the input processor recognizes the command,
// and the action defines how the editor executes the command.
type Command struct {
	Name        string
	BuildExpr   func() vm.Expr
	BuildAction func(Context, CommandParams) Action
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
			Name: "cursor left (left arrow or h)",
			BuildExpr: func() vm.Expr {
				return altExpr(keyExpr(tcell.KeyLeft), runeExpr('h'))
			},
			BuildAction: func(ctx Context, p CommandParams) Action {
				return decorate(CursorLeft)
			},
		},
		{
			Name: "cursor right (right arrow or l)",
			BuildExpr: func() vm.Expr {
				return altExpr(keyExpr(tcell.KeyRight), runeExpr('l'))
			},
			BuildAction: func(ctx Context, p CommandParams) Action {
				return decorate(CursorRight)
			},
		},
		{
			Name: "cursor up (up arrow or k)",
			BuildExpr: func() vm.Expr {
				return altExpr(keyExpr(tcell.KeyUp), runeExpr('k'))
			},
			BuildAction: func(ctx Context, p CommandParams) Action {
				return decorate(CursorUp)
			},
		},
		{
			Name: "cursor down (down arrow or j)",
			BuildExpr: func() vm.Expr {
				return altExpr(keyExpr(tcell.KeyDown), runeExpr('j'))
			},
			BuildAction: func(ctx Context, p CommandParams) Action {
				return decorate(CursorDown)
			},
		},
		{
			Name: "cursor back (backspace)",
			BuildExpr: func() vm.Expr {
				return altExpr(keyExpr(tcell.KeyBackspace), keyExpr(tcell.KeyBackspace2))
			},
			BuildAction: func(ctx Context, p CommandParams) Action {
				return decorate(CursorBack)
			},
		},
		{
			Name: "cursor next word start (w)",
			BuildExpr: func() vm.Expr {
				return cmdExpr("w", "", captureOpts{})
			},
			BuildAction: func(ctx Context, p CommandParams) Action {
				return decorate(CursorNextWordStart)
			},
		},
		{
			Name: "cursor prev word start (b)",
			BuildExpr: func() vm.Expr {
				return cmdExpr("b", "", captureOpts{})
			},
			BuildAction: func(ctx Context, p CommandParams) Action {
				return decorate(CursorPrevWordStart)
			},
		},
		{
			Name: "cursor next word end (e)",
			BuildExpr: func() vm.Expr {
				return cmdExpr("e", "", captureOpts{})
			},
			BuildAction: func(ctx Context, p CommandParams) Action {
				return decorate(CursorNextWordEnd)
			},
		},
		{
			Name: "cursor prev paragraph ({)",
			BuildExpr: func() vm.Expr {
				return cmdExpr("{", "", captureOpts{})
			},
			BuildAction: func(ctx Context, p CommandParams) Action {
				return decorate(CursorPrevParagraph)
			},
		},
		{
			Name: "cursor next paragraph (})",
			BuildExpr: func() vm.Expr {
				return cmdExpr("}", "", captureOpts{})
			},
			BuildAction: func(ctx Context, p CommandParams) Action {
				return decorate(CursorNextParagraph)
			},
		},
		{
			Name: "cursor to next matching char (f{char})",
			BuildExpr: func() vm.Expr {
				return cmdExpr("f", "", captureOpts{count: true, matchChar: true})
			},
			BuildAction: func(ctx Context, p CommandParams) Action {
				return decorate(CursorToNextMatchingChar(p.MatchChar, p.Count, true))
			},
		},
		{
			Name: "cursor to prev matching char (F{char})",
			BuildExpr: func() vm.Expr {
				return cmdExpr("F", "", captureOpts{count: true, matchChar: true})
			},
			BuildAction: func(ctx Context, p CommandParams) Action {
				return decorate(CursorToPrevMatchingChar(p.MatchChar, p.Count, true))
			},
		},
		{
			Name: "cursor till next matching char (t{char})",
			BuildExpr: func() vm.Expr {
				return cmdExpr("t", "", captureOpts{count: true, matchChar: true})
			},
			BuildAction: func(ctx Context, p CommandParams) Action {
				return decorate(CursorToNextMatchingChar(p.MatchChar, p.Count, false))
			},
		},
		{
			Name: "cursor to prev matching char (T{char})",
			BuildExpr: func() vm.Expr {
				return cmdExpr("T", "", captureOpts{count: true, matchChar: true})
			},
			BuildAction: func(ctx Context, p CommandParams) Action {
				return decorate(CursorToPrevMatchingChar(p.MatchChar, p.Count, false))
			},
		},
		{
			Name: "cursor line start (0)",
			BuildExpr: func() vm.Expr {
				return cmdExpr("0", "", captureOpts{})
			},
			BuildAction: func(ctx Context, p CommandParams) Action {
				return decorate(CursorLineStart)
			},
		},
		{
			Name: "cursor line start non-whitespace (^)",
			BuildExpr: func() vm.Expr {
				return cmdExpr("^", "", captureOpts{})
			},
			BuildAction: func(ctx Context, p CommandParams) Action {
				return decorate(CursorLineStartNonWhitespace)
			},
		},
		{
			Name: "cursor line end ($)",
			BuildExpr: func() vm.Expr {
				return cmdExpr("$", "", captureOpts{})
			},
			BuildAction: func(ctx Context, p CommandParams) Action {
				return decorate(CursorLineEnd)
			},
		},
		{
			Name: "cursor start of line num (gg)",
			BuildExpr: func() vm.Expr {
				return cmdExpr("gg", "", captureOpts{count: true})
			},
			BuildAction: func(ctx Context, p CommandParams) Action {
				return decorate(CursorStartOfLineNum(p.Count))
			},
		},
		{
			Name: "cursor start of last line (G)",
			BuildExpr: func() vm.Expr {
				return cmdExpr("G", "", captureOpts{})
			},
			BuildAction: func(ctx Context, p CommandParams) Action {
				return decorate(CursorStartOfLastLine)
			},
		},
		{
			Name: "scroll up (ctrl-u)",
			BuildExpr: func() vm.Expr {
				return keyExpr(tcell.KeyCtrlU)
			},
			BuildAction: func(ctx Context, p CommandParams) Action {
				return decorate(ScrollUp(ctx))
			},
		},
		{
			Name: "scroll down (ctrl-d)",
			BuildExpr: func() vm.Expr {
				return keyExpr(tcell.KeyCtrlD)
			},
			BuildAction: func(ctx Context, p CommandParams) Action {
				return decorate(ScrollDown(ctx))
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

func NormalModeCommands() []Command {
	return append(cursorCommands(), []Command{
		{
			Name: "enter insert mode (i)",
			BuildExpr: func() vm.Expr {
				return cmdExpr("i", "", captureOpts{})
			},
			BuildAction: func(ctx Context, p CommandParams) Action {
				return decorateNormalOrVisual(
					EnterInsertMode,
					addToMacro{lastAction: true, user: true})
			},
		},
		{
			Name: "enter insert mode at start of line (I)",
			BuildExpr: func() vm.Expr {
				return cmdExpr("I", "", captureOpts{})
			},
			BuildAction: func(ctx Context, p CommandParams) Action {
				return decorateNormalOrVisual(
					EnterInsertModeAtStartOfLine,
					addToMacro{lastAction: true, user: true})
			},
		},
		{
			Name: "enter insert mode at next pos (a)",
			BuildExpr: func() vm.Expr {
				return cmdExpr("a", "", captureOpts{})
			},
			BuildAction: func(ctx Context, p CommandParams) Action {
				return decorateNormalOrVisual(
					EnterInsertModeAtNextPos,
					addToMacro{lastAction: true, user: true})
			},
		},
		{
			Name: "enter insert mode at end of line (A)",
			BuildExpr: func() vm.Expr {
				return cmdExpr("A", "", captureOpts{})
			},
			BuildAction: func(ctx Context, p CommandParams) Action {
				return decorateNormalOrVisual(
					EnterInsertModeAtEndOfLine,
					addToMacro{lastAction: true, user: true})
			},
		},
		{
			Name: "begin new line below (o)",
			BuildExpr: func() vm.Expr {
				return cmdExpr("o", "", captureOpts{})
			},
			BuildAction: func(ctx Context, p CommandParams) Action {
				return decorateNormalOrVisual(
					BeginNewLineBelow,
					addToMacro{lastAction: true, user: true})
			},
		},
		{
			Name: "begin new line above (O)",
			BuildExpr: func() vm.Expr {
				return cmdExpr("O", "", captureOpts{})
			},
			BuildAction: func(ctx Context, p CommandParams) Action {
				return decorateNormalOrVisual(
					BeginNewLineAbove,
					addToMacro{lastAction: true, user: true})
			},
		},
		{
			Name: "join lines (J)",
			BuildExpr: func() vm.Expr {
				return cmdExpr("J", "", captureOpts{})
			},
			BuildAction: func(ctx Context, p CommandParams) Action {
				return decorateNormalOrVisual(
					JoinLines,
					addToMacro{lastAction: true, user: true})
			},
		},
		{
			Name: "delete line (dd)",
			BuildExpr: func() vm.Expr {
				return cmdExpr("dd", "", captureOpts{count: true, clipboardPage: true})
			},
			BuildAction: func(ctx Context, p CommandParams) Action {
				return decorateNormalOrVisual(
					DeleteLines(p.Count, p.ClipboardPage),
					addToMacro{lastAction: true, user: true})
			},
		},
		{
			Name: "delete prev char in line (dh)",
			BuildExpr: func() vm.Expr {
				return cmdExpr("d", "h", captureOpts{clipboardPage: true})
			},
			BuildAction: func(ctx Context, p CommandParams) Action {
				return decorateNormalOrVisual(
					DeletePrevCharInLine(p.ClipboardPage),
					addToMacro{lastAction: true, user: true})
			},
		},
		{
			Name: "delete down (dj)",
			BuildExpr: func() vm.Expr {
				return cmdExpr("d", "j", captureOpts{clipboardPage: true})
			},
			BuildAction: func(ctx Context, p CommandParams) Action {
				return decorateNormalOrVisual(
					DeleteDown(p.ClipboardPage),
					addToMacro{lastAction: true, user: true})
			},
		},
		{
			Name: "delete up (dk)",
			BuildExpr: func() vm.Expr {
				return cmdExpr("d", "k", captureOpts{clipboardPage: true})
			},
			BuildAction: func(ctx Context, p CommandParams) Action {
				return decorateNormalOrVisual(
					DeleteUp(p.ClipboardPage),
					addToMacro{lastAction: true, user: true})
			},
		},
		{
			Name: "delete next char in line (dl or x)",
			BuildExpr: func() vm.Expr {
				return altExpr(
					cmdExpr("d", "l", captureOpts{count: true, clipboardPage: true}),
					cmdExpr("x", "", captureOpts{count: true, clipboardPage: true}),
				)
			},
			BuildAction: func(ctx Context, p CommandParams) Action {
				return decorateNormalOrVisual(
					DeleteNextCharInLine(p.Count, p.ClipboardPage),
					addToMacro{lastAction: true, user: true})
			},
		},
		{
			Name: "delete to end of line (d$)",
			BuildExpr: func() vm.Expr {
				return cmdExpr("d", "$", captureOpts{clipboardPage: true})
			},
			BuildAction: func(ctx Context, p CommandParams) Action {
				return decorateNormalOrVisual(
					DeleteToEndOfLine(p.ClipboardPage),
					addToMacro{lastAction: true, user: true})
			},
		},
		{
			Name: "delete to start of line (d0)",
			BuildExpr: func() vm.Expr {
				return cmdExpr("d", "0", captureOpts{clipboardPage: true})
			},
			BuildAction: func(ctx Context, p CommandParams) Action {
				return decorateNormalOrVisual(
					DeleteToStartOfLine(p.ClipboardPage),
					addToMacro{lastAction: true, user: true})
			},
		},
		{
			Name: "delete to start of line non-whitespace (d^)",
			BuildExpr: func() vm.Expr {
				return cmdExpr("d", "^", captureOpts{clipboardPage: true})
			},
			BuildAction: func(ctx Context, p CommandParams) Action {
				return decorateNormalOrVisual(
					DeleteToStartOfLineNonWhitespace(p.ClipboardPage),
					addToMacro{lastAction: true, user: true})
			},
		},
		{
			Name: "delete to end of line (D)",
			BuildExpr: func() vm.Expr {
				return cmdExpr("D", "", captureOpts{clipboardPage: true})
			},
			BuildAction: func(ctx Context, p CommandParams) Action {
				return decorateNormalOrVisual(
					DeleteToEndOfLine(p.ClipboardPage),
					addToMacro{lastAction: true, user: true})
			},
		},
		{
			Name: "delete to next matching char (df{char})",
			BuildExpr: func() vm.Expr {
				return cmdExpr("d", "f", captureOpts{count: true, clipboardPage: true, matchChar: true})
			},
			BuildAction: func(ctx Context, p CommandParams) Action {
				return decorateNormalOrVisual(
					DeleteToNextMatchingChar(p.MatchChar, p.Count, p.ClipboardPage, true),
					addToMacro{lastAction: true, user: true})
			},
		},
		{
			Name: "delete to prev matching char (dF{char})",
			BuildExpr: func() vm.Expr {
				return cmdExpr("d", "F", captureOpts{count: true, clipboardPage: true, matchChar: true})
			},
			BuildAction: func(ctx Context, p CommandParams) Action {
				return decorateNormalOrVisual(
					DeleteToPrevMatchingChar(p.MatchChar, p.Count, p.ClipboardPage, true),
					addToMacro{lastAction: true, user: true})
			},
		},
		{
			Name: "delete till next matching char (dt{char})",
			BuildExpr: func() vm.Expr {
				return cmdExpr("d", "t", captureOpts{count: true, clipboardPage: true, matchChar: true})
			},
			BuildAction: func(ctx Context, p CommandParams) Action {
				return decorateNormalOrVisual(
					DeleteToNextMatchingChar(p.MatchChar, p.Count, p.ClipboardPage, false),
					addToMacro{lastAction: true, user: true})
			},
		},
		{
			Name: "delete till prev matching char (dT{char})",
			BuildExpr: func() vm.Expr {
				return cmdExpr("d", "T", captureOpts{count: true, clipboardPage: true, matchChar: true})
			},
			BuildAction: func(ctx Context, p CommandParams) Action {
				return decorateNormalOrVisual(
					DeleteToPrevMatchingChar(p.MatchChar, p.Count, p.ClipboardPage, false),
					addToMacro{lastAction: true, user: true})
			},
		},
		{
			Name: "delete to start of next word (dw)",
			BuildExpr: func() vm.Expr {
				return cmdExpr("d", "w", captureOpts{clipboardPage: true})
			},
			BuildAction: func(ctx Context, p CommandParams) Action {
				return decorateNormalOrVisual(
					DeleteToStartOfNextWord(p.ClipboardPage),
					addToMacro{lastAction: true, user: true})
			},
		},
		{
			Name: "delete a word (daw)",
			BuildExpr: func() vm.Expr {
				return cmdExpr("d", "aw", captureOpts{clipboardPage: true})
			},
			BuildAction: func(ctx Context, p CommandParams) Action {
				return decorateNormalOrVisual(
					DeleteAWord(p.ClipboardPage),
					addToMacro{lastAction: true, user: true})
			},
		},
		{
			Name: "delete inner word (diw)",
			BuildExpr: func() vm.Expr {
				return cmdExpr("d", "iw", captureOpts{clipboardPage: true})
			},
			BuildAction: func(ctx Context, p CommandParams) Action {
				return decorateNormalOrVisual(
					DeleteInnerWord(p.ClipboardPage),
					addToMacro{lastAction: true, user: true})
			},
		},
		{
			Name: "change a word (caw)",
			BuildExpr: func() vm.Expr {
				return cmdExpr("c", "aw", captureOpts{clipboardPage: true})
			},
			BuildAction: func(ctx Context, p CommandParams) Action {
				return decorateNormalOrVisual(
					ChangeAWord(p.ClipboardPage),
					addToMacro{lastAction: true, user: true})
			},
		},
		{
			Name: "change word and change inner word (cw and ciw)",
			BuildExpr: func() vm.Expr {
				// Unlike "dw", "cw" excludes whitespace after the word by default,
				// so we alias it to "ciw" (change inner word).
				// See https://vimhelp.org/change.txt.html
				return altExpr(
					cmdExpr("c", "w", captureOpts{clipboardPage: true}),
					cmdExpr("c", "iw", captureOpts{clipboardPage: true}),
				)
			},
			BuildAction: func(ctx Context, p CommandParams) Action {
				return decorateNormalOrVisual(
					ChangeInnerWord(p.ClipboardPage),
					addToMacro{lastAction: true, user: true})
			},
		},
		{
			Name: "change to next matching char (cf{char})",
			BuildExpr: func() vm.Expr {
				return cmdExpr("c", "f", captureOpts{count: true, clipboardPage: true, matchChar: true})
			},
			BuildAction: func(ctx Context, p CommandParams) Action {
				return decorateNormalOrVisual(
					ChangeToNextMatchingChar(p.MatchChar, p.Count, p.ClipboardPage, true),
					addToMacro{lastAction: true, user: true})
			},
		},
		{
			Name: "change to prev matching char (cF{char})",
			BuildExpr: func() vm.Expr {
				return cmdExpr("c", "F", captureOpts{count: true, clipboardPage: true, matchChar: true})
			},
			BuildAction: func(ctx Context, p CommandParams) Action {
				return decorateNormalOrVisual(
					ChangeToPrevMatchingChar(p.MatchChar, p.Count, p.ClipboardPage, true),
					addToMacro{lastAction: true, user: true})
			},
		},
		{
			Name: "change till next matching char (ct{char})",
			BuildExpr: func() vm.Expr {
				return cmdExpr("c", "t", captureOpts{count: true, clipboardPage: true, matchChar: true})
			},
			BuildAction: func(ctx Context, p CommandParams) Action {
				return decorateNormalOrVisual(
					ChangeToNextMatchingChar(p.MatchChar, p.Count, p.ClipboardPage, false),
					addToMacro{lastAction: true, user: true})
			},
		},
		{
			Name: "change till prev matching char (cT{char})",
			BuildExpr: func() vm.Expr {
				return cmdExpr("c", "T", captureOpts{count: true, clipboardPage: true, matchChar: true})
			},
			BuildAction: func(ctx Context, p CommandParams) Action {
				return decorateNormalOrVisual(
					ChangeToPrevMatchingChar(p.MatchChar, p.Count, p.ClipboardPage, false),
					addToMacro{lastAction: true, user: true})
			},
		},
		{
			Name: "replace character (r)",
			BuildExpr: func() vm.Expr {
				return cmdExpr("r", "", captureOpts{replaceChar: true})
			},
			BuildAction: func(ctx Context, p CommandParams) Action {
				return decorateNormalOrVisual(
					ReplaceCharacter(p.ReplaceChar),
					addToMacro{lastAction: true, user: true})
			},
		},
		{
			Name: "toggle case (~)",
			BuildExpr: func() vm.Expr {
				return cmdExpr("~", "", captureOpts{})
			},
			BuildAction: func(ctx Context, p CommandParams) Action {
				return decorateNormalOrVisual(
					ToggleCaseAtCursor,
					addToMacro{lastAction: true, user: true})
			},
		},
		{
			Name: "indent (>>)",
			BuildExpr: func() vm.Expr {
				return cmdExpr(">>", "", captureOpts{})
			},
			BuildAction: func(ctx Context, p CommandParams) Action {
				return decorateNormalOrVisual(
					IndentLine,
					addToMacro{lastAction: true, user: true})
			},
		},
		{
			Name: "outdent (<<)",
			BuildExpr: func() vm.Expr {
				return cmdExpr("<<", "", captureOpts{})
			},
			BuildAction: func(ctx Context, p CommandParams) Action {
				return decorateNormalOrVisual(
					OutdentLine,
					addToMacro{lastAction: true, user: true})
			},
		},
		{
			Name: "yank to start of next word (yw)",
			BuildExpr: func() vm.Expr {
				return cmdExpr("y", "w", captureOpts{clipboardPage: true})
			},
			BuildAction: func(ctx Context, p CommandParams) Action {
				return decorateNormalOrVisual(
					CopyToStartOfNextWord(p.ClipboardPage),
					addToMacro{lastAction: true, user: true})
			},
		},
		{
			Name: "yank a word (yaw)",
			BuildExpr: func() vm.Expr {
				return cmdExpr("y", "aw", captureOpts{clipboardPage: true})
			},
			BuildAction: func(ctx Context, p CommandParams) Action {
				return decorateNormalOrVisual(
					CopyAWord(p.ClipboardPage),
					addToMacro{lastAction: true, user: true})
			},
		},
		{
			Name: "yank inner word (yiw)",
			BuildExpr: func() vm.Expr {
				return cmdExpr("y", "iw", captureOpts{clipboardPage: true})
			},
			BuildAction: func(ctx Context, p CommandParams) Action {
				return decorateNormalOrVisual(
					CopyInnerWord(p.ClipboardPage),
					addToMacro{lastAction: true, user: true})
			},
		},
		{
			Name: "yank line (yy)",
			BuildExpr: func() vm.Expr {
				return cmdExpr("yy", "", captureOpts{clipboardPage: true})
			},
			BuildAction: func(ctx Context, p CommandParams) Action {
				return decorateNormalOrVisual(
					CopyLines(p.ClipboardPage),
					addToMacro{lastAction: true, user: true})
			},
		},
		{
			Name: "put after cursor (p)",
			BuildExpr: func() vm.Expr {
				return cmdExpr("p", "", captureOpts{clipboardPage: true})
			},
			BuildAction: func(ctx Context, p CommandParams) Action {
				return decorateNormalOrVisual(
					PasteAfterCursor(p.ClipboardPage),
					addToMacro{lastAction: true, user: true})
			},
		},
		{
			Name: "put before cursor (P)",
			BuildExpr: func() vm.Expr {
				return cmdExpr("P", "", captureOpts{clipboardPage: true})
			},
			BuildAction: func(ctx Context, p CommandParams) Action {
				return decorateNormalOrVisual(
					PasteBeforeCursor(p.ClipboardPage),
					addToMacro{lastAction: true, user: true})
			},
		},
		{
			Name: "show command menu",
			BuildExpr: func() vm.Expr {
				return runeExpr(':')
			},
			BuildAction: func(ctx Context, p CommandParams) Action {
				return decorateNormalOrVisual(
					ShowCommandMenu(ctx),
					addToMacro{})
			},
		},
		{
			Name: "start forward search",
			BuildExpr: func() vm.Expr {
				return runeExpr('/')
			},
			BuildAction: func(ctx Context, p CommandParams) Action {
				return decorateNormalOrVisual(
					StartSearchForward,
					addToMacro{user: true})
			},
		},
		{
			Name: "start backward search",
			BuildExpr: func() vm.Expr {
				return runeExpr('?')
			},
			BuildAction: func(ctx Context, p CommandParams) Action {
				return decorateNormalOrVisual(
					StartSearchBackward,
					addToMacro{user: true})
			},
		},
		{
			Name: "find next match",
			BuildExpr: func() vm.Expr {
				return runeExpr('n')
			},
			BuildAction: func(ctx Context, p CommandParams) Action {
				return decorateNormalOrVisual(
					FindNextMatch,
					addToMacro{user: true})
			},
		},
		{
			Name: "find previous match",
			BuildExpr: func() vm.Expr {
				return runeExpr('N')
			},
			BuildAction: func(ctx Context, p CommandParams) Action {
				return decorateNormalOrVisual(
					FindPrevMatch,
					addToMacro{user: true})
			},
		},
		{
			Name: "undo (u)",
			BuildExpr: func() vm.Expr {
				return runeExpr('u')
			},
			BuildAction: func(ctx Context, p CommandParams) Action {
				return decorateNormalOrVisual(
					Undo,
					addToMacro{user: true})
			},
		},
		{
			Name: "redo (ctrl-r)",
			BuildExpr: func() vm.Expr {
				return keyExpr(tcell.KeyCtrlR)
			},
			BuildAction: func(ctx Context, p CommandParams) Action {
				return decorateNormalOrVisual(
					Redo,
					addToMacro{user: true})
			},
		},
		{
			Name: "enter visual mode charwise (v)",
			BuildExpr: func() vm.Expr {
				return runeExpr('v')
			},
			BuildAction: func(ctx Context, p CommandParams) Action {
				return decorateNormalOrVisual(
					ToggleVisualModeCharwise,
					addToMacro{user: true})
			},
		},
		{
			Name: "enter visual mode linewise (V)",
			BuildExpr: func() vm.Expr {
				return runeExpr('V')
			},
			BuildAction: func(ctx Context, p CommandParams) Action {
				return decorateNormalOrVisual(
					ToggleVisualModeLinewise,
					addToMacro{user: true})
			},
		},
		{
			Name: "repeat last action (.)",
			BuildExpr: func() vm.Expr {
				return cmdExpr(".", "", captureOpts{count: true})
			},
			BuildAction: func(ctx Context, p CommandParams) Action {
				return decorateNormalOrVisual(
					ReplayLastActionMacro(p.Count),
					addToMacro{user: true})
			},
		},
	}...)
}

func VisualModeCommands() []Command {
	return append(cursorCommands(), []Command{
		{
			Name: "toggle visual mode charwise (v)",
			BuildExpr: func() vm.Expr {
				return runeExpr('v')
			},
			BuildAction: func(ctx Context, p CommandParams) Action {
				return decorateNormalOrVisual(
					ToggleVisualModeCharwise,
					addToMacro{user: true})
			},
		},
		{
			Name: "toggle visual mode linewise (V)",
			BuildExpr: func() vm.Expr {
				return runeExpr('V')
			},
			BuildAction: func(ctx Context, p CommandParams) Action {
				return decorateNormalOrVisual(
					ToggleVisualModeLinewise,
					addToMacro{user: true})
			},
		},
		{
			Name: "return to normal mode (esc)",
			BuildExpr: func() vm.Expr {
				return keyExpr(tcell.KeyEscape)
			},
			BuildAction: func(ctx Context, p CommandParams) Action {
				return decorateNormalOrVisual(
					ReturnToNormalMode,
					addToMacro{user: true})
			},
		},
		{
			Name: "show command menu",
			BuildExpr: func() vm.Expr {
				return runeExpr(':')
			},
			BuildAction: func(ctx Context, p CommandParams) Action {
				return decorateNormalOrVisual(
					ShowCommandMenu(ctx),
					addToMacro{})
			},
		},
		{
			Name: "delete selection (x or d)",
			BuildExpr: func() vm.Expr {
				return altExpr(
					cmdExpr("x", "", captureOpts{clipboardPage: true}),
					cmdExpr("d", "", captureOpts{clipboardPage: true}),
				)
			},
			BuildAction: func(ctx Context, p CommandParams) Action {
				return decorateNormalOrVisual(
					DeleteSelectionAndReturnToNormalMode(
						p.ClipboardPage,
						ctx.SelectionMode,
						ctx.SelectionEndLocator,
					), addToMacro{lastAction: true, user: true})
			},
		},
		{
			Name: "change selection (c)",
			BuildExpr: func() vm.Expr {
				return cmdExpr("c", "", captureOpts{clipboardPage: true})
			},
			BuildAction: func(ctx Context, p CommandParams) Action {
				return decorateNormalOrVisual(
					ChangeSelection(
						p.ClipboardPage,
						ctx.SelectionMode,
						ctx.SelectionEndLocator,
					), addToMacro{lastAction: true, user: true})
			},
		},
		{
			Name: "toggle case for selection (~)",
			BuildExpr: func() vm.Expr {
				return cmdExpr("~", "", captureOpts{})
			},
			BuildAction: func(ctx Context, p CommandParams) Action {
				return decorateNormalOrVisual(
					ToggleCaseInSelectionAndReturnToNormalMode(ctx.SelectionEndLocator),
					addToMacro{lastAction: true, user: true})
			},
		},
		{
			Name: "indent selection (>)",
			BuildExpr: func() vm.Expr {
				return cmdExpr(">", "", captureOpts{})
			},
			BuildAction: func(ctx Context, p CommandParams) Action {
				return decorateNormalOrVisual(
					IndentSelectionAndReturnToNormalMode(ctx.SelectionEndLocator),
					addToMacro{lastAction: true, user: true})
			},
		},
		{
			Name: "outdent selection (<)",
			BuildExpr: func() vm.Expr {
				return cmdExpr("<", "", captureOpts{})
			},
			BuildAction: func(ctx Context, p CommandParams) Action {
				return decorateNormalOrVisual(
					OutdentSelectionAndReturnToNormalMode(ctx.SelectionEndLocator),
					addToMacro{lastAction: true, user: true})
			},
		},
		{
			Name: "yank selection (y)",
			BuildExpr: func() vm.Expr {
				return cmdExpr("y", "", captureOpts{clipboardPage: true})
			},
			BuildAction: func(ctx Context, p CommandParams) Action {
				return decorateNormalOrVisual(
					CopySelectionAndReturnToNormalMode(p.ClipboardPage),
					addToMacro{user: true})
			},
		},
	}...)
}

func InsertModeCommands() []Command {
	decorate := func(action Action) Action {
		return func(s *state.EditorState) {
			wrappedAction := func(s *state.EditorState) {
				action(s)
				state.ScrollViewToCursor(s)
			}
			wrappedAction(s)
			state.AddToLastActionMacro(s, state.MacroAction(wrappedAction))
			state.AddToRecordingUserMacro(s, state.MacroAction(wrappedAction))
		}
	}

	return []Command{
		{
			Name: "insert rune",
			BuildExpr: func() vm.Expr {
				return insertExpr
			},
			BuildAction: func(ctx Context, p CommandParams) Action {
				return decorate(InsertRune(p.InsertChar))
			},
		},
		{
			Name: "delete prev char",
			BuildExpr: func() vm.Expr {
				return altExpr(keyExpr(tcell.KeyBackspace), keyExpr(tcell.KeyBackspace2))
			},
			BuildAction: func(ctx Context, p CommandParams) Action {
				return decorate(DeletePrevChar(clipboard.PageNull))
			},
		},
		{
			Name: "insert newline",
			BuildExpr: func() vm.Expr {
				return keyExpr(tcell.KeyEnter)
			},
			BuildAction: func(ctx Context, p CommandParams) Action {
				return decorate(InsertNewlineAndUpdateAutoIndentWhitespace)
			},
		},
		{
			Name: "insert tab",
			BuildExpr: func() vm.Expr {
				return keyExpr(tcell.KeyTab)
			},
			BuildAction: func(ctx Context, p CommandParams) Action {
				return decorate(InsertTab)
			},
		},
		{
			Name: "cursor left",
			BuildExpr: func() vm.Expr {
				return keyExpr(tcell.KeyLeft)
			},
			BuildAction: func(ctx Context, p CommandParams) Action {
				return decorate(CursorLeft)
			},
		},
		{
			Name: "cursor right",
			BuildExpr: func() vm.Expr {
				return keyExpr(tcell.KeyRight)
			},
			BuildAction: func(ctx Context, p CommandParams) Action {
				return decorate(CursorRight)
			},
		},
		{
			Name: "cursor up",
			BuildExpr: func() vm.Expr {
				return keyExpr(tcell.KeyUp)
			},
			BuildAction: func(ctx Context, p CommandParams) Action {
				return decorate(CursorUp)
			},
		},
		{
			Name: "cursor down",
			BuildExpr: func() vm.Expr {
				return keyExpr(tcell.KeyDown)
			},
			BuildAction: func(ctx Context, p CommandParams) Action {
				return decorate(CursorDown)
			},
		},
		{
			Name: "escape to normal mode",
			BuildExpr: func() vm.Expr {
				return keyExpr(tcell.KeyEscape)
			},
			BuildAction: func(ctx Context, p CommandParams) Action {
				return decorate(ReturnToNormalModeAfterInsert)
			},
		},
	}
}

func MenuModeCommands() []Command {
	return []Command{
		{
			Name: "escape to normal mode",
			BuildExpr: func() vm.Expr {
				return keyExpr(tcell.KeyEscape)
			},
			BuildAction: func(ctx Context, p CommandParams) Action {
				return HideMenuAndReturnToNormalMode
			},
		},
		{
			Name: "execute menu item",
			BuildExpr: func() vm.Expr {
				return keyExpr(tcell.KeyEnter)
			},
			BuildAction: func(ctx Context, p CommandParams) Action {
				return ExecuteSelectedMenuItem
			},
		},
		{
			Name: "move menu selection up",
			BuildExpr: func() vm.Expr {
				return keyExpr(tcell.KeyUp)
			},
			BuildAction: func(ctx Context, p CommandParams) Action {
				return MenuSelectionUp
			},
		},
		{
			Name: "move menu selection down",
			BuildExpr: func() vm.Expr {
				return altExpr(keyExpr(tcell.KeyDown), keyExpr(tcell.KeyTab))
			},
			BuildAction: func(ctx Context, p CommandParams) Action {
				return MenuSelectionDown
			},
		},
		{
			Name: "insert char to menu query",
			BuildExpr: func() vm.Expr {
				return insertExpr
			},
			BuildAction: func(ctx Context, p CommandParams) Action {
				return AppendRuneToMenuSearch(p.InsertChar)
			},
		},
		{
			Name: "delete char from menu query",
			BuildExpr: func() vm.Expr {
				return altExpr(keyExpr(tcell.KeyBackspace), keyExpr(tcell.KeyBackspace2))
			},
			BuildAction: func(ctx Context, p CommandParams) Action {
				return DeleteRuneFromMenuSearch
			},
		},
	}
}

func SearchModeCommands() []Command {
	return []Command{
		{
			Name: "escape to normal mode",
			BuildExpr: func() vm.Expr {
				return keyExpr(tcell.KeyEscape)
			},
			BuildAction: func(ctx Context, p CommandParams) Action {
				return AbortSearchAndReturnToNormalMode
			},
		},
		{
			Name: "commit search",
			BuildExpr: func() vm.Expr {
				return keyExpr(tcell.KeyEnter)
			},
			BuildAction: func(ctx Context, p CommandParams) Action {
				return CommitSearchAndReturnToNormalMode
			},
		},
		{
			Name: "insert char to search query",
			BuildExpr: func() vm.Expr {
				return insertExpr
			},
			BuildAction: func(ctx Context, p CommandParams) Action {
				return AppendRuneToSearchQuery(p.InsertChar)
			},
		},
		{
			Name: "delete char from search query",
			BuildExpr: func() vm.Expr {
				return altExpr(keyExpr(tcell.KeyBackspace), keyExpr(tcell.KeyBackspace2))
			},
			BuildAction: func(ctx Context, p CommandParams) Action {
				// This returns the input mode to normal if the search query is empty.
				return DeleteRuneFromSearchQuery
			},
		},
	}
}

func TaskModeCommands() []Command {
	return []Command{
		{
			Name: "cancel task",
			BuildExpr: func() vm.Expr {
				return keyExpr(tcell.KeyEscape)
			},
			BuildAction: func(ctx Context, p CommandParams) Action {
				return state.CancelTaskIfRunning
			},
		},
	}
}
