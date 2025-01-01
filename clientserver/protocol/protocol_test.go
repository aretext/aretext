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
	pipeInReader, pipeInWriter, err := os.Pipe()
	require.NoError(t, err)
	defer pipeInWriter.Close()
	defer pipeInReader.Close()

	pipeOutReader, pipeOutWriter, err := os.Pipe()
	require.NoError(t, err)
	defer pipeOutWriter.Close()
	defer pipeOutReader.Close()

	msg := &StartSessionMsg{
		PipeIn:       pipeInReader,
		PipeOut:      pipeOutWriter,
		TerminalWidth: 128,
		TerminalHeight: 129,
		TerminalEnv:  map[string]string{"TERM": "tmux"},
		DocumentPath: "/test/file",
		WorkingDir:   "/test",
	}

	receivedMsg := simulateSendAndReceive(t, msg)
	receivedStartSessionMsg, ok := receivedMsg.(*StartSessionMsg)
	require.True(t, ok)
	assert.NotNil(t, receivedStartSessionMsg.PipeIn)
	assert.NotNil(t, receivedStartSessionMsg.PipeOut)
	assert.Equal(t, msg.TerminalWidth, receivedStartSessionMsg.TerminalWidth)
	assert.Equal(t, msg.TerminalHeight, receivedStartSessionMsg.TerminalHeight)
	assert.Equal(t, msg.TerminalEnv, receivedStartSessionMsg.TerminalEnv)
	assert.Equal(t, msg.DocumentPath, receivedStartSessionMsg.DocumentPath)
	assert.Equal(t, msg.WorkingDir, receivedStartSessionMsg.WorkingDir)

	sentPipeInInfo, err := pipeInReader.Stat()
	require.NoError(t, err)
	receivedPipeInInfo, err := receivedStartSessionMsg.PipeIn.Stat()
	require.NoError(t, err)
	assert.True(t, os.SameFile(sentPipeInInfo, receivedPipeInInfo))

	sentPipeOutInfo, err := pipeOutWriter.Stat()
	require.NoError(t, err)
	receivedPipeOutInfo, err := receivedStartSessionMsg.PipeOut.Stat()
	require.NoError(t, err)
	assert.True(t, os.SameFile(sentPipeOutInfo, receivedPipeOutInfo))
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
