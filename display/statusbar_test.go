package display

import (
	"testing"

	"github.com/gdamore/tcell/v2"

	"github.com/aretext/aretext/state"
)

func TestDrawStatusBar(t *testing.T) {
	testCases := []struct {
		name             string
		statusMsg        state.StatusMsg
		inputMode        state.InputMode
		filePath         string
		expectedContents [][]rune
	}{
		{
			name:      "normal mode shows file path",
			inputMode: state.InputModeNormal,
			filePath:  "./foo/bar",
			expectedContents: [][]rune{
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
				{'.', '/', 'f', 'o', 'o', '/', 'b', 'a', 'r', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
			},
		},
		{
			name:      "insert mode shows INSERT",
			inputMode: state.InputModeInsert,
			filePath:  "./foo/bar",
			expectedContents: [][]rune{
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
				{'-', '-', ' ', 'I', 'N', 'S', 'E', 'R', 'T', ' ', '-', '-', ' ', ' ', ' ', ' '},
			},
		},
		{
			name:      "visual mode shows VISUAL",
			inputMode: state.InputModeVisual,
			filePath:  "./foo/bar",
			expectedContents: [][]rune{
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
				{'-', '-', ' ', 'V', 'I', 'S', 'U', 'A', 'L', ' ', '-', '-', ' ', ' ', ' ', ' '},
			},
		},
		{
			name:      "menu mode shows file path",
			inputMode: state.InputModeMenu,
			filePath:  "./foo/bar",
			expectedContents: [][]rune{
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
				{'.', '/', 'f', 'o', 'o', '/', 'b', 'a', 'r', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
			},
		},
		{
			name: "status message success",
			statusMsg: state.StatusMsg{
				Text:  "success",
				Style: state.StatusMsgStyleSuccess,
			},
			expectedContents: [][]rune{
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
				{'s', 'u', 'c', 'c', 'e', 's', 's', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
			},
		},
		{
			name: "status message error",
			statusMsg: state.StatusMsg{
				Text:  "error",
				Style: state.StatusMsgStyleError,
			},
			expectedContents: [][]rune{
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
				{'e', 'r', 'r', 'o', 'r', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			withSimScreen(t, func(s tcell.SimulationScreen) {
				s.SetSize(16, 2)
				DrawStatusBar(s, tc.statusMsg, tc.inputMode, tc.filePath)
				s.Sync()
				assertCellContents(t, s, tc.expectedContents)
			})
		})
	}
}
