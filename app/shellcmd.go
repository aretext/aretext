package app

import (
	"io"
	"log"
	"os"
	"os/exec"

	"github.com/google/shlex"
	"github.com/pkg/errors"
)

// RunShellCmd executes a command in a shell and pipes the output to a pager (like `less`).
// If the command exits with non-zero status, an error is returned.
func RunShellCmd(shellCmd string) error {
	runClearCommand()

	// Start a pager process, which will receive the shell command's output.
	// The pager process will take control of the terminal so the user can scroll through the output.
	// We assume that on exit the pager process will return the terminal to its previous configuration.
	pagerStdin, pagerCleanup, err := startPager()
	if err != nil {
		return err
	}
	defer pagerCleanup()

	// Run the command in a shell, piping the output to the pager.
	if err := runCmdInShell(shellCmd, pagerStdin); err != nil {
		return err
	}
	return nil
}

func runClearCommand() {
	clearCmd := exec.Command("clear")
	clearCmd.Stdout = os.Stdout
	clearCmd.Stderr = os.Stderr
	if err := clearCmd.Run(); err != nil {
		log.Printf("Error clearing screen: %v\n", err)
	}
}

func startPager() (io.Writer, func(), error) {
	pager, err := findPagerCmd()
	if err != nil {
		return nil, nil, err
	}

	pagerCmd := exec.Command(pager[0], pager[1:]...)
	pagerStdin, err := pagerCmd.StdinPipe()
	if err != nil {
		return nil, nil, errors.Wrapf(err, "Cmd.StdinPipe")
	}

	pagerCmd.Stdout = os.Stdout
	pagerCmd.Stderr = os.Stderr
	if err := pagerCmd.Start(); err != nil {
		return nil, nil, errors.Wrapf(err, "Cmd.Start")
	}

	cleanupFunc := func() {
		// Close pager stdin so the pager process exits.
		if err := pagerStdin.Close(); err != nil {
			log.Printf("Error closing pager stdin: %v\n", err)
		}

		// Wait for the pager process to exit.
		if err := pagerCmd.Wait(); err != nil {
			log.Printf("Error exiting pager: %v\n", err)
		}
	}

	return pagerStdin, cleanupFunc, nil
}

func runCmdInShell(shellCmd string, pagerStdin io.Writer) error {
	s, err := findShellCmd()
	if err != nil {
		return err
	}

	s = append(s, "-c", shellCmd)
	c := exec.Command(s[0], s[1:]...)
	c.Env = os.Environ()
	c.Stdout = pagerStdin
	c.Stderr = pagerStdin
	if err := c.Run(); err != nil {
		return errors.Wrapf(err, "Cmd.Run")
	}
	return nil
}

func findShellCmd() ([]string, error) {
	const defaultShell = "sh"
	return cmdFromEnvVar("SHELL", defaultShell)
}

func findPagerCmd() ([]string, error) {
	const defaultPager = "less -R"
	return cmdFromEnvVar("PAGER", defaultPager)
}

func cmdFromEnvVar(envVar string, defaultCmd string) ([]string, error) {
	s := os.Getenv(envVar)
	if s == "" {
		s = defaultCmd
	}

	// If defaultCmd != "", then the input string will always have at least one char, so len(parts) > 0.
	parts, err := shlex.Split(s)
	if err != nil {
		return nil, errors.Wrapf(err, "shlex.Split")
	}
	return parts, nil
}
