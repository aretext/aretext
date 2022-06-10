package state

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStartTask(t *testing.T) {
	state := NewEditorState(100, 100, nil, nil)
	StartTask(state, func(ctx context.Context) func(*EditorState) {
		return func(s *EditorState) {
			SetStatusMsg(s, StatusMsg{
				Style: StatusMsgStyleSuccess,
				Text:  "task completed",
			})
		}
	})
	assert.Equal(t, InputModeTask, state.InputMode())

	select {
	case action := <-state.TaskResultChan():
		action(state)
		assert.Equal(t, "task completed", state.StatusMsg().Text)
		assert.Equal(t, InputModeNormal, state.InputMode())
	case <-time.After(5 * time.Second):
		require.Fail(t, "Timed out")
	}
}

func TestCancelTaskIfRunning(t *testing.T) {
	cancelChan := make(chan struct{})
	state := NewEditorState(100, 100, nil, nil)
	StartTask(state, func(ctx context.Context) func(*EditorState) {
		select {
		case <-ctx.Done():
			cancelChan <- struct{}{}
		case <-time.After(5 * time.Second):
			break
		}
		return func(s *EditorState) {}
	})

	assert.Equal(t, InputModeTask, state.InputMode())
	CancelTaskIfRunning(state)
	assert.Nil(t, state.TaskResultChan())
	assert.Equal(t, InputModeNormal, state.InputMode())

	select {
	case <-cancelChan:
		// Successfully cancelled.
		break
	case <-time.After(5 * time.Second):
		require.Fail(t, "Timed out")
	}
}
