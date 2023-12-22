package state

import (
	"fmt"
	"strings"
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

			ShowTextField(state, "test prompt", emptyAction, nil)
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
	ShowTextField(state, "test prompt", emptyAction, nil)
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
	ShowTextField(state, "test prompt", emptyAction, nil)

	for i := 0; i < maxTextFieldLen+5; i++ {
		AppendRuneToTextField(state, 'x')
	}
	assert.Equal(t, maxTextFieldLen, len(state.TextField().InputText()))
}

func TestDeleteRuneFromTextField(t *testing.T) {
	state := NewEditorState(100, 100, nil, nil)
	emptyAction := func(_ *EditorState, _ string) error { return nil }
	ShowTextField(state, "test prompt", emptyAction, nil)
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

	ShowTextField(state, "test prompt", fakeAction, nil)
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
	ShowTextField(state, "test prompt", errorAction, nil)
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

func TestAutocompleteTextField(t *testing.T) {
	state := NewEditorState(100, 100, nil, nil)
	fakeAction := func(state *EditorState, inputText string) error {
		SetStatusMsg(state, StatusMsg{
			Style: StatusMsgStyleSuccess,
			Text:  fmt.Sprintf("Input: %s", inputText),
		})
		return nil
	}

	fakeCandidates := []string{"aa", "ab", "ba", "bb"}
	fakeAutocompleteFunc := func(prefix string) ([]string, error) {
		var suffixes []string
		for _, s := range fakeCandidates {
			if strings.HasPrefix(s, prefix) && len(prefix) < len(s) {
				suffixes = append(suffixes, s[len(prefix):])
			}
		}
		return suffixes, nil
	}

	ShowTextField(state, "test prompt", fakeAction, fakeAutocompleteFunc)
	assert.Equal(t, InputModeTextField, state.InputMode())
	assert.Equal(t, "", state.TextField().InputText())
	assert.Equal(t, "", state.TextField().AutocompleteSuffix())

	AutocompleteTextField(state)
	assert.Equal(t, "", state.TextField().InputText())
	assert.Equal(t, "aa", state.TextField().AutocompleteSuffix())

	AutocompleteTextField(state)
	assert.Equal(t, "", state.TextField().InputText())
	assert.Equal(t, "ab", state.TextField().AutocompleteSuffix())

	AutocompleteTextField(state)
	assert.Equal(t, "", state.TextField().InputText())
	assert.Equal(t, "ba", state.TextField().AutocompleteSuffix())

	AutocompleteTextField(state)
	assert.Equal(t, "", state.TextField().InputText())
	assert.Equal(t, "bb", state.TextField().AutocompleteSuffix())

	AutocompleteTextField(state)
	assert.Equal(t, "", state.TextField().InputText())
	assert.Equal(t, "", state.TextField().AutocompleteSuffix())

	AppendRuneToTextField(state, 'b')
	assert.Equal(t, "b", state.TextField().InputText())
	assert.Equal(t, "", state.TextField().AutocompleteSuffix())

	AutocompleteTextField(state)
	assert.Equal(t, "b", state.TextField().InputText())
	assert.Equal(t, "a", state.TextField().AutocompleteSuffix())

	AutocompleteTextField(state)
	assert.Equal(t, "b", state.TextField().InputText())
	assert.Equal(t, "b", state.TextField().AutocompleteSuffix())

	ExecuteTextFieldAction(state)
	assert.Equal(t, InputModeNormal, state.InputMode())
	assert.Equal(t, "", state.TextField().PromptText())
	assert.Equal(t, "", state.TextField().InputText())
	assert.Equal(t, "Input: bb", state.StatusMsg().Text)
}

func TestAutocompleteTextFieldError(t *testing.T) {
	state := NewEditorState(100, 100, nil, nil)
	emptyAction := func(_ *EditorState, _ string) error { return nil }
	errorAutocompleteFunc := func(prefix string) ([]string, error) {
		return nil, fmt.Errorf("autocomplete error")
	}

	// Show autocomplete error in status msg.
	ShowTextField(state, "test prompt", emptyAction, errorAutocompleteFunc)
	AutocompleteTextField(state)
	assert.Equal(t, InputModeTextField, state.InputMode())
	assert.Equal(t, StatusMsgStyleError, state.StatusMsg().Style)
	assert.Equal(t, "Error occurred during autocomplete: autocomplete error", state.StatusMsg().Text)

	// Typing more clears the status msg.
	AppendRuneToTextField(state, 'a')
	assert.Equal(t, "", state.StatusMsg().Text)

	// Autocomplete again to bring the error back.
	AutocompleteTextField(state)
	assert.Equal(t, StatusMsgStyleError, state.StatusMsg().Style)
	assert.Equal(t, "Error occurred during autocomplete: autocomplete error", state.StatusMsg().Text)

	// Deleting also clears the status msg.
	DeleteRuneFromTextField(state)
	assert.Equal(t, "", state.StatusMsg().Text)
}
