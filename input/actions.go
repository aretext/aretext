package input

import (
	"log"

	"github.com/aretext/aretext/locate"
	"github.com/aretext/aretext/state"
	"github.com/aretext/aretext/text"
	"github.com/gdamore/tcell/v2"
)

func CursorLeft(inputEvents []*tcell.EventKey, count *int64, config Config) Action {
	return func(s *state.EditorState) {
		state.MoveCursor(s, func(params state.LocatorParams) uint64 {
			return locate.PrevCharInLine(params.TextTree, 1, false, params.CursorPos)
		})
	}
}

func CursorBack(inputEvents []*tcell.EventKey, count *int64, config Config) Action {
	return func(s *state.EditorState) {
		state.MoveCursor(s, func(params state.LocatorParams) uint64 {
			return locate.PrevChar(params.TextTree, 1, params.CursorPos)
		})
	}
}

func CursorRight(inputEvents []*tcell.EventKey, count *int64, config Config) Action {
	return func(s *state.EditorState) {
		state.MoveCursor(s, func(params state.LocatorParams) uint64 {
			return locate.NextCharInLine(params.TextTree, 1, false, params.CursorPos)
		})
	}
}

func CursorRightIncludeEndOfLineOrFile(inputEvents []*tcell.EventKey, count *int64, config Config) Action {
	return func(s *state.EditorState) {
		state.MoveCursor(s, func(params state.LocatorParams) uint64 {
			return locate.NextCharInLine(params.TextTree, 1, true, params.CursorPos)
		})
	}
}

func CursorUp(inputEvents []*tcell.EventKey, count *int64, config Config) Action {
	return func(s *state.EditorState) {
		state.MoveCursorToLineAbove(s, 1)
	}
}

func CursorDown(inputEvents []*tcell.EventKey, count *int64, config Config) Action {
	return func(s *state.EditorState) {
		state.MoveCursorToLineBelow(s, 1)
	}
}

func CursorNextWordStart(inputEvents []*tcell.EventKey, count *int64, config Config) Action {
	return func(s *state.EditorState) {
		state.MoveCursor(s, func(params state.LocatorParams) uint64 {
			return locate.NextWordStart(params.TextTree, params.TokenTree, params.CursorPos)
		})
	}
}

func CursorPrevWordStart(inputEvents []*tcell.EventKey, count *int64, config Config) Action {
	return func(s *state.EditorState) {
		state.MoveCursor(s, func(params state.LocatorParams) uint64 {
			return locate.PrevWordStart(params.TextTree, params.TokenTree, params.CursorPos)
		})
	}
}

func CursorNextWordEnd(inputEvents []*tcell.EventKey, count *int64, config Config) Action {
	return func(s *state.EditorState) {
		state.MoveCursor(s, func(params state.LocatorParams) uint64 {
			return locate.NextWordEnd(params.TextTree, params.TokenTree, params.CursorPos)
		})
	}
}

func CursorPrevParagraph(inputEvents []*tcell.EventKey, count *int64, config Config) Action {
	return func(s *state.EditorState) {
		state.MoveCursor(s, func(params state.LocatorParams) uint64 {
			return locate.PrevParagraph(params.TextTree, params.CursorPos)
		})
	}
}

func CursorNextParagraph(inputEvents []*tcell.EventKey, count *int64, config Config) Action {
	return func(s *state.EditorState) {
		state.MoveCursor(s, func(params state.LocatorParams) uint64 {
			return locate.NextParagraph(params.TextTree, params.CursorPos)
		})
	}
}

func ScrollUp(inputEvents []*tcell.EventKey, count *int64, config Config) Action {
	scrollLines := config.ScrollLines
	if scrollLines < 1 {
		scrollLines = 1
	}

	return func(s *state.EditorState) {
		// Move the cursor to the start of a line above, then scroll up.
		// (We don't scroll the view, because that happens automatically after every action.)
		state.MoveCursor(s, func(params state.LocatorParams) uint64 {
			return locate.StartOfLineAbove(params.TextTree, scrollLines, params.CursorPos)
		})
		state.ScrollViewByNumLines(s, text.ReadDirectionBackward, scrollLines)
	}
}

func ScrollDown(inputEvents []*tcell.EventKey, count *int64, config Config) Action {
	scrollLines := config.ScrollLines
	if scrollLines < 1 {
		scrollLines = 1
	}

	return func(s *state.EditorState) {
		// Move the cursor to the start of a line below, then scroll down.
		// (We don't scroll the view, because that happens automatically after every action.)
		state.MoveCursor(s, func(params state.LocatorParams) uint64 {
			return locate.StartOfLineBelow(params.TextTree, scrollLines, params.CursorPos)
		})
		state.ScrollViewByNumLines(s, text.ReadDirectionForward, scrollLines)
	}
}

func CursorLineStart(inputEvents []*tcell.EventKey, count *int64, config Config) Action {
	return func(s *state.EditorState) {
		state.MoveCursor(s, func(params state.LocatorParams) uint64 {
			return locate.PrevLineBoundary(params.TextTree, params.CursorPos)
		})
	}
}

func CursorLineStartNonWhitespace(inputEvents []*tcell.EventKey, count *int64, config Config) Action {
	return func(s *state.EditorState) {
		state.MoveCursor(s, func(params state.LocatorParams) uint64 {
			lineStartPos := locate.PrevLineBoundary(params.TextTree, params.CursorPos)
			return locate.NextNonWhitespaceOrNewline(params.TextTree, lineStartPos)
		})
	}
}

func CursorLineEnd(inputEvents []*tcell.EventKey, count *int64, config Config) Action {
	return func(s *state.EditorState) {
		state.MoveCursor(s, func(params state.LocatorParams) uint64 {
			return locate.NextLineBoundary(params.TextTree, false, params.CursorPos)
		})
	}
}

func CursorLineEndIncludeEndOfLineOrFile(inputEvents []*tcell.EventKey, count *int64, config Config) Action {
	return func(s *state.EditorState) {
		state.MoveCursor(s, func(params state.LocatorParams) uint64 {
			return locate.NextLineBoundary(params.TextTree, true, params.CursorPos)
		})
	}
}

func CursorStartOfLineNum(inputEvents []*tcell.EventKey, count *int64, config Config) Action {
	// Convert 1-indexed count to 0-indexed line num
	var lineNum uint64
	if count != nil && *count > 0 {
		lineNum = uint64(*count - 1)
	}

	return func(s *state.EditorState) {
		state.MoveCursor(s, func(params state.LocatorParams) uint64 {
			lineStartPos := locate.StartOfLineNum(params.TextTree, lineNum)
			return locate.NextNonWhitespaceOrNewline(params.TextTree, lineStartPos)
		})
	}
}

func CursorStartOfLastLine(inputEvents []*tcell.EventKey, count *int64, config Config) Action {
	return func(s *state.EditorState) {
		state.MoveCursor(s, func(params state.LocatorParams) uint64 {
			lineStartPos := locate.StartOfLastLine(params.TextTree)
			return locate.NextNonWhitespaceOrNewline(params.TextTree, lineStartPos)
		})
	}
}

func EnterInsertMode(inputEvents []*tcell.EventKey, count *int64, config Config) Action {
	return func(s *state.EditorState) {
		state.SetStatusMsg(s, state.StatusMsg{})
		state.SetInputMode(s, state.InputModeInsert)
	}
}

func EnterInsertModeAtStartOfLine(inputEvents []*tcell.EventKey, count *int64, config Config) Action {
	return func(s *state.EditorState) {
		state.SetStatusMsg(s, state.StatusMsg{})
		state.SetInputMode(s, state.InputModeInsert)
		CursorLineStartNonWhitespace(nil, nil, config)(s)
	}
}

func EnterInsertModeAtNextPos(inputEvents []*tcell.EventKey, count *int64, config Config) Action {
	return func(s *state.EditorState) {
		state.SetStatusMsg(s, state.StatusMsg{})
		state.SetInputMode(s, state.InputModeInsert)
		CursorRightIncludeEndOfLineOrFile(nil, nil, config)(s)
	}
}

func EnterInsertModeAtEndOfLine(inputEvents []*tcell.EventKey, count *int64, config Config) Action {
	return func(s *state.EditorState) {
		state.SetStatusMsg(s, state.StatusMsg{})
		state.SetInputMode(s, state.InputModeInsert)
		CursorLineEndIncludeEndOfLineOrFile(nil, nil, config)(s)
	}
}

func BeginNewLineBelow(inputEvents []*tcell.EventKey, count *int64, config Config) Action {
	return func(s *state.EditorState) {
		CursorLineEndIncludeEndOfLineOrFile(nil, nil, config)(s)
		state.InsertNewline(s)
		state.SetInputMode(s, state.InputModeInsert)
	}
}

func BeginNewLineAbove(inputEvents []*tcell.EventKey, count *int64, config Config) Action {
	return func(s *state.EditorState) {
		CursorLineStart(nil, nil, config)(s)
		state.InsertNewline(s)
		CursorUp(nil, nil, config)(s)
		state.SetInputMode(s, state.InputModeInsert)
	}
}

func DeleteLine(inputEvents []*tcell.EventKey, count *int64, config Config) Action {
	currentPos := func(params state.LocatorParams) uint64 {
		return params.CursorPos
	}
	return func(s *state.EditorState) {
		state.DeleteLines(s, currentPos, false)
		CursorLineStartNonWhitespace(nil, nil, config)(s)
	}
}

func DeletePrevCharInLine(inputEvents []*tcell.EventKey, count *int64, config Config) Action {
	return func(s *state.EditorState) {
		state.DeleteRunes(s, func(params state.LocatorParams) uint64 {
			return locate.PrevCharInLine(params.TextTree, 1, false, params.CursorPos)
		})
	}
}

func DeleteNextCharInLine(inputEvents []*tcell.EventKey, count *int64, config Config) Action {
	return func(s *state.EditorState) {
		state.DeleteRunes(s, func(params state.LocatorParams) uint64 {
			return locate.NextCharInLine(params.TextTree, 1, true, params.CursorPos)
		})
		state.MoveCursor(s, func(params state.LocatorParams) uint64 {
			return locate.ClosestCharOnLine(params.TextTree, params.CursorPos)
		})
	}
}

func DeleteDown(inputEvents []*tcell.EventKey, count *int64, config Config) Action {
	targetLineLoc := func(params state.LocatorParams) uint64 {
		return locate.StartOfLineBelow(params.TextTree, 1, params.CursorPos)
	}
	return func(s *state.EditorState) {
		state.DeleteLines(s, targetLineLoc, true)
		CursorLineStartNonWhitespace(nil, nil, config)(s)
	}
}

func DeleteUp(inputEvents []*tcell.EventKey, count *int64, config Config) Action {
	targetLineLoc := func(params state.LocatorParams) uint64 {
		return locate.StartOfLineAbove(params.TextTree, 1, params.CursorPos)
	}

	return func(s *state.EditorState) {
		state.DeleteLines(s, targetLineLoc, true)
		CursorLineStartNonWhitespace(nil, nil, config)(s)
	}
}

func DeleteToEndOfLine(inputEvents []*tcell.EventKey, count *int64, config Config) Action {
	return func(s *state.EditorState) {
		state.DeleteRunes(s, func(params state.LocatorParams) uint64 {
			return locate.NextLineBoundary(params.TextTree, true, params.CursorPos)
		})
		state.MoveCursor(s, func(params state.LocatorParams) uint64 {
			return locate.ClosestCharOnLine(params.TextTree, params.CursorPos)
		})
	}
}

func DeleteToStartOfLine(inputEvents []*tcell.EventKey, count *int64, config Config) Action {
	return func(s *state.EditorState) {
		state.DeleteRunes(s, func(params state.LocatorParams) uint64 {
			return locate.PrevLineBoundary(params.TextTree, params.CursorPos)
		})
	}
}

func DeleteToStartOfLineNonWhitespace(inputEvents []*tcell.EventKey, count *int64, config Config) Action {
	return func(s *state.EditorState) {
		state.DeleteRunes(s, func(params state.LocatorParams) uint64 {
			lineStartPos := locate.PrevLineBoundary(params.TextTree, params.CursorPos)
			return locate.NextNonWhitespaceOrNewline(params.TextTree, lineStartPos)
		})
	}
}

func DeleteInnerWord(inputEvents []*tcell.EventKey, count *int64, config Config) Action {
	return func(s *state.EditorState) {
		state.MoveCursor(s, func(params state.LocatorParams) uint64 {
			return locate.CurrentWordStart(params.TextTree, params.TokenTree, params.CursorPos)
		})
		state.DeleteRunes(s, func(params state.LocatorParams) uint64 {
			return locate.CurrentWordEnd(params.TextTree, params.TokenTree, params.CursorPos)
		})
	}
}

func ChangeInnerWord(inputEvents []*tcell.EventKey, count *int64, config Config) Action {
	return func(s *state.EditorState) {
		DeleteInnerWord(nil, nil, config)(s)
		EnterInsertMode(nil, nil, config)(s)
	}
}

func ReplaceCharacter(inputEvents []*tcell.EventKey, count *int64, config Config) Action {
	if len(inputEvents) == 0 {
		// This should never happen if the parser rule is configured correctly.
		panic("Replace chars expects at least one input event")
	}

	lastInput := inputEvents[len(inputEvents)-1]
	if lastInput.Key() != tcell.KeyRune {
		log.Printf("Unsupported input for replace character command\n")
		return nil
	}

	newChar := string([]rune{lastInput.Rune()})
	return func(s *state.EditorState) {
		state.ReplaceChar(s, newChar)
	}
}

func ShowCommandMenu(inputEvents []*tcell.EventKey, count *int64, config Config) Action {
	return func(s *state.EditorState) {
		state.SetStatusMsg(s, state.StatusMsg{})

		// This sets the input mode to menu.
		state.ShowMenu(s, "command", commandMenuItems(config), false, true)
	}
}

func StartSearchForward(inputEvents []*tcell.EventKey, count *int64, config Config) Action {
	return func(s *state.EditorState) {
		state.SetStatusMsg(s, state.StatusMsg{})

		// This sets the input mode to search.
		state.StartSearch(s, text.ReadDirectionForward)
	}
}

func StartSearchBackward(inputEvents []*tcell.EventKey, count *int64, config Config) Action {
	return func(s *state.EditorState) {
		state.SetStatusMsg(s, state.StatusMsg{})

		// This sets the input mode to search.
		state.StartSearch(s, text.ReadDirectionBackward)
	}
}

func FindNextMatch(inputEvents []*tcell.EventKey, count *int64, config Config) Action {
	return func(s *state.EditorState) {
		state.FindNextMatch(s, false)
	}
}

func FindPrevMatch(inputEvents []*tcell.EventKey, count *int64, config Config) Action {
	return func(s *state.EditorState) {
		reverse := true
		state.FindNextMatch(s, reverse)
	}
}
