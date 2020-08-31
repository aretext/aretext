package aretext

import (
	"log"
	"os"

	"github.com/gdamore/tcell"
	"github.com/pkg/errors"
	"github.com/wedaly/aretext/internal/pkg/display"
	"github.com/wedaly/aretext/internal/pkg/exec"
	"github.com/wedaly/aretext/internal/pkg/input"
	"github.com/wedaly/aretext/internal/pkg/repl"
	"github.com/wedaly/aretext/internal/pkg/text"
)

// Editor is a terminal-based text editing program.
type Editor struct {
	path             string
	inputInterpreter *input.Interpreter
	state            *exec.EditorState
	screen           tcell.Screen
	termEventChan    chan tcell.Event
	repl             repl.Repl
	replOutputChan   chan string
	quitChan         chan struct{}
}

// NewEditor instantiates a new editor that uses the provided screen and file path.
func NewEditor(path string, screen tcell.Screen) (*Editor, error) {
	screenWidth, screenHeight := screen.Size()
	state, err := initializeState(path, uint64(screenWidth), uint64(screenHeight))
	if err != nil {
		return nil, errors.Wrapf(err, "initializing tree")
	}
	inputInterpreter := input.NewInterpreter()
	termEventChan := make(chan tcell.Event, 1)
	repl := repl.NewDummyRepl()
	replOutputChan := make(chan string, 1)
	quitChan := make(chan struct{})
	editor := &Editor{
		path,
		inputInterpreter,
		state,
		screen,
		termEventChan,
		repl,
		replOutputChan,
		quitChan,
	}
	return editor, nil
}

func initializeState(path string, viewWidth uint64, viewHeight uint64) (*exec.EditorState, error) {
	file, err := os.Open(path)
	if os.IsNotExist(err) {
		emptyBufferState := exec.NewBufferState(text.NewTree(), 0, 0, 0, viewWidth, viewHeight)
		state := exec.NewEditorState(viewWidth, viewHeight, emptyBufferState)
		return state, nil
	} else if err != nil {
		return nil, errors.Wrapf(err, "opening file at %s", path)
	}
	defer file.Close()

	tree, err := text.NewTreeFromReader(file)
	if err != nil {
		return nil, err
	}

	bufferState := exec.NewBufferState(tree, 0, 0, 0, viewWidth, viewHeight)
	state := exec.NewEditorState(viewWidth, viewHeight, bufferState)
	return state, nil
}

// RunEventLoop processes events and draws to the screen, blocking until the user exits the program.
func (e *Editor) RunEventLoop() {
	display.DrawEditor(e.screen, e.state)
	e.screen.Sync()

	go e.pollTermEvents()
	go e.pollReplOutput()
	go e.runMainEventLoop()

	<-e.quitChan
}

// Quit terminates the event loop.
// Calling this more than once will panic
func (e *Editor) Quit() {
	close(e.quitChan)
}

func (e *Editor) pollTermEvents() {
	for {
		select {
		case <-e.quitChan:
			break
		default:
			event := e.screen.PollEvent()
			e.termEventChan <- event
		}
	}
}

func (e *Editor) pollReplOutput() {
	for {
		select {
		case <-e.quitChan:
			break
		default:
			output, err := e.repl.PollOutput()
			if err != nil {
				log.Fatalf("%s", err) // TODO: handle this gracefully, restart the REPL
			}
			e.replOutputChan <- output
		}
	}
}

func (e *Editor) runMainEventLoop() {
	for {
		select {
		case <-e.quitChan:
			break
		case event := <-e.termEventChan:
			e.handleTermEvent(event)
		case output := <-e.replOutputChan:
			e.handleReplOutput(output)
		}
	}
}

func (e *Editor) handleTermEvent(event tcell.Event) {
	if event, ok := event.(*tcell.EventKey); ok {
		if event.Key() == tcell.KeyCtrlC {
			e.Quit()
			return
		}
	}

	mutator := e.inputInterpreter.ProcessEvent(event, e.inputConfig())
	e.applyMutator(mutator)
}

func (e *Editor) handleReplOutput(output string) {
	mutator := exec.NewCompositeMutator([]exec.Mutator{
		exec.NewOutputReplMutator(output),
		exec.NewScrollToCursorMutator(),
	})
	e.applyMutator(mutator)
}

func (e *Editor) inputConfig() input.Config {
	_, screenHeight := e.screen.Size()
	scrollLines := uint64(screenHeight) / 2
	return input.Config{
		Repl:        e.repl,
		ScrollLines: scrollLines,
	}
}

func (e *Editor) applyMutator(m exec.Mutator) {
	if m == nil {
		return
	}

	m.Mutate(e.state)
	display.DrawEditor(e.screen, e.state)
	e.screen.Show()
}
