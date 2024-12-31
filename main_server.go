package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/adrg/xdg"

	"github.com/aretext/aretext/clientserver"
)

func main() {
	config := clientserver.Config{
		ServerSocketPath: filepath.Join(xdg.RuntimeDir, "aretext.socket"),
		ServerLockPath:   filepath.Join(xdg.RuntimeDir, "aretext.lock"),
	}
	err := clientserver.NewServer(config).Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error running server: %s", err)
		os.Exit(1)
	}
}
