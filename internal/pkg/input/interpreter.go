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
			ModeTypeNormal:     newNormalMode(),
			ModeTypeInsert:     newInsertMode(),
			ModeTypeReplNormal: newReplNormalMode(),
			ModeTypeReplInsert: newReplInsertMode(),
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
	mutators := make([]exec.Mutator, 0, 2)
	for _, event := range inp.splitEscapeSequence(event) {
		mode := inp.modes[inp.currentMode]
		mutator, nextMode := mode.ProcessKeyEvent(event, config)
		inp.currentMode = nextMode
		if mutator != nil {
			mutators = append(mutators, mutator)
		}
	}
	return inp.combineMutators(mutators)
}

func (inp *Interpreter) splitEscapeSequence(event *tcell.EventKey) []*tcell.EventKey {
	// Terminal escape sequences begin with an escape character,
	// so sometimes tcell reports an escape keypress as a modifier on
	// another key.  Tcell uses a 50ms delay to identify individual escape chars,
	// but this strategy doesn't always work (e.g. due to network delays over
	// an SSH connection).
	// Because the escape key is used to return to normal mode, we never
	// want to miss it.  So treat ALL modifiers as an escape key followed
	// by an unmodified keypress.
	if event.Key() == tcell.KeyRune && event.Modifiers() != tcell.ModNone {
		escKeyEvent := tcell.NewEventKey(tcell.KeyEscape, '\x00', tcell.ModNone)
		unmodifiedKeyEvent := tcell.NewEventKey(event.Key(), event.Rune(), tcell.ModNone)
		return []*tcell.EventKey{escKeyEvent, unmodifiedKeyEvent}
	}

	return []*tcell.EventKey{event}
}

func (inp *Interpreter) combineMutators(mutators []exec.Mutator) exec.Mutator {
	switch len(mutators) {
	case 0:
		return nil
	case 1:
		return mutators[0]
	default:
		return exec.NewCompositeMutator(mutators)
	}
}

func (inp *Interpreter) processResizeEvent(event *tcell.EventResize) exec.Mutator {
	width, height := event.Size()
	return exec.NewCompositeMutator([]exec.Mutator{
		exec.NewResizeMutator(uint64(width), uint64(height)),
		exec.NewScrollToCursorMutator(),
	})
}
