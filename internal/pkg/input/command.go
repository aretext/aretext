package input

import (
	"fmt"

	"github.com/wedaly/aretext/internal/pkg/exec"
)

// Command represents an action initiated by the user.
type Command interface {
	fmt.Stringer
}

// QuitCommand is a command to exit the program.
type QuitCommand struct{}

func (c *QuitCommand) String() string {
	return "Quit()"
}

// ExecCommand is a command to modify the state of the editor (text, cursor position, etc).
type ExecCommand struct {
	Mutator exec.Mutator
}

func (c *ExecCommand) String() string {
	return fmt.Sprintf("Exec(%s)", c.Mutator)
}
