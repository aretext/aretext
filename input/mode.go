package input

import (
	"regexp"
	"strconv"

	"log"

	"github.com/aretext/aretext/exec"
	"github.com/aretext/aretext/text"
	"github.com/gdamore/tcell"
)

// Mode represents an input mode, which is a way of interpreting key events.
type Mode interface {
	// ProcessKeyEvent interprets the key event according to this mode.
	// It will return any user-initiated mutator resulting from the keypress
	ProcessKeyEvent(event *tcell.EventKey, config Config) exec.Mutator
}

// normalMode is used for navigating text.
type normalMode struct {
	buffer []rune
}

func newNormalMode() Mode {
	return &normalMode{
		buffer: make([]rune, 0, 8),
	}
}

func (m *normalMode) ProcessKeyEvent(event *tcell.EventKey, config Config) exec.Mutator {
	mutator := m.processKeyEvent(event, config)
	if mutator != nil {
		m.buffer = m.buffer[:0]
	}
	return appendScrollToCursor(mutator)
}

func (m *normalMode) processKeyEvent(event *tcell.EventKey, config Config) exec.Mutator {
	if event.Key() == tcell.KeyRune {
		return m.processRuneKey(event.Rune())
	}
	return m.processSpecialKey(event.Key(), config)
}

func (m *normalMode) processSpecialKey(key tcell.Key, config Config) exec.Mutator {
	switch key {
	case tcell.KeyLeft:
		return m.cursorLeft()
	case tcell.KeyRight:
		return m.cursorRight(false)
	case tcell.KeyUp:
		return m.cursorUp()
	case tcell.KeyDown:
		return m.cursorDown()
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		return m.cursorBack()
	case tcell.KeyCtrlU:
		return m.scrollUp(config.ScrollLines)
	case tcell.KeyCtrlD:
		return m.scrollDown(config.ScrollLines)
	default:
		return nil
	}
}

func (m *normalMode) processRuneKey(r rune) exec.Mutator {
	m.buffer = append(m.buffer, r)
	count, cmd := m.parseSequence(m.buffer)

	switch cmd {
	case ":":
		return m.showCommandMenu()
	case "h":
		return m.cursorLeft()
	case "l":
		return m.cursorRight(false)
	case "k":
		return m.cursorUp()
	case "j":
		return m.cursorDown()
	case "w":
		return m.cursorNextWordStart()
	case "b":
		return m.cursorPrevWordStart()
	case "{":
		return m.cursorPrevParagraph()
	case "}":
		return m.cursorNextParagraph()
	case "x":
		return m.deleteNextCharInLine()
	case "0":
		return m.cursorLineStart()
	case "^":
		return m.cursorLineStartNonWhitespace()
	case "$":
		return m.cursorLineEnd(false)
	case "gg":
		return m.cursorStartOfLineNum(count)
	case "G":
		return m.cursorStartOfLastLine()
	case "i":
		return m.enterInsertMode()
	case "I":
		return m.enterInsertModeAtStartOfLine()
	case "a":
		return m.enterInsertModeAtNextPos()
	case "A":
		return m.enterInsertModeAtEndOfLine()
	case "o":
		return m.beginNewLineBelow()
	case "O":
		return m.beginNewLineAbove()
	case "dd":
		return m.deleteLine()
	case "dh":
		return m.deletePrevCharInLine()
	case "dj":
		return m.deleteDown()
	case "dk":
		return m.deleteUp()
	case "dl":
		return m.deleteNextCharInLine()
	case "d$":
		return m.deleteToEndOfLine()
	case "d0":
		return m.deleteToStartOfLine()
	case "d^":
		return m.deleteToStartOfLineNonWhitespace()
	case "D":
		return m.deleteToEndOfLine()
	default:
		return nil
	}
}

var normalModeSequenceRegex = regexp.MustCompile(`(?P<count>[1-9][0-9]*)?(?P<command>:|h|l|k|j|w|b|\{|\}|x|^[1-9]?0|\^|\$|gg|G|i|I|a|A|o|O|d[dhjkl0$^]|D)$`)

func (m *normalMode) parseSequence(seq []rune) (uint64, string) {
	submatches := normalModeSequenceRegex.FindStringSubmatch(string(seq))

	if submatches == nil {
		return 0, ""
	}

	countStr, cmdStr := submatches[1], submatches[2]
	if len(countStr) > 0 {
		count, err := strconv.ParseUint(countStr, 10, 64)
		if err != nil {
			log.Fatalf("%s", err)
		}
		return count, cmdStr
	}

	return 0, cmdStr
}

func (m *normalMode) showCommandMenu() exec.Mutator {
	// The show menu mutator sets the input mode to menu.
	return exec.NewCompositeMutator([]exec.Mutator{
		exec.NewSetStatusMsgMutator(exec.StatusMsg{}),
		exec.NewShowMenuMutator("command", commandMenuItems, false, true),
	})
}

func (m *normalMode) cursorLeft() exec.Mutator {
	loc := exec.NewCharInLineLocator(text.ReadDirectionBackward, 1, false)
	return exec.NewCursorMutator(loc)
}

func (m *normalMode) cursorBack() exec.Mutator {
	loc := exec.NewPrevCharLocator(1)
	return exec.NewCursorMutator(loc)
}

func (m *normalMode) cursorRight(allowPastEndOfLineOrFile bool) exec.Mutator {
	loc := exec.NewCharInLineLocator(text.ReadDirectionForward, 1, allowPastEndOfLineOrFile)
	return exec.NewCursorMutator(loc)
}

func (m *normalMode) cursorUp() exec.Mutator {
	loc := exec.NewRelativeLineLocator(text.ReadDirectionBackward, 1)
	return exec.NewCursorMutator(loc)
}

func (m *normalMode) cursorDown() exec.Mutator {
	loc := exec.NewRelativeLineLocator(text.ReadDirectionForward, 1)
	return exec.NewCursorMutator(loc)
}

func (m *normalMode) cursorNextWordStart() exec.Mutator {
	loc := exec.NewNextWordStartLocator()
	return exec.NewCursorMutator(loc)
}

func (m *normalMode) cursorPrevWordStart() exec.Mutator {
	loc := exec.NewPrevWordStartLocator()
	return exec.NewCursorMutator(loc)
}

func (m *normalMode) cursorPrevParagraph() exec.Mutator {
	loc := exec.NewPrevParagraphLocator()
	return exec.NewCursorMutator(loc)
}

func (m *normalMode) cursorNextParagraph() exec.Mutator {
	loc := exec.NewNextParagraphLocator()
	return exec.NewCursorMutator(loc)
}

func (m *normalMode) scrollUp(scrollLines uint64) exec.Mutator {
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

func (m *normalMode) scrollDown(scrollLines uint64) exec.Mutator {
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

func (m *normalMode) cursorLineStart() exec.Mutator {
	loc := exec.NewLineBoundaryLocator(text.ReadDirectionBackward, false)
	return exec.NewCursorMutator(loc)
}

func (m *normalMode) cursorLineStartNonWhitespace() exec.Mutator {
	lineStartLoc := exec.NewLineBoundaryLocator(text.ReadDirectionBackward, false)
	firstNonWhitespaceLoc := exec.NewNonWhitespaceOrNewlineLocator(lineStartLoc)
	return exec.NewCursorMutator(firstNonWhitespaceLoc)
}

func (m *normalMode) cursorLineEnd(includeEndOfLineOrFile bool) exec.Mutator {
	loc := exec.NewLineBoundaryLocator(text.ReadDirectionForward, includeEndOfLineOrFile)
	return exec.NewCursorMutator(loc)
}

func (m *normalMode) cursorStartOfLineNum(count uint64) exec.Mutator {
	// Convert 1-indexed count to 0-indexed line num
	var lineNum uint64
	if count > 0 {
		lineNum = count - 1
	}

	lineNumLoc := exec.NewLineNumLocator(lineNum)
	firstNonWhitespaceLoc := exec.NewNonWhitespaceOrNewlineLocator(lineNumLoc)
	return exec.NewCursorMutator(firstNonWhitespaceLoc)
}

func (m *normalMode) cursorStartOfLastLine() exec.Mutator {
	lastLineLoc := exec.NewLastLineLocator()
	firstNonWhitespaceLoc := exec.NewNonWhitespaceOrNewlineLocator(lastLineLoc)
	return exec.NewCursorMutator(firstNonWhitespaceLoc)
}

func (m *normalMode) enterInsertMode() exec.Mutator {
	return exec.NewCompositeMutator([]exec.Mutator{
		exec.NewSetStatusMsgMutator(exec.StatusMsg{}),
		exec.NewSetInputModeMutator(exec.InputModeInsert),
	})
}

func (m *normalMode) enterInsertModeAtStartOfLine() exec.Mutator {
	return exec.NewCompositeMutator([]exec.Mutator{
		exec.NewSetStatusMsgMutator(exec.StatusMsg{}),
		exec.NewSetInputModeMutator(exec.InputModeInsert),
		m.cursorLineStartNonWhitespace(),
	})
}

func (m *normalMode) enterInsertModeAtNextPos() exec.Mutator {
	return exec.NewCompositeMutator([]exec.Mutator{
		exec.NewSetStatusMsgMutator(exec.StatusMsg{}),
		exec.NewSetInputModeMutator(exec.InputModeInsert),
		m.cursorRight(true),
	})
}

func (m *normalMode) enterInsertModeAtEndOfLine() exec.Mutator {
	return exec.NewCompositeMutator([]exec.Mutator{
		exec.NewSetStatusMsgMutator(exec.StatusMsg{}),
		exec.NewSetInputModeMutator(exec.InputModeInsert),
		m.cursorLineEnd(true),
	})
}

func (m *normalMode) beginNewLineBelow() exec.Mutator {
	return exec.NewCompositeMutator([]exec.Mutator{
		m.cursorLineEnd(true),
		exec.NewInsertNewlineMutator(),
		exec.NewSetInputModeMutator(exec.InputModeInsert),
	})
}

func (m *normalMode) beginNewLineAbove() exec.Mutator {
	return exec.NewCompositeMutator([]exec.Mutator{
		m.cursorLineStart(),
		exec.NewInsertNewlineMutator(),
		m.cursorUp(),
		exec.NewSetInputModeMutator(exec.InputModeInsert),
	})
}

func (m *normalMode) deleteLine() exec.Mutator {
	return exec.NewCompositeMutator([]exec.Mutator{
		exec.NewDeleteLinesMutator(exec.NewCurrentCursorLocator(), false),
		m.cursorLineStartNonWhitespace(),
	})
}

func (m *normalMode) deletePrevChar() exec.Mutator {
	loc := exec.NewCharInLineLocator(text.ReadDirectionBackward, 1, true)
	return exec.NewDeleteMutator(loc)
}

func (m *normalMode) deletePrevCharInLine() exec.Mutator {
	loc := exec.NewCharInLineLocator(text.ReadDirectionBackward, 1, false)
	return exec.NewDeleteMutator(loc)
}

func (m *normalMode) deleteNextCharInLine() exec.Mutator {
	loc := exec.NewCharInLineLocator(text.ReadDirectionForward, 1, true)
	return exec.NewCompositeMutator([]exec.Mutator{
		exec.NewDeleteMutator(loc),
		exec.NewCursorMutator(exec.NewOntoLineLocator()),
	})
}

func (m *normalMode) deleteDown() exec.Mutator {
	targetLineLoc := exec.NewRelativeLineStartLocator(text.ReadDirectionForward, 1)
	return exec.NewCompositeMutator([]exec.Mutator{
		exec.NewDeleteLinesMutator(targetLineLoc, true),
		m.cursorLineStartNonWhitespace(),
	})
}

func (m *normalMode) deleteUp() exec.Mutator {
	targetLineLoc := exec.NewRelativeLineStartLocator(text.ReadDirectionBackward, 1)
	return exec.NewCompositeMutator([]exec.Mutator{
		exec.NewDeleteLinesMutator(targetLineLoc, true),
		m.cursorLineStartNonWhitespace(),
	})
}

func (m *normalMode) deleteToEndOfLine() exec.Mutator {
	loc := exec.NewLineBoundaryLocator(text.ReadDirectionForward, true)
	return exec.NewCompositeMutator([]exec.Mutator{
		exec.NewDeleteMutator(loc),
		exec.NewCursorMutator(exec.NewOntoLineLocator()),
	})
}

func (m *normalMode) deleteToStartOfLine() exec.Mutator {
	loc := exec.NewLineBoundaryLocator(text.ReadDirectionBackward, false)
	return exec.NewDeleteMutator(loc)
}

func (m *normalMode) deleteToStartOfLineNonWhitespace() exec.Mutator {
	lineStartLoc := exec.NewLineBoundaryLocator(text.ReadDirectionBackward, false)
	firstNonWhitespaceLoc := exec.NewNonWhitespaceOrNewlineLocator(lineStartLoc)
	return exec.NewDeleteMutator(firstNonWhitespaceLoc)
}

// insertMode is used for inserting characters into text.
type insertMode struct{}

func newInsertMode() Mode {
	return &insertMode{}
}

func (m *insertMode) ProcessKeyEvent(event *tcell.EventKey, config Config) exec.Mutator {
	mutator := m.processKeyEvent(event)
	return appendScrollToCursor(mutator)
}

func (m *insertMode) processKeyEvent(event *tcell.EventKey) exec.Mutator {
	switch event.Key() {
	case tcell.KeyRune:
		return m.insertRune(event.Rune())
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		return m.deletePrevChar()
	case tcell.KeyEnter:
		return m.insertNewline()
	case tcell.KeyTab:
		return m.insertTab()
	default:
		return m.returnToNormalMode()
	}
}

func (m *insertMode) insertRune(r rune) exec.Mutator {
	return exec.NewInsertRuneMutator(r)
}

func (m *insertMode) insertNewline() exec.Mutator {
	return exec.NewInsertNewlineMutator()
}

func (m *insertMode) insertTab() exec.Mutator {
	return exec.NewInsertTabMutator()
}

func (m *insertMode) deletePrevChar() exec.Mutator {
	loc := exec.NewMinPosLocator([]exec.CursorLocator{
		exec.NewCharInLineLocator(text.ReadDirectionBackward, 1, true),
		exec.NewPrevAutoIndentLocator(),
	})
	return exec.NewDeleteMutator(loc)
}

// appendScrollToCursor appends a mutator to scroll the view so the cursor is visible.
func appendScrollToCursor(mutator exec.Mutator) exec.Mutator {
	if mutator == nil {
		return nil
	}

	return exec.NewCompositeMutator([]exec.Mutator{
		mutator,
		exec.NewScrollToCursorMutator(),
	})
}

func (m *insertMode) returnToNormalMode() exec.Mutator {
	loc := exec.NewOntoLineLocator()
	return exec.NewCompositeMutator([]exec.Mutator{
		exec.NewCursorMutator(loc),
		exec.NewSetInputModeMutator(exec.InputModeNormal),
	})
}

// menuMode allows the user to search for and select items in a menu.
type menuMode struct{}

func newMenuMode() Mode {
	return &menuMode{}
}

func (m *menuMode) ProcessKeyEvent(event *tcell.EventKey, config Config) exec.Mutator {
	switch event.Key() {
	case tcell.KeyEscape:
		return m.closeMenu()
	case tcell.KeyEnter:
		return m.executeSelectedMenuItem()
	case tcell.KeyUp:
		return m.menuSelectionUp()
	case tcell.KeyDown:
		return m.menuSelectionDown()
	case tcell.KeyTab:
		return m.menuSelectionDown()
	case tcell.KeyRune:
		return m.appendMenuSearch(event.Rune())
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		return m.deleteMenuSearch()
	default:
		return nil
	}
}

func (m *menuMode) closeMenu() exec.Mutator {
	// Returns to normal mode.
	return exec.NewHideMenuMutator()
}

func (m *menuMode) executeSelectedMenuItem() exec.Mutator {
	// Hides the menu, then executes the menu item action.
	// This usually returns to normal mode, unless the menu item action sets a different mode.
	return exec.NewExecuteSelectedMenuItemMutator()
}

func (m *menuMode) menuSelectionUp() exec.Mutator {
	return exec.NewMoveMenuSelectionMutator(-1)
}

func (m *menuMode) menuSelectionDown() exec.Mutator {
	return exec.NewMoveMenuSelectionMutator(1)
}

func (m *menuMode) appendMenuSearch(r rune) exec.Mutator {
	return exec.NewAppendMenuSearchMutator(r)
}

func (m *menuMode) deleteMenuSearch() exec.Mutator {
	return exec.NewDeleteMenuSearchMutator()
}
