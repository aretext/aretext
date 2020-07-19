package input

import (
	"github.com/wedaly/aretext/internal/pkg/exec"
)

// Command represents an action initiated by the user.
type Command interface{}

// QuitCommand is a command to exit the program.
type QuitCommand struct{}

// ExecCommand is a command to modify the state of the editor (text, cursor position, etc).
type ExecCommand struct {
	Mutator exec.Mutator
}
