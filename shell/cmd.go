package shell

import (
	"log"
	"os"
	"os/exec"

	"github.com/google/shlex"
	"github.com/pkg/errors"
)

// Cmd represents a command to run in the shell.
type Cmd struct {
	cmd string
}

func NewCmd(cmd string) *Cmd {
	return &Cmd{cmd}
}

// Run executes the command in a shell.
// If the command exits with non-zero status, an error is returned.
// This assumes that the tcell screen has been suspended.
func (c *Cmd) Run() error {
	c.clearTerminal()
	return c.runInShell()
}

func (c *Cmd) String() string {
	return c.cmd
}

func (c *Cmd) clearTerminal() {
	clearCmd := exec.Command("clear")
	clearCmd.Stdout = os.Stdout
	clearCmd.Stderr = os.Stderr
	if err := clearCmd.Run(); err != nil {
		log.Printf("Error clearing screen: %v\n", err)
	}
}

func (c *Cmd) runInShell() error {
	s, err := shellProgAndArgs()
	if err != nil {
		return err
	}

	s = append(s, "-c", c.cmd)
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
