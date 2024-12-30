package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/adrg/xdg"

	"github.com/aretext/aretext/client"
)

func main() {
	config := client.Config{
		ServerSocketPath: filepath.Join(xdg.RuntimeDir, "aretext.socket"),
	}
	err := client.NewClient(config).Run("test.txt")
	if err != nil {
		fmt.Fprintf(os.Stderr, "error running client: %s", err)
		os.Exit(1)
	}
}
