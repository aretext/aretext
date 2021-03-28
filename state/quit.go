package state

// Quit sets a flag that terminates the program.
func Quit(state *EditorState) {
	state.fileWatcher.Stop()
	state.quitFlag = true
}
