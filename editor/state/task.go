package state

import (
	"context"
	"log"
)

// TaskFunc is a task that runs asynchronously.
// It accepts a context so that the user can cancel the task if it takes too long.
type TaskFunc func(context.Context) func(*EditorState)

// TaskState represents the state of the currently running task.
type TaskState struct {
	// resultChan receives actions to perform once the task completes.
	// This is used by the main event loop to update the editor state
	// once the task completes (meaning the user didn't cancel it).
	resultChan chan func(*EditorState)

	// cancelFunc is the function to cancel the task's context.
	cancelFunc context.CancelFunc

	// prevInputMode is the input mode to set once the task completes or is cancelled.
	prevInputMode InputMode
}

// StartTask starts a task executing asynchronously in a separate goroutine.
// If the task completes, it will send an action to state.TaskResultChan()
// which the main event loop will receive and execute.
// This will also set the input mode to InputModeTask so that the user
// can press ESC to cancel the task.
func StartTask(state *EditorState, task TaskFunc) {
	CancelTaskIfRunning(state)

	resultChan := make(chan func(*EditorState), 1)
	ctx, cancelFunc := context.WithCancel(context.Background())
	state.task = &TaskState{
		resultChan:    resultChan,
		cancelFunc:    cancelFunc,
		prevInputMode: state.inputMode,
	}
	setInputMode(state, InputModeTask)

	log.Printf("Starting task goroutine...\n")
	go func(ctx context.Context) {
		action := task(ctx)
		resultChan <- func(state *EditorState) {
			prevInputMode := state.task.prevInputMode
			state.task = nil
			setInputMode(state, prevInputMode) // from InputModeTask -> prevInputMode
			action(state)
		}
	}(ctx)
}

// CancelTaskIfRunning cancels the current task if one is running; otherwise, it does nothing.
func CancelTaskIfRunning(state *EditorState) {
	if state.task != nil {
		log.Printf("Cancelling current task...\n")
		prevInputMode := state.task.prevInputMode
		state.task.cancelFunc()
		state.task = nil
		setInputMode(state, prevInputMode) // from InputModeTask -> prevInputMode
	}
}
