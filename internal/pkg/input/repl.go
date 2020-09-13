package input

import (
	"github.com/gdamore/tcell"
	"github.com/wedaly/aretext/internal/pkg/exec"
)

// replNormalMode is used to navigate the REPL buffer.
type replNormalMode struct {
	normalMode Mode
}

func newReplNormalMode() Mode {
	return &replNormalMode{
		normalMode: newNormalMode(),
	}
}

func (m *replNormalMode) ProcessKeyEvent(event *tcell.EventKey, config Config) (exec.Mutator, ModeType) {
	return processReplModeKeyEvent(event, config, m.normalMode)
}

// replInsertMode is used to insert characters into the REPL buffer.
type replInsertMode struct {
	insertMode Mode
}

func newReplInsertMode() Mode {
	return &replInsertMode{
		insertMode: newInsertMode(),
	}
}

func (m *replInsertMode) ProcessKeyEvent(event *tcell.EventKey, config Config) (exec.Mutator, ModeType) {
	return processReplModeKeyEvent(event, config, m.insertMode)
}

// processReplModeKeyEvent handles input event processing common to all REPL modes.
func processReplModeKeyEvent(event *tcell.EventKey, config Config, delegateMode Mode) (exec.Mutator, ModeType) {
	if event.Key() == tcell.KeyCtrlC {
		return exec.NewInterruptReplMutator(config.Repl), ModeTypeReplInsert
	}

	if event.Key() == tcell.KeyCtrlD {
		mutator := exec.NewCompositeMutator([]exec.Mutator{
			exec.NewLayoutMutator(exec.LayoutDocumentOnly),
			exec.NewFocusBufferMutator(exec.BufferIdDocument),
			exec.NewScrollToCursorMutator(),
		})
		return mutator, ModeTypeNormal
	}

	if event.Key() == tcell.KeyEnter {
		return exec.NewSubmitReplMutator(config.Repl), ModeTypeReplInsert
	}

	mutator, nextMode := delegateMode.ProcessKeyEvent(event, config)
	if mutator != nil {
		mutator.RestrictToReplInput()
	}
	return mutator, translateModeToRepl(nextMode)
}

// translateModeToRepl converts regular modes into their REPL equivalents.
func translateModeToRepl(mode ModeType) ModeType {
	switch mode {
	case ModeTypeInsert:
		return ModeTypeReplInsert
	case ModeTypeNormal:
		return ModeTypeReplNormal
	default:
		return mode
	}
}
