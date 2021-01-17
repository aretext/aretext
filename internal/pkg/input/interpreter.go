package input

import (
	"log"

	"github.com/aretext/aretext/internal/pkg/exec"
	"github.com/gdamore/tcell"
)

// Interpreter translates key events to commands.
type Interpreter struct {
	modes map[exec.InputMode]Mode
}

// NewInterpreter creates a new interpreter.
func NewInterpreter() *Interpreter {
	return &Interpreter{
		modes: map[exec.InputMode]Mode{
			exec.InputModeNormal: newNormalMode(),
			exec.InputModeInsert: newInsertMode(),
			exec.InputModeMenu:   newMenuMode(),
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

func (inp *Interpreter) processKeyEvent(event *tcell.EventKey, config Config) exec.Mutator {
	log.Printf("Processing key in mode %s\n", config.InputMode)
	mode := inp.modes[config.InputMode]
	return mode.ProcessKeyEvent(event, config)
}

func (inp *Interpreter) processResizeEvent(event *tcell.EventResize) exec.Mutator {
	log.Printf("Processing resize event\n")
	width, height := event.Size()
	return exec.NewCompositeMutator([]exec.Mutator{
		exec.NewResizeMutator(uint64(width), uint64(height)),
		exec.NewScrollToCursorMutator(),
	})
}
