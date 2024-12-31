package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/adrg/xdg"

	"github.com/aretext/aretext/clientserver"
)

func main() {
	logFile, err := os.Create("client.log")
	if err != nil {
		fmt.Fprintf(os.Stderr, "error creating log file: %s\n", err)
		os.Exit(1)
	}
	defer logFile.Close()
	log.SetOutput(logFile)

	config := clientserver.Config{
		ServerSocketPath: filepath.Join(xdg.RuntimeDir, "aretext.socket"),
	}
	err = clientserver.NewClient(config).Run("test.txt")
	if err != nil {
		fmt.Fprintf(os.Stderr, "error running client: %s\n", err)
		os.Exit(1)
	}
}
