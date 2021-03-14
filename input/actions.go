package input

import (
	"log"

	"github.com/aretext/aretext/exec"
	"github.com/aretext/aretext/text"
	"github.com/gdamore/tcell/v2"
)

func CursorLeft(inputEvents []*tcell.EventKey, count *int64, config Config) exec.Mutator {
	loc := exec.NewCharInLineLocator(text.ReadDirectionBackward, 1, false)
	return exec.NewCursorMutator(loc)
}

func CursorBack(inputEvents []*tcell.EventKey, count *int64, config Config) exec.Mutator {
	loc := exec.NewPrevCharLocator(1)
	return exec.NewCursorMutator(loc)
}

func CursorRight(inputEvents []*tcell.EventKey, count *int64, config Config) exec.Mutator {
	loc := exec.NewCharInLineLocator(text.ReadDirectionForward, 1, false)
	return exec.NewCursorMutator(loc)
}

func CursorRightIncludeEndOfLineOrFile(inputEvents []*tcell.EventKey, count *int64, config Config) exec.Mutator {
	loc := exec.NewCharInLineLocator(text.ReadDirectionForward, 1, true)
	return exec.NewCursorMutator(loc)
}

func CursorUp(inputEvents []*tcell.EventKey, count *int64, config Config) exec.Mutator {
	loc := exec.NewRelativeLineLocator(text.ReadDirectionBackward, 1)
	return exec.NewCursorMutator(loc)
}

func CursorDown(inputEvents []*tcell.EventKey, count *int64, config Config) exec.Mutator {
	loc := exec.NewRelativeLineLocator(text.ReadDirectionForward, 1)
	return exec.NewCursorMutator(loc)
}

func CursorNextWordStart(inputEvents []*tcell.EventKey, count *int64, config Config) exec.Mutator {
	loc := exec.NewNextWordStartLocator()
	return exec.NewCursorMutator(loc)
}

func CursorPrevWordStart(inputEvents []*tcell.EventKey, count *int64, config Config) exec.Mutator {
	loc := exec.NewPrevWordStartLocator()
	return exec.NewCursorMutator(loc)
}

func CursorNextWordEnd(inputEvents []*tcell.EventKey, count *int64, config Config) exec.Mutator {
	loc := exec.NewNextWordEndLocator()
	return exec.NewCursorMutator(loc)
}

func CursorPrevParagraph(inputEvents []*tcell.EventKey, count *int64, config Config) exec.Mutator {
	loc := exec.NewPrevParagraphLocator()
	return exec.NewCursorMutator(loc)
}

func CursorNextParagraph(inputEvents []*tcell.EventKey, count *int64, config Config) exec.Mutator {
	loc := exec.NewNextParagraphLocator()
	return exec.NewCursorMutator(loc)
}

func ScrollUp(inputEvents []*tcell.EventKey, count *int64, config Config) exec.Mutator {
	scrollLines := config.ScrollLines
	if scrollLines < 1 {
		scrollLines = 1
	}

	// Move the cursor to the start of a line above.  We rely on ScrollToCursorMutator
	// (appended later to every mutator) to reposition the view.
	loc := exec.NewRelativeLineStartLocator(text.ReadDirectionBackward, scrollLines)
	return exec.NewCompositeMutator([]exec.Mutator{
		exec.NewCursorMutator(loc),
		exec.NewScrollLinesMutator(text.ReadDirectionBackward, scrollLines),
	})
}

func ScrollDown(inputEvents []*tcell.EventKey, count *int64, config Config) exec.Mutator {
	scrollLines := config.ScrollLines
	if scrollLines < 1 {
		scrollLines = 1
	}

	// Move the cursor to the start of a line below.  We rely on ScrollToCursorMutator
	// (appended later to every mutator) to reposition the view.
	loc := exec.NewRelativeLineStartLocator(text.ReadDirectionForward, scrollLines)
	return exec.NewCompositeMutator([]exec.Mutator{
		exec.NewCursorMutator(loc),
		exec.NewScrollLinesMutator(text.ReadDirectionForward, scrollLines),
	})
}

func ShowCommandMenu(inputEvents []*tcell.EventKey, count *int64, config Config) exec.Mutator {
	// The show menu mutator sets the input mode to menu.
	return exec.NewCompositeMutator([]exec.Mutator{
		exec.NewSetStatusMsgMutator(exec.StatusMsg{}),
		exec.NewShowMenuMutator("command", commandMenuItems, false, true),
	})
}

func CursorLineStart(inputEvents []*tcell.EventKey, count *int64, config Config) exec.Mutator {
	loc := exec.NewLineBoundaryLocator(text.ReadDirectionBackward, false)
	return exec.NewCursorMutator(loc)
}

func CursorLineStartNonWhitespace(inputEvents []*tcell.EventKey, count *int64, config Config) exec.Mutator {
	lineStartLoc := exec.NewLineBoundaryLocator(text.ReadDirectionBackward, false)
	firstNonWhitespaceLoc := exec.NewNonWhitespaceOrNewlineLocator(lineStartLoc)
	return exec.NewCursorMutator(firstNonWhitespaceLoc)
}

func CursorLineEnd(inputEvents []*tcell.EventKey, count *int64, config Config) exec.Mutator {
	loc := exec.NewLineBoundaryLocator(text.ReadDirectionForward, false)
	return exec.NewCursorMutator(loc)
}

func CursorLineEndIncludeEndOfLineOrFile(inputEvents []*tcell.EventKey, count *int64, config Config) exec.Mutator {
	loc := exec.NewLineBoundaryLocator(text.ReadDirectionForward, true)
	return exec.NewCursorMutator(loc)
}

func CursorStartOfLineNum(inputEvents []*tcell.EventKey, count *int64, config Config) exec.Mutator {
	// Convert 1-indexed count to 0-indexed line num
	var lineNum uint64
	if count != nil && *count > 0 {
		lineNum = uint64(*count - 1)
	}

	lineNumLoc := exec.NewLineNumLocator(lineNum)
	firstNonWhitespaceLoc := exec.NewNonWhitespaceOrNewlineLocator(lineNumLoc)
	return exec.NewCursorMutator(firstNonWhitespaceLoc)
}

func CursorStartOfLastLine(inputEvents []*tcell.EventKey, count *int64, config Config) exec.Mutator {
	lastLineLoc := exec.NewLastLineLocator()
	firstNonWhitespaceLoc := exec.NewNonWhitespaceOrNewlineLocator(lastLineLoc)
	return exec.NewCursorMutator(firstNonWhitespaceLoc)
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
	return exec.NewCompositeMutator([]exec.Mutator{
		exec.NewDeleteLinesMutator(exec.NewCurrentCursorLocator(), false),
		CursorLineStartNonWhitespace(nil, nil, config),
	})
}

func DeletePrevChar(inputEvents []*tcell.EventKey, count *int64, config Config) exec.Mutator {
	loc := exec.NewCharInLineLocator(text.ReadDirectionBackward, 1, true)
	return exec.NewDeleteMutator(loc)
}

func DeletePrevCharInLine(inputEvents []*tcell.EventKey, count *int64, config Config) exec.Mutator {
	loc := exec.NewCharInLineLocator(text.ReadDirectionBackward, 1, false)
	return exec.NewDeleteMutator(loc)
}

func DeleteNextCharInLine(inputEvents []*tcell.EventKey, count *int64, config Config) exec.Mutator {
	loc := exec.NewCharInLineLocator(text.ReadDirectionForward, 1, true)
	return exec.NewCompositeMutator([]exec.Mutator{
		exec.NewDeleteMutator(loc),
		exec.NewCursorMutator(exec.NewOntoLineLocator()),
	})
}

func DeleteDown(inputEvents []*tcell.EventKey, count *int64, config Config) exec.Mutator {
	targetLineLoc := exec.NewRelativeLineStartLocator(text.ReadDirectionForward, 1)
	return exec.NewCompositeMutator([]exec.Mutator{
		exec.NewDeleteLinesMutator(targetLineLoc, true),
		CursorLineStartNonWhitespace(nil, nil, config),
	})
}

func DeleteUp(inputEvents []*tcell.EventKey, count *int64, config Config) exec.Mutator {
	targetLineLoc := exec.NewRelativeLineStartLocator(text.ReadDirectionBackward, 1)
	return exec.NewCompositeMutator([]exec.Mutator{
		exec.NewDeleteLinesMutator(targetLineLoc, true),
		CursorLineStartNonWhitespace(nil, nil, config),
	})
}

func DeleteToEndOfLine(inputEvents []*tcell.EventKey, count *int64, config Config) exec.Mutator {
	loc := exec.NewLineBoundaryLocator(text.ReadDirectionForward, true)
	return exec.NewCompositeMutator([]exec.Mutator{
		exec.NewDeleteMutator(loc),
		exec.NewCursorMutator(exec.NewOntoLineLocator()),
	})
}

func DeleteToStartOfLine(inputEvents []*tcell.EventKey, count *int64, config Config) exec.Mutator {
	loc := exec.NewLineBoundaryLocator(text.ReadDirectionBackward, false)
	return exec.NewDeleteMutator(loc)
}

func DeleteToStartOfLineNonWhitespace(inputEvents []*tcell.EventKey, count *int64, config Config) exec.Mutator {
	lineStartLoc := exec.NewLineBoundaryLocator(text.ReadDirectionBackward, false)
	firstNonWhitespaceLoc := exec.NewNonWhitespaceOrNewlineLocator(lineStartLoc)
	return exec.NewDeleteMutator(firstNonWhitespaceLoc)
}

func DeleteInnerWord(inputEvents []*tcell.EventKey, count *int64, config Config) exec.Mutator {
	startLoc := exec.NewCurrentWordStartLocator()
	endLoc := exec.NewCurrentWordEndLocator()
	return exec.NewCompositeMutator([]exec.Mutator{
		exec.NewCursorMutator(startLoc),
		exec.NewDeleteMutator(endLoc),
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
