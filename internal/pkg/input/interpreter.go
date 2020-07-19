package input

import (
	"github.com/gdamore/tcell"
	"github.com/wedaly/aretext/internal/pkg/exec"
	"github.com/wedaly/aretext/internal/pkg/text"
)

// Interpreter translates key events to commands.
type Interpreter struct{}

// NewInterpreter creates a new interpreter.
func NewInterpreter() *Interpreter {
	return &Interpreter{}
}

// ProcessKeyEvent interprets a key input event, producing a command if necessary.
// A return value of nil means no command occurred.
func (inp *Interpreter) ProcessKeyEvent(event *tcell.EventKey) Command {
	switch event.Key() {
	case tcell.KeyEscape, tcell.KeyEnter, tcell.KeyCtrlC:
		return &QuitCommand{}
	case tcell.KeyLeft:
		return inp.cursorLeftCmd()
	case tcell.KeyRight:
		return inp.cursorRightCmd()
	default:
		return nil
	}
}

func (inp *Interpreter) cursorLeftCmd() Command {
	loc := exec.NewCharInLineLocator(text.ReadDirectionBackward, 1)
	mutator := exec.NewCursorMutator(loc)
	return &ExecCommand{mutator}
}

func (inp *Interpreter) cursorRightCmd() Command {
	loc := exec.NewCharInLineLocator(text.ReadDirectionForward, 1)
	mutator := exec.NewCursorMutator(loc)
	return &ExecCommand{mutator}
}
