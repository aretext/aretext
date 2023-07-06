package shellcmd

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"runtime"
	"unicode/utf8"
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
		return "", fmt.Errorf("Shell command output is not valid UTF-8")
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
	prog := shellProg()
	commandArg := commandArgForShellProg(prog)
	cmd := exec.CommandContext(ctx, prog, commandArg, shellCmd)
	cmd.Env = env
	cmd.Stdin = stdin
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Cmd.Run: %w", err)
	}
	return nil
}

func shellProg() string {
	if s := os.Getenv("ARETEXT_SHELL"); s != "" {
		return s
	} else if s := os.Getenv("SHELL"); s != "" {
		return s
	} else if runtime.GOOS == "windows" {
		// Default to powershell on Windows.
		return "powershell.exe"
	} else {
		// Default to sh on Linux and macOS.
		return "sh"
	}
}

func commandArgForShellProg(s string) string {
	if s == "powershell.exe" {
		return "-Command"
	} else if s == "cmd.exe" {
		return "/c"
	} else {
		// This works for most (all?) shells on Linux/macOS,
		// including sh, bash, zsh, and fish.
		return "-c"
	}
}
