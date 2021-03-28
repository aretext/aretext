package state

// ScheduleShellCmd schedules a shell command to be executed by the editor.
func ScheduleShellCmd(state *EditorState, shellCmd string) {
	state.scheduledShellCmd = shellCmd
}
