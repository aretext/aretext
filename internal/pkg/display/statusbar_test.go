package display

import (
	"testing"

	"github.com/gdamore/tcell"
	"github.com/wedaly/aretext/internal/pkg/exec"
	"github.com/wedaly/aretext/internal/pkg/input"
)

func TestDrawStatusBar(t *testing.T) {
	testCases := []struct {
		name             string
		statusMsg        exec.StatusMsg
		inputMode        input.ModeType
		filePath         string
		expectedContents [][]rune
	}{
		{
			name:      "normal mode shows file path",
			inputMode: input.ModeTypeNormal,
			filePath:  "./foo/bar",
			expectedContents: [][]rune{
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
				{'.', '/', 'f', 'o', 'o', '/', 'b', 'a', 'r', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
			},
		},
		{
			name:      "insert mode shows INSERT",
			inputMode: input.ModeTypeInsert,
			filePath:  "./foo/bar",
			expectedContents: [][]rune{
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
				{'-', '-', ' ', 'I', 'N', 'S', 'E', 'R', 'T', ' ', '-', '-', ' ', ' ', ' ', ' '},
			},
		},
		{
			name:      "menu mode shows file path",
			inputMode: input.ModeTypeMenu,
			filePath:  "./foo/bar",
			expectedContents: [][]rune{
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
				{'.', '/', 'f', 'o', 'o', '/', 'b', 'a', 'r', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
			},
		},
		{
			name: "status message success",
			statusMsg: exec.StatusMsg{
				Text:  "success",
				Style: exec.StatusMsgStyleSuccess,
			},
			expectedContents: [][]rune{
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
				{'s', 'u', 'c', 'c', 'e', 's', 's', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
			},
		},
		{
			name: "status message error",
			statusMsg: exec.StatusMsg{
				Text:  "error",
				Style: exec.StatusMsgStyleError,
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
