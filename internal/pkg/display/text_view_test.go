package display

import (
	"testing"

	"github.com/gdamore/tcell"

	"github.com/stretchr/testify/require"
	"github.com/wedaly/aretext/internal/pkg/text"
)

func textViewWithString(t *testing.T, screen tcell.Screen, s string) *TextView {
	screenWidth, screenHeight := screen.Size()
	region := NewScreenRegion(screen, 0, 0, screenWidth, screenHeight)
	tree, err := text.NewTreeFromString(s)
	require.NoError(t, err)
	return NewTextView(tree, region)
}

func TestTextViewDraw(t *testing.T) {
	testCases := []struct {
		name             string
		inputString      string
		expectedContents [][]rune
	}{
		{
			name:        "empty",
			inputString: "",
			expectedContents: [][]rune{
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
			},
		},
		{
			name:        "short line",
			inputString: "abc",
			expectedContents: [][]rune{
				{'a', 'b', 'c', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
			},
		},
		{
			name:        "wrapping line",
			inputString: "abcdefghijklmnopqrstuvwxyz",
			expectedContents: [][]rune{
				{'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j'},
				{'k', 'l', 'm', 'n', 'o', 'p', 'q', 'r', 's', 't'},
				{'u', 'v', 'w', 'x', 'y', 'z', ' ', ' ', ' ', ' '},
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
			},
		},
		{
			name:        "newline",
			inputString: "abc\ndefghi\njkl",
			expectedContents: [][]rune{
				{'a', 'b', 'c', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
				{'d', 'e', 'f', 'g', 'h', 'i', ' ', ' ', ' ', ' '},
				{'j', 'k', 'l', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
			},
		},
		{
			name:        "full-width characters, no wrapping",
			inputString: "abc界xyz",
			expectedContents: [][]rune{
				{'a', 'b', 'c', '界', 'X', 'x', 'y', 'z', ' ', ' '},
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
			},
		},
		{
			name:        "full-width characters wrapped at end of line",
			inputString: "abcdefghi界jklmn",
			expectedContents: [][]rune{
				{'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', ' '},
				{'界', 'X', 'j', 'k', 'l', 'm', 'n', ' ', ' ', ' '},
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			withSimScreen(t, func(s tcell.SimulationScreen) {
				s.SetSize(10, 10)
				view := textViewWithString(t, s, tc.inputString)
				view.Draw()
				s.Sync()
				assertCellContents(t, s, tc.expectedContents)
			})
		})
	}
}

func TestTextViewDrawSizeTooSmall(t *testing.T) {
	withSimScreen(t, func(s tcell.SimulationScreen) {
		s.SetSize(1, 2)
		view := textViewWithString(t, s, "abcd1234")
		view.Draw()
		s.Sync()
		assertCellContents(t, s, [][]rune{
			{'~'},
			{'~'},
		})
	})
}
