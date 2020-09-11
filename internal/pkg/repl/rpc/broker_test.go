package rpc

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wedaly/aretext/internal/pkg/exec"
	"github.com/wedaly/aretext/internal/pkg/text"
)

func TestTaskBroker(t *testing.T) {
	var emptyMsg EmptyMsg
	emptyMsgData, err := json.Marshal(&emptyMsg)
	require.NoError(t, err)

	broker := NewTaskBroker()
	replyChan, err := broker.ExecuteAsync("quit", emptyMsgData)
	require.NoError(t, err)

	go func() {
		task := broker.PollTask()
		emptyState := exec.NewEditorState(0, 0, exec.NewBufferState(text.NewTree(), 0, 0, 0, 0, 0))
		task.SendResponse(emptyState)
	}()

	select {
	case replyData := <-replyChan:
		var replyMsg QuitResultMsg
		err = json.Unmarshal(replyData, &replyMsg)
		require.NoError(t, err)
		assert.Equal(t, QuitResultMsg{Accepted: true}, replyMsg)

	case <-time.After(time.Second * 10):
		assert.Fail(t, "Timed out waiting for reply")
	}
}

func TestTaskBrokerInvalidEndpoint(t *testing.T) {
	broker := NewTaskBroker()
	_, err := broker.ExecuteAsync("invalid", []byte{})
	assert.EqualError(t, err, "Invalid endpoint")
}
