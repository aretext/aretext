package state

import (
	"fmt"

	"github.com/aretext/aretext/editor/text"
)

// Same as Linux PATH_MAX.
const maxTextFieldLen = 4096

// TextFieldAction is the action to perform with the text input by the user.
type TextFieldAction func(*EditorState, string) error

// TextFieldAutocompleteFunc retrieves autocomplete suffixes for a given prefix.
// It is acceptable to return an empty slice if there are no autocompletions,
// but every string in the slice must have non-zero length.
type TextFieldAutocompleteFunc func(prefix string) ([]string, error)

// TextFieldState represents the state of the text field.
// This is used to enter text such as the file path
// when creating a new file from within the editor.
type TextFieldState struct {
	promptText            string
	inputText             text.RuneStack
	action                TextFieldAction
	prevInputMode         InputMode
	autocompleteFunc      TextFieldAutocompleteFunc // Set to nil to disable autocompletion.
	autocompleteSuffixes  []string
	autocompleteSuffixIdx int
}

func (s *TextFieldState) PromptText() string {
	return s.promptText
}

func (s *TextFieldState) InputText() string {
	return s.inputText.String()
}

func (s *TextFieldState) AutocompleteSuffix() string {
	if s.autocompleteSuffixIdx < len(s.autocompleteSuffixes) {
		return s.autocompleteSuffixes[s.autocompleteSuffixIdx]
	} else {
		return ""
	}
}

func (s *TextFieldState) applyAutocomplete() {
	for _, r := range s.AutocompleteSuffix() {
		s.inputText.Push(r)
	}
	s.autocompleteSuffixes = nil
	s.autocompleteSuffixIdx = 0
}

func ShowTextField(state *EditorState, promptText string, action TextFieldAction, autocompleteFunc TextFieldAutocompleteFunc) {
	state.textfield = &TextFieldState{
		promptText:       promptText,
		action:           action,
		prevInputMode:    state.inputMode,
		autocompleteFunc: autocompleteFunc,
	}
	setInputMode(state, InputModeTextField)
}

func HideTextField(state *EditorState) {
	prevInputMode := state.textfield.prevInputMode
	state.textfield = &TextFieldState{}
	setInputMode(state, prevInputMode)
}

func AppendRuneToTextField(state *EditorState, r rune) {
	state.textfield.applyAutocomplete()
	inputText := &state.textfield.inputText
	if inputText.Len() < maxTextFieldLen {
		inputText.Push(r)
	}
	SetStatusMsg(state, StatusMsg{})
}

func DeleteRuneFromTextField(state *EditorState) {
	state.textfield.applyAutocomplete()
	state.textfield.inputText.Pop()
	SetStatusMsg(state, StatusMsg{})
}

func ExecuteTextFieldAction(state *EditorState) {
	state.textfield.applyAutocomplete()
	action := state.textfield.action
	inputText := state.textfield.InputText()
	err := action(state, inputText)
	if err != nil {
		// If the action failed, show the error as a status message,
		// but remain in the input mode so the user can edit/retry.
		SetStatusMsg(state, StatusMsg{
			Style: StatusMsgStyleError,
			Text:  err.Error(),
		})
		return
	}

	// The action completed successfully, so hide the text field.
	HideTextField(state)
}

// AutocompleteTextField performs autocompletion on the text field input.
// If there are multiple matching suffixes, repeated invocations will cycle
// through the options (including the original input).
func AutocompleteTextField(state *EditorState) {
	tf := state.textfield
	if tf.autocompleteFunc == nil {
		// Autocomplete disabled.
		return
	}

	// If we already have autocomplete suffixes, cycle through them.
	if len(tf.autocompleteSuffixes) > 0 {
		tf.autocompleteSuffixIdx = (tf.autocompleteSuffixIdx + 1) % len(tf.autocompleteSuffixes)
		return
	}

	// Otherwise, retrieve suffixes for the current prefix.
	prefix := tf.inputText.String()
	suffixes, err := tf.autocompleteFunc(prefix)
	if err != nil {
		SetStatusMsg(state, StatusMsg{
			Style: StatusMsgStyleError,
			Text:  fmt.Sprintf("Error occurred during autocomplete: %s", err),
		})
		return
	}

	SetStatusMsg(state, StatusMsg{})

	if len(suffixes) > 0 {
		tf.autocompleteSuffixes = append(suffixes, "") // Last item is always "" to show just the prefix.
		tf.autocompleteSuffixIdx = 0
	}
}
