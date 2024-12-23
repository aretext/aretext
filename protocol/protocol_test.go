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

func TestSendAndReceive(t *testing.T) {
	tmpDir := t.TempDir()
	socketPath := filepath.Join(tmpDir, "test.socket")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	fakePtsPath := filepath.Join(tmpDir, "pts")
	fakePts, err := os.Create(fakePtsPath)
	require.NoError(t, err)
	defer fakePts.Close()

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

	msg := &ClientHelloMsg{
		FilePath:    "/test/file",
		WorkingDir:  "/test",
		TerminalEnv: []string{"TERM=tmux"},
		Pts:         fakePts,
	}
	err = SendMessage(clientConn, msg)
	assert.NoError(t, err)

	select {
	case err := <-errChan:
		t.Fatalf("Error sending message: %v", err)
	case receivedMsg := <-recvChan:
		assert.Equal(t, msg, receivedMsg)
	case <-time.After(5 * time.Second):
		t.Fatal("Timeout waiting for message")
	}
}
