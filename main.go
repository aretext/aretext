package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"runtime/pprof"

	"github.com/aretext/aretext/app"
	"github.com/gdamore/tcell"
)

var logpath = flag.String("log", "", "log to file")
var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")
var noconfig = flag.Bool("noconfig", false, "use default configuration instead of loading it from $HOME/.config/aretext")

func main() {
	flag.Usage = printUsage
	flag.Parse()

	log.SetFlags(log.Ltime | log.Lmicroseconds | log.Llongfile)
	if *logpath != "" {
		logFile, err := os.Create(*logpath)
		if err != nil {
			exitWithError(err)
		}
		defer logFile.Close()
		log.SetOutput(logFile)
	} else {
		log.SetOutput(ioutil.Discard)
	}

	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			exitWithError(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	path := flag.Arg(0)
	err := runEditor(path)
	if err != nil {
		exitWithError(err)
	}
}

func printUsage() {
	f := flag.CommandLine.Output()
	fmt.Fprintf(f, "Usage: %s [options...] [path]\n", os.Args[0])
	flag.PrintDefaults()
}

func runEditor(path string) error {
	configRuleSet, err := app.LoadOrCreateConfig(*noconfig)
	if err != nil {
		return err
	}

	screen, err := tcell.NewScreen()
	if err != nil {
		return err
	}

	if err := screen.Init(); err != nil {
		return err
	}
	defer screen.Fini()

	editor := app.NewEditor(screen, path, configRuleSet)
	editor.RunEventLoop()
	return nil
}

func exitWithError(err error) {
	fmt.Fprintf(os.Stderr, "%v\n", err)
	os.Exit(1)
}
