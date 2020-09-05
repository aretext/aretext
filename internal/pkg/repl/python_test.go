package repl

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPythonRepl(t *testing.T) {
	// Ensure that the pyaretext package is in the default PYTHONPATH.
	os.Chdir("../../..")

	repl := NewPythonRepl()
	err := repl.Start()
	require.NoError(t, err)

	defer func() {
		err := repl.Terminate()
		require.NoError(t, err)
	}()

	output, err := repl.PollOutput()
	require.NoError(t, err)
	assert.Contains(t, output, "Aretext REPL")
	assert.Contains(t, output, ">>>")

	err = repl.SubmitInput("print('sum={}'.format(1+7))")
	require.NoError(t, err)

	output, err = repl.PollOutput()
	require.NoError(t, err)
	assert.Equal(t, "sum=8\n>>> ", output)

}
