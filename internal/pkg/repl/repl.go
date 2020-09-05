package repl

// Repl provides a read-eval-print loop for users to control the editor.
// All methods must be thread-safe.
type Repl interface {

	// Start begins running the read-eval-print loop.
	// A REPL can be started at most once.
	Start() error

	// Terminate stops the read-eval-print loop.
	// Once a Repl is terminated, it will no longer execute commands
	// or send output.
	Terminate() error

	// Interrupt attempts to stop the currently running command.
	Interrupt() error

	// SubmitInput sends user input to the REPL.
	// It will not wait for the REPL to evaluate the input, but it may block if
	// another input is being evaluated.
	SubmitInput(s string) error

	// PollOutput blocks until the REPL produces some output.
	PollOutput() (string, error)
}
