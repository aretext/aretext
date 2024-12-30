package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/adrg/xdg"

	"github.com/aretext/aretext/client"
)

func main() {
	logFile, err := os.Create("client.log")
	if err != nil {
		fmt.Fprintf(os.Stderr, "error creating log file: %s", err)
		os.Exit(1)
	}
	defer logFile.Close()
	log.SetOutput(logFile)

	config := client.Config{
		ServerSocketPath: filepath.Join(xdg.RuntimeDir, "aretext.socket"),
	}
	err = client.NewClient(config).Run("test.txt")
	if err != nil {
		fmt.Fprintf(os.Stderr, "error running client: %s", err)
		os.Exit(1)
	}
}
