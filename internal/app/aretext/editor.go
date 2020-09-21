package aretext

import (
	"fmt"
	"log"
	"time"

	"github.com/gdamore/tcell"
	"github.com/pkg/errors"
	"github.com/wedaly/aretext/internal/pkg/display"
	"github.com/wedaly/aretext/internal/pkg/exec"
	"github.com/wedaly/aretext/internal/pkg/file"
	"github.com/wedaly/aretext/internal/pkg/input"
	"github.com/wedaly/aretext/internal/pkg/repl"
	"github.com/wedaly/aretext/internal/pkg/repl/rpc"
)

const fileWatcherPollInterval = time.Second

// Editor is a terminal-based text editing program.
type Editor struct {
	inputInterpreter *input.Interpreter
	state            *exec.EditorState
	screen           tcell.Screen
	termEventChan    chan tcell.Event
	repl             repl.Repl
	rpcTaskBroker    *rpc.TaskBroker
	rpcServer        *rpc.Server
}

// NewEditor instantiates a new editor that uses the provided screen.
func NewEditor(screen tcell.Screen) (*Editor, error) {
	screenWidth, screenHeight := screen.Size()
	state := exec.NewEditorState(uint64(screenWidth), uint64(screenHeight))
	inputInterpreter := input.NewInterpreter()
	termEventChan := make(chan tcell.Event, 1)

	rpcTaskBroker := rpc.NewTaskBroker()
	rpcServer, err := rpc.NewServer(rpcTaskBroker)
	if err != nil {
		return nil, errors.Wrapf(err, "creating RPC server")
	}

	apiConfig := repl.NewApiConfig(rpcServer.Addr(), rpcServer.ApiKey())
	repl := repl.NewPythonRepl(apiConfig)
	if err := repl.Start(); err != nil {
		return nil, errors.Wrapf(err, "starting REPL")
	}

	editor := &Editor{
		inputInterpreter,
		state,
		screen,
		termEventChan,
		repl,
		rpcTaskBroker,
		rpcServer,
	}
	return editor, nil
}

// LoadInitialFile loads the initial file into the editor.
// This should be called before starting the event loop.
func (e *Editor) LoadInitialFile(path string) error {
	tree, watcher, err := file.Load(path, fileWatcherPollInterval)
	if err != nil {
		return errors.Wrapf(err, "loading file at %s", path)
	}

	exec.NewLoadDocumentMutator(tree, watcher).Mutate(e.state)
	return nil
}

// RunEventLoop processes events and draws to the screen, blocking until the user exits the program.
func (e *Editor) RunEventLoop() {
	display.DrawEditor(e.screen, e.state)
	e.screen.Sync()

	go e.rpcServer.ListenAndServe()
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
		case output, ok := <-e.repl.OutputChan():
			if ok {
				e.handleReplOutput(output)
			} else {
				e.restartRepl()
			}
		case task := <-e.rpcTaskBroker.TaskChan():
			e.handleRpcTask(task)
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

func (e *Editor) handleReplOutput(output string) {
	log.Printf("Sending REPL output to buffer: '%s'\n", output)
	mutator := exec.NewCompositeMutator([]exec.Mutator{
		exec.NewOutputReplMutator(output),
		exec.NewScrollToCursorMutator(),
	})
	e.applyMutator(mutator)
}

func (e *Editor) restartRepl() {
	log.Printf("Terminating REPL...\n")
	if err := e.repl.Terminate(); err != nil {
		log.Printf("Error terminating REPL: %v\n", err)
	}
	log.Printf("REPL terminated\n")
	log.Printf("Starting new REPL...\n")
	apiConfig := repl.NewApiConfig(e.rpcServer.Addr(), e.rpcServer.ApiKey())
	e.repl = repl.NewPythonRepl(apiConfig)
	if err := e.repl.Start(); err != nil {
		log.Fatalf("Error starting REPL: %v\n", err)
	}
	log.Printf("New REPL started\n")
}

func (e *Editor) handleRpcTask(task rpc.Task) {
	log.Printf("Executing RPC task %s\n", task.String())
	task.ExecuteAndSendResponse(e.state)
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

	if err := e.repl.Terminate(); err != nil {
		log.Printf("Error terminating REPL: %v", err)
	}

	if err := e.rpcServer.Terminate(); err != nil {
		log.Printf("Error terminating RPC server: %v", err)
	}
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
