package input

import (
	"github.com/gdamore/tcell"
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

// ProcessKeyEvent interprets a key input event, producing a command if necessary.
// A return value of nil means no command occurred.
func (inp *Interpreter) ProcessKeyEvent(event *tcell.EventKey) Command {
	if event.Key() == tcell.KeyCtrlC {
		return &QuitCommand{}
	}

	mode := inp.modes[inp.currentMode]
	cmd, nextMode := mode.ProcessKeyEvent(event)
	inp.currentMode = nextMode
	return cmd
}

// Mode returns the current input mode of the interpreter.
func (inp *Interpreter) Mode() ModeType {
	return inp.currentMode
}
