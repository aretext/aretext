package display

import (
	"testing"

	"github.com/gdamore/tcell"

	"github.com/wedaly/aretext/internal/pkg/exec"
)

// DrawMenu(screen, menu)

func TestDrawMenu(t *testing.T) {
	testCases := []struct {
		name             string
		buildMenu        func() *exec.MenuState
		expectedContents [][]rune
	}{
		{
			name: "not visible",
			buildMenu: func() *exec.MenuState {
				return &exec.MenuState{}
			},
			expectedContents: [][]rune{
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
			},
		},
		{
			name: "visible, initial state with prompt",
			buildMenu: func() *exec.MenuState {
				state := exec.NewEditorState(100, 100, nil, nil)
				mutator := exec.NewShowMenuMutator("test", nil)
				mutator.Mutate(state)
				return state.Menu()
			},
			expectedContents: [][]rune{
				{':', 't', 'e', 's', 't', ' ', ' ', ' ', ' ', ' '},
				{'─', '─', '─', '─', '─', '─', '─', '─', '─', '─'},
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
			},
		},
		{
			name: "visible, query with no results",
			buildMenu: func() *exec.MenuState {
				state := exec.NewEditorState(100, 100, nil, nil)
				mutator := exec.NewCompositeMutator([]exec.Mutator{
					exec.NewShowMenuMutator("test", nil),
					exec.NewAppendMenuSearchMutator('a'),
					exec.NewAppendMenuSearchMutator('b'),
					exec.NewAppendMenuSearchMutator('c'),
				})
				mutator.Mutate(state)
				return state.Menu()
			},
			expectedContents: [][]rune{
				{':', 'a', 'b', 'c', ' ', ' ', ' ', ' ', ' ', ' '},
				{'─', '─', '─', '─', '─', '─', '─', '─', '─', '─'},
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
			},
		},
		{
			name: "visible, query with results, first selected",
			buildMenu: func() *exec.MenuState {
				state := exec.NewEditorState(100, 100, nil, nil)
				mutator := exec.NewCompositeMutator([]exec.Mutator{
					exec.NewShowMenuMutator("test", []exec.MenuItem{
						{Name: "test first"},
						{Name: "test second"},
						{Name: "test third"},
					}),
					exec.NewAppendMenuSearchMutator('t'),
				})
				mutator.Mutate(state)
				return state.Menu()
			},
			expectedContents: [][]rune{
				{':', 't', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
				{' ', ' ', '>', ' ', 't', 'e', 's', 't', ' ', 'f'},
				{' ', ' ', ' ', ' ', 't', 'e', 's', 't', ' ', 's'},
				{' ', ' ', ' ', ' ', 't', 'e', 's', 't', ' ', 't'},
				{'─', '─', '─', '─', '─', '─', '─', '─', '─', '─'},
			},
		},
		{
			name: "visible, query with many results, second-to-last selected",
			buildMenu: func() *exec.MenuState {
				state := exec.NewEditorState(100, 100, nil, nil)
				mutator := exec.NewCompositeMutator([]exec.Mutator{
					exec.NewShowMenuMutator("test", []exec.MenuItem{
						{Name: "test 1"},
						{Name: "test 2"},
						{Name: "test 3"},
						{Name: "test 4"},
						{Name: "test 5"},
						{Name: "test 6"},
						{Name: "test 7"},
						{Name: "test 8"},
						{Name: "test 9"},
					}),
					exec.NewAppendMenuSearchMutator('t'),
					exec.NewMoveMenuSelectionMutator(-2),
				})
				mutator.Mutate(state)
				return state.Menu()
			},
			expectedContents: [][]rune{
				{':', 't', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
				{' ', ' ', ' ', ' ', 't', 'e', 's', 't', ' ', '5'},
				{' ', ' ', ' ', ' ', 't', 'e', 's', 't', ' ', '6'},
				{' ', ' ', ' ', ' ', 't', 'e', 's', 't', ' ', '7'},
				{' ', ' ', '>', ' ', 't', 'e', 's', 't', ' ', '8'},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			withSimScreen(t, func(s tcell.SimulationScreen) {
				s.SetSize(10, 5)
				menu := tc.buildMenu()
				DrawMenu(s, menu)
				s.Sync()
				assertCellContents(t, s, tc.expectedContents)
			})
		})
	}
}
