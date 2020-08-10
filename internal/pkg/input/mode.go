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
	// It will return any user-initiated command resulting from the keypress
	// as well as the next input mode (which could be the same as the current mode).
	ProcessKeyEvent(event *tcell.EventKey) (Command, ModeType)
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

func (m *normalMode) ProcessKeyEvent(event *tcell.EventKey) (cmd Command, mode ModeType) {
	defer func() {
		if cmd != nil {
			m.buffer = m.buffer[:0]
		}
	}()

	if event.Key() == tcell.KeyRune {
		return m.processRuneKey(event.Rune())
	}
	return m.processSpecialKey(event.Key())
}

func (m *normalMode) processSpecialKey(key tcell.Key) (Command, ModeType) {
	switch key {
	case tcell.KeyLeft:
		return m.cursorLeftCmd(), ModeTypeNormal
	case tcell.KeyRight:
		return m.cursorRightCmd(false), ModeTypeNormal
	case tcell.KeyUp:
		return m.cursorUpCmd(), ModeTypeNormal
	case tcell.KeyDown:
		return m.cursorDownCmd(), ModeTypeNormal
	default:
		return nil, ModeTypeNormal
	}
}

func (m *normalMode) processRuneKey(r rune) (Command, ModeType) {
	m.buffer = append(m.buffer, r)
	count, cmd := m.parseSequence(m.buffer)

	switch cmd {
	case "h":
		return m.cursorLeftCmd(), ModeTypeNormal
	case "l":
		return m.cursorRightCmd(false), ModeTypeNormal
	case "k":
		return m.cursorUpCmd(), ModeTypeNormal
	case "j":
		return m.cursorDownCmd(), ModeTypeNormal
	case "x":
		return m.deleteNextCharCmd(), ModeTypeNormal
	case "0":
		return m.cursorLineStartCmd(), ModeTypeNormal
	case "^":
		return m.cursorLineStartNonWhitespaceCmd(), ModeTypeNormal
	case "$":
		return m.cursorLineEndCmd(false), ModeTypeNormal
	case "gg":
		return m.cursorStartOfLineNumCmd(count), ModeTypeNormal
	case "G":
		return m.cursorStartOfLastLineCmd(), ModeTypeNormal
	case "i":
		return nil, ModeTypeInsert
	case "I":
		return m.cursorLineStartNonWhitespaceCmd(), ModeTypeInsert
	case "a":
		return m.cursorRightCmd(true), ModeTypeInsert
	case "A":
		return m.cursorLineEndCmd(true), ModeTypeInsert
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

func (m *normalMode) cursorLeftCmd() Command {
	loc := exec.NewCharInLineLocator(text.ReadDirectionBackward, 1, false)
	mutator := exec.NewCursorMutator(loc)
	return &ExecCommand{mutator}
}

func (m *normalMode) cursorRightCmd(allowPastEndOfLineOrFile bool) Command {
	loc := exec.NewCharInLineLocator(text.ReadDirectionForward, 1, allowPastEndOfLineOrFile)
	mutator := exec.NewCursorMutator(loc)
	return &ExecCommand{mutator}
}

func (m *normalMode) cursorUpCmd() Command {
	loc := exec.NewRelativeLineLocator(text.ReadDirectionBackward, 1)
	mutator := exec.NewCursorMutator(loc)
	return &ExecCommand{mutator}
}

func (m *normalMode) cursorDownCmd() Command {
	loc := exec.NewRelativeLineLocator(text.ReadDirectionForward, 1)
	mutator := exec.NewCursorMutator(loc)
	return &ExecCommand{mutator}
}

func (m *normalMode) cursorLineStartCmd() Command {
	loc := exec.NewLineBoundaryLocator(text.ReadDirectionBackward, false)
	mutator := exec.NewCursorMutator(loc)
	return &ExecCommand{mutator}
}

func (m *normalMode) cursorLineStartNonWhitespaceCmd() Command {
	lineStartLoc := exec.NewLineBoundaryLocator(text.ReadDirectionBackward, false)
	firstNonWhitespaceLoc := exec.NewNonWhitespaceLocator(text.ReadDirectionForward)
	mutator := exec.NewCompositeMutator([]exec.Mutator{
		exec.NewCursorMutator(lineStartLoc),
		exec.NewCursorMutator(firstNonWhitespaceLoc),
	})
	return &ExecCommand{mutator}
}

func (m *normalMode) cursorLineEndCmd(includeEndOfLineOrFile bool) Command {
	loc := exec.NewLineBoundaryLocator(text.ReadDirectionForward, includeEndOfLineOrFile)
	mutator := exec.NewCursorMutator(loc)
	return &ExecCommand{mutator}
}

func (m *normalMode) cursorStartOfLineNumCmd(count uint64) Command {
	// Convert 1-indexed count to 0-indexed line num
	var lineNum uint64
	if count > 0 {
		lineNum = count - 1
	}

	lineNumLoc := exec.NewLineNumLocator(lineNum)
	firstNonWhitespaceLoc := exec.NewNonWhitespaceLocator(text.ReadDirectionForward)
	mutator := exec.NewCompositeMutator([]exec.Mutator{
		exec.NewCursorMutator(lineNumLoc),
		exec.NewCursorMutator(firstNonWhitespaceLoc),
	})
	return &ExecCommand{mutator}
}

func (m *normalMode) cursorStartOfLastLineCmd() Command {
	lastLineLoc := exec.NewLastLineLocator()
	firstNonWhitespaceLoc := exec.NewNonWhitespaceLocator(text.ReadDirectionForward)
	mutator := exec.NewCompositeMutator([]exec.Mutator{
		exec.NewCursorMutator(lastLineLoc),
		exec.NewCursorMutator(firstNonWhitespaceLoc),
	})
	return &ExecCommand{mutator}
}

func (m *normalMode) deleteNextCharCmd() Command {
	loc := exec.NewCharInLineLocator(text.ReadDirectionForward, 1, true)
	mutator := exec.NewCompositeMutator([]exec.Mutator{
		exec.NewDeleteMutator(loc),
		exec.NewCursorMutator(exec.NewOntoLineLocator()),
	})
	return &ExecCommand{mutator}
}

// insertMode is used for inserting characters into text.
type insertMode struct {
}

func newInsertMode() Mode {
	return &insertMode{}
}

func (m *insertMode) ProcessKeyEvent(event *tcell.EventKey) (Command, ModeType) {
	switch event.Key() {
	case tcell.KeyRune:
		return m.insertCmd(event.Rune()), ModeTypeInsert
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		return m.deletePrevCharCmd(), ModeTypeInsert
	case tcell.KeyEnter:
		return m.insertCmd('\n'), ModeTypeInsert
	case tcell.KeyTab:
		return m.insertCmd('\t'), ModeTypeInsert
	default:
		return m.moveCursorOntoLineCmd(), ModeTypeNormal
	}
}

func (m *insertMode) insertCmd(r rune) Command {
	mutator := exec.NewInsertRuneMutator(r)
	return &ExecCommand{mutator}
}

func (m *insertMode) deletePrevCharCmd() Command {
	loc := exec.NewCharInLineLocator(text.ReadDirectionBackward, 1, true)
	mutator := exec.NewDeleteMutator(loc)
	return &ExecCommand{mutator}
}

func (m *insertMode) moveCursorOntoLineCmd() Command {
	loc := exec.NewOntoLineLocator()
	mutator := exec.NewCursorMutator(loc)
	return &ExecCommand{mutator}
}
