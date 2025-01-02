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
	"golang.org/x/sys/unix"
)

func TestSendAndReceiveStartSessionMsg(t *testing.T) {
	fds, err := unix.Socketpair(unix.AF_UNIX, unix.SOCK_STREAM, 0)
	require.NoError(t, err)

	clientTtySocket := os.NewFile(uintptr(fds[0]), "")
	serverTtySocket := os.NewFile(uintptr(fds[1]), "")
	defer clientTtySocket.Close()
	defer serverTtySocket.Close()

	msg := &StartSessionMsg{
		TtyFd:            int(serverTtySocket.Fd()),
		TerminalWidth:  128,
		TerminalHeight: 129,
		TerminalEnv:    map[string]string{"TERM": "tmux"},
		DocumentPath:   "/test/file",
		WorkingDir:     "/test",
	}

	receivedMsg := simulateSendAndReceive(t, msg)
	receivedStartSessionMsg, ok := receivedMsg.(*StartSessionMsg)
	require.True(t, ok)
	assert.Equal(t, msg.TerminalWidth, receivedStartSessionMsg.TerminalWidth)
	assert.Equal(t, msg.TerminalHeight, receivedStartSessionMsg.TerminalHeight)
	assert.Equal(t, msg.TerminalEnv, receivedStartSessionMsg.TerminalEnv)
	assert.Equal(t, msg.DocumentPath, receivedStartSessionMsg.DocumentPath)
	assert.Equal(t, msg.WorkingDir, receivedStartSessionMsg.WorkingDir)

	sentTtyInfo, err := serverTtySocket.Stat()
	require.NoError(t, err)
	receivedTtyFile := os.NewFile(uintptr(receivedStartSessionMsg.TtyFd), "")
	receivedTtyInfo, err := receivedTtyFile.Stat()
	assert.True(t, os.SameFile(sentTtyInfo, receivedTtyInfo))
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
