package input

import (
	"github.com/aretext/aretext/clipboard"
	"github.com/aretext/aretext/locate"
	"github.com/aretext/aretext/selection"
	"github.com/aretext/aretext/state"
)

// Action is a function that mutates the editor state.
type Action func(*state.EditorState)

// EmptyAction is an action that does nothing.
func EmptyAction(s *state.EditorState) {}

func CursorLeft(s *state.EditorState) {
	state.MoveCursor(s, func(params state.LocatorParams) uint64 {
		return locate.PrevCharInLine(params.TextTree, 1, false, params.CursorPos)
	})
}

func CursorBack(s *state.EditorState) {
	state.MoveCursor(s, func(params state.LocatorParams) uint64 {
		return locate.PrevChar(params.TextTree, 1, params.CursorPos)
	})
}

func CursorRight(s *state.EditorState) {
	state.MoveCursor(s, func(params state.LocatorParams) uint64 {
		return locate.NextCharInLine(params.TextTree, 1, false, params.CursorPos)
	})
}

func CursorRightIncludeEndOfLineOrFile(s *state.EditorState) {
	state.MoveCursor(s, func(params state.LocatorParams) uint64 {
		return locate.NextCharInLine(params.TextTree, 1, true, params.CursorPos)
	})
}

func CursorUp(s *state.EditorState) {
	state.MoveCursorToLineAbove(s, 1)
}

func CursorDown(s *state.EditorState) {
	state.MoveCursorToLineBelow(s, 1)
}

func CursorNextWordStart(s *state.EditorState) {
	state.MoveCursor(s, func(params state.LocatorParams) uint64 {
		return locate.NextWordStart(params.TextTree, params.CursorPos)
	})
}

func CursorPrevWordStart(s *state.EditorState) {
	state.MoveCursor(s, func(params state.LocatorParams) uint64 {
		return locate.PrevWordStart(params.TextTree, params.CursorPos)
	})
}

func CursorNextWordEnd(s *state.EditorState) {
	state.MoveCursor(s, func(params state.LocatorParams) uint64 {
		return locate.NextWordEnd(params.TextTree, params.CursorPos)
	})
}

func CursorPrevParagraph(s *state.EditorState) {
	state.MoveCursor(s, func(params state.LocatorParams) uint64 {
		return locate.PrevParagraph(params.TextTree, params.CursorPos)
	})
}

func CursorNextParagraph(s *state.EditorState) {
	state.MoveCursor(s, func(params state.LocatorParams) uint64 {
		return locate.NextParagraph(params.TextTree, params.CursorPos)
	})
}

func CursorToNextMatchingChar(char rune, count uint64, includeChar bool) Action {
	return func(s *state.EditorState) {
		state.MoveCursor(s, func(params state.LocatorParams) uint64 {
			found, pos := locate.NextMatchingCharInLine(params.TextTree, char, count, includeChar, params.CursorPos)
			if !found {
				pos = params.CursorPos
			}
			return pos
		})
	}
}

func CursorToPrevMatchingChar(char rune, count uint64, includeChar bool) Action {
	return func(s *state.EditorState) {
		state.MoveCursor(s, func(params state.LocatorParams) uint64 {
			found, pos := locate.PrevMatchingCharInLine(params.TextTree, char, count, includeChar, params.CursorPos)
			if !found {
				pos = params.CursorPos
			}
			return pos
		})
	}
}

func ScrollUp(config Config) Action {
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
		state.ScrollViewByNumLines(s, state.ScrollDirectionBackward, scrollLines)
	}
}

func ScrollDown(config Config) Action {
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
		state.ScrollViewByNumLines(s, state.ScrollDirectionForward, scrollLines)
	}
}

func CursorLineStart(s *state.EditorState) {
	state.MoveCursor(s, func(params state.LocatorParams) uint64 {
		return locate.PrevLineBoundary(params.TextTree, params.CursorPos)
	})
}

func CursorLineStartNonWhitespace(s *state.EditorState) {
	state.MoveCursor(s, func(params state.LocatorParams) uint64 {
		lineStartPos := locate.PrevLineBoundary(params.TextTree, params.CursorPos)
		return locate.NextNonWhitespaceOrNewline(params.TextTree, lineStartPos)
	})
}

func CursorLineEnd(s *state.EditorState) {
	state.MoveCursor(s, func(params state.LocatorParams) uint64 {
		return locate.NextLineBoundary(params.TextTree, false, params.CursorPos)
	})
}

func CursorLineEndIncludeEndOfLineOrFile(s *state.EditorState) {
	state.MoveCursor(s, func(params state.LocatorParams) uint64 {
		return locate.NextLineBoundary(params.TextTree, true, params.CursorPos)
	})
}

func CursorStartOfLineNum(count uint64) Action {
	// Convert 1-indexed count to 0-indexed line num
	lineNum := count
	if lineNum > 0 {
		lineNum--
	}

	return func(s *state.EditorState) {
		state.MoveCursor(s, func(params state.LocatorParams) uint64 {
			lineStartPos := locate.StartOfLineNum(params.TextTree, lineNum)
			return locate.NextNonWhitespaceOrNewline(params.TextTree, lineStartPos)
		})
	}
}

func CursorStartOfLastLine(s *state.EditorState) {
	state.MoveCursor(s, func(params state.LocatorParams) uint64 {
		lineStartPos := locate.StartOfLastLine(params.TextTree)
		return locate.NextNonWhitespaceOrNewline(params.TextTree, lineStartPos)
	})
}

func EnterInsertMode(s *state.EditorState) {
	state.SetInputMode(s, state.InputModeInsert)
}

func EnterInsertModeAtStartOfLine(s *state.EditorState) {
	state.SetInputMode(s, state.InputModeInsert)
	CursorLineStartNonWhitespace(s)
}

func EnterInsertModeAtNextPos(s *state.EditorState) {
	state.SetInputMode(s, state.InputModeInsert)
	CursorRightIncludeEndOfLineOrFile(s)
}

func EnterInsertModeAtEndOfLine(s *state.EditorState) {
	state.SetInputMode(s, state.InputModeInsert)
	CursorLineEndIncludeEndOfLineOrFile(s)
}

func ReturnToNormalMode(s *state.EditorState) {
	state.SetInputMode(s, state.InputModeNormal)
}

func ReturnToNormalModeAfterInsert(s *state.EditorState) {
	state.ClearAutoIndentWhitespaceLine(s, func(params state.LocatorParams) uint64 {
		return locate.StartOfLineAtPos(params.TextTree, params.CursorPos)
	})
	CursorLeft(s)
	state.SetInputMode(s, state.InputModeNormal)
}

func InsertRune(r rune) Action {
	return func(s *state.EditorState) {
		state.InsertRune(s, r)
	}
}

func InsertNewlineAndUpdateAutoIndentWhitespace(s *state.EditorState) {
	state.InsertNewline(s)
	state.ClearAutoIndentWhitespaceLine(s, func(params state.LocatorParams) uint64 {
		return locate.StartOfLineAbove(params.TextTree, 1, params.CursorPos)
	})
}

func InsertTab(s *state.EditorState) {
	state.InsertTab(s)
}

func DeletePrevChar(clipboardPage clipboard.PageId) Action {
	return func(s *state.EditorState) {
		state.DeleteRunes(s, func(params state.LocatorParams) uint64 {
			prevInLinePos := locate.PrevCharInLine(params.TextTree, 1, true, params.CursorPos)
			prevAutoIndentPos := locate.PrevAutoIndent(
				params.TextTree,
				params.AutoIndentEnabled,
				params.TabSize,
				params.CursorPos)
			if prevInLinePos < prevAutoIndentPos {
				return prevInLinePos
			} else {
				return prevAutoIndentPos
			}
		}, clipboardPage)
	}
}

func BeginNewLineBelow(s *state.EditorState) {
	CursorLineEndIncludeEndOfLineOrFile(s)
	state.InsertNewline(s)
	state.SetInputMode(s, state.InputModeInsert)
}

func BeginNewLineAbove(s *state.EditorState) {
	state.BeginNewLineAbove(s)
	EnterInsertMode(s)
}

func JoinLines(s *state.EditorState) {
	state.JoinLines(s)
}

func DeleteLines(count uint64, clipboardPage clipboard.PageId) Action {
	if count > 0 {
		count--
	}
	return func(s *state.EditorState) {
		targetLoc := func(params state.LocatorParams) uint64 {
			return locate.StartOfLineBelow(params.TextTree, count, params.CursorPos)
		}
		state.DeleteLines(s, targetLoc, false, false, clipboardPage)
		CursorLineStartNonWhitespace(s)
	}
}

func DeletePrevCharInLine(clipboardPage clipboard.PageId) Action {
	return func(s *state.EditorState) {
		state.DeleteRunes(s, func(params state.LocatorParams) uint64 {
			return locate.PrevCharInLine(params.TextTree, 1, false, params.CursorPos)
		}, clipboardPage)
	}
}

func DeleteNextCharInLine(count uint64, clipboardPage clipboard.PageId) Action {
	return func(s *state.EditorState) {
		state.DeleteRunes(s, func(params state.LocatorParams) uint64 {
			return locate.NextCharInLine(params.TextTree, count, true, params.CursorPos)
		}, clipboardPage)
		state.MoveCursor(s, func(params state.LocatorParams) uint64 {
			return locate.ClosestCharOnLine(params.TextTree, params.CursorPos)
		})
	}
}

func DeleteToNextMatchingChar(char rune, count uint64, clipboardPage clipboard.PageId, includeChar bool) Action {
	return func(s *state.EditorState) {
		state.DeleteRunes(s, func(params state.LocatorParams) uint64 {
			found, pos := locate.NextMatchingCharInLine(params.TextTree, char, count, includeChar, params.CursorPos)
			if !found {
				// No character matched in this line, so don't delete anything.
				return params.CursorPos
			}
			// Delete up to and including `pos`
			return locate.NextCharInLine(params.TextTree, 1, true, pos)
		}, clipboardPage)
		state.MoveCursor(s, func(params state.LocatorParams) uint64 {
			return locate.ClosestCharOnLine(params.TextTree, params.CursorPos)
		})
	}
}

func DeleteToPrevMatchingChar(char rune, count uint64, clipboardPage clipboard.PageId, includeChar bool) Action {
	return func(s *state.EditorState) {
		state.DeleteRunes(s, func(params state.LocatorParams) uint64 {
			found, pos := locate.PrevMatchingCharInLine(params.TextTree, char, count, includeChar, params.CursorPos)
			if !found {
				pos = params.CursorPos
			}
			return pos
		}, clipboardPage)
	}
}

func DeleteDown(clipboardPage clipboard.PageId) Action {
	return func(s *state.EditorState) {
		targetLineLoc := func(params state.LocatorParams) uint64 {
			return locate.StartOfLineBelow(params.TextTree, 1, params.CursorPos)
		}
		state.DeleteLines(s, targetLineLoc, true, false, clipboardPage)
		CursorLineStartNonWhitespace(s)
	}
}

func DeleteUp(clipboardPage clipboard.PageId) Action {
	return func(s *state.EditorState) {
		targetLineLoc := func(params state.LocatorParams) uint64 {
			return locate.StartOfLineAbove(params.TextTree, 1, params.CursorPos)
		}
		state.DeleteLines(s, targetLineLoc, true, false, clipboardPage)
		CursorLineStartNonWhitespace(s)
	}
}

func DeleteToEndOfLine(clipboardPage clipboard.PageId) Action {
	return func(s *state.EditorState) {
		state.DeleteRunes(s, func(params state.LocatorParams) uint64 {
			return locate.NextLineBoundary(params.TextTree, true, params.CursorPos)
		}, clipboardPage)
		state.MoveCursor(s, func(params state.LocatorParams) uint64 {
			return locate.ClosestCharOnLine(params.TextTree, params.CursorPos)
		})
	}
}

func DeleteToStartOfLine(clipboardPage clipboard.PageId) Action {
	return func(s *state.EditorState) {
		state.DeleteRunes(s, func(params state.LocatorParams) uint64 {
			return locate.PrevLineBoundary(params.TextTree, params.CursorPos)
		}, clipboardPage)
	}
}

func DeleteToStartOfLineNonWhitespace(clipboardPage clipboard.PageId) Action {
	return func(s *state.EditorState) {
		state.DeleteRunes(s, func(params state.LocatorParams) uint64 {
			lineStartPos := locate.PrevLineBoundary(params.TextTree, params.CursorPos)
			return locate.NextNonWhitespaceOrNewline(params.TextTree, lineStartPos)
		}, clipboardPage)
	}
}

func DeleteToStartOfNextWord(clipboardPage clipboard.PageId) Action {
	return func(s *state.EditorState) {
		state.DeleteRunes(s, func(params state.LocatorParams) uint64 {
			return locate.NextWordStartInLine(params.TextTree, params.CursorPos)
		}, clipboardPage)
		state.MoveCursor(s, func(params state.LocatorParams) uint64 {
			return locate.ClosestCharOnLine(params.TextTree, params.CursorPos)
		})
	}
}

func DeleteAWord(clipboardPage clipboard.PageId) Action {
	return func(s *state.EditorState) {
		state.MoveCursor(s, func(params state.LocatorParams) uint64 {
			return locate.CurrentWordStart(params.TextTree, params.CursorPos)
		})
		state.DeleteRunes(s, func(params state.LocatorParams) uint64 {
			return locate.CurrentWordEndWithTrailingWhitespace(params.TextTree, params.CursorPos)
		}, clipboardPage)
	}
}

func DeleteInnerWord(clipboardPage clipboard.PageId) Action {
	return func(s *state.EditorState) {
		state.MoveCursor(s, func(params state.LocatorParams) uint64 {
			return locate.CurrentWordStart(params.TextTree, params.CursorPos)
		})
		state.DeleteRunes(s, func(params state.LocatorParams) uint64 {
			return locate.CurrentWordEnd(params.TextTree, params.CursorPos)
		}, clipboardPage)
	}
}

func ChangeToStartOfNextWord(clipboardPage clipboard.PageId) Action {
	return func(s *state.EditorState) {
		state.DeleteRunes(s, func(params state.LocatorParams) uint64 {
			return locate.NextWordStartInLine(params.TextTree, params.CursorPos)
		}, clipboardPage)
		EnterInsertMode(s)
	}
}

func ChangeAWord(clipboardPage clipboard.PageId) Action {
	deleteAWordAction := DeleteAWord(clipboardPage)
	return func(s *state.EditorState) {
		deleteAWordAction(s)
		EnterInsertMode(s)
	}
}

func ChangeInnerWord(clipboardPage clipboard.PageId) Action {
	deleteInnerWordAction := DeleteInnerWord(clipboardPage)
	return func(s *state.EditorState) {
		deleteInnerWordAction(s)
		EnterInsertMode(s)
	}
}

func ChangeToNextMatchingChar(char rune, count uint64, clipboardPage clipboard.PageId, includeChar bool) Action {
	deleteToNextMatchingCharAction := DeleteToNextMatchingChar(char, count, clipboardPage, includeChar)
	return func(s *state.EditorState) {
		deleteToNextMatchingCharAction(s)
		EnterInsertMode(s)
	}
}

func ChangeToPrevMatchingChar(char rune, count uint64, clipboardPage clipboard.PageId, includeChar bool) Action {
	deleteToPrevMatchingCharAction := DeleteToPrevMatchingChar(char, count, clipboardPage, includeChar)
	return func(s *state.EditorState) {
		deleteToPrevMatchingCharAction(s)
		EnterInsertMode(s)
	}
}

func ReplaceCharacter(newChar rune) Action {
	return func(s *state.EditorState) {
		state.ReplaceChar(s, newChar)
	}
}

func ToggleCaseAtCursor(s *state.EditorState) {
	state.ToggleCaseAtCursor(s)
}

func IndentLine(s *state.EditorState) {
	targetLineLoc := func(p state.LocatorParams) uint64 { return p.CursorPos }
	state.IndentLines(s, targetLineLoc)
}

func OutdentLine(s *state.EditorState) {
	targetLineLoc := func(p state.LocatorParams) uint64 { return p.CursorPos }
	state.OutdentLines(s, targetLineLoc)
}

func CopyToStartOfNextWord(clipboardPage clipboard.PageId) Action {
	return func(s *state.EditorState) {
		startLoc := func(params state.LocatorParams) uint64 {
			return params.CursorPos
		}
		endLoc := func(params state.LocatorParams) uint64 {
			return locate.NextWordStartInLine(params.TextTree, params.CursorPos)
		}
		state.CopyRegion(s, clipboardPage, startLoc, endLoc)
	}
}

func CopyAWord(clipboardPage clipboard.PageId) Action {
	return func(s *state.EditorState) {
		startLoc := func(params state.LocatorParams) uint64 {
			return locate.CurrentWordStart(params.TextTree, params.CursorPos)
		}
		endLoc := func(params state.LocatorParams) uint64 {
			return locate.CurrentWordEndWithTrailingWhitespace(params.TextTree, params.CursorPos)
		}
		state.CopyRegion(s, clipboardPage, startLoc, endLoc)
	}
}

func CopyInnerWord(clipboardPage clipboard.PageId) Action {
	return func(s *state.EditorState) {
		startLoc := func(params state.LocatorParams) uint64 {
			return locate.CurrentWordStart(params.TextTree, params.CursorPos)
		}
		endLoc := func(params state.LocatorParams) uint64 {
			return locate.CurrentWordEnd(params.TextTree, params.CursorPos)
		}
		state.CopyRegion(s, clipboardPage, startLoc, endLoc)
	}
}

func CopyLines(clipboardPage clipboard.PageId) Action {
	return func(s *state.EditorState) {
		state.CopyLine(s, clipboardPage)
	}
}

func PasteAfterCursor(clipboardPage clipboard.PageId) Action {
	return func(s *state.EditorState) {
		state.PasteAfterCursor(s, clipboardPage)
	}
}

func PasteBeforeCursor(clipboardPage clipboard.PageId) Action {
	return func(s *state.EditorState) {
		state.PasteBeforeCursor(s, clipboardPage)
	}
}

func ShowCommandMenu(config Config) Action {
	return func(s *state.EditorState) {
		// This sets the input mode to menu.
		state.ShowMenu(s, state.MenuStyleCommand, menuItems(config))
	}
}

func ShowFileMenu(config Config) Action {
	return func(s *state.EditorState) {
		state.ShowFileMenu(s, config.DirPatternsToHide)
	}
}

func HideMenuAndReturnToNormalMode(s *state.EditorState) {
	state.HideMenu(s)
}

func ExecuteSelectedMenuItem(s *state.EditorState) {
	// Hides the menu, then executes the menu item action.
	// This usually returns to normal mode, unless the menu item action sets a different mode.
	state.ExecuteSelectedMenuItem(s)
}

func MenuSelectionUp(s *state.EditorState) {
	state.MoveMenuSelection(s, -1)
}

func MenuSelectionDown(s *state.EditorState) {
	state.MoveMenuSelection(s, 1)
}

func AppendRuneToMenuSearch(r rune) Action {
	return func(s *state.EditorState) {
		state.AppendRuneToMenuSearch(s, r)
	}
}

func DeleteRuneFromMenuSearch(s *state.EditorState) {
	state.DeleteRuneFromMenuSearch(s)
}

func StartSearchForward(s *state.EditorState) {
	// This sets the input mode to search.
	state.StartSearch(s, state.SearchDirectionForward)
}

func StartSearchBackward(s *state.EditorState) {
	// This sets the input mode to search.
	state.StartSearch(s, state.SearchDirectionBackward)
}

func AbortSearchAndReturnToNormalMode(s *state.EditorState) {
	state.CompleteSearch(s, false)
}

func CommitSearchAndReturnToNormalMode(s *state.EditorState) {
	state.CompleteSearch(s, true)
}

func AppendRuneToSearchQuery(r rune) Action {
	return func(s *state.EditorState) {
		state.AppendRuneToSearchQuery(s, r)
	}
}

func DeleteRuneFromSearchQuery(s *state.EditorState) {
	state.DeleteRuneFromSearchQuery(s)
}

func FindNextMatch(s *state.EditorState) {
	state.FindNextMatch(s, false)
}

func FindPrevMatch(s *state.EditorState) {
	reverse := true
	state.FindNextMatch(s, reverse)
}

func Undo(s *state.EditorState) {
	state.Undo(s)
}

func Redo(s *state.EditorState) {
	state.Redo(s)
}

func ToggleVisualModeCharwise(s *state.EditorState) {
	state.ToggleVisualMode(s, selection.ModeChar)
}

func ToggleVisualModeLinewise(s *state.EditorState) {
	state.ToggleVisualMode(s, selection.ModeLine)
}

func DeleteSelection(clipboardPage clipboard.PageId, selectionMode selection.Mode, selectionEndLoc state.Locator, replaceWithEmptyLine bool) Action {
	return func(s *state.EditorState) {
		state.MoveCursorToStartOfSelection(s)
		if selectionMode == selection.ModeChar {
			state.DeleteRunes(s, selectionEndLoc, clipboardPage)
		} else if selectionMode == selection.ModeLine {
			state.DeleteLines(s, selectionEndLoc, false, replaceWithEmptyLine, clipboardPage)
		}
	}
}

func DeleteSelectionAndReturnToNormalMode(clipboardPage clipboard.PageId, selectionMode selection.Mode, selectionEndLoc state.Locator) Action {
	deleteSelectionAction := DeleteSelection(clipboardPage, selectionMode, selectionEndLoc, false)
	return func(s *state.EditorState) {
		deleteSelectionAction(s)
		ReturnToNormalMode(s)
	}
}

func ToggleCaseInSelectionAndReturnToNormalMode(selectionEndLoc state.Locator) Action {
	return func(s *state.EditorState) {
		state.MoveCursorToStartOfSelection(s)
		state.ToggleCaseInSelection(s, selectionEndLoc)
		ReturnToNormalMode(s)
	}
}

func IndentSelectionAndReturnToNormalMode(selectionEndLoc state.Locator) Action {
	return func(s *state.EditorState) {
		state.MoveCursorToStartOfSelection(s)
		state.IndentLines(s, selectionEndLoc)
		ReturnToNormalMode(s)
	}
}

func OutdentSelectionAndReturnToNormalMode(selectionEndLoc state.Locator) Action {
	return func(s *state.EditorState) {
		state.MoveCursorToStartOfSelection(s)
		state.OutdentLines(s, selectionEndLoc)
		ReturnToNormalMode(s)
	}
}

func ChangeSelection(clipboardPage clipboard.PageId, selectionMode selection.Mode, selectionEndLoc state.Locator) Action {
	deleteSelectionAction := DeleteSelection(clipboardPage, selectionMode, selectionEndLoc, true)
	return func(s *state.EditorState) {
		deleteSelectionAction(s)
		EnterInsertMode(s)
	}
}

func CopySelectionAndReturnToNormalMode(clipboardPage clipboard.PageId) Action {
	return func(s *state.EditorState) {
		state.CopySelection(s, clipboardPage)
		ReturnToNormalMode(s)
	}
}

func ReplayLastActionMacro(count uint64) Action {
	return func(s *state.EditorState) {
		state.ReplayLastActionMacro(s, count)
	}
}
