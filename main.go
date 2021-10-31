package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"runtime/pprof"

	"github.com/gdamore/tcell/v2"

	"github.com/aretext/aretext/app"
)

// These variables are set automatically as part of the release process.
// Please do NOT modify the following lines
var (
	version = "dev"
	commit  = ""
)

var logpath = flag.String("log", "", "log to file")
var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")
var editconfig = flag.Bool("editconfig", false, "open the aretext configuration file")
var noconfig = flag.Bool("noconfig", false, "force default configuration")
var versionFlag = flag.Bool("version", false, "print version")

func main() {
	flag.Usage = printUsage
	flag.Parse()

	if *versionFlag {
		fmt.Printf("%s @ %s\n", version, commit)
		return
	}

	log.SetFlags(log.Ltime | log.Lmicroseconds | log.Llongfile)
	if *logpath != "" {
		logFile, err := os.Create(*logpath)
		if err != nil {
			exitWithError(err)
		}
		defer logFile.Close()
		log.SetOutput(logFile)
	} else {
		log.SetOutput(io.Discard)
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
	if *editconfig {
		configPath, err := app.ConfigPath()
		if err != nil {
			exitWithError(err)
		}
		path = configPath
	}

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
	log.Printf("aretext (version: %s, commit: %s)\n", version, commit)
	log.Printf("path arg: '%s'\n", path)
	log.Printf("$TERM env var: '%s'\n", os.Getenv("TERM"))

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
