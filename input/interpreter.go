package input

import (
	"embed"
	"fmt"
	"log"
	"strings"

	"github.com/gdamore/tcell/v2"

	"github.com/aretext/aretext/input/engine"
	"github.com/aretext/aretext/state"
)

// Interpreter translates key events to commands.
type Interpreter struct {
	modes            map[state.InputMode]*mode
	inBracketedPaste bool
	pasteBuffer      strings.Builder
}

// NewInterpreter creates a new interpreter.
func NewInterpreter() *Interpreter {
	return &Interpreter{
		modes: map[state.InputMode]*mode{
			// normal mode is used for navigating text.
			state.InputModeNormal: {
				name:     "normal",
				commands: NormalModeCommands(),
				runtime:  runtimeForMode(NormalModePath),
			},

			// insert mode is used for inserting characters into the document.
			state.InputModeInsert: {
				name:     "insert",
				commands: InsertModeCommands(),
				runtime:  runtimeForMode(InsertModePath),
			},

			// visual mode is used to visually select a region of the document.
			state.InputModeVisual: {
				name:     "visual",
				commands: VisualModeCommands(),
				runtime:  runtimeForMode(VisualModePath),
			},

			// menu mode allows the user to search for and select items in a menu.
			state.InputModeMenu: {
				name:     "menu",
				commands: MenuModeCommands(),
				runtime:  runtimeForMode(MenuModePath),
			},

			// search mode is used to search the document for a substring.
			state.InputModeSearch: {
				name:     "search",
				commands: SearchModeCommands(),
				runtime:  runtimeForMode(SearchModePath),
			},

			// task mode is used while a task is running asynchronously.
			// This allows the user to cancel the task if it takes too long.
			state.InputModeTask: {
				name:     "task",
				commands: TaskModeCommands(),
				runtime:  runtimeForMode(TaskModePath),
			},
		},
	}
}

// ProcessEvent interprets a terminal input event as an action.
// (If there is no action, then EmptyAction will be returned.)
func (inp *Interpreter) ProcessEvent(event tcell.Event, ctx Context) Action {
	switch event := event.(type) {
	case *tcell.EventKey:
		if inp.inBracketedPaste {
			return inp.processPasteKey(event)
		}
		return inp.processKeyEvent(event, ctx)
	case *tcell.EventPaste:
		if event.Start() {
			return inp.processPasteStart()
		} else {
			return inp.processPasteEnd(ctx)
		}
	case *tcell.EventResize:
		return inp.processResizeEvent(event)
	default:
		return EmptyAction
	}
}

func (inp *Interpreter) processKeyEvent(event *tcell.EventKey, ctx Context) Action {
	log.Printf("Processing key %s in mode %s\n", event.Name(), ctx.InputMode)
	mode := inp.modes[ctx.InputMode]
	return mode.ProcessKeyEvent(event, ctx)
}

func (inp *Interpreter) processPasteStart() Action {
	inp.inBracketedPaste = true
	return EmptyAction
}

func (inp *Interpreter) processPasteKey(event *tcell.EventKey) Action {
	switch event.Key() {
	// Most terminals send KeyEnter (ASCII code 13 = carriage return)
	// but alacritty sends KeyLF (ASCII code 10 = line feed).
	case tcell.KeyEnter, tcell.KeyLF:
		inp.pasteBuffer.WriteRune('\n')
	case tcell.KeyTab:
		inp.pasteBuffer.WriteRune('\t')
	case tcell.KeyRune:
		inp.pasteBuffer.WriteRune(event.Rune())
	}
	return EmptyAction
}

func (inp *Interpreter) processPasteEnd(ctx Context) Action {
	text := inp.pasteBuffer.String()
	inp.inBracketedPaste = false
	inp.pasteBuffer.Reset()

	switch ctx.InputMode {
	case state.InputModeInsert:
		return InsertFromBracketedPaste(text)
	case state.InputModeNormal, state.InputModeVisual:
		return ShowStatusMsgBracketedPasteWrongMode
	case state.InputModeMenu:
		return BracketedPasteIntoMenuSearch(text)
	case state.InputModeSearch:
		return BracketedPasteIntoSearchQuery(text)
	default:
		return EmptyAction
	}
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

const (
	NormalModePath = "generated/normal.bin"
	InsertModePath = "generated/insert.bin"
	VisualModePath = "generated/visual.bin"
	MenuModePath   = "generated/menu.bin"
	SearchModePath = "generated/search.bin"
	TaskModePath   = "generated/task.bin"
)

//go:generate go run generate.go
//go:embed generated/*
var generatedFiles embed.FS

// runtimeForMode loads a state machine for an input mode.
// The state machine is serialized and embedded in the aretext binary.
// See input/generate.go for the code that compiles the state machines.
func runtimeForMode(path string) *engine.Runtime {
	// This should be long enough for any valid input sequence,
	// but not long enough that count params can overflow uint64.
	const maxInputLen = 64

	data, err := generatedFiles.ReadFile(path)
	if err != nil {
		log.Fatalf("Could not read %s: %s", path, err)
	}

	stateMachine, err := engine.Deserialize(data)
	if err != nil {
		log.Fatalf("Could not deserialize state machine %s: %s", path, err)
	}

	return engine.NewRuntime(stateMachine, maxInputLen)
}

// mode is an editor input mode.
// Each mode has its own rules for interpreting user input.
type mode struct {
	name        string
	commands    []Command
	runtime     *engine.Runtime
	inputBuffer strings.Builder
}

func (m *mode) ProcessKeyEvent(event *tcell.EventKey, ctx Context) Action {
	engineEvent := eventKeyToEngineEvent(event)
	if event.Key() == tcell.KeyRune {
		m.inputBuffer.WriteRune(event.Rune())
	}

	action := EmptyAction
	result := m.runtime.ProcessEvent(engineEvent)
	if result.Decision == engine.DecisionAccept {
		command := m.commands[result.CmdId]
		params := capturesToCommandParams(result.Captures)
		log.Printf(
			"%s mode accepted input for command %q with params %+v and ctx %+v\n",
			m.name, command.Name,
			params, ctx,
		)

		if err := m.validateParams(command, params); err != nil {
			action = func(s *state.EditorState) {
				state.SetStatusMsg(s, state.StatusMsg{
					Style: state.StatusMsgStyleError,
					Text:  err.Error(),
				})
			}
		} else {
			action = command.BuildAction(ctx, params)
		}
	}

	if result.Decision != engine.DecisionWait {
		m.inputBuffer.Reset()
	}

	return action
}

func (m *mode) validateParams(command Command, params CommandParams) error {
	if command.MaxCount > 0 && params.Count > command.MaxCount {
		return fmt.Errorf("count must be less than or equal to %d", command.MaxCount)
	}
	return nil
}

func (m *mode) InputBufferString() string {
	return m.inputBuffer.String()
}
