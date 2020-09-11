package repl

import (
	"net"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func withReplAndListener(t *testing.T, f func(*testing.T, Repl, net.Listener)) {
	// Ensure that the pyaretext package is in the default PYTHONPATH.
	dir, err := os.Getwd()
	require.NoError(t, err)
	err = os.Chdir("../../..")
	require.NoError(t, err)
	defer os.Chdir(dir)

	// Start a fake RPC server for the client to connect to.
	listener, err := net.Listen("tcp", "127.0.0.1:")
	require.NoError(t, err)
	defer listener.Close()

	// Start the Python REPL.
	apiKey := "abcd1234"
	apiConfig := NewApiConfig(listener.Addr(), apiKey)
	repl := NewPythonRepl(apiConfig)
	err = repl.Start()
	require.NoError(t, err)

	// Terminate the Python REPL once the test finishes.
	defer func() {
		err := repl.Terminate()
		require.NoError(t, err)
	}()

	f(t, repl, listener)
}

func TestPythonReplOutput(t *testing.T) {
	withReplAndListener(t, func(t *testing.T, repl Repl, Listener net.Listener) {
		output := <-repl.OutputChan()
		assert.Contains(t, output, "Aretext REPL")
		assert.Contains(t, output, ">>>")

		err := repl.SubmitInput("print('sum={}'.format(1+7))")
		require.NoError(t, err)

		var sb strings.Builder
		for {
			select {
			case s := <-repl.OutputChan():
				_, err := sb.WriteString(s)
				require.NoError(t, err)

				if sb.String() == "sum=8\n>>> " {
					t.Logf("Found expected output string: '%s'\n", sb.String())
					return
				}

			case <-time.After(time.Second * 10):
				assert.Failf(t, "Timed out waiting for full output, got '%s'\n", sb.String())
				return
			}

		}
	})
}

func TestPythonReplConnectToRpcServer(t *testing.T) {
	withReplAndListener(t, func(t *testing.T, repl Repl, listener net.Listener) {
		// Detect when a client connects to the RPC server.
		acceptedChan := make(chan struct{}, 0)
		go func() {
			conn, err := listener.Accept()
			if err == nil {
				defer conn.Close()
				close(acceptedChan)
			}
		}()

		// Consume and discard all output from the REPL.
		go func() {
			for {
				output, ok := <-repl.OutputChan()
				if ok {
					t.Logf("Output from REPL: %s\n", output)
				} else {
					t.Log("REPL output channel closed")
					return
				}
			}
		}()

		// Attempt to connect to the RPC server.
		t.Log("Submitting input to REPL...")
		err := repl.SubmitInput("from pyaretext.api.rpcclient import DEFAULT_CLIENT")
		require.NoError(t, err)

		// Verify that the client connected.
		t.Log("Waiting for client to connect to RPC server...")
		select {
		case <-acceptedChan:
			t.Log("Client connected successfully\n")
		case <-time.After(time.Second * 10):
			assert.Fail(t, "Timed out waiting for client to connect\n")
		}
	})
}
