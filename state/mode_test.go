package state

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/aretext/aretext/selection"
)

func TestNormalToVisualMode(t *testing.T) {
	state := NewEditorState(100, 100, nil)
	SetInputMode(state, InputModeNormal)
	ToggleVisualMode(state, selection.ModeChar)
	assert.Equal(t, InputModeVisual, state.inputMode)
	assert.Equal(t, selection.ModeChar, state.documentBuffer.selector.Mode())
}

func TestVisualModeToNormalMode(t *testing.T) {
	state := NewEditorState(100, 100, nil)
	ToggleVisualMode(state, selection.ModeChar)
	SetInputMode(state, InputModeNormal)
	assert.Equal(t, InputModeNormal, state.inputMode)
	assert.Equal(t, selection.ModeNone, state.documentBuffer.selector.Mode())
}

func TestToggleVisualModeSameSelectionMode(t *testing.T) {
	testCases := []struct {
		name          string
		selectionMode selection.Mode
	}{
		{
			name:          "charwise",
			selectionMode: selection.ModeChar,
		},
		{
			name:          "linewise",
			selectionMode: selection.ModeLine,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			state := NewEditorState(100, 100, nil)
			ToggleVisualMode(state, tc.selectionMode)
			ToggleVisualMode(state, tc.selectionMode)
			assert.Equal(t, InputModeNormal, state.inputMode)
			assert.Equal(t, selection.ModeNone, state.documentBuffer.selector.Mode())
		})
	}
}

func TestToggleVisualModeDifferentSelectionMode(t *testing.T) {
	testCases := []struct {
		name                string
		firstSelectionMode  selection.Mode
		secondSelectionMode selection.Mode
	}{
		{
			name:                "charwise to linewise",
			firstSelectionMode:  selection.ModeChar,
			secondSelectionMode: selection.ModeLine,
		},
		{
			name:                "linewise to charwise",
			firstSelectionMode:  selection.ModeLine,
			secondSelectionMode: selection.ModeChar,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			state := NewEditorState(100, 100, nil)
			ToggleVisualMode(state, tc.firstSelectionMode)
			ToggleVisualMode(state, tc.secondSelectionMode)
			assert.Equal(t, InputModeVisual, state.inputMode)
			assert.Equal(t, tc.secondSelectionMode, state.documentBuffer.selector.Mode())
		})
	}
}
