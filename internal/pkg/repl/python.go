package repl

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"
	"unicode/utf8"

	"github.com/pkg/errors"
)

type pythonRepl struct {
	mu         sync.Mutex
	started    bool
	cmd        *exec.Cmd
	stdinPipe  io.Writer
	stdoutPipe io.Reader
	stderrPipe io.Reader
	outputChan chan rune
}

// NewPythonRepl returns an instance of a Python REPL.
func NewPythonRepl() Repl {
	return &pythonRepl{
		outputChan: make(chan rune, 1024),
	}
}

// Start begins running the Python REPL.
func (r *pythonRepl) Start() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.started {
		return errors.New("Repl has already been started")
	}

	cmd := exec.Command("python", "-u", "-m", "pyaretext")

	stdinPipe, err := cmd.StdinPipe()
	if err != nil {
		return errors.Wrapf(err, "cmd.StdinPipe()")
	}

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return errors.Wrapf(err, "cmd.StdoutPipe()")
	}

	// Execute the Python interpreter as a subprocess, with buffering disabled.
	// The pyaretext module implements the REPL and a Python API for interacting
	// with the aretext host program.
	// We assume that the pyaretext Python module is on the PYTHONPATH
	// (either because the module is in the same directory as the module or
	// the module has been installed using setup.py); if not, we'll get an error here.
	if err := cmd.Start(); err != nil {
		return errors.Wrapf(err, "cmd.Start()")
	}

	r.started = true
	r.cmd = cmd
	r.stdinPipe = stdinPipe

	// Start goroutines to read command output and write runes to outputChan.
	// We do these in a separate goroutine because the reads can block.
	go r.readCmdOutput(stdoutPipe)

	return nil
}

func (r *pythonRepl) readCmdOutput(pipe io.Reader) {
	scanner := bufio.NewScanner(pipe)
	scanner.Split(bufio.ScanRunes)
	for scanner.Scan() {
		c, _ := utf8.DecodeRune(scanner.Bytes())
		r.outputChan <- c
	}
	if err := scanner.Err(); err != nil {
		log.Printf("Error reading cmd output: %v", err)
	}
	close(r.outputChan)
}

// Terminate stops the REPL.
func (r *pythonRepl) Terminate() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if !r.started {
		return errors.New("Repl was not started")
	}

	if err := r.cmd.Process.Kill(); err != nil {
		return errors.Wrapf(err, "cmd.Process.Kill()")
	}

	if err := r.cmd.Wait(); err != nil && !exitedWithSigterm(err) {
		return errors.Wrapf(err, "cmd.Wait()")
	}

	return nil
}

// exitedWithSigterm returns whether the error was an caused by a SIGTERM signal.
func exitedWithSigterm(err error) bool {
	exitErr, ok := err.(*exec.ExitError)
	return ok && exitErr.ExitCode() == -1
}

// Interrupt stops the currently running computation in the REPL.
func (r *pythonRepl) Interrupt() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if !r.started {
		return errors.New("Repl was not started")
	}

	// Send a SIGINT signal to the Python interpreter.
	// This will trigger a KeyboardInterrupt exception in the Python code,
	// effectively interrupting the currently executing program.
	// The program can continue running as long as it captures and handles the exception.
	if err := r.cmd.Process.Signal(os.Interrupt); err != nil {
		return errors.Wrapf(err, "cmd.Process.Signal()")
	}

	return nil
}

// SubmitInput sends an input string to the REPL.
func (r *pythonRepl) SubmitInput(s string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if !r.started {
		return errors.New("Repl was not started")
	}

	if _, err := fmt.Fprintf(r.stdinPipe, "%s\n", s); err != nil {
		return errors.Wrapf(err, "write to cmd stdin pipe")
	}

	return nil
}

// PollOutput blocks until output is available from the REPL.
// The output is guaranteed to be valid UTF-8 (invalid bytes will be represented
// by the unicode replacement character).
func (r *pythonRepl) PollOutput() (string, error) {
	// We do NOT acquire the mutex here to avoid a potential deadlock.
	// 1) SubmitInput acquires the mutex, then blocks writing to stdin
	// 2) Python subprocess blocks writing to stdout.
	// 3) PollOutput blocks waiting on the mutex, so stdout remains blocked.
	//
	// To avoid the deadlock, do NOT acquire the mutex in this method.
	// This means that we can always unblock stdout by calling PollOutput(),
	// even if other methods have acquired the mutex.
	//
	// This is safe because PollOutput uses only the outputChan, which is thread-safe.

	var sb strings.Builder

	// Block until we have at least one character to output.
	c, ok := <-r.outputChan
	if !ok {
		return "", errors.New("Repl output channel closed")
	}
	if _, err := sb.WriteRune(c); err != nil {
		return "", errors.Wrapf(err, "WriteRune()")
	}

	// Read up to 1024 bytes from the outputChan.
	// If there aren't any immediately available, return what we have so far.
	for {
		select {
		case c, ok := <-r.outputChan:
			if !ok {
				// Output channel is closed, so return what we have so far.
				return sb.String(), nil
			}
			if _, err := sb.WriteRune(c); err != nil {
				return "", errors.Wrapf(err, "WriteRune()")
			}
			if sb.Len() >= 1024 {
				// We've accumulated enough runes for now, so return what we have so far.
				// This bounds the maximum memory allocated by the string builder.
				// It also prevents us from looping indefinitely if the Python program spams stdout.
				return sb.String(), nil
			}

		default:
			// No more runes to read now, so return what we have so far.
			return sb.String(), nil
		}
	}
}
