package aretext

import (
	"os"

	"github.com/gdamore/tcell"
	"github.com/pkg/errors"
	"github.com/wedaly/aretext/internal/pkg/display"
	"github.com/wedaly/aretext/internal/pkg/exec"
	"github.com/wedaly/aretext/internal/pkg/input"
	"github.com/wedaly/aretext/internal/pkg/text"
)

// Editor is a terminal-based text editing program.
type Editor struct {
	path             string
	inputInterpreter *input.Interpreter
	execState        *exec.State
	textView         *display.TextView
	screen           tcell.Screen
	termEventChan    chan tcell.Event
	quitChan         chan struct{}
}

// NewEditor instantiates a new editor that uses the provided screen and file path.
func NewEditor(path string, screen tcell.Screen) (*Editor, error) {
	screenWidth, screenHeight := screen.Size()
	screenRegion := display.NewScreenRegion(screen, 0, 0, screenWidth, screenHeight)
	execState, err := initializeExecState(path, uint64(screenWidth), uint64(screenHeight))
	if err != nil {
		return nil, errors.Wrapf(err, "initializing tree")
	}
	inputInterpreter := input.NewInterpreter()
	textView := display.NewTextView(execState, screenRegion)
	termEventChan := make(chan tcell.Event, 1)
	quitChan := make(chan struct{})
	editor := &Editor{
		path,
		inputInterpreter,
		execState,
		textView,
		screen,
		termEventChan,
		quitChan,
	}
	return editor, nil
}

func initializeExecState(path string, viewWidth uint64, viewHeight uint64) (*exec.State, error) {
	file, err := os.Open(path)
	if os.IsNotExist(err) {
		emptyState := exec.NewState(text.NewTree(), 0, viewWidth, viewHeight)
		return emptyState, nil
	} else if err != nil {
		return nil, errors.Wrapf(err, "opening file at %s", path)
	}
	defer file.Close()

	tree, err := text.NewTreeFromReader(file)
	if err != nil {
		return nil, err
	}

	execState := exec.NewState(tree, 0, viewWidth, viewHeight)
	return execState, nil
}

// RunEventLoop processes events and draws to the screen, blocking until the user exits the program.
func (e *Editor) RunEventLoop() {
	e.redraw()
	e.screen.Sync()

	go e.pollTermEvents()
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

func (e *Editor) runMainEventLoop() {
	for {
		select {
		case <-e.quitChan:
			break
		case event := <-e.termEventChan:
			e.handleTermEvent(event)
		}
	}
}

func (e *Editor) handleTermEvent(event tcell.Event) {
	switch event := event.(type) {
	case *tcell.EventKey:
		e.handleKeyEvent(event)
	case *tcell.EventResize:
		e.handleResizeEvent(event)
	}
}

func (e *Editor) handleKeyEvent(event *tcell.EventKey) {
	// Terminal escape sequences begin with an escape character,
	// so sometimes tcell reports an escape keypress as a modifier on
	// another key.  Tcell uses a 50ms delay to identify individual escape chars,
	// but this strategy doesn't always work (e.g. due to network delays over
	// an SSH connection).
	// Because the escape key is used to return to normal mode, we never
	// want to miss it.  So treat ALL modifiers as an escape key followed
	// by an unmodified keypress.
	if event.Modifiers() != tcell.ModNone {
		escKeyEvent, unmodifiedKeyEvent := e.splitEscapeSequence(event)
		e.handleKeyEvent(escKeyEvent)
		e.handleKeyEvent(unmodifiedKeyEvent)
		return
	}

	if event.Key() == tcell.KeyCtrlC {
		e.Quit()
		return
	}

	mutator := e.inputInterpreter.ProcessKeyEvent(event, e.inputConfig())
	if mutator != nil {
		mutator.Mutate(e.execState)
		e.redraw()
		e.screen.Show()
	}
}

func (e *Editor) splitEscapeSequence(event *tcell.EventKey) (*tcell.EventKey, *tcell.EventKey) {
	escKeyEvent := tcell.NewEventKey(tcell.KeyEscape, '\x00', tcell.ModNone)
	unmodifiedKeyEvent := tcell.NewEventKey(event.Key(), event.Rune(), tcell.ModNone)
	return escKeyEvent, unmodifiedKeyEvent
}

func (e *Editor) handleResizeEvent(event *tcell.EventResize) {
	screenWidth, screenHeight := e.screen.Size()
	e.execState.SetViewSize(uint64(screenWidth), uint64(screenHeight))
	exec.NewScrollToCursorMutator().Mutate(e.execState)
	e.textView.Resize(screenWidth, screenHeight)
	e.redraw()
	e.screen.Sync()
}

func (e *Editor) redraw() {
	e.screen.Clear()
	e.textView.Draw()
}

func (e *Editor) inputConfig() input.Config {
	_, screenHeight := e.screen.Size()
	scrollLines := uint64(screenHeight) / 2
	return input.Config{
		ScrollLines: scrollLines,
	}
}
