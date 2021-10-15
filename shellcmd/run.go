package shellcmd

import (
	"bytes"
	"context"
	"io"
	"log"
	"os"
	"os/exec"
	"unicode/utf8"

	"github.com/google/shlex"
	"github.com/pkg/errors"
)

// RunSilent runs the command and discards any output.
func RunSilent(ctx context.Context, cmd string, env []string) error {
	return runInShell(ctx, cmd, env, nil, nil, nil)
}

// RunInTerminal runs the command using inputs and outputs of the current process.
func RunInTerminal(ctx context.Context, cmd string, env []string) error {
	clearTerminal(ctx)
	return runInShell(ctx, cmd, env, os.Stdin, os.Stdout, os.Stderr)
}

// RunAndCaptureOutput runs the command and returns its stdout as a byte slice.
// If the output is not valid UTF-8 text, this returns an error.
func RunAndCaptureOutput(ctx context.Context, cmd string, env []string) (string, error) {
	var buf bytes.Buffer
	stdin, stdout, stderr := io.Reader(nil), &buf, io.Writer(nil)
	err := runInShell(ctx, cmd, env, stdin, stdout, stderr)
	if err != nil {
		return "", err
	}

	if !utf8.Valid(buf.Bytes()) {
		return "", errors.New("Shell command output is not valid UTF-8")
	}

	return buf.String(), nil
}

func clearTerminal(ctx context.Context) {
	clearCmd := exec.CommandContext(ctx, "clear")
	clearCmd.Stdout = os.Stdout
	clearCmd.Stderr = os.Stderr
	if err := clearCmd.Run(); err != nil {
		log.Printf("Error clearing screen: %v\n", err)
	}
}

func runInShell(ctx context.Context, shellCmd string, env []string, stdin io.Reader, stdout io.Writer, stderr io.Writer) error {
	s, err := shellProgAndArgs()
	if err != nil {
		return err
	}

	s = append(s, "-c", shellCmd)
	cmd := exec.CommandContext(ctx, s[0], s[1:]...)
	cmd.Env = env
	cmd.Stdin = stdin
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	if err := cmd.Run(); err != nil {
		return errors.Wrap(err, "Cmd.Run")
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
		return nil, errors.Wrap(err, "shlex.Split")
	}
	return parts, nil
}
