package state

import "github.com/aretext/aretext/text"

// Same as Linux PATH_MAX.
const maxTextFieldLen = 4096

// TextFieldAction is the action to perform with the text input by the user.
type TextFieldAction func(*EditorState, string) error

// TextFieldState represents the state of the text field.
// This is used to enter text such as the file path
// when creating a new file from within the editor.
type TextFieldState struct {
	promptText    string
	inputText     text.RuneStack
	action        TextFieldAction
	prevInputMode InputMode
}

func (s *TextFieldState) PromptText() string {
	return s.promptText
}

func (s *TextFieldState) InputText() string {
	return s.inputText.String()
}

func ShowTextField(state *EditorState, promptText string, action TextFieldAction) {
	state.textfield = &TextFieldState{
		promptText:    promptText,
		action:        action,
		prevInputMode: state.inputMode,
	}
	setInputMode(state, InputModeTextField)
}

func HideTextField(state *EditorState) {
	prevInputMode := state.textfield.prevInputMode
	state.textfield = &TextFieldState{}
	setInputMode(state, prevInputMode)
}

func AppendRuneToTextField(state *EditorState, r rune) {
	inputText := &state.textfield.inputText
	if inputText.Len() < maxTextFieldLen {
		inputText.Push(r)
	}
}

func DeleteRuneFromTextField(state *EditorState) {
	state.textfield.inputText.Pop()
}

func ExecuteTextFieldAction(state *EditorState) {
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
