package state

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/aretext/aretext/config"
	"github.com/aretext/aretext/selection"
)

func runShellCmdAndApplyAction(t *testing.T, state *EditorState, cmd string, mode string) {
	RunShellCmd(state, cmd, mode)
	if mode == config.CmdModeTerminal {
		return // executes synchronously
	}

	// Wait for asynchronous task to complete and apply resulting action.
	select {
	case action := <-state.TaskResultChan():
		action(state)

	case <-time.After(5 * time.Second):
		require.Fail(t, "Timed out")
	}
}

func TestRunShellCmd(t *testing.T) {
	setupShellCmdTest(t, func(state *EditorState, dir string) {
		p := filepath.Join(dir, "test-output.txt")
		cmd := fmt.Sprintf(`printf "hello" > %s`, p)
		runShellCmdAndApplyAction(t, state, cmd, config.CmdModeSilent)
		data, err := os.ReadFile(p)
		require.NoError(t, err)
		assert.Equal(t, "hello", string(data))
	})
}

func TestRunShellCmdFilePathEnvVar(t *testing.T) {
	setupShellCmdTest(t, func(state *EditorState, dir string) {
		filePath := filepath.Join(dir, "test-input.txt")
		os.WriteFile(filePath, []byte("xyz"), 0644)
		LoadDocument(state, filePath, true, func(LocatorParams) uint64 { return 0 })

		p := filepath.Join(dir, "test-output.txt")
		cmd := fmt.Sprintf(`printenv FILEPATH > %s`, p)
		runShellCmdAndApplyAction(t, state, cmd, config.CmdModeSilent)
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

				p := filepath.Join(dir, "test-output.txt")
				cmd := fmt.Sprintf(`printenv WORD > %s`, p)
				runShellCmdAndApplyAction(t, state, cmd, config.CmdModeSilent)
				data, err := os.ReadFile(p)
				require.NoError(t, err)
				assert.Equal(t, tc.expectedWordEnvVar+"\n", string(data))
			})
		})
	}
}

func TestRunShellCmdLineAndColumnEnvVars(t *testing.T) {
	testCases := []struct {
		name                 string
		text                 string
		cursorPos            uint64
		expectedLineEnvVar   string
		expectedColumnEnvVar string
	}{
		{
			name:                 "empty document",
			text:                 "",
			cursorPos:            0,
			expectedLineEnvVar:   "1",
			expectedColumnEnvVar: "1",
		},
		{
			name:                 "single line",
			text:                 "abc",
			cursorPos:            0,
			expectedLineEnvVar:   "1",
			expectedColumnEnvVar: "1",
		},
		{
			name:                 "multiple lines, cursor on first line",
			text:                 "abc\ndef\nghi",
			cursorPos:            2,
			expectedLineEnvVar:   "1",
			expectedColumnEnvVar: "3",
		},
		{
			name:                 "multiple lines, cursor on second line",
			text:                 "abc\ndef\nghi",
			cursorPos:            4,
			expectedLineEnvVar:   "2",
			expectedColumnEnvVar: "1",
		},
		{
			name:                 "multiple lines, cursor on last line",
			text:                 "abc\ndef\nghi",
			cursorPos:            10,
			expectedLineEnvVar:   "3",
			expectedColumnEnvVar: "3",
		},
		{
			name:                 "line with multi-byte unicode",
			text:                 "\U0010AAAA abcd",
			cursorPos:            1,
			expectedLineEnvVar:   "1",
			expectedColumnEnvVar: "5", // column counts bytes, not runes.
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			setupShellCmdTest(t, func(state *EditorState, dir string) {
				for _, r := range tc.text {
					InsertRune(state, r)
				}
				MoveCursor(state, func(p LocatorParams) uint64 { return tc.cursorPos })

				p := filepath.Join(dir, "test-output.txt")
				cmd := fmt.Sprintf(`printenv LINE > %s; printenv COLUMN >> %s`, p, p)
				runShellCmdAndApplyAction(t, state, cmd, config.CmdModeSilent)
				data, err := os.ReadFile(p)
				require.NoError(t, err)
				expected := fmt.Sprintf("%s\n%s\n", tc.expectedLineEnvVar, tc.expectedColumnEnvVar)
				assert.Equal(t, expected, string(data))
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

		p := filepath.Join(dir, "test-output.txt")
		cmd := fmt.Sprintf(`printenv SELECTION > %s`, p)
		runShellCmdAndApplyAction(t, state, cmd, config.CmdModeSilent)
		data, err := os.ReadFile(p)
		require.NoError(t, err)
		assert.Equal(t, "foobar\n", string(data))
	})
}

func TestRunShellCmdInsertIntoDocument(t *testing.T) {
	testCases := []struct {
		name              string
		documentText      string
		insertedText      string
		cursorPos         uint64
		expectedCursorPos uint64
		expectedText      string
	}{
		{
			name:              "insert into empty document",
			documentText:      "",
			insertedText:      "hello world",
			cursorPos:         0,
			expectedCursorPos: 10,
			expectedText:      "hello world",
		},
		{
			name:              "insert into document with text",
			documentText:      "foo bar",
			insertedText:      "hello world",
			cursorPos:         3,
			expectedCursorPos: 14,
			expectedText:      "foo hello worldbar",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			setupShellCmdTest(t, func(state *EditorState, dir string) {
				// Setup initial state.
				for _, r := range tc.documentText {
					InsertRune(state, r)
				}
				MoveCursor(state, func(p LocatorParams) uint64 { return tc.cursorPos })

				// Create test file with content
				p := filepath.Join(dir, "test-output.txt")
				err := os.WriteFile(p, []byte(tc.insertedText), 0644)
				require.NoError(t, err)

				// Execute command to insert contents of text file.
				cmd := fmt.Sprintf("cat %s", p)
				runShellCmdAndApplyAction(t, state, cmd, config.CmdModeInsert)

				// Check the document state.
				s := state.documentBuffer.textTree.String()
				cursorPos := state.documentBuffer.cursor.position
				assert.Equal(t, tc.expectedText, s)
				assert.Equal(t, tc.expectedCursorPos, cursorPos)
				assert.Equal(t, InputModeNormal, state.InputMode())
			})
		})
	}
}

func TestRunShellCmdInsertIntoDocumentThenUndo(t *testing.T) {
	setupShellCmdTest(t, func(state *EditorState, dir string) {
		// Setup initial state.
		for _, r := range "abcd" {
			InsertRune(state, r)
		}
		MoveCursor(state, func(p LocatorParams) uint64 { return 2 })

		// Create test file with content
		p := filepath.Join(dir, "test-output.txt")
		err := os.WriteFile(p, []byte("xyz"), 0644)
		require.NoError(t, err)

		// Execute command to insert contents of text file.
		cmd := fmt.Sprintf("cat %s", p)
		runShellCmdAndApplyAction(t, state, cmd, config.CmdModeInsert)

		// Undo the last action.
		Undo(state)

		// Check the document state.
		s := state.documentBuffer.textTree.String()
		cursorPos := state.documentBuffer.cursor.position
		assert.Equal(t, "abcd", s)
		assert.Equal(t, uint64(2), cursorPos)
		assert.Equal(t, InputModeNormal, state.InputMode())
	})
}

func TestRunShellCmdInsertIntoDocumentWithSelection(t *testing.T) {
	testCases := []struct {
		name              string
		documentText      string
		insertedText      string
		selectionMode     selection.Mode
		cursorStartPos    uint64
		cursorEndPos      uint64
		expectedCursorPos uint64
		expectedText      string
	}{
		{
			name:              "charwise selection",
			documentText:      "foobar",
			insertedText:      "hello world",
			selectionMode:     selection.ModeChar,
			cursorStartPos:    3,
			cursorEndPos:      4,
			expectedCursorPos: 13,
			expectedText:      "foohello worldr",
		},
		{
			name:              "linewise selection",
			documentText:      "foo\nbar\nbaz\nbat",
			insertedText:      "hello world",
			selectionMode:     selection.ModeLine,
			cursorStartPos:    5,
			cursorEndPos:      9,
			expectedCursorPos: 14,
			expectedText:      "foo\nhello world\nbat",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			setupShellCmdTest(t, func(state *EditorState, dir string) {
				for _, r := range tc.documentText {
					InsertRune(state, r)
				}
				MoveCursor(state, func(p LocatorParams) uint64 { return tc.cursorStartPos })
				ToggleVisualMode(state, tc.selectionMode)
				MoveCursor(state, func(p LocatorParams) uint64 { return tc.cursorEndPos })

				p := filepath.Join(dir, "test-output.txt")
				err := os.WriteFile(p, []byte(tc.insertedText), 0644)
				require.NoError(t, err)
				cmd := fmt.Sprintf("cat %s", p)
				runShellCmdAndApplyAction(t, state, cmd, config.CmdModeInsert)
				s := state.documentBuffer.textTree.String()
				cursorPos := state.documentBuffer.cursor.position
				assert.Equal(t, tc.expectedText, s)
				assert.Equal(t, tc.expectedCursorPos, cursorPos)
				assert.Equal(t, InputModeNormal, state.InputMode())
			})
		})
	}
}

func TestRunShellCmdInsertChoiceMenu(t *testing.T) {
	setupShellCmdTest(t, func(state *EditorState, dir string) {
		// Run a command that outputs two lines.
		cmd := "printf 'abc\nxyz'"
		runShellCmdAndApplyAction(t, state, cmd, config.CmdModeInsertChoice)

		// Verify that the insert choice menu loads with the two lines.
		assert.Equal(t, InputModeMenu, state.InputMode())
		menuItems, _ := state.Menu().SearchResults()
		require.Equal(t, 2, len(menuItems))
		assert.Equal(t, "abc", menuItems[0].Name)
		assert.Equal(t, "xyz", menuItems[1].Name)

		// Execute the first menu item and verify the text is inserted.
		ExecuteSelectedMenuItem(state)
		s := state.documentBuffer.textTree.String()
		cursorPos := state.documentBuffer.cursor.position
		assert.Equal(t, "abc", s)
		assert.Equal(t, uint64(2), cursorPos)
		assert.Equal(t, InputModeNormal, state.InputMode())
	})
}

func TestRunShellCmdFileLocationsMenu(t *testing.T) {
	setupShellCmdTest(t, func(state *EditorState, dir string) {
		// Create a test file to load.
		p := filepath.Join(dir, "test-file.txt")
		err := os.WriteFile(p, []byte("ab\ncd\nef\ngh"), 0644)
		require.NoError(t, err)

		// Populate the location list with a single file location.
		cmd := fmt.Sprintf("echo '%s:2:cd'", p)
		runShellCmdAndApplyAction(t, state, cmd, config.CmdModeFileLocations)

		// Verify that the location list menu opens.
		assert.Equal(t, InputModeMenu, state.InputMode())
		menuItems, _ := state.Menu().SearchResults()
		require.Equal(t, 1, len(menuItems))
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

func TestRunShellCmdWorkingDirMenu(t *testing.T) {
	setupShellCmdTest(t, func(state *EditorState, dir string) {
		// Save the original working dir so we can restore it later.
		originalWorkingDir, err := os.Getwd()
		require.NoError(t, err)
		defer os.Chdir(originalWorkingDir)

		// Populate the menu with a path to a temp dir.
		dirPath := t.TempDir()
		dirPath, err = filepath.EvalSymlinks(dirPath)
		require.NoError(t, err)
		cmd := fmt.Sprintf("echo '%s'", dirPath)
		runShellCmdAndApplyAction(t, state, cmd, config.CmdModeWorkingDir)

		// Verify that the menu shows the path.
		assert.Equal(t, InputModeMenu, state.InputMode())
		menuItems, _ := state.Menu().SearchResults()
		require.Equal(t, 1, len(menuItems))
		assert.Equal(t, dirPath, menuItems[0].Name)

		// Execute the menu item and verify that the working directory changes.
		ExecuteSelectedMenuItem(state)
		workingDir, err := os.Getwd()
		require.NoError(t, err)
		assert.Equal(t, dirPath, workingDir)
	})
}

func setupShellCmdTest(t *testing.T, f func(*EditorState, string)) {
	oldShellEnv := os.Getenv("SHELL")
	defer os.Setenv("SHELL", oldShellEnv)
	os.Setenv("SHELL", "")

	oldAretextShellEnv := os.Getenv("ARETEXT_SHELL")
	defer os.Setenv("ARETEXT_SHELL", oldAretextShellEnv)
	os.Setenv("ARETEXT_SHELL", "")

	suspendScreenFunc := func(f func() error) error { return f() }
	state := NewEditorState(100, 100, nil, suspendScreenFunc)

	dir := t.TempDir()

	f(state, dir)
}
