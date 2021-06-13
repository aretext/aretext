package app

import (
	"fmt"
	"log"
	"path/filepath"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/pkg/errors"

	"github.com/aretext/aretext/config"
	"github.com/aretext/aretext/display"
	"github.com/aretext/aretext/input"
	"github.com/aretext/aretext/state"
)

// Editor is a terminal-based text editing program.
type Editor struct {
	inputInterpreter *input.Interpreter
	editorState      *state.EditorState
	screen           tcell.Screen
	termEventChan    chan tcell.Event
}

// NewEditor instantiates a new editor that uses the provided screen.
func NewEditor(screen tcell.Screen, path string, configRuleSet config.RuleSet) *Editor {
	screenWidth, screenHeight := screen.Size()
	editorState := state.NewEditorState(
		uint64(screenWidth),
		uint64(screenHeight),
		configRuleSet,
		suspendScreenFunc(screen),
	)
	inputInterpreter := input.NewInterpreter()
	termEventChan := make(chan tcell.Event, 1)
	editor := &Editor{inputInterpreter, editorState, screen, termEventChan}

	// Attempt to load the file.
	// If it doesn't exist, this will start with an empty document
	// that the user can edit and save to the specified path.
	state.LoadDocument(
		editorState,
		effectivePath(path),
		false,
		func(state.LocatorParams) uint64 { return 0 },
	)

	return editor
}

func effectivePath(path string) string {
	if path == "" {
		// If no path is specified, set a default that is probably unique.
		// The user can treat this as a scratchpad or discard it and open another file.
		path = fmt.Sprintf("untitled-%d.txt", time.Now().Unix())
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		log.Printf("Error converting '%s' to absolute path: %v", path, errors.Wrap(err, "filepath.Abs"))
		return path
	}

	return absPath
}

// RunEventLoop processes events and draws to the screen, blocking until the user exits the program.
func (e *Editor) RunEventLoop() {
	e.redraw(true)
	go e.pollTermEvents()
	e.runMainEventLoop()
	e.shutdown()
}

func (e *Editor) pollTermEvents() {
	for {
		event := e.screen.PollEvent()
		e.termEventChan <- event
	}
}

func (e *Editor) runMainEventLoop() {
	for {
		select {
		case event := <-e.termEventChan:
			e.handleTermEvent(event)

		case <-e.editorState.FileWatcher().ChangedChan():
			e.handleFileChanged()
		}

		if e.editorState.QuitFlag() {
			log.Printf("Quit flag set, exiting event loop...\n")
			return
		}

		e.redraw(false)
	}
}

func (e *Editor) handleTermEvent(event tcell.Event) {
	log.Printf("Handling terminal event %s\n", describeTermEvent(event))
	actionFunc := e.inputInterpreter.ProcessEvent(event, e.inputConfig())
	actionFunc(e.editorState)
}

func (e *Editor) handleFileChanged() {
	log.Printf("File change detected, reloading file...\n")
	state.AbortIfUnsavedChanges(e.editorState, state.ReloadDocument, false)
}

func (e *Editor) shutdown() {
	e.editorState.FileWatcher().Stop()
}

func (e *Editor) inputConfig() input.Config {
	_, screenHeight := e.screen.Size()
	scrollLines := uint64(screenHeight) / 2
	return input.Config{
		InputMode:         e.editorState.InputMode(),
		ScrollLines:       scrollLines,
		DirPatternsToHide: e.editorState.DirPatternsToHide(),
	}
}

func (e *Editor) redraw(sync bool) {
	inputMode := e.editorState.InputMode()
	inputBufferString := e.inputInterpreter.InputBufferString(inputMode)
	display.DrawEditor(e.screen, e.editorState, inputBufferString)
	if sync {
		e.screen.Sync()
	} else {
		e.screen.Show()
	}
}

func suspendScreenFunc(screen tcell.Screen) state.SuspendScreenFunc {
	return func(f func() error) error {
		// Suspend input processing and reset the terminal to its original state.
		if err := screen.Suspend(); err != nil {
			return errors.Wrap(err, "screen.Suspend()")
		}

		// Ensure screen is resumed after executing the function.
		defer func() {
			if err := screen.Resume(); err != nil {
				log.Printf("Error resuming screen: %v\n", err)
			}
		}()

		// Execute the function.
		return f()
	}
}

func describeTermEvent(event tcell.Event) string {
	switch event := event.(type) {
	case *tcell.EventKey:
		if event.Key() == tcell.KeyRune {
			return fmt.Sprintf("EventKey rune %q with modifiers %v", event.Rune(), event.Modifiers())
		} else {
			return fmt.Sprintf("EventKey %v with modifiers %v", event.Key(), event.Modifiers())
		}

	case *tcell.EventResize:
		width, height := event.Size()
		return fmt.Sprintf("EventResize with width %d and height %d", width, height)

	default:
		return "OtherEvent"
	}
}
