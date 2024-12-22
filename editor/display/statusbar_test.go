package display

import (
	"path/filepath"
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/stretchr/testify/require"

	"github.com/aretext/aretext/editor/state"
)

func TestDrawStatusBar(t *testing.T) {
	testCases := []struct {
		name                 string
		statusMsg            state.StatusMsg
		inputMode            state.InputMode
		inputBufferString    string
		isRecordingUserMacro bool
		filePath             string
		expectedContents     [][]rune
	}{
		{
			name:      "normal mode shows file path",
			inputMode: state.InputModeNormal,
			filePath:  "./foo/bar",
			expectedContents: [][]rune{
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
				{'f', 'o', 'o', '/', 'b', 'a', 'r', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
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
				{'f', 'o', 'o', '/', 'b', 'a', 'r', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
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
		{
			name:              "input buffer",
			inputBufferString: `"aya`,
			expectedContents: [][]rune{
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
				{'"', 'a', 'y', 'a', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
			},
		},
		{
			name:                 "recording user macro",
			inputMode:            state.InputModeNormal,
			isRecordingUserMacro: true,
			expectedContents: [][]rune{
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
				{'R', 'e', 'c', 'o', 'r', 'd', 'i', 'n', 'g', ' ', 'm', 'a', 'c', 'r', 'o', '.'},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			absFilePath, err := filepath.Abs(tc.filePath)
			require.NoError(t, err)

			withSimScreen(t, func(s tcell.SimulationScreen) {
				s.SetSize(16, 2)
				palette := NewPalette()
				DrawStatusBar(
					s,
					palette,
					tc.statusMsg,
					tc.inputMode,
					tc.inputBufferString,
					tc.isRecordingUserMacro,
					absFilePath,
				)
				s.Sync()
				assertCellContents(t, s, tc.expectedContents)
			})
		})
	}
}
