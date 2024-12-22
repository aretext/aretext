package state

import (
	"io"
	"log"
)

func init() {
	// Suppress noisy log output from these tests.
	log.SetOutput(io.Discard)
}
