package input

import (
	"regexp"
	"strconv"

	"github.com/gdamore/tcell"
	"github.com/wedaly/aretext/internal/pkg/exec"
	"github.com/wedaly/aretext/internal/pkg/text"
)

type ModeType int

const (
	ModeTypeNormal = ModeType(iota)
	ModeTypeInsert
)

// Mode represents an input mode, which is a way of interpreting key events.
type Mode interface {
	// ProcessKeyEvent interprets the key event according to this mode.
	// It will return any user-initiated mutator resulting from the keypress
	// as well as the next input mode (which could be the same as the current mode).
	ProcessKeyEvent(event *tcell.EventKey, config Config) (exec.Mutator, ModeType)
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

func (m *normalMode) ProcessKeyEvent(event *tcell.EventKey, config Config) (exec.Mutator, ModeType) {
	mutator, nextMode := m.processKeyEvent(event, config)
	if mutator != nil {
		m.buffer = m.buffer[:0]
	}

	return appendScrollToCursor(mutator), nextMode
}

func (m *normalMode) processKeyEvent(event *tcell.EventKey, config Config) (mutator exec.Mutator, mode ModeType) {
	if event.Key() == tcell.KeyRune {
		return m.processRuneKey(event.Rune())
	}
	return m.processSpecialKey(event.Key(), config)
}

func (m *normalMode) processSpecialKey(key tcell.Key, config Config) (exec.Mutator, ModeType) {
	switch key {
	case tcell.KeyLeft:
		return m.cursorLeft(), ModeTypeNormal
	case tcell.KeyRight:
		return m.cursorRight(false), ModeTypeNormal
	case tcell.KeyUp:
		return m.cursorUp(), ModeTypeNormal
	case tcell.KeyDown:
		return m.cursorDown(), ModeTypeNormal
	case tcell.KeyCtrlU:
		return m.scrollUp(config.ScrollLines), ModeTypeNormal
	case tcell.KeyCtrlD:
		return m.scrollDown(config.ScrollLines), ModeTypeNormal
	default:
		return nil, ModeTypeNormal
	}
}

func (m *normalMode) processRuneKey(r rune) (exec.Mutator, ModeType) {
	m.buffer = append(m.buffer, r)
	count, cmd := m.parseSequence(m.buffer)

	switch cmd {
	case "h":
		return m.cursorLeft(), ModeTypeNormal
	case "l":
		return m.cursorRight(false), ModeTypeNormal
	case "k":
		return m.cursorUp(), ModeTypeNormal
	case "j":
		return m.cursorDown(), ModeTypeNormal
	case "x":
		return m.deleteNextChar(), ModeTypeNormal
	case "0":
		return m.cursorLineStart(), ModeTypeNormal
	case "^":
		return m.cursorLineStartNonWhitespace(), ModeTypeNormal
	case "$":
		return m.cursorLineEnd(false), ModeTypeNormal
	case "gg":
		return m.cursorStartOfLineNum(count), ModeTypeNormal
	case "G":
		return m.cursorStartOfLastLine(), ModeTypeNormal
	case "i":
		return nil, ModeTypeInsert
	case "I":
		return m.cursorLineStartNonWhitespace(), ModeTypeInsert
	case "a":
		return m.cursorRight(true), ModeTypeInsert
	case "A":
		return m.cursorLineEnd(true), ModeTypeInsert
	default:
		return nil, ModeTypeNormal
	}
}

var normalModeSequenceRegex = regexp.MustCompile(`(?P<count>[1-9][0-9]*)?(?P<command>h|l|k|j|x|^[1-9]?0|\^|\$|gg|G|i|I|a|A)$`)

func (m *normalMode) parseSequence(seq []rune) (uint64, string) {
	submatches := normalModeSequenceRegex.FindStringSubmatch(string(seq))

	if submatches == nil {
		return 0, ""
	}

	countStr, cmdStr := submatches[1], submatches[2]
	if len(countStr) > 0 {
		count, err := strconv.ParseUint(countStr, 10, 64)
		if err != nil {
			panic(err)
		}
		return count, cmdStr
	}

	return 0, cmdStr
}

func (m *normalMode) cursorLeft() exec.Mutator {
	loc := exec.NewCharInLineLocator(text.ReadDirectionBackward, 1, false)
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

func (m *normalMode) scrollUp(scrollLines uint64) exec.Mutator {
	if scrollLines < 1 {
		scrollLines = 1
	}

	// Move the cursor to the start of a line above.  We rely on ScrollToCursorMutator
	// (appended later to every mutator) to reposition the view.
	lineAboveLoc := exec.NewRelativeLineLocator(text.ReadDirectionBackward, scrollLines)
	startOfLineLoc := exec.NewLineBoundaryLocator(text.ReadDirectionBackward, false)
	return exec.NewCompositeMutator([]exec.Mutator{
		exec.NewCursorMutator(lineAboveLoc),
		exec.NewCursorMutator(startOfLineLoc),
	})
}

func (m *normalMode) scrollDown(scrollLines uint64) exec.Mutator {
	if scrollLines < 1 {
		scrollLines = 1
	}

	// Move the cursor to the start of a line below.  We rely on ScrollToCursorMutator
	// (appended later to every mutator) to reposition the view.
	lineBelowLoc := exec.NewRelativeLineLocator(text.ReadDirectionForward, scrollLines)
	startOfLineLoc := exec.NewLineBoundaryLocator(text.ReadDirectionBackward, false)
	return exec.NewCompositeMutator([]exec.Mutator{
		exec.NewCursorMutator(lineBelowLoc),
		exec.NewCursorMutator(startOfLineLoc),
	})
}

func (m *normalMode) cursorLineStart() exec.Mutator {
	loc := exec.NewLineBoundaryLocator(text.ReadDirectionBackward, false)
	return exec.NewCursorMutator(loc)
}

func (m *normalMode) cursorLineStartNonWhitespace() exec.Mutator {
	lineStartLoc := exec.NewLineBoundaryLocator(text.ReadDirectionBackward, false)
	firstNonWhitespaceLoc := exec.NewNonWhitespaceOrNewlineLocator()
	return exec.NewCompositeMutator([]exec.Mutator{
		exec.NewCursorMutator(lineStartLoc),
		exec.NewCursorMutator(firstNonWhitespaceLoc),
	})
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
	firstNonWhitespaceLoc := exec.NewNonWhitespaceOrNewlineLocator()
	return exec.NewCompositeMutator([]exec.Mutator{
		exec.NewCursorMutator(lineNumLoc),
		exec.NewCursorMutator(firstNonWhitespaceLoc),
	})
}

func (m *normalMode) cursorStartOfLastLine() exec.Mutator {
	lastLineLoc := exec.NewLastLineLocator()
	firstNonWhitespaceLoc := exec.NewNonWhitespaceOrNewlineLocator()
	return exec.NewCompositeMutator([]exec.Mutator{
		exec.NewCursorMutator(lastLineLoc),
		exec.NewCursorMutator(firstNonWhitespaceLoc),
	})
}

func (m *normalMode) deleteNextChar() exec.Mutator {
	loc := exec.NewCharInLineLocator(text.ReadDirectionForward, 1, true)
	return exec.NewCompositeMutator([]exec.Mutator{
		exec.NewDeleteMutator(loc),
		exec.NewCursorMutator(exec.NewOntoLineLocator()),
	})
}

// insertMode is used for inserting characters into text.
type insertMode struct {
}

func newInsertMode() Mode {
	return &insertMode{}
}

func (m *insertMode) ProcessKeyEvent(event *tcell.EventKey, config Config) (exec.Mutator, ModeType) {
	mutator, nextMode := m.processKeyEvent(event)
	return appendScrollToCursor(mutator), nextMode
}

func (m *insertMode) processKeyEvent(event *tcell.EventKey) (exec.Mutator, ModeType) {
	switch event.Key() {
	case tcell.KeyRune:
		return m.insert(event.Rune()), ModeTypeInsert
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		return m.deletePrevChar(), ModeTypeInsert
	case tcell.KeyEnter:
		return m.insert('\n'), ModeTypeInsert
	case tcell.KeyTab:
		return m.insert('\t'), ModeTypeInsert
	default:
		return m.moveCursorOntoLine(), ModeTypeNormal
	}
}

func (m *insertMode) insert(r rune) exec.Mutator {
	return exec.NewInsertRuneMutator(r)
}

func (m *insertMode) deletePrevChar() exec.Mutator {
	loc := exec.NewCharInLineLocator(text.ReadDirectionBackward, 1, true)
	return exec.NewDeleteMutator(loc)
}

func (m *insertMode) moveCursorOntoLine() exec.Mutator {
	loc := exec.NewOntoLineLocator()
	return exec.NewCursorMutator(loc)
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
