package aretext

import (
	"fmt"
	"log"
	"path/filepath"
	"time"

	"github.com/gdamore/tcell"
	"github.com/pkg/errors"
	"github.com/wedaly/aretext/internal/pkg/config"
	"github.com/wedaly/aretext/internal/pkg/display"
	"github.com/wedaly/aretext/internal/pkg/exec"
	"github.com/wedaly/aretext/internal/pkg/input"
)

// Editor is a terminal-based text editing program.
type Editor struct {
	inputInterpreter *input.Interpreter
	state            *exec.EditorState
	screen           tcell.Screen
	termEventChan    chan tcell.Event
}

// NewEditor instantiates a new editor that uses the provided screen.
func NewEditor(screen tcell.Screen, path string, configRuleSet config.RuleSet) *Editor {
	screenWidth, screenHeight := screen.Size()
	state := exec.NewEditorState(uint64(screenWidth), uint64(screenHeight), configRuleSet)
	inputInterpreter := input.NewInterpreter()
	termEventChan := make(chan tcell.Event, 1)
	editor := &Editor{inputInterpreter, state, screen, termEventChan}

	// Attempt to load the file.
	// If it doesn't exist, this will start with an empty document
	// that the user can edit and save to the specified path.
	loadMutator := exec.NewLoadDocumentMutator(effectivePath(path), false, true)
	editor.applyMutator(loadMutator)

	if path == "" {
		// If the user didn't specify a path, automatically open the "find files" menu
		// so they can search for a file to open.
		editor.applyMutator(input.ShowFileMenuMutator())
	}

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
		log.Printf("Error converting '%s' to absolute path: %v", path, errors.Wrapf(err, "filepath.Abs"))
		return path
	}

	return absPath
}

// RunEventLoop processes events and draws to the screen, blocking until the user exits the program.
func (e *Editor) RunEventLoop() {
	display.DrawEditor(e.screen, e.state)
	e.screen.Sync()

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

		case <-e.state.FileWatcher().ChangedChan():
			e.handleFileChanged()
		}

		if e.state.QuitFlag() {
			log.Printf("Quit flag set, exiting event loop...\n")
			return
		}

		e.redraw()
	}
}

func (e *Editor) handleTermEvent(event tcell.Event) {
	log.Printf("Handling terminal event %s\n", describeTermEvent(event))
	for _, event := range splitEscapeSequence(event) {
		mutator := e.inputInterpreter.ProcessEvent(event, e.inputConfig())
		e.applyMutator(mutator)
	}
}

func (e *Editor) handleFileChanged() {
	log.Printf("File change detected, reloading file...\n")
	const showStatusFlag = false
	mutator := exec.NewAbortIfUnsavedChangesMutator(
		exec.NewReloadDocumentMutator(showStatusFlag),
		showStatusFlag,
	)
	e.applyMutator(mutator)
}

func (e *Editor) shutdown() {
	e.state.FileWatcher().Stop()
}

func (e *Editor) inputConfig() input.Config {
	_, screenHeight := e.screen.Size()
	scrollLines := uint64(screenHeight) / 2
	return input.Config{
		InputMode:   e.state.InputMode(),
		ScrollLines: scrollLines,
	}
}

func (e *Editor) applyMutator(m exec.Mutator) {
	if m == nil {
		log.Printf("No mutator to apply\n")
		return
	}

	log.Printf("Applying mutator '%s'\n", m.String())
	m.Mutate(e.state)
}

func (e *Editor) redraw() {
	display.DrawEditor(e.screen, e.state)
	e.screen.Show()
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

func splitEscapeSequence(event tcell.Event) []tcell.Event {
	eventKey, ok := event.(*tcell.EventKey)
	if !ok {
		return []tcell.Event{event}
	}

	if eventKey.Key() != tcell.KeyRune || eventKey.Modifiers() == tcell.ModNone {
		return []tcell.Event{event}
	}

	// Terminal escape sequences begin with an escape character,
	// so sometimes tcell reports an escape keypress as a modifier on
	// another key.  Tcell uses a 50ms delay to identify individual escape chars,
	// but this strategy doesn't always work (e.g. due to network delays over
	// an SSH connection).
	// Because the escape key is used to return to normal mode, we never
	// want to miss it.  So treat ALL modifiers as an escape key followed
	// by an unmodified keypress.
	escKeyEvent := tcell.NewEventKey(tcell.KeyEscape, '\x00', tcell.ModNone)
	unmodifiedKeyEvent := tcell.NewEventKey(eventKey.Key(), eventKey.Rune(), tcell.ModNone)
	return []tcell.Event{escKeyEvent, unmodifiedKeyEvent}
}
