package state

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/aretext/aretext/undo"
)

func TestQuit(t *testing.T) {
	testCases := []struct {
		name              string
		force             bool
		hasUnsavedChanges bool
		expectQuitFlag    bool
	}{
		{
			name:           "no force, no unsaved changes",
			expectQuitFlag: true,
		},
		{
			name:           "force, no unsaved changes",
			force:          true,
			expectQuitFlag: true,
		},
		{
			name:              "no force, unsaved changes",
			hasUnsavedChanges: true,
			expectQuitFlag:    false,
		},
		{
			name:              "force, unsaved changes",
			force:             true,
			hasUnsavedChanges: true,
			expectQuitFlag:    true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			state := NewEditorState(100, 100, nil, nil)

			if tc.hasUnsavedChanges {
				state.documentBuffer.undoLog.BeginEntry(0)
				state.documentBuffer.undoLog.TrackOp(undo.InsertOp(0, "a"))
				state.documentBuffer.undoLog.CommitEntry(1)
			}

			if tc.force {
				Quit(state)
			} else {
				AbortIfUnsavedChanges(state, "Document has unsaved changes", Quit)
			}

			assert.Equal(t, tc.expectQuitFlag, state.QuitFlag())
			if !tc.expectQuitFlag {
				assert.Equal(t, StatusMsgStyleError, state.statusMsg.Style)
				assert.Contains(t, state.statusMsg.Text, "Document has unsaved changes")
			}
		})
	}
}
