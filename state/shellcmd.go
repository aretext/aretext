package state

import (
	"log"
	"os"
	"os/exec"

	"github.com/google/shlex"
	"github.com/pkg/errors"
)

// SuspendScreenFunc suspends the screen, executes a function, then resumes the screen.
// This is allows the shell command to take control of the terminal.
type SuspendScreenFunc func(func()) error

// RunShellCmd executes the command in a shell.
// It suspends the screen to give the command control of the terminal
// then resumes once the command completes.
func RunShellCmd(state *EditorState, shellCmd string) {
	log.Printf("Running shell command: '%s'\n", shellCmd)
	err := state.suspendScreenFunc(func() {
		clearTerminal()
		err := runInShell(shellCmd)
		if err != nil {
			SetStatusMsg(state, StatusMsg{
				Style: StatusMsgStyleError,
				Text:  err.Error(),
			})
			return
		}
		SetStatusMsg(state, StatusMsg{
			Style: StatusMsgStyleSuccess,
			Text:  "Shell command completed successfully",
		})
	})
	if err != nil {
		log.Printf("Error suspending the screen: %v\n", err)
	}
}

func clearTerminal() {
	clearCmd := exec.Command("clear")
	clearCmd.Stdout = os.Stdout
	clearCmd.Stderr = os.Stderr
	if err := clearCmd.Run(); err != nil {
		log.Printf("Error clearing screen: %v\n", err)
	}
}

func runInShell(shellCmd string) error {
	s, err := shellProgAndArgs()
	if err != nil {
		return err
	}

	s = append(s, "-c", shellCmd)
	cmd := exec.Command(s[0], s[1:]...)
	cmd.Env = os.Environ()

	// Allow the shell to take over stdin/stdout/stderr.
	// This assumes that the tcell screen has been suspended.
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return errors.Wrapf(err, "Cmd.Run")
	}
	return nil
}

const defaultShell = "sh"

func shellProgAndArgs() ([]string, error) {
	s := os.Getenv("SHELL")
	if s == "" {
		s = defaultShell
	}

	// The $SHELL env var might include command line args for the shell command.
	// These args need to be passed separately to exec.Command, so split them here.
	parts, err := shlex.Split(s)
	if err != nil {
		return nil, errors.Wrapf(err, "shlex.Split")
	}
	return parts, nil
}
