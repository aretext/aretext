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
	quitChan         chan struct{}
}

// NewEditor instantiates a new editor that uses the provided screen and file path.
func NewEditor(path string, screen tcell.Screen) (*Editor, error) {
	execState, err := initializeExecState(path)
	if err != nil {
		return nil, errors.Wrapf(err, "initializing tree")
	}
	inputInterpreter := input.NewInterpreter()
	screenWidth, screenHeight := screen.Size()
	screenRegion := display.NewScreenRegion(screen, 0, 0, screenWidth, screenHeight)
	textView := display.NewTextView(execState, screenRegion)
	quitChan := make(chan struct{})
	editor := &Editor{path, inputInterpreter, execState, textView, screen, quitChan}
	return editor, nil
}

func initializeExecState(path string) (*exec.State, error) {
	file, err := os.Open(path)
	if os.IsNotExist(err) {
		emptyState := exec.NewState(text.NewTree(), 0)
		return emptyState, nil
	} else if err != nil {
		return nil, errors.Wrapf(err, "opening file at %s", path)
	}
	defer file.Close()

	tree, err := text.NewTreeFromReader(file)
	if err != nil {
		return nil, err
	}

	execState := exec.NewState(tree, 0)
	return execState, nil
}

// RunEventLoop processes events and draws to the screen, blocking until the user exits the program.
func (e *Editor) RunEventLoop() {
	e.redraw()
	e.screen.Sync()

	go func() {
		for {
			select {
			case <-e.quitChan:
				break
			default:
				event := e.screen.PollEvent()
				e.handleEvent(event)
			}
		}
	}()

	<-e.quitChan
}

// Quit terminates the event loop.
// Calling this more than once will panic
func (e *Editor) Quit() {
	close(e.quitChan)
}

func (e *Editor) handleEvent(event tcell.Event) {
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

	switch cmd := e.inputInterpreter.ProcessKeyEvent(event).(type) {
	case *input.QuitCommand:
		e.Quit()
	case *input.ExecCommand:
		cmd.Mutator.Mutate(e.execState)
		e.textView.ScrollToCursor()
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
	e.textView.Resize(screenWidth, screenHeight)
	e.textView.ScrollToCursor()
	e.redraw()
	e.screen.Sync()
}

func (e *Editor) redraw() {
	e.screen.Clear()
	e.textView.Draw()
}
