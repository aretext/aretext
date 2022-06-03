package state

// ToggleShowTabs shows or hides tab characters in the document.
func ToggleShowTabs(s *EditorState) {
	toggleFlagAndSetStatus(s, &s.documentBuffer.showTabs, "Showing tabs", "Hiding tabs")
}

// ToggleShowSpaces shows or hides space characters in the document.
func ToggleShowSpaces(s *EditorState) {
	toggleFlagAndSetStatus(s, &s.documentBuffer.showSpaces, "Showing spaces", "Hiding spaces")
}

// ToggleTabExpand toggles whether tabs should be expanded to spaces.
func ToggleTabExpand(s *EditorState) {
	toggleFlagAndSetStatus(s, &s.documentBuffer.tabExpand, "Enabled tab expand", "Disabled tab expand")
}

// ToggleShowLineNumbers shows or hides line numbers in the left margin.
func ToggleShowLineNumbers(s *EditorState) {
	toggleFlagAndSetStatus(s, &s.documentBuffer.showLineNum, "Showing line numbers", "Hiding line numbers")
}

// ToggleAutoIndent enables or disables auto-indent.
func ToggleAutoIndent(s *EditorState) {
	toggleFlagAndSetStatus(s, &s.documentBuffer.autoIndent, "Enabled auto-indent", "Disabled auto-indent")
}

func toggleFlagAndSetStatus(s *EditorState, flagValue *bool, enabledMsg string, disabledMsg string) {
	*flagValue = !(*flagValue)

	var msg string
	if *flagValue {
		msg = enabledMsg
	} else {
		msg = disabledMsg
	}

	SetStatusMsg(s, StatusMsg{
		Style: StatusMsgStyleSuccess,
		Text:  msg,
	})
}
