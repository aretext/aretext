package app

import (
	"fmt"
	"log"
	"path/filepath"
	"time"

	"github.com/gdamore/tcell/v2"

	"github.com/aretext/aretext/config"
	"github.com/aretext/aretext/display"
	"github.com/aretext/aretext/input"
	"github.com/aretext/aretext/locate"
	"github.com/aretext/aretext/state"
)

// Editor is a terminal-based text editing program.
type Editor struct {
	inputInterpreter  *input.Interpreter
	editorState       *state.EditorState
	screen            tcell.Screen
	palette           *display.Palette
	documentLoadCount int
	termEventChan     chan tcell.Event
	quitChan          chan struct{}
}

// NewEditor instantiates a new editor that uses the provided screen.
func NewEditor(screen tcell.Screen, path string, lineNum uint64, configRuleSet config.RuleSet) *Editor {
	screenWidth, screenHeight := screen.Size()
	editorState := state.NewEditorState(
		uint64(screenWidth),
		uint64(screenHeight),
		configRuleSet,
		suspendScreenFunc(screen),
	)
	inputInterpreter := input.NewInterpreter()
	palette := display.NewPalette()
	documentLoadCount := editorState.DocumentLoadCount()
	termEventChan := make(chan tcell.Event, 1)
	quitChan := make(chan struct{}, 1)
	editor := &Editor{
		inputInterpreter,
		editorState,
		screen,
		palette,
		documentLoadCount,
		termEventChan,
		quitChan,
	}

	// Attempt to load the file.
	// If it doesn't exist, this will start with an empty document
	// that the user can edit and save to the specified path.
	state.LoadDocument(
		editorState,
		effectivePath(path),
		false,
		func(p state.LocatorParams) uint64 {
			return locate.StartOfLineNum(p.TextTree, lineNum)
		},
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
		log.Printf("Error converting %q to absolute path: %v", path, fmt.Errorf("filepath.Abs: %w", err))
		return path
	}

	return absPath
}

// RunEventLoop processes events and draws to the screen, blocking until the user exits the program.
func (e *Editor) RunEventLoop() {
	e.redraw(true)
	go e.screen.ChannelEvents(e.termEventChan, e.quitChan)
	e.runMainEventLoop()
	e.shutdown()
}

func (e *Editor) runMainEventLoop() {
	var inBracketedPaste bool
	for {
		select {
		case event := <-e.termEventChan:
			e.handleTermEvent(event)
			if pasteEvent, ok := event.(*tcell.EventPaste); ok {
				inBracketedPaste = pasteEvent.Start()
			}

		case actionFunc := <-e.editorState.TaskResultChan():
			log.Printf("Task completed, executing resulting action...\n")
			actionFunc(e.editorState)

		case <-e.editorState.FileWatcher().ChangedChan():
			e.handleFileChanged()
		}

		e.handleIfDocumentLoaded()

		if e.editorState.QuitFlag() {
			log.Printf("Quit flag set, exiting event loop...\n")
			return
		}

		// Redraw unless there are pending terminal events to process first
		// or we're in the middle of a bracketed paste.
		// This helps avoid the overhead of redrawing after every keypress
		// if the user pastes a lot of text into the terminal emulator.
		if len(e.termEventChan) == 0 && !inBracketedPaste {
			e.redraw(false)
		}
	}
}

func (e *Editor) handleTermEvent(event tcell.Event) {
	inputCtx := input.ContextFromEditorState(e.editorState)
	actionFunc := e.inputInterpreter.ProcessEvent(event, inputCtx)
	actionFunc(e.editorState)
}

func (e *Editor) handleFileChanged() {
	log.Printf("File change detected, reloading file...\n")
	state.AbortIfUnsavedChanges(e.editorState, "", state.ReloadDocument)
}

func (e *Editor) handleIfDocumentLoaded() {
	documentLoadCount := e.editorState.DocumentLoadCount()
	if documentLoadCount != e.documentLoadCount {
		log.Printf("Detected document loaded, updating editor")

		// Reset the input interpreter, which may have state from the prev document.
		e.inputInterpreter = input.NewInterpreter()

		// Update palette, since the configuration might have changed.
		styles := e.editorState.Styles()
		e.palette = display.NewPalette().ApplyConfigOverrides(styles)

		// Store the new document load count so we know when the next document loads.
		e.documentLoadCount = documentLoadCount
	}
}

func (e *Editor) shutdown() {
	e.editorState.FileWatcher().Stop()
	e.quitChan <- struct{}{}
}

func (e *Editor) redraw(sync bool) {
	inputMode := e.editorState.InputMode()
	inputBufferString := e.inputInterpreter.InputBufferString(inputMode)
	display.DrawEditor(e.screen, e.palette, e.editorState, inputBufferString)
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
			return fmt.Errorf("screen.Suspend: %w", err)
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
