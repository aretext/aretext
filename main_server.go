package main

import (
	"fmt"
	"os"

	"github.com/aretext/aretext/server"
)

func main() {
	config := server.Config{
		SocketPath: "/var/run/aretext.socket",
		LockPath:   "/var/run/aretext.lock",
	}
	err := server.NewServer(config).Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error running server: %s", err)
		os.Exit(1)
	}
}
