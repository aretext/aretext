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
	// ExecuteAndSendResponse applies mutations to the editor state and sends a reply to the client.
	// This should be called at most once.
	ExecuteAndSendResponse(*exec.EditorState)
}

// TaskPoller retrieves tasks available for execution.
type TaskPoller interface {
	// PollTask returns the next available task, blocking until one is available.
	PollTask() Task
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

func (t *quitTask) ExecuteAndSendResponse(state *exec.EditorState) {
	defer close(t.replyChan)
	// TODO: quit the editor
	fmt.Printf("DEBUG: received quit RPC\n")
	t.replyChan <- QuitResultMsg{Accepted: true}
}
