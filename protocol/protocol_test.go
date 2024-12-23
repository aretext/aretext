package protocol

import (
	"context"
	"os"
	"path/filepath"
	"testing"

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

	messages := []Message{
		ClientHelloMsg{
			FilePath:    "/test/file",
			WorkingDir:  "/test",
			TerminalEnv: []string{"TERM=tmux"},
			Pts:         fakePts,
		},
	}

	go runServer(ctx, socketPath, messages)

}
