package state

import (
	"io"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/aretext/aretext/syntax"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createTestFile(t *testing.T, contents string) (path string, cleanup func()) {
	f, err := ioutil.TempFile(os.TempDir(), "aretext-")
	require.NoError(t, err)
	defer f.Close()

	_, err = io.WriteString(f, contents)
	require.NoError(t, err)

	cleanup = func() { os.Remove(f.Name()) }
	return f.Name(), cleanup
}

func TestLoadDocumentShowStatus(t *testing.T) {
	// Start with an empty document.
	state := NewEditorState(100, 100, nil)
	assert.Equal(t, "", state.documentBuffer.textTree.String())
	assert.Equal(t, "", state.FileWatcher().Path())

	// Load a document.
	path, cleanup := createTestFile(t, "abcd")
	LoadDocument(state, path, true)
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
	LoadDocument(state, path, true)
	defer state.fileWatcher.Stop()
	assert.Contains(t, state.statusMsg.Text, "Could not open")
	assert.Equal(t, StatusMsgStyleError, state.statusMsg.Style)
}

func TestLoadDocumentSameFile(t *testing.T) {
	// Load the initial document.
	path, cleanup := createTestFile(t, "abcd\nefghi\njklmnop\nqrst")
	defer cleanup()
	state := NewEditorState(5, 3, nil)
	LoadDocument(state, path, true)
	state.documentBuffer.cursor.position = 22

	// Scroll to cursor at end of document.
	ScrollViewToCursor(state)
	assert.Equal(t, uint64(16), state.documentBuffer.view.textOrigin)

	// Set the syntax.
	SetSyntax(state, syntax.LanguageJson)
	assert.Equal(t, syntax.LanguageJson, state.documentBuffer.syntaxLanguage)

	// Update the file with shorter text and reload.
	err := ioutil.WriteFile(path, []byte("ab"), 0644)
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
	state := NewEditorState(5, 3, nil)
	LoadDocument(state, path, true)
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
	LoadDocument(state, path2, true)
	defer state.fileWatcher.Stop()

	// Expect that the cursor, view, and syntax are reset.
	assert.Equal(t, "ab", state.documentBuffer.textTree.String())
	assert.Equal(t, uint64(0), state.documentBuffer.cursor.position)
	assert.Equal(t, uint64(0), state.documentBuffer.view.textOrigin)
	assert.Equal(t, syntax.LanguageUndefined, state.documentBuffer.syntaxLanguage)
}

func TestSaveDocument(t *testing.T) {
	// Start with an empty document.
	state := NewEditorState(100, 100, nil)

	// Load an existing document.
	path, cleanup := createTestFile(t, "")
	defer cleanup()
	LoadDocument(state, path, true)
	defer state.fileWatcher.Stop()

	// Modify and save the document
	InsertRune(state, 'x')
	SaveDocument(state, true)
	defer state.fileWatcher.Stop()

	// Expect a success message.
	assert.Contains(t, state.statusMsg.Text, "Saved")
	assert.Equal(t, StatusMsgStyleSuccess, state.statusMsg.Style)

	// Check that the changes were persisted
	contents, err := ioutil.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, "x\n", string(contents))
}

func TestAbortIfFileChanged(t *testing.T) {
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
			state := NewEditorState(100, 100, nil)
			LoadDocument(state, path, true)

			// Modify the file.
			if tc.didChange {
				f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
				require.NoError(t, err)
				defer f.Close()
				_, err = io.WriteString(f, "test")
				require.NoError(t, err)

				// Wait for the watcher to detect the change.
				select {
				case <-state.fileWatcher.ChangedChan():
					break
				case <-time.After(time.Second * 10):
					assert.Fail(t, "Timed out waiting for change")
					return
				}
			}

			// Attempt an operation, but abort if the file changed.
			AbortIfFileChanged(state, func(state *EditorState) {
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
				assert.Equal(t, state.statusMsg.Text, "Operation executed")
			}
		})
	}
}
