package rpc

import (
	"fmt"
	"os"
	"runtime/pprof"

	"github.com/pkg/errors"
	"github.com/wedaly/aretext/internal/pkg/exec"
	"github.com/wedaly/aretext/internal/pkg/syntax"
)

// Task performs work scheduled by a remote procedure call.
// Every task MUST have a constructor of the form:
//   func New<Task>(msg <RequestMsg>, replyChan chan <ResponseMsg>) (Task, error)
// The task MUST send a response to replyChan when ExecuteAndSendResponse is called.
type Task interface {
	fmt.Stringer

	// ExecuteAndSendResponse executes the task and sends a response to the client.
	// This should be called at most once.
	ExecuteAndSendResponse(*exec.EditorState)
}

// TaskSource retrieves tasks available for execution.
type TaskSource interface {
	// TaskChan returns a channel of tasks available for execution.
	TaskChan() chan Task
}

// AsyncExecutor schedules tasks for asynchronous execution.
type AsyncExecutor interface {
	// ApiVersion returns an identifier for the API version supported by this executor.
	ApiVersion() string

	// ExecuteAsync schedules a task triggered by an RPC.
	// It returns a channel that receives the response once the task has been executed.
	// Implementations MUST send a response to the channel.
	// Implementations MAY block if at least one other task is awaiting execution.
	ExecuteAsync(endpoint string, data []byte) (chan []byte, error)
}

// setSyntaxTask sets the syntax of the current document.
type setSyntaxTask struct {
	params    SetSyntaxMsg
	replyChan chan OpResultMsg
}

func NewSetSyntaxTask(params SetSyntaxMsg, replyChan chan OpResultMsg) (Task, error) {
	task := setSyntaxTask{params: params, replyChan: replyChan}
	return &task, nil
}

func (t *setSyntaxTask) ExecuteAndSendResponse(state *exec.EditorState) {
	defer close(t.replyChan)

	language, err := syntax.LanguageFromString(t.params.Language)
	if err != nil {
		t.replyChan <- OpResultMsg{
			Success:     false,
			Description: err.Error(),
		}
		return
	}

	exec.NewSetSyntaxMutator(language).Mutate(state)
	t.replyChan <- OpResultMsg{
		Success:     true,
		Description: fmt.Sprintf("Set syntax to %s", language),
	}
}

func (t *setSyntaxTask) String() string {
	return fmt.Sprintf("SetSyntaxTask(%s)", t.params.Language)
}

// profileMemoryTask writes a heap profile to a file.
type profileMemoryTask struct {
	params    ProfileMemoryMsg
	replyChan chan OpResultMsg
}

func NewProfileMemoryTask(params ProfileMemoryMsg, replyChan chan OpResultMsg) (Task, error) {
	task := &profileMemoryTask{
		params:    params,
		replyChan: replyChan,
	}
	return task, nil
}

func (t *profileMemoryTask) ExecuteAndSendResponse(_ *exec.EditorState) {
	defer close(t.replyChan)
	if err := t.profileMemory(); err != nil {
		t.replyChan <- OpResultMsg{
			Success:     false,
			Description: err.Error(),
		}
		return
	}
	t.replyChan <- OpResultMsg{
		Success:     true,
		Description: fmt.Sprintf("Memory profile written to %s", t.params.Path),
	}
}

func (t *profileMemoryTask) profileMemory() error {
	f, err := os.Create(t.params.Path)
	if err != nil {
		return errors.Wrapf(err, "os.Create()")
	}
	defer f.Close()

	err = pprof.WriteHeapProfile(f)
	if err != nil {
		return errors.Wrapf(err, "pprof.WriteHeapProfile()")
	}

	return nil
}

func (t *profileMemoryTask) String() string {
	return fmt.Sprintf("ProfileMemoryTask(%s)", t.params.Path)
}

// quitTask terminates the editor.
type quitTask struct {
	replyChan chan OpResultMsg
}

func NewQuitTask(_ EmptyMsg, replyChan chan OpResultMsg) (Task, error) {
	return &quitTask{replyChan}, nil
}

func (t *quitTask) ExecuteAndSendResponse(state *exec.EditorState) {
	exec.NewQuitMutator().Mutate(state)
	defer close(t.replyChan)
	t.replyChan <- OpResultMsg{Success: true}
}

func (t *quitTask) String() string {
	return "QuitTask"
}
