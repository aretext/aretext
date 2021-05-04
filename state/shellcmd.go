package state

import (
	"log"

	"github.com/aretext/aretext/shell"
)

// ScheduleShellCmd schedules a shell command to be executed by the editor.
func ScheduleShellCmd(state *EditorState, shellCmd string) {
	log.Printf("Scheduled shell command: '%s'\n", shellCmd)
	state.scheduledShellCmd = shell.NewCmd(shellCmd)
}
