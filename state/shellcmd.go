package state

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"

	"github.com/aretext/aretext/clipboard"
	"github.com/aretext/aretext/text/utf8"
	"github.com/google/shlex"
	"github.com/pkg/errors"
)

// SuspendScreenFunc suspends the screen, executes a function, then resumes the screen.
// This is allows the shell command to take control of the terminal.
type SuspendScreenFunc func(func() error) error

// ShellCmdMode controls how the command's input and output are handled.
type ShellCmdMode int

const (
	ShellCmdModeSilent   = ShellCmdMode(iota) // accepts no input and any output is discarded.
	ShellCmdModeTerminal                      // takes control of the terminal.
	ShellCmdModeInsert                        // output is inserted into the document at the cursor position, replacing any selection.
)

// RunShellCmd executes the command in a shell.
func RunShellCmd(state *EditorState, shellCmd string, mode ShellCmdMode) {
	log.Printf("Running shell command: '%s'\n", shellCmd)
	env := envVars(state)

	var err error
	switch mode {
	case ShellCmdModeSilent:
		err = runInShellWithModeSilent(shellCmd, env)
	case ShellCmdModeTerminal:
		err = runInShellWithModeTerminal(state, shellCmd, env)
	case ShellCmdModeInsert:
		err = runInShellWithModeInsert(state, shellCmd, env)
	default:
		panic("Unrecognized shell cmd mode")
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

func runInShellWithModeSilent(shellCmd string, env []string) error {
	return runInShell(shellCmd, env, nil, nil, nil)
}

func runInShellWithModeTerminal(state *EditorState, shellCmd string, env []string) error {
	return state.suspendScreenFunc(func() error {
		clearTerminal()
		return runInShell(shellCmd, env, os.Stdin, os.Stdout, os.Stderr)
	})
}

func runInShellWithModeInsert(state *EditorState, shellCmd string, env []string) error {
	var buf bytes.Buffer
	stdin, stdout, stderr := io.Reader(nil), &buf, &buf
	err := runInShell(shellCmd, env, stdin, stdout, stderr)
	if err != nil {
		return err
	}

	v := utf8.NewValidator()
	isValid := v.ValidateBytes(buf.Bytes()) && v.ValidateEnd()
	if !isValid {
		return errors.New("Shell command output is not valid UTF-8")
	}

	state.clipboard.Set(clipboard.PageShellCmdOutput, clipboard.PageContent{
		Text: string(buf.Bytes()),
	})
	DeleteSelection(state, false)
	PasteBeforeCursor(state, clipboard.PageShellCmdOutput)
	SetInputMode(state, InputModeNormal)
	return nil
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
