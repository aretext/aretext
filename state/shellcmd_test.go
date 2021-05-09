package state

import (
	"fmt"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/aretext/aretext/selection"
)

func TestRunShellCmd(t *testing.T) {
	withStateAndTmpDir(t, func(state *EditorState, dir string) {
		p := path.Join(dir, "test-output.txt")
		cmd := fmt.Sprintf(`printf "hello" > %s`, p)
		RunShellCmd(state, cmd)
		data, err := os.ReadFile(p)
		require.NoError(t, err)
		assert.Equal(t, "hello", string(data))
	})
}

func TestRunShellCmdWithSelection(t *testing.T) {
	withStateAndTmpDir(t, func(state *EditorState, dir string) {
		for _, r := range "foobar" {
			InsertRune(state, r)
		}
		ToggleVisualMode(state, selection.ModeLine)

		p := path.Join(dir, "test-output.txt")
		cmd := fmt.Sprintf(`printf "$SELECTION" > %s`, p)
		RunShellCmd(state, cmd)
		data, err := os.ReadFile(p)
		require.NoError(t, err)
		assert.Equal(t, "foobar", string(data))
	})
}

func withStateAndTmpDir(t *testing.T, f func(*EditorState, string)) {
	suspendScreenFunc := func(f func() error) error { return f() }
	state := NewEditorState(100, 100, nil, suspendScreenFunc)

	dir, err := os.MkdirTemp("", "aretext")
	require.NoError(t, err)
	defer os.RemoveAll(dir)

	f(state, dir)
}
