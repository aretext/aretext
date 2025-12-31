package state

import "github.com/aretext/aretext/config"

// ToggleShowTabs shows or hides tab characters in the document.
func ToggleShowTabs(s *EditorState) {
	toggleFlagAndSetStatus(s, &s.documentBuffer.showTabs, "Showing tabs", "Hiding tabs")
}

// ToggleShowSpaces shows or hides space characters in the document.
func ToggleShowSpaces(s *EditorState) {
	toggleFlagAndSetStatus(s, &s.documentBuffer.showSpaces, "Showing spaces", "Hiding spaces")
}

// ToggleShowUnicode shows or hides escaped unicode codepoints in the document.
func ToggleShowUnicode(s *EditorState) {
	toggleFlagAndSetStatus(s, &s.documentBuffer.showUnicode, "Showing unicode", "Hiding unicode")
}

// ToggleTabExpand toggles whether tabs should be expanded to spaces.
func ToggleTabExpand(s *EditorState) {
	toggleFlagAndSetStatus(s, &s.documentBuffer.tabExpand, "Enabled tab expand", "Disabled tab expand")
}

// ToggleShowLineNumbers shows or hides line numbers in the left margin.
func ToggleShowLineNumbers(s *EditorState) {
	toggleFlagAndSetStatus(s, &s.documentBuffer.showLineNum, "Showing line numbers", "Hiding line numbers")
}

// SetLineNumberMode sets the line number mode.
func SetLineNumberMode(s *EditorState, mode config.LineNumberMode) {
	switch mode {
	case config.LineNumberModeAbsolute:
		s.documentBuffer.lineNumberMode = config.LineNumberModeAbsolute
		SetStatusMsg(s, StatusMsg{
			Style: StatusMsgStyleSuccess,
			Text:  "Showing absolute line numbers",
		})
	case config.LineNumberModeRelative:
		s.documentBuffer.lineNumberMode = config.LineNumberModeRelative
		SetStatusMsg(s, StatusMsg{
			Style: StatusMsgStyleSuccess,
			Text:  "Showing relative line numbers",
		})
	default:
		SetStatusMsg(s, StatusMsg{
			Style: StatusMsgStyleError,
			Text:  "Line number mode not supported: " + string(mode),
		})
	}
}

// ToggleLineNumberMode toggles the line number mode between absolute and relative.
func ToggleLineNumberMode(s *EditorState) {
	switch s.documentBuffer.lineNumberMode {
	case config.LineNumberModeAbsolute:
		SetLineNumberMode(s, config.LineNumberModeRelative)
		SetStatusMsg(s, StatusMsg{
			Style: StatusMsgStyleSuccess,
			Text:  "Showing relative line numbers",
		})
	case config.LineNumberModeRelative:
		SetLineNumberMode(s, config.LineNumberModeAbsolute)
		SetStatusMsg(s, StatusMsg{
			Style: StatusMsgStyleSuccess,
			Text:  "Hiding line numbers",
		})
	}
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
