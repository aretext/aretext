package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/adrg/xdg"

	"github.com/aretext/aretext/server"
)

func main() {
	config := server.Config{
		SocketPath: filepath.Join(xdg.RuntimeDir, "aretext.socket"),
		LockPath:   filepath.Join(xdg.RuntimeDir, "aretext.lock"),
	}
	err := server.NewServer(config).Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error running server: %s", err)
		os.Exit(1)
	}
}
