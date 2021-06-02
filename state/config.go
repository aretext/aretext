package state

// ToggleShowTabs shows or hides tab characters in the document.
func ToggleShowTabs(s *EditorState) {
	s.documentBuffer.showTabs = !s.documentBuffer.showTabs
}

// ToggleShowTrailingSpaces shows or hides trailing spaces at the end of each line.
func ToggleShowTrailingSpaces(s *EditorState) {
	s.documentBuffer.showTrailingSpaces = !s.documentBuffer.showTrailingSpaces
}

// ToggleShowLineNumbers shows or hides line numbers in the left margin.
func ToggleShowLineNumbers(s *EditorState) {
	s.documentBuffer.showLineNum = !s.documentBuffer.showLineNum
}
