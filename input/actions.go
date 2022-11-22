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

func CursorLeft(count uint64) Action {
	return func(s *state.EditorState) {
		state.MoveCursor(s, func(params state.LocatorParams) uint64 {
			return locate.PrevCharInLine(params.TextTree, count, false, params.CursorPos)
		})
	}
}

func CursorBack(count uint64) Action {
	return func(s *state.EditorState) {
		state.MoveCursor(s, func(params state.LocatorParams) uint64 {
			return locate.PrevChar(params.TextTree, count, params.CursorPos)
		})
	}
}

func CursorRight(count uint64) Action {
	return func(s *state.EditorState) {
		state.MoveCursor(s, func(params state.LocatorParams) uint64 {
			return locate.NextCharInLine(params.TextTree, count, false, params.CursorPos)
		})
	}
}

func CursorRightIncludeEndOfLineOrFile(s *state.EditorState) {
	state.MoveCursor(s, func(params state.LocatorParams) uint64 {
		return locate.NextCharInLine(params.TextTree, 1, true, params.CursorPos)
	})
}

func CursorUp(count uint64) Action {
	return func(s *state.EditorState) {
		state.MoveCursorToLineAbove(s, count)
	}
}

func CursorDown(count uint64) Action {
	return func(s *state.EditorState) {
		state.MoveCursorToLineBelow(s, count)
	}
}

func CursorNextLine(count uint64) Action {
	return func(s *state.EditorState) {
		state.MoveCursorToLineBelow(s, count)
		state.MoveCursor(s, func(params state.LocatorParams) uint64 {
			lineStartPos := locate.PrevLineBoundary(params.TextTree, params.CursorPos)
			return locate.NextNonWhitespaceOrNewline(params.TextTree, lineStartPos)
		})
	}
}

func CursorNextWordStart(count uint64, withPunctuation bool) Action {
	return func(s *state.EditorState) {
		state.MoveCursor(s, func(params state.LocatorParams) uint64 {
			return locate.NextWordStart(params.TextTree, params.CursorPos, count, withPunctuation, false)
		})
	}
}

func CursorPrevWordStart(count uint64, withPunctuation bool) Action {
	return func(s *state.EditorState) {
		state.MoveCursor(s, func(params state.LocatorParams) uint64 {
			return locate.PrevWordStart(params.TextTree, params.CursorPos, count, withPunctuation)
		})
	}
}

func CursorNextWordEnd(count uint64, withPunctuation bool) Action {
	return func(s *state.EditorState) {
		state.MoveCursor(s, func(params state.LocatorParams) uint64 {
			return locate.NextWordEnd(params.TextTree, params.CursorPos, count, withPunctuation)
		})
	}
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

func ScrollUp(ctx Context, half bool) Action {
	scrollLines := ctx.ScrollLines
	if scrollLines < 1 {
		scrollLines = 1
	} else if half {
		scrollLines /= 2
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

func ScrollDown(ctx Context, half bool) Action {
	scrollLines := ctx.ScrollLines
	if scrollLines < 1 {
		scrollLines = 1
	} else if half {
		scrollLines /= 2
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

func CursorMatchingCodeBlockDelimiter(s *state.EditorState) {
	state.MoveCursor(s, func(params state.LocatorParams) uint64 {
		matchPos, hasMatch := locate.MatchingCodeBlockDelimiter(params.TextTree, params.SyntaxParser, params.CursorPos)
		if hasMatch {
			return matchPos
		} else {
			return params.CursorPos
		}
	})
}

func CursorPrevUnmatchedOpenBrace(s *state.EditorState) {
	state.MoveCursor(s, func(params state.LocatorParams) uint64 {
		matchPos, hasMatch := locate.PrevUnmatchedOpenDelimiter(locate.BracePair, params.TextTree, params.SyntaxParser, params.CursorPos)
		if hasMatch {
			return matchPos
		} else {
			return params.CursorPos
		}
	})
}

func CursorNextUnmatchedCloseBrace(s *state.EditorState) {
	state.MoveCursor(s, func(params state.LocatorParams) uint64 {
		matchPos, hasMatch := locate.NextUnmatchedCloseDelimiter(locate.BracePair, params.TextTree, params.SyntaxParser, params.CursorPos)
		if hasMatch {
			return matchPos
		} else {
			return params.CursorPos
		}
	})
}

func CursorPrevUnmatchedOpenParen(s *state.EditorState) {
	state.MoveCursor(s, func(params state.LocatorParams) uint64 {
		matchPos, hasMatch := locate.PrevUnmatchedOpenDelimiter(locate.ParenPair, params.TextTree, params.SyntaxParser, params.CursorPos)
		if hasMatch {
			return matchPos
		} else {
			return params.CursorPos
		}
	})
}

func CursorNextUnmatchedCloseParen(s *state.EditorState) {
	state.MoveCursor(s, func(params state.LocatorParams) uint64 {
		matchPos, hasMatch := locate.NextUnmatchedCloseDelimiter(locate.ParenPair, params.TextTree, params.SyntaxParser, params.CursorPos)
		if hasMatch {
			return matchPos
		} else {
			return params.CursorPos
		}
	})
}

func DeleteParenBlock(includeParens bool, clipboardPage clipboard.PageId) Action {
	return func(s *state.EditorState) {
		state.DeleteRange(s, func(params state.LocatorParams) (uint64, uint64) {
			return locate.DelimitedBlock(locate.ParenPair, params.TextTree, params.SyntaxParser, includeParens, params.CursorPos)
		}, clipboardPage)
	}
}

func DeleteBraceBlock(includeBraces bool, clipboardPage clipboard.PageId) Action {
	return func(s *state.EditorState) {
		state.DeleteRange(s, func(params state.LocatorParams) (uint64, uint64) {
			return locate.DelimitedBlock(locate.BracePair, params.TextTree, params.SyntaxParser, includeBraces, params.CursorPos)
		}, clipboardPage)
	}
}

func DeleteAngleBlock(includeAngleBrackets bool, clipboardPage clipboard.PageId) Action {
	return func(s *state.EditorState) {
		state.DeleteRange(s, func(params state.LocatorParams) (uint64, uint64) {
			return locate.DelimitedBlock(locate.AnglePair, params.TextTree, params.SyntaxParser, includeAngleBrackets, params.CursorPos)
		}, clipboardPage)
	}
}

func ChangeParenBlock(includeParens bool, clipboardPage clipboard.PageId) Action {
	return func(s *state.EditorState) {
		startPos, endPos := state.DeleteRange(s, func(params state.LocatorParams) (uint64, uint64) {
			return locate.DelimitedBlock(locate.ParenPair, params.TextTree, params.SyntaxParser, includeParens, params.CursorPos)
		}, clipboardPage)

		if startPos == endPos {
			// Not within a paren block.
			return
		}

		EnterInsertMode(s)
	}
}

func ChangeBraceBlock(includeBraces bool, clipboardPage clipboard.PageId) Action {
	return func(s *state.EditorState) {
		startPos, endPos := state.DeleteRange(s, func(params state.LocatorParams) (uint64, uint64) {
			return locate.DelimitedBlock(locate.BracePair, params.TextTree, params.SyntaxParser, includeBraces, params.CursorPos)
		}, clipboardPage)

		if startPos == endPos {
			// Not within a brace block.
			return
		}

		if !includeBraces {
			state.InsertNewline(s)
			state.InsertNewline(s)
			state.MoveCursorToLineAbove(s, 1)
		}

		EnterInsertMode(s)
	}
}

func ChangeAngleBlock(includeAngleBrackets bool, clipboardPage clipboard.PageId) Action {
	return func(s *state.EditorState) {
		startPos, endPos := state.DeleteRange(s, func(params state.LocatorParams) (uint64, uint64) {
			return locate.DelimitedBlock(locate.AnglePair, params.TextTree, params.SyntaxParser, includeAngleBrackets, params.CursorPos)
		}, clipboardPage)

		if startPos == endPos {
			// Not within a paren block.
			return
		}

		EnterInsertMode(s)
	}
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
	state.MoveCursor(s, func(params state.LocatorParams) uint64 {
		return locate.PrevCharInLine(params.TextTree, 1, false, params.CursorPos)
	})
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
		state.DeleteToPos(s, func(params state.LocatorParams) uint64 {
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
		state.DeleteToPos(s, func(params state.LocatorParams) uint64 {
			return locate.PrevCharInLine(params.TextTree, 1, false, params.CursorPos)
		}, clipboardPage)
	}
}

func DeleteNextCharInLine(count uint64, clipboardPage clipboard.PageId) Action {
	return func(s *state.EditorState) {
		state.DeleteToPos(s, func(params state.LocatorParams) uint64 {
			return locate.NextCharInLine(params.TextTree, count, true, params.CursorPos)
		}, clipboardPage)
		state.MoveCursor(s, func(params state.LocatorParams) uint64 {
			return locate.ClosestCharOnLine(params.TextTree, params.CursorPos)
		})
	}
}

func DeleteToNextMatchingChar(char rune, count uint64, clipboardPage clipboard.PageId, includeChar bool) Action {
	return func(s *state.EditorState) {
		state.DeleteToPos(s, func(params state.LocatorParams) uint64 {
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
		state.DeleteToPos(s, func(params state.LocatorParams) uint64 {
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
		state.DeleteToPos(s, func(params state.LocatorParams) uint64 {
			return locate.NextLineBoundary(params.TextTree, true, params.CursorPos)
		}, clipboardPage)
		state.MoveCursor(s, func(params state.LocatorParams) uint64 {
			return locate.ClosestCharOnLine(params.TextTree, params.CursorPos)
		})
	}
}

func DeleteToStartOfLine(clipboardPage clipboard.PageId) Action {
	return func(s *state.EditorState) {
		state.DeleteToPos(s, func(params state.LocatorParams) uint64 {
			return locate.PrevLineBoundary(params.TextTree, params.CursorPos)
		}, clipboardPage)
	}
}

func DeleteToStartOfLineNonWhitespace(clipboardPage clipboard.PageId) Action {
	return func(s *state.EditorState) {
		state.DeleteToPos(s, func(params state.LocatorParams) uint64 {
			lineStartPos := locate.PrevLineBoundary(params.TextTree, params.CursorPos)
			return locate.NextNonWhitespaceOrNewline(params.TextTree, lineStartPos)
		}, clipboardPage)
	}
}

func DeleteToStartOfNextWord(count uint64, clipboardPage clipboard.PageId, withPunctuation bool) Action {
	return func(s *state.EditorState) {
		state.DeleteToPos(s, func(params state.LocatorParams) uint64 {
			endPos := locate.NextWordStart(params.TextTree, params.CursorPos, count, withPunctuation, true)
			if endPos == params.CursorPos {
				// The cursor didn't move, so we're on an empty line.
				// Attempt to delete the newline at the end of the line.
				endPos = locate.StartOfLineBelow(params.TextTree, 1, params.CursorPos)
			}
			return endPos
		}, clipboardPage)
		state.MoveCursor(s, func(params state.LocatorParams) uint64 {
			pos := locate.NextNonWhitespaceOrNewline(params.TextTree, params.CursorPos)
			return locate.ClosestCharOnLine(params.TextTree, pos)
		})
	}
}

func DeleteAWord(count uint64, clipboardPage clipboard.PageId) Action {
	return func(s *state.EditorState) {
		state.DeleteRange(s, func(params state.LocatorParams) (uint64, uint64) {
			return locate.WordObject(params.TextTree, params.CursorPos, count)
		}, clipboardPage)
		state.MoveCursor(s, func(params state.LocatorParams) uint64 {
			return locate.ClosestCharOnLine(params.TextTree, params.CursorPos)
		})
	}
}

func DeleteInnerWord(count uint64, clipboardPage clipboard.PageId) Action {
	return func(s *state.EditorState) {
		state.DeleteRange(s, func(params state.LocatorParams) (uint64, uint64) {
			return locate.InnerWordObject(params.TextTree, params.CursorPos, count)
		}, clipboardPage)
		state.MoveCursor(s, func(params state.LocatorParams) uint64 {
			return locate.ClosestCharOnLine(params.TextTree, params.CursorPos)
		})
	}
}

func ChangeWord(count uint64, clipboardPage clipboard.PageId) Action {
	return func(s *state.EditorState) {
		state.DeleteToPos(s, func(params state.LocatorParams) uint64 {
			// Unlike "dw", "cw" within a word excludes whitespace after the word by default.
			// See https://vimhelp.org/change.txt.html
			_, endPos := locate.InnerWordObject(params.TextTree, params.CursorPos, count)
			return endPos
		}, clipboardPage)
		EnterInsertMode(s)
	}
}

func ChangeAWord(count uint64, clipboardPage clipboard.PageId) Action {
	return func(s *state.EditorState) {
		state.DeleteRange(s, func(params state.LocatorParams) (uint64, uint64) {
			return locate.WordObject(params.TextTree, params.CursorPos, count)
		}, clipboardPage)
		EnterInsertMode(s)
	}
}

func ChangeInnerWord(count uint64, clipboardPage clipboard.PageId) Action {
	return func(s *state.EditorState) {
		state.DeleteRange(s, func(params state.LocatorParams) (uint64, uint64) {
			return locate.InnerWordObject(params.TextTree, params.CursorPos, count)
		}, clipboardPage)
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

func IndentLine(count uint64) Action {
	return func(s *state.EditorState) {
		targetLineLoc := func(p state.LocatorParams) uint64 {
			return locate.StartOfLineBelow(p.TextTree, count-1, p.CursorPos)
		}
		state.IndentLines(s, targetLineLoc, 1)
	}
}

func OutdentLine(count uint64) Action {
	return func(s *state.EditorState) {
		targetLineLoc := func(p state.LocatorParams) uint64 {
			return locate.StartOfLineBelow(p.TextTree, count-1, p.CursorPos)
		}
		state.OutdentLines(s, targetLineLoc, 1)
	}
}

func CopyToStartOfNextWord(count uint64, clipboardPage clipboard.PageId, withPunctuation bool) Action {
	return func(s *state.EditorState) {
		state.CopyRange(s, clipboardPage, func(params state.LocatorParams) (uint64, uint64) {
			startPos := params.CursorPos
			endPos := locate.NextWordStart(params.TextTree, params.CursorPos, count, withPunctuation, true)
			return startPos, endPos
		})
	}
}

func CopyAWord(count uint64, clipboardPage clipboard.PageId) Action {
	return func(s *state.EditorState) {
		state.CopyRange(s, clipboardPage, func(params state.LocatorParams) (uint64, uint64) {
			return locate.WordObject(params.TextTree, params.CursorPos, count)
		})
	}
}

func CopyInnerWord(count uint64, clipboardPage clipboard.PageId) Action {
	return func(s *state.EditorState) {
		state.CopyRange(s, clipboardPage, func(params state.LocatorParams) (uint64, uint64) {
			return locate.InnerWordObject(params.TextTree, params.CursorPos, count)
		})
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

func ShowCommandMenu(ctx Context) Action {
	return func(s *state.EditorState) {
		// This sets the input mode to menu.
		state.ShowMenu(s, state.MenuStyleCommand, menuItems(ctx))
	}
}

func ShowFileMenu(ctx Context) Action {
	return func(s *state.EditorState) {
		state.ShowFileMenu(s, ctx.DirPatternsToHide)
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

func SearchWordUnderCursor(direction state.SearchDirection, count uint64) Action {
	return func(s *state.EditorState) {
		state.SearchWordUnderCursor(s, direction, count)
	}
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
			state.DeleteToPos(s, selectionEndLoc, clipboardPage)
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

func IndentSelectionAndReturnToNormalMode(selectionEndLoc state.Locator, count uint64) Action {
	return func(s *state.EditorState) {
		state.MoveCursorToStartOfSelection(s)
		state.IndentLines(s, selectionEndLoc, count)
		ReturnToNormalMode(s)
	}
}

func OutdentSelectionAndReturnToNormalMode(selectionEndLoc state.Locator, count uint64) Action {
	return func(s *state.EditorState) {
		state.MoveCursorToStartOfSelection(s)
		state.OutdentLines(s, selectionEndLoc, count)
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

func SelectInnerWord(count uint64) Action {
	return func(s *state.EditorState) {
		state.SelectRange(s, func(params state.LocatorParams) (uint64, uint64) {
			return locate.InnerWordObject(params.TextTree, params.CursorPos, count)
		})
	}
}

func SelectAWord(count uint64) Action {
	return func(s *state.EditorState) {
		state.SelectRange(s, func(params state.LocatorParams) (uint64, uint64) {
			return locate.WordObject(params.TextTree, params.CursorPos, count)
		})
	}
}

func SelectParenBlock(includeParens bool) Action {
	return func(s *state.EditorState) {
		state.SelectRange(s, func(params state.LocatorParams) (uint64, uint64) {
			return locate.DelimitedBlock(locate.ParenPair, params.TextTree, params.SyntaxParser, includeParens, params.CursorPos)
		})
	}
}

func SelectBraceBlock(includeBraces bool) Action {
	return func(s *state.EditorState) {
		state.SelectRange(s, func(params state.LocatorParams) (uint64, uint64) {
			return locate.DelimitedBlock(locate.BracePair, params.TextTree, params.SyntaxParser, includeBraces, params.CursorPos)
		})
	}
}

func SelectAngleBlock(includeAngleBrackets bool) Action {
	return func(s *state.EditorState) {
		state.SelectRange(s, func(params state.LocatorParams) (uint64, uint64) {
			return locate.DelimitedBlock(locate.AnglePair, params.TextTree, params.SyntaxParser, includeAngleBrackets, params.CursorPos)
		})
	}
}

func ReplayLastActionMacro(count uint64) Action {
	return func(s *state.EditorState) {
		state.ReplayLastActionMacro(s, count)
	}
}
