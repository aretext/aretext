package display

import (
	"testing"

	"github.com/gdamore/tcell"

	"github.com/stretchr/testify/require"
	"github.com/wedaly/aretext/internal/pkg/exec"
	"github.com/wedaly/aretext/internal/pkg/file"
	"github.com/wedaly/aretext/internal/pkg/text"
)

func TestDrawEditorLayoutDocumentOnly(t *testing.T) {
	withSimScreen(t, func(s tcell.SimulationScreen) {
		s.SetSize(5, 5)
		tree, err := text.NewTreeFromString("abcd")
		require.NoError(t, err)
		state := exec.NewEditorState(5, 5)
		exec.NewLoadDocumentMutator(tree, file.NewEmptyWatcher()).Mutate(state)
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
		state := exec.NewEditorState(5, 5)
		exec.NewCompositeMutator([]exec.Mutator{
			exec.NewLoadDocumentMutator(tree, file.NewEmptyWatcher()),
			exec.NewLayoutMutator(exec.LayoutDocumentAndRepl),
		}).Mutate(state)
		state.Buffer(exec.BufferIdRepl).TextTree().InsertAtPosition(0, '>')
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
				state := exec.NewEditorState(5, 1)
				exec.NewCompositeMutator([]exec.Mutator{
					exec.NewLoadDocumentMutator(tree, file.NewEmptyWatcher()),
					exec.NewLayoutMutator(tc.layout),
				}).Mutate(state)
				state.Buffer(exec.BufferIdRepl).TextTree().InsertAtPosition(0, '>')
				DrawEditor(s, state)
				s.Sync()
				assertCellContents(t, s, tc.expectedContents)
			})
		})
	}
}
