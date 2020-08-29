package display

import (
	"testing"

	"github.com/gdamore/tcell"

	"github.com/stretchr/testify/require"
	"github.com/wedaly/aretext/internal/pkg/exec"
	"github.com/wedaly/aretext/internal/pkg/text"
)

func TestDrawEditorLayoutDocumentOnly(t *testing.T) {
	withSimScreen(t, func(s tcell.SimulationScreen) {
		s.SetSize(5, 5)
		tree, err := text.NewTreeFromString("abcd")
		require.NoError(t, err)
		state := exec.NewEditorState(5, 5, exec.NewBufferState(tree, 0, 0, 0, 5, 5))
		DrawEditor(s, state)
		s.Sync()
		assertCellContents(t, s, [][]rune{
			{'a', 'b', 'c', 'd', ' '},
			{' ', ' ', ' ', ' ', ' '},
			{' ', ' ', ' ', ' ', ' '},
			{' ', ' ', ' ', ' ', ' '},
			{' ', ' ', ' ', ' ', ' '},
		})
	})
}

func TestDrawEditorLayoutDocumentAndRepl(t *testing.T) {
	withSimScreen(t, func(s tcell.SimulationScreen) {
		s.SetSize(5, 5)
		tree, err := text.NewTreeFromString("abcd")
		require.NoError(t, err)
		state := exec.NewEditorState(5, 5, exec.NewBufferState(tree, 0, 0, 0, 5, 5))
		exec.NewLayoutMutator(exec.LayoutDocumentAndRepl).Mutate(state)
		state.ReplBuffer().Tree().InsertAtPosition(0, '>')
		DrawEditor(s, state)
		s.Sync()
		assertCellContents(t, s, [][]rune{
			{'a', 'b', 'c', 'd', ' '},
			{' ', ' ', ' ', ' ', ' '},
			{'─', '─', '─', '─', '─'},
			{'>', ' ', ' ', ' ', ' '},
			{' ', ' ', ' ', ' ', ' '},
		})
	})
}

func TestDrawEditorSingleLine(t *testing.T) {
	testCases := []struct {
		name             string
		layout           exec.Layout
		expectedContents [][]rune
	}{
		{
			name:   "LayoutDocumentOnly",
			layout: exec.LayoutDocumentOnly,
			expectedContents: [][]rune{
				{'a', 'b', 'c', 'd', ' '},
			},
		},
		{
			name:   "LayoutDocumentAndRepl",
			layout: exec.LayoutDocumentAndRepl,
			expectedContents: [][]rune{
				{'>', ' ', ' ', ' ', ' '},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			withSimScreen(t, func(s tcell.SimulationScreen) {
				s.SetSize(5, 1)
				tree, err := text.NewTreeFromString("abcd")
				require.NoError(t, err)
				state := exec.NewEditorState(5, 1, exec.NewBufferState(tree, 0, 0, 0, 5, 1))
				exec.NewLayoutMutator(tc.layout).Mutate(state)
				state.ReplBuffer().Tree().InsertAtPosition(0, '>')
				DrawEditor(s, state)
				s.Sync()
				assertCellContents(t, s, tc.expectedContents)
			})
		})
	}
}
