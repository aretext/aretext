package state

import (
	"fmt"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/aretext/aretext/config"
	"github.com/aretext/aretext/selection"
)

func TestRunShellCmd(t *testing.T) {
	testCases := []struct {
		name string
		mode string
	}{
		{
			name: "mode silent",
			mode: config.CmdModeSilent,
		},
		{
			name: "mode terminal",
			mode: config.CmdModeTerminal,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			setupShellCmdTest(t, func(state *EditorState, dir string) {
				p := path.Join(dir, "test-output.txt")
				cmd := fmt.Sprintf(`printf "hello" > %s`, p)
				RunShellCmd(state, cmd, tc.mode)
				data, err := os.ReadFile(p)
				require.NoError(t, err)
				assert.Equal(t, "hello", string(data))
			})
		})
	}
}

func TestRunShellCmdFilePathEnvVar(t *testing.T) {
	setupShellCmdTest(t, func(state *EditorState, dir string) {
		filePath := path.Join(dir, "test-input.txt")
		os.WriteFile(filePath, []byte("xyz"), 0644)
		LoadDocument(state, filePath, true, func(LocatorParams) uint64 { return 0 })

		p := path.Join(dir, "test-output.txt")
		cmd := fmt.Sprintf(`printenv FILEPATH > %s`, p)
		RunShellCmd(state, cmd, config.CmdModeSilent)
		data, err := os.ReadFile(p)
		require.NoError(t, err)
		assert.Equal(t, filePath+"\n", string(data))
	})
}

func TestRunShellCmdWordEnvVar(t *testing.T) {
	testCases := []struct {
		name               string
		text               string
		cursorPos          uint64
		expectedWordEnvVar string
	}{
		{
			name:               "empty document",
			text:               "",
			cursorPos:          0,
			expectedWordEnvVar: "",
		},
		{
			name:               "non-empty word",
			text:               "abcd  xyz  123",
			cursorPos:          7,
			expectedWordEnvVar: "xyz",
		},
		{
			name:               "whitespace between words",
			text:               "abcd  xyz  123",
			cursorPos:          4,
			expectedWordEnvVar: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			setupShellCmdTest(t, func(state *EditorState, dir string) {
				for _, r := range tc.text {
					InsertRune(state, r)
				}
				MoveCursor(state, func(p LocatorParams) uint64 { return tc.cursorPos })

				p := path.Join(dir, "test-output.txt")
				cmd := fmt.Sprintf(`printenv WORD > %s`, p)
				RunShellCmd(state, cmd, config.CmdModeSilent)
				data, err := os.ReadFile(p)
				require.NoError(t, err)
				assert.Equal(t, tc.expectedWordEnvVar+"\n", string(data))
			})
		})
	}
}

func TestRunShellCmdWithSelection(t *testing.T) {
	setupShellCmdTest(t, func(state *EditorState, dir string) {
		for _, r := range "foobar" {
			InsertRune(state, r)
		}
		ToggleVisualMode(state, selection.ModeLine)

		p := path.Join(dir, "test-output.txt")
		cmd := fmt.Sprintf(`printenv SELECTION > %s`, p)
		RunShellCmd(state, cmd, config.CmdModeSilent)
		data, err := os.ReadFile(p)
		require.NoError(t, err)
		assert.Equal(t, "foobar\n", string(data))
	})
}

func TestRunShellCmdInsertIntoDocument(t *testing.T) {
	setupShellCmdTest(t, func(state *EditorState, dir string) {
		p := path.Join(dir, "test-output.txt")
		err := os.WriteFile(p, []byte("hello world"), 0644)
		require.NoError(t, err)
		cmd := fmt.Sprintf("cat %s", p)
		RunShellCmd(state, cmd, config.CmdModeInsert)
		s := state.documentBuffer.textTree.String()
		cursorPos := state.documentBuffer.cursor.position
		assert.Equal(t, "hello world", s)
		assert.Equal(t, uint64(10), cursorPos)
		assert.Equal(t, InputModeNormal, state.InputMode())
	})
}

func TestRunShellCmdInsertIntoDocumentWithSelection(t *testing.T) {
	setupShellCmdTest(t, func(state *EditorState, dir string) {
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
		RunShellCmd(state, cmd, config.CmdModeInsert)
		s := state.documentBuffer.textTree.String()
		cursorPos := state.documentBuffer.cursor.position
		assert.Equal(t, "foohello worldr", s)
		assert.Equal(t, uint64(13), cursorPos)
		assert.Equal(t, InputModeNormal, state.InputMode())
	})
}

func TestRunShellCmdFileLocationsMenu(t *testing.T) {
	setupShellCmdTest(t, func(state *EditorState, dir string) {
		// Create a test file to load.
		p := path.Join(dir, "test-file.txt")
		err := os.WriteFile(p, []byte("ab\ncd\nef\ngh"), 0644)
		require.NoError(t, err)

		// Populate the location list with a single file location.
		cmd := fmt.Sprintf("echo '%s:2:cd'", p)
		RunShellCmd(state, cmd, config.CmdModeFileLocations)
		assert.Equal(t, InputModeMenu, state.InputMode())
		assert.True(t, state.Menu().Visible())

		// Verify that the location list menu opens.
		menuItems, _ := state.Menu().SearchResults()
		assert.Equal(t, 1, len(menuItems))
		expectedName := fmt.Sprintf("%s:2  cd", p)
		assert.Equal(t, expectedName, menuItems[0].Name)

		// Execute the menu item and verify that the document loads.
		ExecuteSelectedMenuItem(state)
		assert.Equal(t, p, state.fileWatcher.Path())
		assert.Equal(t, uint64(3), state.documentBuffer.cursor.position)
		text := state.documentBuffer.textTree.String()
		assert.Equal(t, "ab\ncd\nef\ngh", text)
	})
}

func setupShellCmdTest(t *testing.T, f func(*EditorState, string)) {
	oldShellEnv := os.Getenv("SHELL")
	defer os.Setenv("SHELL", oldShellEnv)
	os.Setenv("SHELL", "")

	suspendScreenFunc := func(f func() error) error { return f() }
	state := NewEditorState(100, 100, nil, suspendScreenFunc)

	dir, err := os.MkdirTemp("", "aretext")
	require.NoError(t, err)
	defer os.RemoveAll(dir)

	f(state, dir)
}
