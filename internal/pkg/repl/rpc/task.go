package rpc

import (
	"fmt"

	"github.com/wedaly/aretext/internal/pkg/exec"
)

// Task performs work scheduled by a remote procedure call.
// Every task MUST have a constructor of the form:
//   func New<Task>(msg <RequestMsg>, replyChan chan <ResponseMsg>) (Task, error)
// The task MUST send a response to replyChan when ExecuteAndSendResponse is called.
type Task interface {
	fmt.Stringer

	// Mutator returns a mutator to update the editor state.
	Mutator() exec.Mutator

	// SendResponse sends a reply to the client based on the updated editor state.
	// This should be called at most once.
	SendResponse(*exec.EditorState)
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

// quitTask terminates the editor.
type quitTask struct {
	replyChan chan QuitResultMsg
}

func NewQuitTask(_ EmptyMsg, replyChan chan QuitResultMsg) (Task, error) {
	return &quitTask{replyChan}, nil
}

func (t *quitTask) Mutator() exec.Mutator {
	return exec.NewQuitMutator()
}

func (t *quitTask) SendResponse(state *exec.EditorState) {
	defer close(t.replyChan)
	t.replyChan <- QuitResultMsg{Accepted: true}
}

func (t *quitTask) String() string {
	return "QuitTask"
}
