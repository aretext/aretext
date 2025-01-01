package protocol

import (
	"context"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSendAndReceiveStartSessionMsg(t *testing.T) {
	fakePtsPath := filepath.Join(t.TempDir(), "pts")
	fakePts, err := os.Create(fakePtsPath)
	require.NoError(t, err)
	defer fakePts.Close()

	msg := &StartSessionMsg{
		DocumentPath: "/test/file",
		WorkingDir:   "/test",
		TerminalEnv:  map[string]string{"TERM": "tmux"},
		Pts:          fakePts,
	}

	receivedMsg := simulateSendAndReceive(t, msg)
	receivedStartSessionMsg, ok := receivedMsg.(*StartSessionMsg)
	require.True(t, ok)
	assert.Equal(t, msg.DocumentPath, receivedStartSessionMsg.DocumentPath)
	assert.Equal(t, msg.WorkingDir, receivedStartSessionMsg.WorkingDir)
	assert.Equal(t, msg.TerminalEnv, receivedStartSessionMsg.TerminalEnv)
	assert.NotNil(t, receivedStartSessionMsg.Pts)

	sentPtsFileInfo, err := msg.Pts.Stat()
	require.NoError(t, err)
	receivedPtsFileInfo, err := receivedStartSessionMsg.Pts.Stat()
	require.NoError(t, err)
	assert.True(t, os.SameFile(sentPtsFileInfo, receivedPtsFileInfo))
}

func TestSendAndReceiveResizeTerminalMsg(t *testing.T) {
	msg := &ResizeTerminalMsg{
		Width:  123,
		Height: 456,
	}
	receivedMsg := simulateSendAndReceive(t, msg)
	assert.Equal(t, msg, receivedMsg)
}

func simulateSendAndReceive(t *testing.T, msgToSend Message) Message {
	socketPath := filepath.Join(t.TempDir(), "test.socket")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	readyChan := make(chan struct{})
	recvChan := make(chan Message, 1)
	errChan := make(chan error, 1)
	go func(ctx context.Context, socketPath string) {
		serverAddr, err := net.ResolveUnixAddr("unix", socketPath)
		if err != nil {
			t.Fatalf("Failed to resolve server address: %v", err)
		}

		listener, err := net.ListenUnix("unix", serverAddr)
		if err != nil {
			t.Fatalf("Failed to create listener: %v", err)
		}
		defer listener.Close()

		readyChan <- struct{}{}

		conn, err := listener.AcceptUnix()
		if err != nil {
			errChan <- err
			return
		}
		defer conn.Close()

		msg, err := ReceiveMessage(conn)
		if err != nil {
			errChan <- err
			return
		}
		recvChan <- msg
	}(ctx, socketPath)

	select {
	case <-readyChan:
	}

	clientAddr, err := net.ResolveUnixAddr("unix", socketPath)
	require.NoError(t, err)
	clientConn, err := net.DialUnix("unix", nil, clientAddr)
	require.NoError(t, err)
	defer clientConn.Close()

	err = SendMessage(clientConn, msgToSend)
	assert.NoError(t, err)

	select {
	case err := <-errChan:
		t.Fatalf("Error sending message: %v", err)
	case receivedMsg := <-recvChan:
		return receivedMsg
	case <-time.After(500 * time.Second): // TODO: lower this
		t.Fatal("Timeout waiting for message")
	}

	return nil
}
