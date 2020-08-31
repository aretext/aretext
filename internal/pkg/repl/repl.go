package repl

// Repl provides a read-eval-print loop for users to control the editor.
type Repl interface {

	// SubmitInput sends user input to the REPL.
	// It will not wait for the REPL to evaluate the input, but it may block if
	// another input is being evaluated.
	SubmitInput(s string)

	// PollOutput blocks until the REPL produces some output.
	PollOutput() (string, error)
}
