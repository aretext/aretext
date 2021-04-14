package app

import (
	"log"
	"os"
	"os/exec"

	"github.com/google/shlex"
	"github.com/pkg/errors"
)

// RunShellCmd executes a command in a shell.
// If the command exits with non-zero status, an error is returned.
// This assumes that the tcell screen has been suspended.
func RunShellCmd(shellCmd string) error {
	runClearCommand()
	return runCmdInShell(shellCmd)
}

func runClearCommand() {
	clearCmd := exec.Command("clear")
	clearCmd.Stdout = os.Stdout
	clearCmd.Stderr = os.Stderr
	if err := clearCmd.Run(); err != nil {
		log.Printf("Error clearing screen: %v\n", err)
	}
}

func runCmdInShell(shellCmd string) error {
	s, err := findShellCmd()
	if err != nil {
		return err
	}

	s = append(s, "-c", shellCmd)
	c := exec.Command(s[0], s[1:]...)
	c.Env = os.Environ()

	// Allow the shell to take over stdin/stdout/stderr.
	// This assumes that the tcell screen has been suspended.
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr

	if err := c.Run(); err != nil {
		return errors.Wrapf(err, "Cmd.Run")
	}
	return nil
}

const defaultShell = "sh"

func findShellCmd() ([]string, error) {
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
