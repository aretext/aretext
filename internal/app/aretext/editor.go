package aretext

import (
	"github.com/gdamore/tcell"
)

// Editor is a terminal-based text editing program.
type Editor struct {
	screen   tcell.Screen
	quitChan chan struct{}
}

// NewEditor instantiates a new editor that uses the provided screen.
func NewEditor(screen tcell.Screen) *Editor {
	quitChan := make(chan struct{})
	return &Editor{screen, quitChan}
}

// RunEventLoop processes events and draws to the screen, blocking until the user exits the program.
func (e *Editor) RunEventLoop() {
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
	e.screen.Sync()
}
