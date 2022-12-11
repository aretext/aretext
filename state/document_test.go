package state

import (
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/aretext/aretext/config"
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

	// Update the file with shorter text and reload.
	err := os.WriteFile(path, []byte("ab"), 0644)
	require.NoError(t, err)
	ReloadDocument(state)
	defer state.fileWatcher.Stop()

	// Expect that the cursor moved back to the end of the text,
	// and the view scrolled to make the cursor visible.
	assert.Equal(t, "ab", state.documentBuffer.textTree.String())
	assert.Equal(t, uint64(1), state.documentBuffer.cursor.position)
	assert.Equal(t, uint64(0), state.documentBuffer.view.textOrigin)
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

func TestLoadDocumentIncrementLoadCount(t *testing.T) {
	// Start with an empty document.
	state := NewEditorState(100, 100, nil, nil)
	assert.Equal(t, state.DocumentLoadCount(), 0)

	// Load a document.
	path, cleanup := createTestFile(t, "abcd")
	LoadDocument(state, path, true, startOfDocLocator)
	defer state.fileWatcher.Stop()
	defer cleanup()

	// Expect that the load count was bumped.
	assert.Equal(t, state.DocumentLoadCount(), 1)
}

func TestReloadDocumentAlignCursorAndScroll(t *testing.T) {
	// Load the initial document.
	initialText := "abcd\nefghi\njklmnop\nqrst"
	path, cleanup := createTestFile(t, initialText)
	defer cleanup()
	state := NewEditorState(5, 3, nil, nil)
	LoadDocument(state, path, true, startOfDocLocator)
	state.documentBuffer.cursor.position = 14

	// Scroll to cursor at end of document.
	ScrollViewToCursor(state)
	assert.Equal(t, uint64(5), state.documentBuffer.view.textOrigin)

	// Add some lines to the beginning of the document.
	insertedText := "123\n456\n789\nqrs\ntuv\nwx\nyz\n"
	err := os.WriteFile(path, []byte(insertedText+initialText), 0644)
	require.NoError(t, err)

	// Reload the document.
	ReloadDocument(state)
	defer state.fileWatcher.Stop()

	// Expect that the cursor and scroll position moved to the
	// equivalent line in the new document.
	assert.Equal(t, insertedText+initialText, state.documentBuffer.textTree.String())
	assert.Equal(t, uint64(40), state.documentBuffer.cursor.position)
	assert.Equal(t, uint64(31), state.documentBuffer.view.textOrigin)
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

func TestReloadDocumentPreserveSearchQueryAndDirection(t *testing.T) {
	testCases := []struct {
		name           string
		direction      SearchDirection
		completeSearch bool
	}{
		{
			name:           "search forward, complete search",
			direction:      SearchDirectionForward,
			completeSearch: true,
		},
		{
			name:           "search backward, complete search",
			direction:      SearchDirectionBackward,
			completeSearch: true,
		},
		{
			name:           "search forward, incomplete search",
			direction:      SearchDirectionForward,
			completeSearch: false,
		},
		{
			name:           "search backward, incomplete search",
			direction:      SearchDirectionBackward,
			completeSearch: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Load the initial document.
			path, cleanup := createTestFile(t, "abcd\nefghi\njklmnop\nqrst")
			defer cleanup()
			state := NewEditorState(5, 3, nil, nil)
			LoadDocument(state, path, true, startOfDocLocator)

			// Text search.
			StartSearch(state, tc.direction)
			AppendRuneToSearchQuery(state, 'e')
			AppendRuneToSearchQuery(state, 'f')
			AppendRuneToSearchQuery(state, 'g')
			if tc.completeSearch {
				CompleteSearch(state, true)
			}

			// Update the file with shorter text and reload.
			err := os.WriteFile(path, []byte("abcefghijk"), 0644)
			require.NoError(t, err)
			ReloadDocument(state)
			defer state.fileWatcher.Stop()

			// Expect we're in normal mode after reload.
			assert.Equal(t, InputModeNormal, state.InputMode())

			// Expect that the search query and direction are preserved.
			expectedSearch := searchState{query: "efg", direction: tc.direction}
			if tc.completeSearch {
				expectedSearch.history = []string{"efg"}
			}
			assert.Equal(t, expectedSearch, state.documentBuffer.search)
		})
	}
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

func TestSaveDocumentIfUnsavedChanges(t *testing.T) {
	// Start with an empty document.
	state := NewEditorState(100, 100, nil, nil)
	path := filepath.Join(t.TempDir(), "test-save-document-if-unsaved-changes.txt")
	LoadDocument(state, path, false, func(LocatorParams) uint64 { return 0 })
	defer state.fileWatcher.Stop()

	// Save the document. The file should be created even though the document is empty.
	SaveDocumentIfUnsavedChanges(state)
	_, err := os.Stat(path)
	require.NoError(t, err)

	// Change the document on disk so we can detect if the file changes on next save.
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_TRUNC, 0644)
	require.NoError(t, err)
	defer f.Close()
	_, err = io.WriteString(f, "abcdefg")
	require.NoError(t, err)

	// Save again, but expect that the file is NOT saved since there are no unsaved changes.
	SaveDocumentIfUnsavedChanges(state)
	contents, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, "abcdefg", string(contents))

	// Modify and save the document, then check that the file was changed.
	InsertRune(state, 'x')
	SaveDocumentIfUnsavedChanges(state)
	contents, err = os.ReadFile(path)
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
	dir := t.TempDir()

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

func TestDeduplicateCustomMenuItems(t *testing.T) {
	// Configure custom menu items with duplicate names.
	configRuleSet := config.RuleSet{
		{
			Name:    "customMenuCommands",
			Pattern: "**",
			Config: map[string]any{
				"menuCommands": []any{
					map[string]any{
						"name":     "foo",
						"shellCmd": "echo 'foo'",
						"mode":     "insert",
					},
					map[string]any{
						"name":     "bar",
						"shellCmd": "echo 'bar'",
						"mode":     "insert",
					},
					map[string]any{
						"name":     "foo", // duplicate
						"shellCmd": "echo 'foo2'",
						"mode":     "insert",
					},
				},
			},
		},
	}

	// Load the document.
	path, cleanup := createTestFile(t, "")
	state := NewEditorState(100, 100, configRuleSet, nil)
	LoadDocument(state, path, true, startOfDocLocator)
	defer cleanup()

	// Show the menu and search for 'f', which should match all custom items.
	ShowMenu(state, MenuStyleCommand, nil)
	AppendRuneToMenuSearch(state, 'f')

	// Expect that the search results include only two items,
	// since "foo" was deduplicated.
	results, selectedIdx := state.Menu().SearchResults()
	assert.Equal(t, len(results), 2)
	assert.Equal(t, selectedIdx, 0)
	assert.Equal(t, results[0].Name, "foo")
	assert.Equal(t, results[1].Name, "bar")

	// Execute the "foo" item and wait for the shell cmd to complete.
	ExecuteSelectedMenuItem(state)
	select {
	case action := <-state.TaskResultChan():
		action(state)
	case <-time.After(5 * time.Second):
		require.Fail(t, "Timed out")
	}

	// Check that the second "foo" command was invoked, which should
	// have inserted "foo2" into the document.
	text := state.DocumentBuffer().TextTree().String()
	assert.Equal(t, text, "foo2\n")
}
