package state

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type actionLogEntry struct {
	name                 string
	isReplayingUserMacro bool
}

type actionLogger struct {
	logEntries []actionLogEntry
}

func (a *actionLogger) buildAction(name string) MacroAction {
	return func(s *EditorState) {
		a.logEntries = append(a.logEntries, actionLogEntry{
			name:                 name,
			isReplayingUserMacro: s.macroState.isReplayingUserMacro,
		})
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
	ReplayRecordedUserMacro(state)
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

func TestLastActionIsUserMacro(t *testing.T) {
	var logger actionLogger
	state := NewEditorState(100, 100, nil, nil)
	ToggleUserMacroRecording(state)
	AddToRecordingUserMacro(state, logger.buildAction("a"))
	AddToRecordingUserMacro(state, logger.buildAction("b"))
	ToggleUserMacroRecording(state)
	ReplayRecordedUserMacro(state)
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
	ReplayRecordedUserMacro(state)
	expected := []actionLogEntry{
		{name: "a", isReplayingUserMacro: true},
		{name: "b", isReplayingUserMacro: true},
	}
	assert.Equal(t, expected, logger.logEntries)
}

func TestReplayWithNoUserMacroRecorded(t *testing.T) {
	var logger actionLogger
	state := NewEditorState(100, 100, nil, nil)
	ReplayRecordedUserMacro(state)
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
	ReplayRecordedUserMacro(state)

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
	AddToRecordingUserMacro(state, func(s *EditorState) {
		BeginUndoEntry(s)
		setInputMode(s, InputModeInsert)
		InsertRune(s, 'a')
	})
	AddToRecordingUserMacro(state, func(s *EditorState) {
		CommitUndoEntry(s)
		setInputMode(s, InputModeNormal)
	})
	AddToRecordingUserMacro(state, func(s *EditorState) {
		BeginUndoEntry(s)
		setInputMode(s, InputModeInsert)
		InsertRune(s, 'b')
	})
	AddToRecordingUserMacro(state, func(s *EditorState) {
		CommitUndoEntry(s)
		setInputMode(s, InputModeNormal)
	})
	ToggleUserMacroRecording(state)

	// Replay the macro.
	ReplayRecordedUserMacro(state)
	assert.Equal(t, "ab", state.documentBuffer.textTree.String())

	// Undo to the last checkpoint, which should be at the start of the macro.
	Undo(state)
	assert.Equal(t, "", state.documentBuffer.textTree.String())
}
