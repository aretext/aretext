package state

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/aretext/aretext/editor/menu"
	"github.com/aretext/aretext/editor/selection"
)

func TestShowMenu(t *testing.T) {
	state := NewEditorState(100, 100, nil, nil)
	items := []menu.Item{
		{Name: "test item 1"},
		{Name: "test item 2"},
	}
	ShowMenu(state, MenuStyleCommand, items)
	assert.Equal(t, InputModeMenu, state.InputMode())
	assert.Equal(t, MenuStyleCommand, state.Menu().Style())
	assert.Equal(t, "", state.Menu().SearchQuery())

	results, selectedIdx := state.Menu().SearchResults()
	assert.Equal(t, 0, selectedIdx)
	assert.Equal(t, 0, len(results))
}

func TestHideMenu(t *testing.T) {
	state := NewEditorState(100, 100, nil, nil)
	items := []menu.Item{
		{Name: "test item"},
	}
	ShowMenu(state, MenuStyleCommand, items)
	HideMenu(state)
	assert.Equal(t, InputModeNormal, state.InputMode())
}

func TestShowMenuFromVisualMode(t *testing.T) {
	state := NewEditorState(100, 100, nil, nil)
	ToggleVisualMode(state, selection.ModeChar)
	ShowMenu(state, MenuStyleCommand, nil)
	assert.Equal(t, InputModeMenu, state.inputMode)
	HideMenu(state)
	assert.Equal(t, InputModeVisual, state.inputMode)
}

func TestSelectAndExecuteMenuItem(t *testing.T) {
	state := NewEditorState(100, 100, nil, nil)
	items := []menu.Item{
		{
			Name:   "test item",
			Action: func(s *EditorState) {},
		},
		{
			Name:   "quit",
			Action: Quit,
		},
	}
	ShowMenu(state, MenuStyleCommand, items)
	AppendRuneToMenuSearch(state, 'q') // search for "q", should match "quit"
	ExecuteSelectedMenuItem(state)
	assert.Equal(t, InputModeNormal, state.InputMode())
	assert.Equal(t, "", state.Menu().SearchQuery())
	assert.True(t, state.QuitFlag())
}

func TestMoveMenuSelection(t *testing.T) {
	testCases := []struct {
		name              string
		items             []menu.Item
		searchRune        rune
		moveDeltas        []int
		expectSelectedIdx int
	}{
		{
			name:              "empty results, move up",
			items:             []menu.Item{},
			searchRune:        't',
			moveDeltas:        []int{-1},
			expectSelectedIdx: 0,
		},
		{
			name:              "empty results, move down",
			items:             []menu.Item{},
			searchRune:        't',
			moveDeltas:        []int{1},
			expectSelectedIdx: 0,
		},
		{
			name: "single result, move up",
			items: []menu.Item{
				{Name: "test"},
			},
			searchRune:        't',
			moveDeltas:        []int{1},
			expectSelectedIdx: 0,
		},
		{
			name: "single result, move down",
			items: []menu.Item{
				{Name: "test"},
			},
			searchRune:        't',
			moveDeltas:        []int{1},
			expectSelectedIdx: 0,
		},
		{
			name: "multiple results, move down and up",
			items: []menu.Item{
				{Name: "test1"},
				{Name: "test2"},
				{Name: "test3"},
			},
			searchRune:        't',
			moveDeltas:        []int{2, -1},
			expectSelectedIdx: 1,
		},
		{
			name: "multiple results, move up and wraparound",
			items: []menu.Item{
				{Name: "test1"},
				{Name: "test2"},
				{Name: "test3"},
				{Name: "test4"},
			},
			searchRune:        't',
			moveDeltas:        []int{-1},
			expectSelectedIdx: 3,
		},
		{
			name: "multiple results, move down and wraparound",
			items: []menu.Item{
				{Name: "test1"},
				{Name: "test2"},
				{Name: "test3"},
				{Name: "test4"},
			},
			searchRune:        't',
			moveDeltas:        []int{3, 1},
			expectSelectedIdx: 0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			state := NewEditorState(100, 100, nil, nil)
			ShowMenu(state, MenuStyleCommand, tc.items)
			AppendRuneToMenuSearch(state, tc.searchRune)
			for _, delta := range tc.moveDeltas {
				MoveMenuSelection(state, delta)
			}
			_, selectedIdx := state.Menu().SearchResults()
			assert.Equal(t, tc.expectSelectedIdx, selectedIdx)
		})
	}
}

func TestAppendRuneToMenuSearch(t *testing.T) {
	state := NewEditorState(100, 100, nil, nil)
	ShowMenu(state, MenuStyleCommand, nil)
	AppendRuneToMenuSearch(state, 'a')
	AppendRuneToMenuSearch(state, 'b')
	AppendRuneToMenuSearch(state, 'c')
	assert.Equal(t, "abc", state.Menu().SearchQuery())
}

func TestDeleteRuneFromMenuSearch(t *testing.T) {
	testCases := []struct {
		name        string
		searchQuery string
		numDeleted  int
		expectQuery string
	}{
		{
			name:        "delete from empty query",
			searchQuery: "",
			numDeleted:  1,
			expectQuery: "",
		},
		{
			name:        "delete ascii from end of query",
			searchQuery: "abc",
			numDeleted:  2,
			expectQuery: "a",
		},
		{
			name:        "delete non-ascii unicode from end of query",
			searchQuery: "£፴",
			numDeleted:  1,
			expectQuery: "£",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			state := NewEditorState(100, 100, nil, nil)
			ShowMenu(state, MenuStyleCommand, nil)
			for _, r := range tc.searchQuery {
				AppendRuneToMenuSearch(state, r)
			}
			for i := 0; i < tc.numDeleted; i++ {
				DeleteRuneFromMenuSearch(state)
			}
			assert.Equal(t, tc.expectQuery, state.Menu().SearchQuery())
		})
	}
}

func TestShowFileMenu(t *testing.T) {
	paths := []string{
		"a/foo.txt",
		"a/b/bar.txt",
		"c/baz.txt",
	}
	withTempDirPaths(t, paths, func(dir string) {
		// Show the file menu.
		state := NewEditorState(100, 100, nil, nil)
		ShowFileMenu(state, nil)
		completeTaskOrTimeout(t, state)

		// Verify that the menu shows file paths.
		items, selectedIdx := state.Menu().SearchResults()
		require.Equal(t, 3, len(items))
		assert.Equal(t, 0, selectedIdx)
		assert.Equal(t, "a/b/bar.txt", items[0].Name)
		assert.Equal(t, "a/foo.txt", items[1].Name)
		assert.Equal(t, "c/baz.txt", items[2].Name)

		// Execute the second item and verify that it opens the file.
		MoveMenuSelection(state, 1)
		ExecuteSelectedMenuItem(state)
		assert.Equal(t, "Opened a/foo.txt", state.StatusMsg().Text)
		assert.Equal(t, "a/foo.txt content", state.DocumentBuffer().TextTree().String())
	})
}

func TestShowFileLocationsMenu(t *testing.T) {
	// These are NOT in lexicographic order.
	items := []menu.Item{
		{Name: "foo.txt:3 foo"},
		{Name: "bar.txt:2 bar"},
		{Name: "baz.txt:123 baz"},
	}

	// Show the menu with style FileLocation
	state := NewEditorState(100, 100, nil, nil)
	ShowMenu(state, MenuStyleFileLocation, items)

	// Verify that the menu shows items in their original order.
	items, selectedIdx := state.Menu().SearchResults()
	require.Equal(t, 3, len(items))
	assert.Equal(t, 0, selectedIdx)
	assert.Equal(t, "foo.txt:3 foo", items[0].Name)
	assert.Equal(t, "bar.txt:2 bar", items[1].Name)
	assert.Equal(t, "baz.txt:123 baz", items[2].Name)
}

func TestShowChildDirsMenu(t *testing.T) {
	paths := []string{
		"root.txt",
		"a/foo.txt",
		"a/b/bar.txt",
		"c/baz.txt",
	}
	withTempDirPaths(t, paths, func(dir string) {
		// Show the child dirs menu.
		state := NewEditorState(100, 100, nil, nil)
		ShowChildDirsMenu(state, nil)
		completeTaskOrTimeout(t, state)

		// Verify that the menu shows subdirectory paths.
		items, selectedIdx := state.Menu().SearchResults()
		require.Equal(t, 3, len(items))
		assert.Equal(t, 0, selectedIdx)
		assert.Equal(t, "./a", items[0].Name)
		assert.Equal(t, "./a/b", items[1].Name)
		assert.Equal(t, "./c", items[2].Name)

		// Execute the second item and verify that the working directory changed.
		MoveMenuSelection(state, 1)
		ExecuteSelectedMenuItem(state)
		assert.Contains(t, state.StatusMsg().Text, "Changed working directory")
		workingDir, err := os.Getwd()
		require.NoError(t, err)
		assert.Equal(t, "b", filepath.Base(workingDir))
	})
}

func TestShowParentDirsMenu(t *testing.T) {
	withTempDirPaths(t, nil, func(dir string) {
		// This path may be different than the temp directory due to symlinks (macOS)
		originalWorkingDir, err := os.Getwd()
		require.NoError(t, err)

		// Show the parent dirs menu.
		state := NewEditorState(100, 100, nil, nil)
		ShowParentDirsMenu(state)

		// Verify that the menu shows parent directory paths.
		// This depends on the randomly chosen tempdir, so
		// we check that the paths are in descending order by length.
		items, selectedIdx := state.Menu().SearchResults()
		assert.Greater(t, len(items), 0)
		assert.Equal(t, 0, selectedIdx)
		for i := 1; i < len(items); i++ {
			assert.Less(t, len(items[i].Name), len(items[i-1].Name))
		}

		// Execute the first item and verify that the working directory changed.
		// The new working directory should be a parent of the current directory.
		ExecuteSelectedMenuItem(state)
		assert.Contains(t, state.StatusMsg().Text, "Changed working directory")
		workingDir, err := os.Getwd()
		require.NoError(t, err)
		assert.Less(t, len(workingDir), len(originalWorkingDir))
		assert.Contains(t, originalWorkingDir, workingDir)
	})
}

func withTempDirPaths(t *testing.T, paths []string, f func(string)) {
	// Reset the original working directory after the test.
	originalWd, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(originalWd)

	// Change the current working directory to a tempdir.
	dir := t.TempDir()
	err = os.Chdir(dir)
	require.NoError(t, err)

	// Create paths in the tempdir.
	for _, p := range paths {
		err = os.MkdirAll(filepath.Dir(p), 0755)
		require.NoError(t, err)
		err = os.WriteFile(p, []byte(p+" content"), 0644)
		require.NoError(t, err)
	}

	// Run the test.
	f(dir)
}

func completeTaskOrTimeout(t *testing.T, state *EditorState) {
	select {
	case action := <-state.TaskResultChan():
		action(state)
	case <-time.After(10 * time.Second):
		require.Fail(t, "Timed out")
	}
}
