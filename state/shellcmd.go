package state

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"

	"github.com/google/shlex"
	"github.com/pkg/errors"
)

// SuspendScreenFunc suspends the screen, executes a function, then resumes the screen.
// This is allows the shell command to take control of the terminal.
type SuspendScreenFunc func(func() error) error

// ShellCmdOutput controls the destination of the shell command's output.
type ShellCmdOutput int

const (
	ShellCmdOutputNone = ShellCmdOutput(iota)
	ShellCmdOutputTerminal
)

// RunShellCmd executes the command in a shell.
// It suspends the screen to give the command control of the terminal
// then resumes once the command completes.
func RunShellCmd(state *EditorState, shellCmd string, output ShellCmdOutput) {
	log.Printf("Running shell command: '%s'\n", shellCmd)
	env := envVars(state)

	var err error
	switch output {
	case ShellCmdOutputNone:
		err = runInShellWithOutputNone(shellCmd, env)
	case ShellCmdOutputTerminal:
		err = runInShellWithOutputTerminal(state, shellCmd, env)
	default:
		panic("Unrecognized shell cmd output")
	}

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
}

func envVars(state *EditorState) []string {
	env := os.Environ()

	selection, _ := copySelectionText(state.documentBuffer)
	if len(selection) > 0 {
		env = append(env, fmt.Sprintf("SELECTION=%s", selection))
	}

	return env
}

func runInShellWithOutputNone(shellCmd string, env []string) error {
	return runInShell(shellCmd, env, nil, nil, nil)
}

func runInShellWithOutputTerminal(state *EditorState, shellCmd string, env []string) error {
	return state.suspendScreenFunc(func() error {
		clearTerminal()
		return runInShell(shellCmd, env, os.Stdin, os.Stdout, os.Stderr)
	})
}

func clearTerminal() {
	clearCmd := exec.Command("clear")
	clearCmd.Stdout = os.Stdout
	clearCmd.Stderr = os.Stderr
	if err := clearCmd.Run(); err != nil {
		log.Printf("Error clearing screen: %v\n", err)
	}
}

func runInShell(shellCmd string, env []string, stdin io.Reader, stdout io.Writer, stderr io.Writer) error {
	s, err := shellProgAndArgs()
	if err != nil {
		return err
	}

	s = append(s, "-c", shellCmd)
	cmd := exec.Command(s[0], s[1:]...)
	cmd.Env = env
	cmd.Stdin = stdin
	cmd.Stdout = stdout
	cmd.Stderr = stderr

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
