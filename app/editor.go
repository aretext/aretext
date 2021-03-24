package app

import (
	"fmt"
	"log"
	"path/filepath"
	"time"

	"github.com/aretext/aretext/config"
	"github.com/aretext/aretext/display"
	"github.com/aretext/aretext/exec"
	"github.com/aretext/aretext/input"
	"github.com/gdamore/tcell/v2"
	"github.com/pkg/errors"
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

		e.executeScheduledShellCmd()
		e.redraw()
	}
}

func (e *Editor) handleTermEvent(event tcell.Event) {
	log.Printf("Handling terminal event %s\n", describeTermEvent(event))
	mutator := e.inputInterpreter.ProcessEvent(event, e.inputConfig())
	e.applyMutator(mutator)
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

func (e *Editor) executeScheduledShellCmd() {
	sc := e.state.ScheduledShellCmd()
	if sc == "" {
		return
	}

	e.state.ClearScheduledShellCmd()

	// Suspend input processing and reset the terminal to its original state
	// while executing the shell command.
	if err := e.screen.Suspend(); err != nil {
		log.Printf("Error suspending the screen: %v\n", errors.Wrapf(err, "Screen.Suspend()"))
		return
	}

	// Run the shell command and pipe the output to a pager.
	err := RunShellCmd(sc)
	if err != nil {
		exec.NewSetStatusMsgMutator(exec.StatusMsg{
			Style: exec.StatusMsgStyleError,
			Text:  err.Error(),
		}).Mutate(e.state)
	}

	if err := e.screen.Resume(); err != nil {
		log.Printf("Error resuming the screen: %v\n", errors.Wrapf(err, "Screen.Resume()"))
	}
}

func (e *Editor) shutdown() {
	e.state.FileWatcher().Stop()
}

func (e *Editor) inputConfig() input.Config {
	_, screenHeight := e.screen.Size()
	scrollLines := uint64(screenHeight) / 2
	return input.Config{
		InputMode:      e.state.InputMode(),
		ScrollLines:    scrollLines,
		DirNamesToHide: e.state.DirNamesToHide(),
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
