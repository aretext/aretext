package input

import (
	"github.com/gdamore/tcell"
	"github.com/wedaly/aretext/internal/pkg/exec"
)

// Interpreter translates key events to commands.
type Interpreter struct {
	currentMode ModeType
	modes       map[ModeType]Mode
}

// NewInterpreter creates a new interpreter.
func NewInterpreter() *Interpreter {
	return &Interpreter{
		currentMode: ModeTypeNormal,
		modes: map[ModeType]Mode{
			ModeTypeNormal: newNormalMode(),
			ModeTypeInsert: newInsertMode(),
		},
	}
}

// ProcessKeyEvent interprets a key input event, producing a mutator if necessary.
// A return value of nil means no mutator occurred.
func (inp *Interpreter) ProcessKeyEvent(event *tcell.EventKey) exec.Mutator {
	mode := inp.modes[inp.currentMode]
	mutator, nextMode := mode.ProcessKeyEvent(event)
	inp.currentMode = nextMode
	return mutator
}

// Mode returns the current input mode of the interpreter.
func (inp *Interpreter) Mode() ModeType {
	return inp.currentMode
}
