package main

import (
	"fmt"
	"os"

	"github.com/gdamore/tcell"
	"github.com/wedaly/aretext/internal/app/aretext"
)

func main() {
	screen, err := tcell.NewScreen()
	if err != nil {
		exitWithError(err)
	}

	if err := screen.Init(); err != nil {
		exitWithError(err)
	} else {
		defer screen.Fini()
	}

	editor := aretext.NewEditor(screen)
	editor.RunEventLoop()
}

func exitWithError(err error) {
	fmt.Fprintf(os.Stderr, "%v\n", err)
	os.Exit(1)
}
