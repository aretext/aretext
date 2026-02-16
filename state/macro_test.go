package state

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type actionLogEntry struct {
	name                 string
	isReplayingUserMacro bool
}

type actionLogger struct {
	logEntries []actionLogEntry
}

func (a *actionLogger) buildAction(name string) MacroAction {
	return func(s *EditorState) error {
		a.logEntries = append(a.logEntries, actionLogEntry{
			name:                 name,
			isReplayingUserMacro: s.macroState.isReplayingUserMacro,
		})
		return nil
	}
}

func (a *actionLogger) clear() {
	a.logEntries = nil
}

func TestLastActionMacro(t *testing.T) {
	var logger actionLogger
	state := NewEditorState(100, 100, nil, nil)
	ReplayLastActionMacro(state, 1)
	assert.Equal(t, 0, len(logger.logEntries))

	AddToLastActionMacro(state, logger.buildAction("a"))
	AddToLastActionMacro(state, logger.buildAction("b"))
	ReplayLastActionMacro(state, 1)
	expected := []actionLogEntry{{name: "a"}, {name: "b"}}
	assert.Equal(t, expected, logger.logEntries)

	logger.clear()
	ClearLastActionMacro(state)
	ReplayLastActionMacro(state, 1)
	assert.Equal(t, 0, len(logger.logEntries))

	AddToLastActionMacro(state, logger.buildAction("c"))
	ReplayLastActionMacro(state, 1)
	assert.Equal(t, []actionLogEntry{{name: "c"}}, logger.logEntries)
}

func TestLastActionMacroWithCount(t *testing.T) {
	var logger actionLogger
	state := NewEditorState(100, 100, nil, nil)
	AddToLastActionMacro(state, logger.buildAction("a"))
	AddToLastActionMacro(state, logger.buildAction("b"))
	ReplayLastActionMacro(state, 3) // Repeat 3 times.
	expected := []actionLogEntry{
		{name: "a"}, {name: "b"},
		{name: "a"}, {name: "b"},
		{name: "a"}, {name: "b"},
	}
	assert.Equal(t, expected, logger.logEntries)
}

func TestRecordAndReplayUserMacro(t *testing.T) {
	var logger actionLogger
	state := NewEditorState(100, 100, nil, nil)

	ToggleUserMacroRecording(state)
	assert.Equal(t, StatusMsg{
		Style: StatusMsgStyleSuccess,
		Text:  "Started recording macro",
	}, state.StatusMsg())

	AddToRecordingUserMacro(state, logger.buildAction("a"))
	AddToRecordingUserMacro(state, logger.buildAction("b"))

	ToggleUserMacroRecording(state)
	assert.Equal(t, StatusMsg{
		Style: StatusMsgStyleSuccess,
		Text:  "Recorded macro",
	}, state.StatusMsg())

	assert.Equal(t, 0, len(logger.logEntries))
	ReplayRecordedUserMacro(state, 1)
	assert.Equal(t, StatusMsg{
		Style: StatusMsgStyleSuccess,
		Text:  "Replayed macro",
	}, state.StatusMsg())
	expected := []actionLogEntry{
		{name: "a", isReplayingUserMacro: true},
		{name: "b", isReplayingUserMacro: true},
	}
	assert.Equal(t, expected, logger.logEntries)
}

func TestRecordAndReplayUserMacroWithCount(t *testing.T) {
	var logger actionLogger
	state := NewEditorState(100, 100, nil, nil)

	ToggleUserMacroRecording(state)
	assert.Equal(t, StatusMsg{
		Style: StatusMsgStyleSuccess,
		Text:  "Started recording macro",
	}, state.StatusMsg())

	AddToRecordingUserMacro(state, logger.buildAction("a"))
	AddToRecordingUserMacro(state, logger.buildAction("b"))

	ToggleUserMacroRecording(state)
	assert.Equal(t, StatusMsg{
		Style: StatusMsgStyleSuccess,
		Text:  "Recorded macro",
	}, state.StatusMsg())

	assert.Equal(t, 0, len(logger.logEntries))
	ReplayRecordedUserMacro(state, 3) // Replay three times.
	assert.Equal(t, StatusMsg{
		Style: StatusMsgStyleSuccess,
		Text:  "Replayed macro",
	}, state.StatusMsg())
	expected := []actionLogEntry{
		// First time.
		{name: "a", isReplayingUserMacro: true},
		{name: "b", isReplayingUserMacro: true},
		// Second time.
		{name: "a", isReplayingUserMacro: true},
		{name: "b", isReplayingUserMacro: true},
		// Third time.
		{name: "a", isReplayingUserMacro: true},
		{name: "b", isReplayingUserMacro: true},
	}
	assert.Equal(t, expected, logger.logEntries)

	// Replay last action should repeat it three more times.
	ReplayLastActionMacro(state, 1)
	expected = append(expected, expected...)
	assert.Equal(t, expected, logger.logEntries)
}

func TestReplayUserMacroHandlesActionError(t *testing.T) {
	var logger actionLogger
	state := NewEditorState(100, 100, nil, nil)
	expectedErr := errors.New("test macro action failed")

	ToggleUserMacroRecording(state)
	AddToRecordingUserMacro(state, logger.buildAction("before error"))
	AddToRecordingUserMacro(state, func(s *EditorState) error {
		logger.logEntries = append(logger.logEntries, actionLogEntry{
			name:                 "error",
			isReplayingUserMacro: s.macroState.isReplayingUserMacro,
		})
		return expectedErr
	})
	AddToRecordingUserMacro(state, logger.buildAction("after error"))
	ToggleUserMacroRecording(state)

	ReplayRecordedUserMacro(state, 1)
	assert.Equal(t, StatusMsg{
		Style: StatusMsgStyleError,
		Text:  expectedErr.Error(),
	}, state.StatusMsg())
	assert.False(t, state.macroState.isReplayingUserMacro)
	assert.Equal(t, []actionLogEntry{
		{name: "before error", isReplayingUserMacro: true},
		{name: "error", isReplayingUserMacro: true},
	}, logger.logEntries)

	logger.clear()
	err := ReplayLastActionMacro(state, 1)
	require.ErrorIs(t, err, expectedErr)
	assert.False(t, state.macroState.isReplayingUserMacro)
	assert.Equal(t, []actionLogEntry{
		{name: "before error", isReplayingUserMacro: true},
		{name: "error", isReplayingUserMacro: true},
	}, logger.logEntries)
}

func TestLastActionIsUserMacro(t *testing.T) {
	var logger actionLogger
	state := NewEditorState(100, 100, nil, nil)
	ToggleUserMacroRecording(state)
	AddToRecordingUserMacro(state, logger.buildAction("a"))
	AddToRecordingUserMacro(state, logger.buildAction("b"))
	ToggleUserMacroRecording(state)
	ReplayRecordedUserMacro(state, 1)
	ReplayLastActionMacro(state, 1)
	expected := []actionLogEntry{
		{name: "a", isReplayingUserMacro: true},
		{name: "b", isReplayingUserMacro: true},
		{name: "a", isReplayingUserMacro: true},
		{name: "b", isReplayingUserMacro: true},
	}
	assert.Equal(t, expected, logger.logEntries)
}

func TestCancelUserMacro(t *testing.T) {
	var logger actionLogger
	state := NewEditorState(100, 100, nil, nil)

	// Record user macro
	ToggleUserMacroRecording(state)
	AddToRecordingUserMacro(state, logger.buildAction("a"))
	AddToRecordingUserMacro(state, logger.buildAction("b"))
	ToggleUserMacroRecording(state)

	// Start recording another macro, then immediately stop.
	ToggleUserMacroRecording(state)
	ToggleUserMacroRecording(state)

	// Status msg should say the macro was cancelled.
	assert.Equal(t, StatusMsg{
		Style: StatusMsgStyleSuccess,
		Text:  "Cancelled macro recording",
	}, state.StatusMsg())

	// Original macro should be preserved.
	ReplayRecordedUserMacro(state, 1)
	expected := []actionLogEntry{
		{name: "a", isReplayingUserMacro: true},
		{name: "b", isReplayingUserMacro: true},
	}
	assert.Equal(t, expected, logger.logEntries)
}

func TestReplayWithNoUserMacroRecorded(t *testing.T) {
	var logger actionLogger
	state := NewEditorState(100, 100, nil, nil)
	ReplayRecordedUserMacro(state, 1)
	assert.Equal(t, 0, len(logger.logEntries))
	assert.Equal(t, StatusMsg{
		Style: StatusMsgStyleError,
		Text:  "No macro has been recorded",
	}, state.StatusMsg())
}

func TestReplayUserMacroWhileRecordingUserMacro(t *testing.T) {
	var logger actionLogger
	state := NewEditorState(100, 100, nil, nil)

	// Record a macro.
	ToggleUserMacroRecording(state)
	AddToRecordingUserMacro(state, logger.buildAction("a"))
	ToggleUserMacroRecording(state)

	// Record a second macro, try to replay while recording.
	ToggleUserMacroRecording(state)
	ReplayRecordedUserMacro(state, 1)

	// Expect an error status.
	assert.Equal(t, StatusMsg{
		Style: StatusMsgStyleError,
		Text:  "Cannot replay a macro while recording a macro",
	}, state.StatusMsg())
}

func TestReplayLastActionWhileRecordingUserMacro(t *testing.T) {
	state := NewEditorState(100, 100, nil, nil)

	// Start recording a user macro.
	ToggleUserMacroRecording(state)

	// Attempt to replay the last action.
	ReplayLastActionMacro(state, 1)

	// Expect an error status.
	assert.Equal(t, StatusMsg{
		Style: StatusMsgStyleError,
		Text:  "Cannot repeat the last action while recording a macro",
	}, state.StatusMsg())
}

func TestReplayCheckpointUndo(t *testing.T) {
	state := NewEditorState(100, 100, nil, nil)

	// Record a macro switching normal -> insert -> normal -> insert -> normal mode.
	ToggleUserMacroRecording(state)
	AddToRecordingUserMacro(state, func(s *EditorState) error {
		BeginUndoEntry(s)
		setInputMode(s, InputModeInsert)
		InsertRune(s, 'a')
		return nil
	})
	AddToRecordingUserMacro(state, func(s *EditorState) error {
		CommitUndoEntry(s)
		setInputMode(s, InputModeNormal)
		return nil
	})
	AddToRecordingUserMacro(state, func(s *EditorState) error {
		BeginUndoEntry(s)
		setInputMode(s, InputModeInsert)
		InsertRune(s, 'b')
		return nil
	})
	AddToRecordingUserMacro(state, func(s *EditorState) error {
		CommitUndoEntry(s)
		setInputMode(s, InputModeNormal)
		return nil
	})
	ToggleUserMacroRecording(state)

	// Replay the macro.
	ReplayRecordedUserMacro(state, 1)
	assert.Equal(t, "ab", state.documentBuffer.textTree.String())

	// Undo to the last checkpoint, which should be at the start of the macro.
	Undo(state)
	assert.Equal(t, "", state.documentBuffer.textTree.String())
}
