package state

// ToggleShowLineNumbers shows or hides line numbers in the left margin.
func ToggleShowLineNumbers(s *EditorState) {
	s.documentBuffer.showLineNum = !s.documentBuffer.showLineNum
}
