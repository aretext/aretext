package aretext

import (
	"os"

	"github.com/gdamore/tcell"
	"github.com/pkg/errors"
	"github.com/wedaly/aretext/internal/pkg/display"
	"github.com/wedaly/aretext/internal/pkg/text"
)

// Editor is a terminal-based text editing program.
type Editor struct {
	path     string
	tree     *text.Tree
	textView *display.TextView
	screen   tcell.Screen
	quitChan chan struct{}
}

// NewEditor instantiates a new editor that uses the provided screen and file path.
func NewEditor(path string, screen tcell.Screen) (*Editor, error) {
	tree, err := initializeTree(path)
	if err != nil {
		return nil, errors.Wrapf(err, "initializing tree")
	}

	screenWidth, screenHeight := screen.Size()
	textView := display.NewTextView(tree, display.NewScreenRegion(screen, 0, 0, screenWidth, screenHeight))
	quitChan := make(chan struct{})
	editor := &Editor{path, tree, textView, screen, quitChan}
	return editor, nil
}

func initializeTree(path string) (*text.Tree, error) {
	file, err := os.Open(path)
	if os.IsNotExist(err) {
		return text.NewTree(), nil
	} else if err != nil {
		return nil, errors.Wrapf(err, "opening file at %s", path)
	}
	defer file.Close()
	return text.NewTreeFromReader(file)
}

// RunEventLoop processes events and draws to the screen, blocking until the user exits the program.
func (e *Editor) RunEventLoop() {
	e.redrawAndSync()

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

// Stop terminates the event loop.
// Calling this more than once will panic
func (e *Editor) Stop() {
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
	switch event.Key() {
	case tcell.KeyEscape, tcell.KeyEnter, tcell.KeyCtrlC:
		close(e.quitChan)
	}
}

func (e *Editor) handleResizeEvent(event *tcell.EventResize) {
	screenWidth, screenHeight := e.screen.Size()
	e.textView.Resize(screenWidth, screenHeight)
	e.redrawAndSync()
}

func (e *Editor) redrawAndSync() {
	e.screen.Clear()
	e.textView.Draw()
	e.screen.Sync()
}
