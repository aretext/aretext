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

func TestSendAndReceiveClientHelloMsg(t *testing.T) {
	fakePtsPath := filepath.Join(t.TempDir(), "pts")
	fakePts, err := os.Create(fakePtsPath)
	require.NoError(t, err)
	defer fakePts.Close()

	msg := &ClientHelloMsg{
		DocumentPath: "/test/file",
		WorkingDir:   "/test",
		TerminalEnv:  []string{"TERM=tmux"},
		Pts:          fakePts,
	}

	receivedMsg := simulateSendAndReceive(t, msg)
	receivedClientHelloMsg, ok := receivedMsg.(*ClientHelloMsg)
	require.True(t, ok)
	assert.Equal(t, msg.DocumentPath, receivedClientHelloMsg.DocumentPath)
	assert.Equal(t, msg.WorkingDir, receivedClientHelloMsg.WorkingDir)
	assert.Equal(t, msg.TerminalEnv, receivedClientHelloMsg.TerminalEnv)
	assert.NotNil(t, receivedClientHelloMsg.Pts)

	sentPtsFileInfo, err := msg.Pts.Stat()
	require.NoError(t, err)
	receivedPtsFileInfo, err := receivedClientHelloMsg.Pts.Stat()
	require.NoError(t, err)
	assert.True(t, os.SameFile(sentPtsFileInfo, receivedPtsFileInfo))
}

func TestSendAndReceiveClientGoodbyeMsg(t *testing.T) {
	msg := &ClientGoodbyeMsg{
		Reason: "Test reason",
	}
	receivedMsg := simulateSendAndReceive(t, msg)
	assert.Equal(t, msg, receivedMsg)
}

func TestSendAndReceiveServerHelloMsg(t *testing.T) {
	msg := &ServerHelloMsg{
		ClientId: 123,
	}
	receivedMsg := simulateSendAndReceive(t, msg)
	assert.Equal(t, msg, receivedMsg)
}

func TestSendAndReceiveServerGoodbyeMsg(t *testing.T) {
	msg := &ServerGoodbyeMsg{
		Reason: "Test reason",
	}
	receivedMsg := simulateSendAndReceive(t, msg)
	assert.Equal(t, msg, receivedMsg)
}

func TestSendAndReceiveTerminalResizeMsg(t *testing.T) {
	msg := &TerminalResizeMsg{
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
