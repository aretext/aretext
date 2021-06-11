package state

// ToggleShowTabs shows or hides tab characters in the document.
func ToggleShowTabs(s *EditorState) {
	s.documentBuffer.showTabs = !s.documentBuffer.showTabs
}

// ToggleShowLineNumbers shows or hides line numbers in the left margin.
func ToggleShowLineNumbers(s *EditorState) {
	s.documentBuffer.showLineNum = !s.documentBuffer.showLineNum
}

// ToggleAutoIndent enables or disables auto-indent.
func ToggleAutoIndent(s *EditorState) {
	s.documentBuffer.autoIndent = !s.documentBuffer.autoIndent
}
