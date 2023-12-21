package state

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestShowAndHideTextField(t *testing.T) {
	testCases := []struct {
		name          string
		fromInputMode InputMode
	}{
		{
			name:          "from normal mode",
			fromInputMode: InputModeNormal,
		},
		{
			name:          "from visual mode",
			fromInputMode: InputModeVisual,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			state := NewEditorState(100, 100, nil, nil)
			setInputMode(state, tc.fromInputMode)
			emptyAction := func(_ *EditorState, _ string) error { return nil }

			ShowTextField(state, "test prompt", emptyAction)
			assert.Equal(t, InputModeTextField, state.InputMode())
			assert.Equal(t, "test prompt", state.TextField().PromptText())
			assert.Equal(t, "", state.TextField().InputText())

			HideTextField(state)
			assert.Equal(t, tc.fromInputMode, state.InputMode())
			assert.Equal(t, "", state.TextField().PromptText())
			assert.Equal(t, "", state.TextField().InputText())
		})
	}
}

func TestAppendRuneToTextField(t *testing.T) {
	state := NewEditorState(100, 100, nil, nil)
	emptyAction := func(_ *EditorState, _ string) error { return nil }
	ShowTextField(state, "test prompt", emptyAction)
	assert.Equal(t, InputModeTextField, state.InputMode())
	assert.Equal(t, "test prompt", state.TextField().PromptText())

	AppendRuneToTextField(state, 'a')
	assert.Equal(t, "a", state.TextField().InputText())
	AppendRuneToTextField(state, 'b')
	assert.Equal(t, "ab", state.TextField().InputText())
	AppendRuneToTextField(state, 'c')
	assert.Equal(t, "abc", state.TextField().InputText())
}

func TestAppendRuneToTextFieldMaxLimit(t *testing.T) {
	state := NewEditorState(100, 100, nil, nil)
	emptyAction := func(_ *EditorState, _ string) error { return nil }
	ShowTextField(state, "test prompt", emptyAction)

	for i := 0; i < maxTextFieldLen+5; i++ {
		AppendRuneToTextField(state, 'x')
	}
	assert.Equal(t, maxTextFieldLen, len(state.TextField().InputText()))
}

func TestDeleteRuneFromTextField(t *testing.T) {
	state := NewEditorState(100, 100, nil, nil)
	emptyAction := func(_ *EditorState, _ string) error { return nil }
	ShowTextField(state, "test prompt", emptyAction)
	assert.Equal(t, InputModeTextField, state.InputMode())
	assert.Equal(t, "test prompt", state.TextField().PromptText())

	AppendRuneToTextField(state, 'a')
	AppendRuneToTextField(state, 'b')
	AppendRuneToTextField(state, 'c')

	DeleteRuneFromTextField(state)
	assert.Equal(t, "ab", state.TextField().InputText())
	DeleteRuneFromTextField(state)
	assert.Equal(t, "a", state.TextField().InputText())
	DeleteRuneFromTextField(state)
	assert.Equal(t, "", state.TextField().InputText())

	// Delete from empty input is a no-op.
	DeleteRuneFromTextField(state)
	assert.Equal(t, "", state.TextField().InputText())
}

func TestExecuteTextFieldActionSuccess(t *testing.T) {
	state := NewEditorState(100, 100, nil, nil)

	fakeAction := func(state *EditorState, inputText string) error {
		SetStatusMsg(state, StatusMsg{
			Style: StatusMsgStyleSuccess,
			Text:  fmt.Sprintf("Input: %s", inputText),
		})
		return nil
	}

	ShowTextField(state, "test prompt", fakeAction)
	AppendRuneToTextField(state, 'a')
	AppendRuneToTextField(state, 'b')
	AppendRuneToTextField(state, 'c')
	ExecuteTextFieldAction(state)

	assert.Equal(t, InputModeNormal, state.InputMode())
	assert.Equal(t, "", state.TextField().PromptText())
	assert.Equal(t, "", state.TextField().InputText())
	assert.Equal(t, "Input: abc", state.StatusMsg().Text)
}

func TestExecuteTextFieldActionError(t *testing.T) {
	state := NewEditorState(100, 100, nil, nil)
	errorAction := func(_ *EditorState, _ string) error {
		return fmt.Errorf("TEST ERROR")
	}
	ShowTextField(state, "test prompt", errorAction)
	AppendRuneToTextField(state, 'a')
	AppendRuneToTextField(state, 'b')
	AppendRuneToTextField(state, 'c')
	ExecuteTextFieldAction(state)

	assert.Equal(t, InputModeTextField, state.InputMode())
	assert.Equal(t, "test prompt", state.TextField().PromptText())
	assert.Equal(t, "abc", state.TextField().InputText())
	assert.Equal(t, StatusMsgStyleError, state.StatusMsg().Style)
	assert.Equal(t, "TEST ERROR", state.StatusMsg().Text)
}
