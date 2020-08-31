package input

import (
	"github.com/wedaly/aretext/internal/pkg/repl"
)

// Config controls how user input is interpreted.
type Config struct {
	// Repl is the read-eval-print-loop where user input should be sent.
	Repl repl.Repl

	// ScrollLines is the number of lines to scroll up or down with Ctrl-U / Ctrl-D.
	ScrollLines uint64
}
