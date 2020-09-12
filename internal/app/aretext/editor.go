package aretext

import (
	"fmt"
	"log"
	"os"

	"github.com/gdamore/tcell"
	"github.com/pkg/errors"
	"github.com/wedaly/aretext/internal/pkg/display"
	"github.com/wedaly/aretext/internal/pkg/exec"
	"github.com/wedaly/aretext/internal/pkg/input"
	"github.com/wedaly/aretext/internal/pkg/repl"
	"github.com/wedaly/aretext/internal/pkg/repl/rpc"
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
	rpcTaskBroker    *rpc.TaskBroker
	rpcServer        *rpc.Server
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
		path,
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

	go e.rpcServer.ListenAndServe()
	go e.pollTermEvents()

	e.runMainEventLoop()

	if err := e.repl.Terminate(); err != nil {
		log.Printf("Error terminating REPL: %v", err)
	}

	if err := e.rpcServer.Terminate(); err != nil {
		log.Printf("Error terminating RPC server: %v", err)
	}
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
		}

		if e.state.QuitFlag() {
			log.Printf("Quit flag set, exiting event loop...\n")
			return
		}
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
	e.applyMutator(task.Mutator())
	task.SendResponse(e.state)
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
