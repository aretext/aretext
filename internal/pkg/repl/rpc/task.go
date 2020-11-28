package rpc

import (
	"fmt"
	"os"
	"runtime/pprof"

	"github.com/pkg/errors"
	"github.com/wedaly/aretext/internal/pkg/exec"
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

type profileMemoryTask struct {
	path      string
	replyChan chan ProfileMemoryResponseMsg
}

func NewProfileMemoryTask(req ProfileMemoryRequestMsg, replyChan chan ProfileMemoryResponseMsg) (Task, error) {
	task := &profileMemoryTask{
		path:      req.Path,
		replyChan: replyChan,
	}
	return task, nil
}

func (t *profileMemoryTask) ExecuteAndSendResponse(_ *exec.EditorState) {
	defer close(t.replyChan)
	if err := t.profileMemory(); err != nil {
		t.replyChan <- ProfileMemoryResponseMsg{
			Succeeded: false,
			Error:     err.Error(),
		}
		return
	}
	t.replyChan <- ProfileMemoryResponseMsg{Succeeded: true}
}

func (t *profileMemoryTask) profileMemory() error {
	f, err := os.Create(t.path)
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
	return fmt.Sprintf("ProfileMemoryTask(%s)", t.path)
}

// quitTask terminates the editor.
type quitTask struct {
	replyChan chan QuitResultMsg
}

func NewQuitTask(_ EmptyMsg, replyChan chan QuitResultMsg) (Task, error) {
	return &quitTask{replyChan}, nil
}

func (t *quitTask) ExecuteAndSendResponse(state *exec.EditorState) {
	exec.NewQuitMutator().Mutate(state)
	defer close(t.replyChan)
	t.replyChan <- QuitResultMsg{Accepted: true}
}

func (t *quitTask) String() string {
	return "QuitTask"
}
