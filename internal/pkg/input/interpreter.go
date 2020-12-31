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
			ModeTypeMenu:   newMenuMode(),
		},
	}
}

// ProcessEvent interprets a terminal input event, producing a mutator if necessary.
// A return value of nil means no mutator occurred.
func (inp *Interpreter) ProcessEvent(event tcell.Event, config Config) exec.Mutator {
	switch event := event.(type) {
	case *tcell.EventKey:
		return inp.processKeyEvent(event, config)
	case *tcell.EventResize:
		return inp.processResizeEvent(event)
	default:
		return nil
	}
}

// Mode returns the current input mode of the interpreter.
func (inp *Interpreter) Mode() ModeType {
	return inp.currentMode
}

func (inp *Interpreter) processKeyEvent(event *tcell.EventKey, config Config) exec.Mutator {
	mode := inp.modes[inp.currentMode]
	mutator, nextMode := mode.ProcessKeyEvent(event, config)
	inp.currentMode = nextMode
	return mutator
}

func (inp *Interpreter) processResizeEvent(event *tcell.EventResize) exec.Mutator {
	width, height := event.Size()
	return exec.NewCompositeMutator([]exec.Mutator{
		exec.NewResizeMutator(uint64(width), uint64(height)),
		exec.NewScrollToCursorMutator(),
	})
}
