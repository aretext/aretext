package main

import (
	"fmt"
	"os"

	"github.com/aretext/aretext/client"
)

func main() {
	config := client.Config{
		ServerSocketPath: "/var/run/aretext.socket",
	}
	err := client.NewClient(config).Run("test.txt")
	if err != nil {
		fmt.Fprintf(os.Stderr, "error running client: %s", err)
		os.Exit(1)
	}
}
