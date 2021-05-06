package state

import (
	"testing"

	"github.com/aretext/aretext/menu"
	"github.com/aretext/aretext/selection"
	"github.com/aretext/aretext/syntax"
	"github.com/stretchr/testify/assert"
)

func TestShowMenu(t *testing.T) {
	state := NewEditorState(100, 100, nil, nil)
	loadItems := func() []menu.Item {
		return []menu.Item{
			{Name: "test item 1"},
			{Name: "test item 2"},
		}
	}
	ShowMenu(state, MenuStyleCommand, loadItems)
	assert.True(t, state.Menu().Visible())
	assert.Equal(t, MenuStyleCommand, state.Menu().Style())
	assert.Equal(t, "", state.Menu().SearchQuery())

	results, selectedIdx := state.Menu().SearchResults()
	assert.Equal(t, 0, selectedIdx)
	assert.Equal(t, 0, len(results))
}

func TestHideMenu(t *testing.T) {
	state := NewEditorState(100, 100, nil, nil)
	loadItems := func() []menu.Item {
		return []menu.Item{
			{Name: "test item"},
		}
	}
	ShowMenu(state, MenuStyleCommand, loadItems)
	HideMenu(state)
	assert.False(t, state.Menu().Visible())
}

func TestShowMenuFromVisualMode(t *testing.T) {
	state := NewEditorState(100, 100, nil, nil)
	loadItems := func() []menu.Item { return nil }
	ToggleVisualMode(state, selection.ModeChar)
	ShowMenu(state, MenuStyleCommand, loadItems)
	assert.Equal(t, InputModeMenu, state.inputMode)
	HideMenu(state)
	assert.Equal(t, InputModeVisual, state.inputMode)
}

func TestSelectAndExecuteMenuItem(t *testing.T) {
	state := NewEditorState(100, 100, nil, nil)
	loadItems := func() []menu.Item {
		return []menu.Item{
			{
				Name: "set syntax json",
				Action: func(s *EditorState) {
					SetSyntax(s, syntax.LanguageJson)
				},
			},
			{
				Name:   "quit",
				Action: Quit,
			},
		}
	}
	ShowMenu(state, MenuStyleCommand, loadItems)
	AppendRuneToMenuSearch(state, 'q') // search for "q", should match "quit"
	ExecuteSelectedMenuItem(state)
	assert.False(t, state.Menu().Visible())
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
			loadItems := func() []menu.Item { return tc.items }
			ShowMenu(state, MenuStyleCommand, loadItems)
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
	loadItems := func() []menu.Item { return nil }
	ShowMenu(state, MenuStyleCommand, loadItems)
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
			loadItems := func() []menu.Item { return nil }
			ShowMenu(state, MenuStyleCommand, loadItems)
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
