package input

import (
	"embed"
	"fmt"
	"log"
	"strings"

	"github.com/gdamore/tcell/v2"

	"github.com/aretext/aretext/input/vm"
	"github.com/aretext/aretext/state"
)

// Interpreter translates key events to commands.
type Interpreter struct {
	modes map[state.InputMode]*mode
}

// NewInterpreter creates a new interpreter.
func NewInterpreter() *Interpreter {
	return &Interpreter{
		modes: map[state.InputMode]*mode{
			// normal mode is used for navigating text.
			state.InputModeNormal: {
				name:     "normal",
				commands: NormalModeCommands(),
				runtime:  runtimeForMode(NormalModeProgramPath),
			},

			// insert mode is used for inserting characters into the document.
			state.InputModeInsert: {
				name:     "insert",
				commands: InsertModeCommands(),
				runtime:  runtimeForMode(InsertModeProgramPath),
			},

			// visual mode is used to visually select a region of the document.
			state.InputModeVisual: {
				name:     "visual",
				commands: VisualModeCommands(),
				runtime:  runtimeForMode(VisualModeProgramPath),
			},

			// menu mode allows the user to search for and select items in a menu.
			state.InputModeMenu: {
				name:     "menu",
				commands: MenuModeCommands(),
				runtime:  runtimeForMode(MenuModeProgramPath),
			},

			// search mode is used to search the document for a substring.
			state.InputModeSearch: {
				name:     "search",
				commands: SearchModeCommands(),
				runtime:  runtimeForMode(SearchModeProgramPath),
			},

			// task mode is used while a task is running asynchronously.
			// This allows the user to cancel the task if it takes too long.
			state.InputModeTask: {
				name:     "task",
				commands: TaskModeCommands(),
				runtime:  runtimeForMode(TaskModeProgramPath),
			},
		},
	}
}

// ProcessEvent interprets a terminal input event as an action.
// (If there is no action, then EmptyAction will be returned.)
func (inp *Interpreter) ProcessEvent(event tcell.Event, ctx Context) Action {
	switch event := event.(type) {
	case *tcell.EventKey:
		return inp.processKeyEvent(event, ctx)
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
	NormalModeProgramPath = "generated/normal.bin"
	InsertModeProgramPath = "generated/insert.bin"
	VisualModeProgramPath = "generated/visual.bin"
	MenuModeProgramPath   = "generated/menu.bin"
	SearchModeProgramPath = "generated/search.bin"
	TaskModeProgramPath   = "generated/task.bin"
)

//go:generate go run generate.go
//go:embed generated/*
var generatedFiles embed.FS

// mustLoadProgram loads an input VM program from an embedded file.
// This avoids the overhead of compiling VM programs from command expressions
// on program startup.
// See input/generate.go for the code that compiles the programs.
func mustLoadProgram(path string) vm.Program {
	data, err := generatedFiles.ReadFile(path)
	if err != nil {
		log.Fatalf("Could not read generated program file %s: %s", path, err)
	}
	return vm.DeserializeProgram(data)
}

func runtimeForMode(programPath string) *vm.Runtime {
	// This should be long enough for any valid input sequence,
	// but not long enough that count params can overflow uint64.
	const maxInputLen = 64

	// Load the pre-compiled program and initialize the runtime.
	program := mustLoadProgram(programPath)
	return vm.NewRuntime(program, maxInputLen)
}

// mode is an editor input mode.
// Each mode has its own rules for interpreting user input.
type mode struct {
	name        string
	commands    []Command
	runtime     *vm.Runtime
	eventBuffer []vm.Event
	inputBuffer strings.Builder
}

func (m *mode) ProcessKeyEvent(event *tcell.EventKey, ctx Context) Action {
	vmEvent := eventKeyToVmEvent(event)
	m.eventBuffer = append(m.eventBuffer, vmEvent)
	if event.Key() == tcell.KeyRune {
		m.inputBuffer.WriteRune(event.Rune())
	}

	action := EmptyAction
	result := m.runtime.ProcessEvent(vmEvent)
	if result.Accepted {
		for _, capture := range result.Captures {
			// The capture ID encodes the command associated with this input.
			// See input/generate.go for details.
			if int(capture.Id) < len(m.commands) {
				command := m.commands[capture.Id]
				params := capturesToCommandParams(result.Captures, m.eventBuffer)
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

				break
			}
		}
	}

	if result.Reset {
		m.eventBuffer = m.eventBuffer[:0]
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
