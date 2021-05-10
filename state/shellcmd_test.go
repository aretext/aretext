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
	testCases := []struct {
		name   string
		output ShellCmdOutput
	}{
		{
			name:   "output none",
			output: ShellCmdOutputNone,
		},
		{
			name:   "output terminal",
			output: ShellCmdOutputTerminal,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			withStateAndTmpDir(t, func(state *EditorState, dir string) {
				p := path.Join(dir, "test-output.txt")
				cmd := fmt.Sprintf(`printf "hello" > %s`, p)
				RunShellCmd(state, cmd, tc.output)
				data, err := os.ReadFile(p)
				require.NoError(t, err)
				assert.Equal(t, "hello", string(data))
			})
		})
	}
}

func TestRunShellCmdWithSelection(t *testing.T) {
	withStateAndTmpDir(t, func(state *EditorState, dir string) {
		for _, r := range "foobar" {
			InsertRune(state, r)
		}
		ToggleVisualMode(state, selection.ModeLine)

		p := path.Join(dir, "test-output.txt")
		cmd := fmt.Sprintf(`printenv SELECTION > %s`, p)
		RunShellCmd(state, cmd, ShellCmdOutputNone)
		data, err := os.ReadFile(p)
		require.NoError(t, err)
		assert.Equal(t, "foobar\n", string(data))
	})
}

func TestRunShellCmdInsertIntoDocument(t *testing.T) {
	withStateAndTmpDir(t, func(state *EditorState, dir string) {
		p := path.Join(dir, "test-output.txt")
		err := os.WriteFile(p, []byte("hello world"), 0644)
		require.NoError(t, err)
		cmd := fmt.Sprintf("cat %s", p)
		RunShellCmd(state, cmd, ShellCmdOutputDocumentInsert)
		s := state.documentBuffer.textTree.String()
		cursorPos := state.documentBuffer.cursor.position
		assert.Equal(t, "hello world", s)
		assert.Equal(t, uint64(10), cursorPos)
		assert.Equal(t, InputModeNormal, state.InputMode())
	})
}

func TestRunShellCmdInsertIntoDocumentWithSelection(t *testing.T) {
	withStateAndTmpDir(t, func(state *EditorState, dir string) {
		for _, r := range "foobar" {
			InsertRune(state, r)
		}
		MoveCursor(state, func(p LocatorParams) uint64 { return 3 })
		ToggleVisualMode(state, selection.ModeChar)
		MoveCursor(state, func(p LocatorParams) uint64 { return 4 })

		p := path.Join(dir, "test-output.txt")
		err := os.WriteFile(p, []byte("hello world"), 0644)
		require.NoError(t, err)
		cmd := fmt.Sprintf("cat %s", p)
		RunShellCmd(state, cmd, ShellCmdOutputDocumentInsert)
		s := state.documentBuffer.textTree.String()
		cursorPos := state.documentBuffer.cursor.position
		assert.Equal(t, "foohello worldr", s)
		assert.Equal(t, uint64(13), cursorPos)
		assert.Equal(t, InputModeNormal, state.InputMode())
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
