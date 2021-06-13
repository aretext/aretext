package input

import (
	"log"

	"github.com/gdamore/tcell/v2"

	"github.com/aretext/aretext/state"
)

// Interpreter translates key events to commands.
type Interpreter struct {
	macroRecorder *MacroRecorder
	modes         map[state.InputMode]Mode
}

// NewInterpreter creates a new interpreter.
func NewInterpreter() *Interpreter {
	return &Interpreter{
		macroRecorder: NewMacroRecorder(),
		modes: map[state.InputMode]Mode{
			state.InputModeNormal: newNormalMode(),
			state.InputModeInsert: &insertMode{},
			state.InputModeMenu:   &menuMode{},
			state.InputModeSearch: &searchMode{},
			state.InputModeVisual: newVisualMode(),
		},
	}
}

// ProcessEvent interprets a terminal input event as an action.
// (If there is no action, then EmptyAction will be returned.)
func (inp *Interpreter) ProcessEvent(event tcell.Event, config Config) Action {
	switch event := event.(type) {
	case *tcell.EventKey:
		return inp.processKeyEvent(event, config)
	case *tcell.EventResize:
		return inp.processResizeEvent(event)
	default:
		return EmptyAction
	}
}

func (inp *Interpreter) processKeyEvent(event *tcell.EventKey, config Config) Action {
	log.Printf("Processing key in mode %s\n", config.InputMode)
	mode := inp.modes[config.InputMode]
	return mode.ProcessKeyEvent(event, inp.macroRecorder, config)
}

func (inp *Interpreter) processResizeEvent(event *tcell.EventResize) Action {
	log.Printf("Processing resize event\n")
	width, height := event.Size()
	return func(s *state.EditorState) {
		state.ResizeView(s, uint64(width), uint64(height))
		state.ScrollViewToCursor(s)
	}
}

// InputBufferString returns a string describing buffered input events.
// It can be displayed to the user to help them understand the input state.
func (inp *Interpreter) InputBufferString(mode state.InputMode) string {
	return inp.modes[mode].InputBufferString()
}
