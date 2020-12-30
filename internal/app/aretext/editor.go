package aretext

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gdamore/tcell"
	"github.com/pkg/errors"
	"github.com/wedaly/aretext/internal/pkg/display"
	"github.com/wedaly/aretext/internal/pkg/exec"
	"github.com/wedaly/aretext/internal/pkg/file"
	"github.com/wedaly/aretext/internal/pkg/input"
	"github.com/wedaly/aretext/internal/pkg/text"
)

const fileWatcherPollInterval = time.Second

// Editor is a terminal-based text editing program.
type Editor struct {
	inputInterpreter *input.Interpreter
	state            *exec.EditorState
	screen           tcell.Screen
	termEventChan    chan tcell.Event
}

// NewEditor instantiates a new editor that uses the provided screen.
func NewEditor(screen tcell.Screen, path string) (*Editor, error) {
	tree, watcher, err := file.Load(path, fileWatcherPollInterval)
	if os.IsNotExist(err) {
		tree = text.NewTree()
		watcher = file.NewWatcher(fileWatcherPollInterval, path, time.Time{}, 0, "")
	} else if err != nil {
		return nil, errors.Wrapf(err, "loading file at %s", path)
	}

	screenWidth, screenHeight := screen.Size()
	state := exec.NewEditorState(uint64(screenWidth), uint64(screenHeight), tree, watcher)
	inputInterpreter := input.NewInterpreter()
	termEventChan := make(chan tcell.Event, 1)
	editor := &Editor{inputInterpreter, state, screen, termEventChan}
	return editor, nil
}

// RunEventLoop processes events and draws to the screen, blocking until the user exits the program.
func (e *Editor) RunEventLoop() {
	display.DrawEditor(e.screen, e.state, e.inputInterpreter.Mode())
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
	mutator := e.inputInterpreter.ProcessEvent(event, e.inputConfig())
	e.applyMutator(mutator)
}

func (e *Editor) handleFileChanged() {
	path := e.state.FileWatcher().Path()
	log.Printf("File change detected, reloading file from '%s'\n", path)
	tree, watcher, err := file.Load(path, fileWatcherPollInterval)
	if err != nil {
		log.Printf("Error reloading file '%s': %v\n", path, err)
		return
	}

	e.applyMutator(exec.NewLoadDocumentMutator(tree, watcher))
	log.Printf("Successfully reloaded file '%s' into editor\n", path)
}

func (e *Editor) shutdown() {
	e.state.FileWatcher().Stop()
}

func (e *Editor) inputConfig() input.Config {
	_, screenHeight := e.screen.Size()
	scrollLines := uint64(screenHeight) / 2
	return input.Config{
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
	display.DrawEditor(e.screen, e.state, e.inputInterpreter.Mode())
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
