package repl

import (
	"fmt"
)

// dummyRepl is a REPL implementation that echos the user's input.
type dummyRepl struct {
	outputChan chan string
}

func NewDummyRepl() Repl {
	outputChan := make(chan string, 1)
	outputChan <- ">>> "
	return &dummyRepl{outputChan}
}

func (r *dummyRepl) SubmitInput(s string) {
	r.outputChan <- fmt.Sprintf("[DUMMY] %s\n>>> ", s)
}

func (r *dummyRepl) PollOutput() (string, error) {
	s := <-r.outputChan
	return s, nil
}
