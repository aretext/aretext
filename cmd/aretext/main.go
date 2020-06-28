package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/gdamore/tcell"
	"github.com/wedaly/aretext/internal/app/aretext"
)

func main() {
	flag.Usage = printUsage
	flag.Parse()
	if len(flag.Args()) == 0 {
		flag.Usage()
		os.Exit(1)
	}

	screen, err := tcell.NewScreen()
	if err != nil {
		exitWithError(err)
	}

	if err := screen.Init(); err != nil {
		exitWithError(err)
	} else {
		defer screen.Fini()
	}

	path := flag.Arg(0)
	editor, err := aretext.NewEditor(path, screen)
	if err != nil {
		exitWithError(err)
	}

	editor.RunEventLoop()
}

func printUsage() {
	f := flag.CommandLine.Output()
	fmt.Fprintf(f, "Usage: %s [OPTIONS] path\n", os.Args[0])
	flag.PrintDefaults()
}

func exitWithError(err error) {
	fmt.Fprintf(os.Stderr, "%v\n", err)
	os.Exit(1)
}
