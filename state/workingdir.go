package state

import (
	"fmt"
	"log"
	"os"
)

// SetWorkingDirectory changes the working directory to the specified path.
func SetWorkingDirectory(s *EditorState, dirPath string) {
	err := os.Chdir(dirPath)
	if err != nil {
		log.Printf("Error changing working directory to %q: %s", dirPath, err)
		SetStatusMsg(s, StatusMsg{
			Style: StatusMsgStyleError,
			Text:  fmt.Sprintf("Error changing working directory: %s", err),
		})
		return
	}

	log.Printf("Changed working directory to %q", dirPath)
	SetStatusMsg(s, StatusMsg{
		Style: StatusMsgStyleSuccess,
		Text:  fmt.Sprintf("Changed working directory to \"%s\"", dirPath),
	})
}
