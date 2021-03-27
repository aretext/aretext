package input

import (
	"log"

	"github.com/aretext/aretext/exec"
	"github.com/aretext/aretext/locate"
	"github.com/aretext/aretext/text"
	"github.com/gdamore/tcell/v2"
)

func CursorLeft(inputEvents []*tcell.EventKey, count *int64, config Config) exec.Mutator {
	return exec.NewCursorMutator(func(params exec.LocatorParams) uint64 {
		return locate.PrevCharInLine(params.TextTree, 1, false, params.CursorPos)
	})
}

func CursorBack(inputEvents []*tcell.EventKey, count *int64, config Config) exec.Mutator {
	return exec.NewCursorMutator(func(params exec.LocatorParams) uint64 {
		return locate.PrevChar(params.TextTree, 1, params.CursorPos)
	})
}

func CursorRight(inputEvents []*tcell.EventKey, count *int64, config Config) exec.Mutator {
	return exec.NewCursorMutator(func(params exec.LocatorParams) uint64 {
		return locate.NextCharInLine(params.TextTree, 1, false, params.CursorPos)
	})
}

func CursorRightIncludeEndOfLineOrFile(inputEvents []*tcell.EventKey, count *int64, config Config) exec.Mutator {
	return exec.NewCursorMutator(func(params exec.LocatorParams) uint64 {
		return locate.NextCharInLine(params.TextTree, 1, true, params.CursorPos)
	})
}

func CursorUp(inputEvents []*tcell.EventKey, count *int64, config Config) exec.Mutator {
	return exec.NewCursorLineAboveMutator(1)
}

func CursorDown(inputEvents []*tcell.EventKey, count *int64, config Config) exec.Mutator {
	return exec.NewCursorLineBelowMutator(1)
}

func CursorNextWordStart(inputEvents []*tcell.EventKey, count *int64, config Config) exec.Mutator {
	return exec.NewCursorMutator(func(params exec.LocatorParams) uint64 {
		return locate.NextWordStart(params.TextTree, params.TokenTree, params.CursorPos)
	})
}

func CursorPrevWordStart(inputEvents []*tcell.EventKey, count *int64, config Config) exec.Mutator {
	return exec.NewCursorMutator(func(params exec.LocatorParams) uint64 {
		return locate.PrevWordStart(params.TextTree, params.TokenTree, params.CursorPos)
	})
}

func CursorNextWordEnd(inputEvents []*tcell.EventKey, count *int64, config Config) exec.Mutator {
	return exec.NewCursorMutator(func(params exec.LocatorParams) uint64 {
		return locate.NextWordEnd(params.TextTree, params.TokenTree, params.CursorPos)
	})
}

func CursorPrevParagraph(inputEvents []*tcell.EventKey, count *int64, config Config) exec.Mutator {
	return exec.NewCursorMutator(func(params exec.LocatorParams) uint64 {
		return locate.PrevParagraph(params.TextTree, params.CursorPos)
	})
}

func CursorNextParagraph(inputEvents []*tcell.EventKey, count *int64, config Config) exec.Mutator {
	return exec.NewCursorMutator(func(params exec.LocatorParams) uint64 {
		return locate.NextParagraph(params.TextTree, params.CursorPos)
	})
}

func ScrollUp(inputEvents []*tcell.EventKey, count *int64, config Config) exec.Mutator {
	scrollLines := config.ScrollLines
	if scrollLines < 1 {
		scrollLines = 1
	}

	// Move the cursor to the start of a line above, then scroll up.
	// We rely on ScrollToCursorMutator (appended later to every mutator)
	// to reposition the view.
	return exec.NewCompositeMutator([]exec.Mutator{
		exec.NewCursorMutator(func(params exec.LocatorParams) uint64 {
			return locate.StartOfLineAbove(params.TextTree, scrollLines, params.CursorPos)
		}),
		exec.NewScrollLinesMutator(text.ReadDirectionBackward, scrollLines),
	})
}

func ScrollDown(inputEvents []*tcell.EventKey, count *int64, config Config) exec.Mutator {
	scrollLines := config.ScrollLines
	if scrollLines < 1 {
		scrollLines = 1
	}

	// Move the cursor to the start of a line below, then scroll down.
	// We rely on ScrollToCursorMutator (appended later to every mutator)
	// to reposition the view.
	return exec.NewCompositeMutator([]exec.Mutator{
		exec.NewCursorMutator(func(params exec.LocatorParams) uint64 {
			return locate.StartOfLineBelow(params.TextTree, scrollLines, params.CursorPos)
		}),
		exec.NewScrollLinesMutator(text.ReadDirectionForward, scrollLines),
	})
}

func CursorLineStart(inputEvents []*tcell.EventKey, count *int64, config Config) exec.Mutator {
	return exec.NewCursorMutator(func(params exec.LocatorParams) uint64 {
		return locate.PrevLineBoundary(params.TextTree, params.CursorPos)
	})
}

func CursorLineStartNonWhitespace(inputEvents []*tcell.EventKey, count *int64, config Config) exec.Mutator {
	return exec.NewCursorMutator(func(params exec.LocatorParams) uint64 {
		lineStartPos := locate.PrevLineBoundary(params.TextTree, params.CursorPos)
		return locate.NextNonWhitespaceOrNewline(params.TextTree, lineStartPos)
	})
}

func CursorLineEnd(inputEvents []*tcell.EventKey, count *int64, config Config) exec.Mutator {
	return exec.NewCursorMutator(func(params exec.LocatorParams) uint64 {
		return locate.NextLineBoundary(params.TextTree, false, params.CursorPos)
	})
}

func CursorLineEndIncludeEndOfLineOrFile(inputEvents []*tcell.EventKey, count *int64, config Config) exec.Mutator {
	return exec.NewCursorMutator(func(params exec.LocatorParams) uint64 {
		return locate.NextLineBoundary(params.TextTree, true, params.CursorPos)
	})
}

func CursorStartOfLineNum(inputEvents []*tcell.EventKey, count *int64, config Config) exec.Mutator {
	// Convert 1-indexed count to 0-indexed line num
	var lineNum uint64
	if count != nil && *count > 0 {
		lineNum = uint64(*count - 1)
	}

	return exec.NewCursorMutator(func(params exec.LocatorParams) uint64 {
		lineStartPos := locate.StartOfLineNum(params.TextTree, lineNum)
		return locate.NextNonWhitespaceOrNewline(params.TextTree, lineStartPos)
	})
}

func CursorStartOfLastLine(inputEvents []*tcell.EventKey, count *int64, config Config) exec.Mutator {
	return exec.NewCursorMutator(func(params exec.LocatorParams) uint64 {
		lineStartPos := locate.StartOfLastLine(params.TextTree)
		return locate.NextNonWhitespaceOrNewline(params.TextTree, lineStartPos)
	})
}

func EnterInsertMode(inputEvents []*tcell.EventKey, count *int64, config Config) exec.Mutator {
	return exec.NewCompositeMutator([]exec.Mutator{
		exec.NewSetStatusMsgMutator(exec.StatusMsg{}),
		exec.NewSetInputModeMutator(exec.InputModeInsert),
	})
}

func EnterInsertModeAtStartOfLine(inputEvents []*tcell.EventKey, count *int64, config Config) exec.Mutator {
	return exec.NewCompositeMutator([]exec.Mutator{
		exec.NewSetStatusMsgMutator(exec.StatusMsg{}),
		exec.NewSetInputModeMutator(exec.InputModeInsert),
		CursorLineStartNonWhitespace(nil, nil, config),
	})
}

func EnterInsertModeAtNextPos(inputEvents []*tcell.EventKey, count *int64, config Config) exec.Mutator {
	return exec.NewCompositeMutator([]exec.Mutator{
		exec.NewSetStatusMsgMutator(exec.StatusMsg{}),
		exec.NewSetInputModeMutator(exec.InputModeInsert),
		CursorRightIncludeEndOfLineOrFile(nil, nil, config),
	})
}

func EnterInsertModeAtEndOfLine(inputEvents []*tcell.EventKey, count *int64, config Config) exec.Mutator {
	return exec.NewCompositeMutator([]exec.Mutator{
		exec.NewSetStatusMsgMutator(exec.StatusMsg{}),
		exec.NewSetInputModeMutator(exec.InputModeInsert),
		CursorLineEndIncludeEndOfLineOrFile(nil, nil, config),
	})
}

func BeginNewLineBelow(inputEvents []*tcell.EventKey, count *int64, config Config) exec.Mutator {
	return exec.NewCompositeMutator([]exec.Mutator{
		CursorLineEndIncludeEndOfLineOrFile(nil, nil, config),
		exec.NewInsertNewlineMutator(),
		exec.NewSetInputModeMutator(exec.InputModeInsert),
	})
}

func BeginNewLineAbove(inputEvents []*tcell.EventKey, count *int64, config Config) exec.Mutator {
	return exec.NewCompositeMutator([]exec.Mutator{
		CursorLineStart(nil, nil, config),
		exec.NewInsertNewlineMutator(),
		CursorUp(nil, nil, config),
		exec.NewSetInputModeMutator(exec.InputModeInsert),
	})
}

func DeleteLine(inputEvents []*tcell.EventKey, count *int64, config Config) exec.Mutator {
	currentPos := func(params exec.LocatorParams) uint64 {
		return params.CursorPos
	}
	return exec.NewCompositeMutator([]exec.Mutator{
		exec.NewDeleteLinesMutator(currentPos, false),
		CursorLineStartNonWhitespace(nil, nil, config),
	})
}

func DeletePrevCharInLine(inputEvents []*tcell.EventKey, count *int64, config Config) exec.Mutator {
	return exec.NewDeleteMutator(func(params exec.LocatorParams) uint64 {
		return locate.PrevCharInLine(params.TextTree, 1, false, params.CursorPos)
	})
}

func DeleteNextCharInLine(inputEvents []*tcell.EventKey, count *int64, config Config) exec.Mutator {
	return exec.NewCompositeMutator([]exec.Mutator{
		exec.NewDeleteMutator(func(params exec.LocatorParams) uint64 {
			return locate.NextCharInLine(params.TextTree, 1, true, params.CursorPos)
		}),
		exec.NewCursorMutator(func(params exec.LocatorParams) uint64 {
			return locate.ClosestCharOnLine(params.TextTree, params.CursorPos)
		}),
	})
}

func DeleteDown(inputEvents []*tcell.EventKey, count *int64, config Config) exec.Mutator {
	targetLineLoc := func(params exec.LocatorParams) uint64 {
		return locate.StartOfLineBelow(params.TextTree, 1, params.CursorPos)
	}
	return exec.NewCompositeMutator([]exec.Mutator{
		exec.NewDeleteLinesMutator(targetLineLoc, true),
		CursorLineStartNonWhitespace(nil, nil, config),
	})
}

func DeleteUp(inputEvents []*tcell.EventKey, count *int64, config Config) exec.Mutator {
	targetLineLoc := func(params exec.LocatorParams) uint64 {
		return locate.StartOfLineAbove(params.TextTree, 1, params.CursorPos)
	}
	return exec.NewCompositeMutator([]exec.Mutator{
		exec.NewDeleteLinesMutator(targetLineLoc, true),
		CursorLineStartNonWhitespace(nil, nil, config),
	})
}

func DeleteToEndOfLine(inputEvents []*tcell.EventKey, count *int64, config Config) exec.Mutator {
	return exec.NewCompositeMutator([]exec.Mutator{
		exec.NewDeleteMutator(func(params exec.LocatorParams) uint64 {
			return locate.NextLineBoundary(params.TextTree, true, params.CursorPos)
		}),
		exec.NewCursorMutator(func(params exec.LocatorParams) uint64 {
			return locate.ClosestCharOnLine(params.TextTree, params.CursorPos)
		}),
	})
}

func DeleteToStartOfLine(inputEvents []*tcell.EventKey, count *int64, config Config) exec.Mutator {
	return exec.NewDeleteMutator(func(params exec.LocatorParams) uint64 {
		return locate.PrevLineBoundary(params.TextTree, params.CursorPos)
	})
}

func DeleteToStartOfLineNonWhitespace(inputEvents []*tcell.EventKey, count *int64, config Config) exec.Mutator {
	return exec.NewDeleteMutator(func(params exec.LocatorParams) uint64 {
		lineStartPos := locate.PrevLineBoundary(params.TextTree, params.CursorPos)
		return locate.NextNonWhitespaceOrNewline(params.TextTree, lineStartPos)
	})
}

func DeleteInnerWord(inputEvents []*tcell.EventKey, count *int64, config Config) exec.Mutator {
	return exec.NewCompositeMutator([]exec.Mutator{
		exec.NewCursorMutator(func(params exec.LocatorParams) uint64 {
			return locate.CurrentWordStart(params.TextTree, params.TokenTree, params.CursorPos)
		}),
		exec.NewDeleteMutator(func(params exec.LocatorParams) uint64 {
			return locate.CurrentWordEnd(params.TextTree, params.TokenTree, params.CursorPos)
		}),
	})
}

func ChangeInnerWord(inputEvents []*tcell.EventKey, count *int64, config Config) exec.Mutator {
	return exec.NewCompositeMutator([]exec.Mutator{
		DeleteInnerWord(nil, nil, config),
		EnterInsertMode(nil, nil, config),
	})
}

func ReplaceCharacter(inputEvents []*tcell.EventKey, count *int64, config Config) exec.Mutator {
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
	return exec.NewReplaceCharMutator(newChar)
}

func ShowCommandMenu(inputEvents []*tcell.EventKey, count *int64, config Config) exec.Mutator {
	// The show menu mutator sets the input mode to menu.
	return exec.NewCompositeMutator([]exec.Mutator{
		exec.NewSetStatusMsgMutator(exec.StatusMsg{}),
		exec.NewShowMenuMutator("command", commandMenuItems(config), false, true),
	})
}

func StartSearchForward(inputEvents []*tcell.EventKey, count *int64, config Config) exec.Mutator {
	// The start search mutator sets the input mode to search.
	return exec.NewCompositeMutator([]exec.Mutator{
		exec.NewSetStatusMsgMutator(exec.StatusMsg{}),
		exec.NewStartSearchMutator(text.ReadDirectionForward),
	})
}

func StartSearchBackward(inputEvents []*tcell.EventKey, count *int64, config Config) exec.Mutator {
	// The start search mutator sets the input mode to search.
	return exec.NewCompositeMutator([]exec.Mutator{
		exec.NewSetStatusMsgMutator(exec.StatusMsg{}),
		exec.NewStartSearchMutator(text.ReadDirectionBackward),
	})
}

func FindNextMatch(inputEvents []*tcell.EventKey, count *int64, config Config) exec.Mutator {
	return exec.NewFindNextMatchMutator(false)
}

func FindPrevMatch(inputEvents []*tcell.EventKey, count *int64, config Config) exec.Mutator {
	reverse := true
	return exec.NewFindNextMatchMutator(reverse)
}
