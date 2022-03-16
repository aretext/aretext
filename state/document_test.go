package state

import (
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/aretext/aretext/syntax"
)

func createTestFile(t *testing.T, contents string) (path string, cleanup func()) {
	f, err := os.CreateTemp(os.TempDir(), "aretext-")
	require.NoError(t, err)
	defer f.Close()

	_, err = io.WriteString(f, contents)
	require.NoError(t, err)

	cleanup = func() { os.Remove(f.Name()) }
	return f.Name(), cleanup
}

func startOfDocLocator(LocatorParams) uint64 { return 0 }

func TestLoadDocumentShowStatus(t *testing.T) {
	// Start with an empty document.
	state := NewEditorState(100, 100, nil, nil)
	assert.Equal(t, "", state.documentBuffer.textTree.String())
	assert.Equal(t, "", state.FileWatcher().Path())

	// Load a document.
	path, cleanup := createTestFile(t, "abcd")
	LoadDocument(state, path, true, startOfDocLocator)
	defer state.fileWatcher.Stop()

	// Expect that the text and watcher are installed.
	assert.Equal(t, "abcd", state.documentBuffer.textTree.String())
	assert.Equal(t, path, state.FileWatcher().Path())

	// Expect success message.
	assert.Contains(t, state.statusMsg.Text, "Opened")
	assert.Equal(t, StatusMsgStyleSuccess, state.statusMsg.Style)

	// Delete the test file.
	cleanup()

	// Load a non-existent path, expect error msg.
	LoadDocument(state, path, true, startOfDocLocator)
	defer state.fileWatcher.Stop()
	assert.Contains(t, state.statusMsg.Text, "Could not open")
	assert.Equal(t, StatusMsgStyleError, state.statusMsg.Style)
}

func TestLoadDocumentSameFile(t *testing.T) {
	// Load the initial document.
	path, cleanup := createTestFile(t, "abcd\nefghi\njklmnop\nqrst")
	defer cleanup()
	state := NewEditorState(5, 3, nil, nil)
	LoadDocument(state, path, true, startOfDocLocator)
	state.documentBuffer.cursor.position = 22

	// Scroll to cursor at end of document.
	ScrollViewToCursor(state)
	assert.Equal(t, uint64(16), state.documentBuffer.view.textOrigin)

	// Set the syntax.
	SetSyntax(state, syntax.LanguageJson)
	assert.Equal(t, syntax.LanguageJson, state.documentBuffer.syntaxLanguage)

	// Update the file with shorter text and reload.
	err := os.WriteFile(path, []byte("ab"), 0644)
	require.NoError(t, err)
	ReloadDocument(state)
	defer state.fileWatcher.Stop()

	// Expect that the cursor moved back to the end of the text,
	// the view scrolled to make the cursor visible,
	// and the syntax language is preserved.
	assert.Equal(t, "ab", state.documentBuffer.textTree.String())
	assert.Equal(t, uint64(1), state.documentBuffer.cursor.position)
	assert.Equal(t, uint64(0), state.documentBuffer.view.textOrigin)
	assert.Equal(t, syntax.LanguageJson, state.documentBuffer.syntaxLanguage)
}

func TestLoadDocumentDifferentFile(t *testing.T) {
	// Load the initial document.
	path, cleanup := createTestFile(t, "abcd\nefghi\njklmnop\nqrst")
	defer cleanup()
	state := NewEditorState(5, 3, nil, nil)
	LoadDocument(state, path, true, startOfDocLocator)
	state.documentBuffer.cursor.position = 22

	// Scroll to cursor at end of document.
	ScrollViewToCursor(state)
	assert.Equal(t, uint64(16), state.documentBuffer.view.textOrigin)

	// Set the syntax.
	SetSyntax(state, syntax.LanguageJson)
	assert.Equal(t, syntax.LanguageJson, state.documentBuffer.syntaxLanguage)

	// Load a new document with a shorter text.
	path2, cleanup2 := createTestFile(t, "ab")
	defer cleanup2()
	LoadDocument(state, path2, true, startOfDocLocator)
	defer state.fileWatcher.Stop()

	// Expect that the cursor, view, and syntax are reset.
	assert.Equal(t, "ab", state.documentBuffer.textTree.String())
	assert.Equal(t, uint64(0), state.documentBuffer.cursor.position)
	assert.Equal(t, uint64(0), state.documentBuffer.view.textOrigin)
	assert.Equal(t, syntax.LanguagePlaintext, state.documentBuffer.syntaxLanguage)
}

func TestLoadPrevDocument(t *testing.T) {
	// Load the initial document.
	path, cleanup := createTestFile(t, "abcd\nefghi\njklmnop\nqrst")
	defer cleanup()
	state := NewEditorState(5, 3, nil, nil)
	LoadDocument(state, path, true, startOfDocLocator)
	MoveCursor(state, func(LocatorParams) uint64 {
		return 7
	})

	// Load another document.
	path2, cleanup2 := createTestFile(t, "xyz")
	defer cleanup2()
	LoadDocument(state, path2, true, startOfDocLocator)
	assert.Equal(t, "xyz", state.documentBuffer.textTree.String())

	// Return to the previous document.
	LoadPrevDocument(state)
	assert.Equal(t, "abcd\nefghi\njklmnop\nqrst", state.documentBuffer.textTree.String())
	assert.Equal(t, path, state.fileWatcher.Path())
	assert.Equal(t, uint64(7), state.documentBuffer.cursor.position)
}

func TestLoadNextDocument(t *testing.T) {
	// Load the initial document.
	path, cleanup := createTestFile(t, "abcd\nefghi\njklmnop\nqrst")
	defer cleanup()
	state := NewEditorState(5, 3, nil, nil)
	LoadDocument(state, path, true, startOfDocLocator)
	MoveCursor(state, func(LocatorParams) uint64 { return 7 })

	// Load another document.
	path2, cleanup2 := createTestFile(t, "qrs\ntuv\nwxyz")
	defer cleanup2()
	LoadDocument(state, path2, true, startOfDocLocator)
	assert.Equal(t, path2, state.fileWatcher.Path())
	MoveCursor(state, func(LocatorParams) uint64 { return 5 })

	// Return to the previous document.
	LoadPrevDocument(state)
	assert.Equal(t, path, state.fileWatcher.Path())

	// Forward to the next document.
	LoadNextDocument(state)
	assert.Equal(t, path2, state.fileWatcher.Path())
	assert.Equal(t, "qrs\ntuv\nwxyz", state.documentBuffer.textTree.String())
	assert.Equal(t, uint64(5), state.documentBuffer.cursor.position)
}

func TestLoadDocumentIncrementConfigVersion(t *testing.T) {
	// Start with an empty document.
	state := NewEditorState(100, 100, nil, nil)
	assert.Equal(t, state.ConfigVersion(), 0)

	// Load a document.
	path, cleanup := createTestFile(t, "abcd")
	LoadDocument(state, path, true, startOfDocLocator)
	defer state.fileWatcher.Stop()
	defer cleanup()

	// Expect that the config version was bumped.
	assert.Equal(t, state.ConfigVersion(), 1)
}

func TestReloadDocumentWithMenuOpen(t *testing.T) {
	// Load the initial document.
	path, cleanup := createTestFile(t, "abcd\nefghi\njklmnop\nqrst")
	defer cleanup()
	state := NewEditorState(5, 3, nil, nil)
	LoadDocument(state, path, true, startOfDocLocator)

	// Open the command menu
	ShowMenu(state, MenuStyleCommand, nil)
	assert.Equal(t, state.InputMode(), InputModeMenu)

	// Update the file with shorter text and reload.
	err := os.WriteFile(path, []byte("ab"), 0644)
	require.NoError(t, err)
	ReloadDocument(state)
	defer state.fileWatcher.Stop()

	// Expect that the input mode is normal and the menu is hidden.
	assert.Equal(t, "ab", state.documentBuffer.textTree.String())
	assert.Equal(t, InputModeNormal, state.InputMode())
	assert.False(t, state.Menu().Visible())
}

func TestSaveDocument(t *testing.T) {
	// Start with an empty document.
	state := NewEditorState(100, 100, nil, nil)

	// Load an existing document.
	path, cleanup := createTestFile(t, "")
	defer cleanup()
	LoadDocument(state, path, true, startOfDocLocator)
	defer state.fileWatcher.Stop()

	// Modify and save the document
	InsertRune(state, 'x')
	SaveDocument(state)
	defer state.fileWatcher.Stop()

	// Expect a success message.
	assert.Contains(t, state.statusMsg.Text, "Saved")
	assert.Equal(t, StatusMsgStyleSuccess, state.statusMsg.Style)

	// Check that the changes were persisted
	contents, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, "x\n", string(contents))
}

func TestAbortIfFileExistsWithChangedContent(t *testing.T) {
	testCases := []struct {
		name        string
		didChange   bool
		expectAbort bool
	}{
		{
			name:        "no changes should commit",
			didChange:   false,
			expectAbort: false,
		},
		{
			name:        "changes should abort",
			didChange:   true,
			expectAbort: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Load the initial document.
			path, cleanup := createTestFile(t, "")
			defer cleanup()
			state := NewEditorState(100, 100, nil, nil)
			LoadDocument(state, path, true, startOfDocLocator)

			// Modify the file.
			if tc.didChange {
				f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
				require.NoError(t, err)
				defer f.Close()
				_, err = io.WriteString(f, "test")
				require.NoError(t, err)
			}

			// Attempt an operation, but abort if the file changed.
			AbortIfFileExistsWithChangedContent(state, func(state *EditorState) {
				SetStatusMsg(state, StatusMsg{
					Style: StatusMsgStyleSuccess,
					Text:  "Operation executed",
				})
			})

			if tc.expectAbort {
				assert.Equal(t, StatusMsgStyleError, state.statusMsg.Style)
				assert.Contains(t, state.statusMsg.Text, "changed since last save")
			} else {
				assert.Equal(t, StatusMsgStyleSuccess, state.statusMsg.Style)
				assert.Equal(t, "Operation executed", state.statusMsg.Text)
			}
		})
	}
}

func TestAbortIfFileExistsWithChangedContentNewFile(t *testing.T) {
	dir, err := os.MkdirTemp("", "aretext")
	require.NoError(t, err)
	defer os.RemoveAll(dir)

	path := filepath.Join(dir, "aretext-does-not-exist")
	state := NewEditorState(100, 100, nil, nil)
	LoadDocument(state, path, false, startOfDocLocator)

	// File doesn't exist on disk, so the operation should succeed.
	AbortIfFileExistsWithChangedContent(state, func(state *EditorState) {
		SetStatusMsg(state, StatusMsg{
			Style: StatusMsgStyleSuccess,
			Text:  "Operation executed",
		})
	})
	assert.Equal(t, StatusMsgStyleSuccess, state.statusMsg.Style)
	assert.Equal(t, "Operation executed", state.statusMsg.Text)

	// Create the file on disk.
	f, err := os.Create(path)
	require.NoError(t, err)
	defer f.Close()

	_, err = io.WriteString(f, "abcd")
	require.NoError(t, err)

	// Now the operation should abort.
	AbortIfFileExistsWithChangedContent(state, func(state *EditorState) {
		SetStatusMsg(state, StatusMsg{
			Style: StatusMsgStyleSuccess,
			Text:  "Operation executed",
		})
	})
	assert.Equal(t, StatusMsgStyleError, state.statusMsg.Style)
	assert.Contains(t, state.statusMsg.Text, "changed since last save")
}

func TestAbortIfFileExistsWithChangedContentFileDeleted(t *testing.T) {
	// Load the initial document.
	path, cleanup := createTestFile(t, "abc")
	state := NewEditorState(100, 100, nil, nil)
	LoadDocument(state, path, true, startOfDocLocator)

	// Delete the file
	cleanup()

	// The operation should succeed, since the file does not exist.
	AbortIfFileExistsWithChangedContent(state, func(state *EditorState) {
		SetStatusMsg(state, StatusMsg{
			Style: StatusMsgStyleSuccess,
			Text:  "Operation executed",
		})
	})
	assert.Equal(t, StatusMsgStyleSuccess, state.statusMsg.Style)
	assert.Equal(t, "Operation executed", state.statusMsg.Text)
}
