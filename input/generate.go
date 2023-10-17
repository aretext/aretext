//go:build ignore

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"github.com/gdamore/tcell/v2"

	"github.com/aretext/aretext/input"
	"github.com/aretext/aretext/input/engine"
)

// This generates state machines for each of the editor input modes.
// These are compiled, written to disk, then embedded in the aretext binary
// so they can be quickly loaded on program startup.
func main() {
	generate(input.NormalModePath, input.NormalModeCommands())
	generate(input.InsertModePath, input.InsertModeCommands())
	generate(input.VisualModePath, input.VisualModeCommands())
	generate(input.MenuModePath, input.MenuModeCommands())
	generate(input.SearchModePath, input.SearchModeCommands())
	generate(input.TaskModePath, input.TaskModeCommands())
}

func generate(path string, commands []input.Command) {
	fmt.Printf("Generating input state machine %s\n", path)

	cmdExprs := make([]engine.CmdExpr, 0, len(commands))
	for i, cmd := range commands {
		cmdExprs = append(cmdExprs, engine.CmdExpr{
			CmdId: engine.CmdId(i),
			Expr:  cmd.BuildExpr(),
		})
	}

	sm, err := engine.Compile(cmdExprs)
	if err != nil {
		fmt.Printf("Error compiling commands for %s: %s", path, err)
		os.Exit(1)
	}

	data := engine.Serialize(sm)
	if err := os.WriteFile(path, data, 0644); err != nil {
		fmt.Printf("Error writing file %s: %s", path, err)
		os.Exit(1)
	}

	// Hack: serialize state machine to JSON.
	jsonPath := fmt.Sprintf("%s.json", strings.TrimSuffix(path, filepath.Ext(path)))
	fmt.Printf("Generating %s\n", jsonPath)
	enricher := &JsonEnricher{
		Commands: commands,
	}
	jsonData, err := engine.SerializeJson(sm, enricher)
	if err != nil {
		fmt.Printf("Error serializing json %s: %s", jsonPath, err)
		os.Exit(1)
	}
	if err := os.WriteFile(jsonPath, jsonData, 0644); err != nil {
		fmt.Printf("Error writing file %s: %s", jsonPath, err)
		os.Exit(1)
	}
}

type JsonEnricher struct {
	Commands []input.Command
}

func (je *JsonEnricher) NameForCmd(cmdId engine.CmdId) string {
	cmd := je.Commands[int(cmdId)]
	return cmd.Name
}

func (je *JsonEnricher) NameForCapture(captureId engine.CaptureId) string {
	base := engine.CaptureId(0)
	switch captureId {
	case base:
		return "verbCount"
	case base + 1:
		return "objectCount"
	case base + 2:
		return "clipboardPage"
	case base + 3:
		return "matchChar"
	case base + 4:
		return "replaceChar"
	case base + 5:
		return "insertChar"
	default:
		return ""
	}
}

type jsonEvent struct {
	DisplayName    string
	KeyName        string
	StartCodepoint rune
	EndCodepoint   rune
}

func (je *JsonEnricher) Event(start, end engine.Event) interface{} {
	startKey := engineEventToKey(start)
	startRune := engineEventToRune(start)
	endRune := engineEventToRune(end)

	event := jsonEvent{
		StartCodepoint: startRune,
		EndCodepoint:   endRune,
	}

	if startKey != tcell.KeyRune {
		keyName := tcell.KeyNames[startKey]
		event.DisplayName = keyName
		event.KeyName = keyName
	} else if startRune == endRune {
		event.DisplayName = fmt.Sprintf("%c", startRune)
	} else if startRune == 0 && endRune == 255 {
		event.DisplayName = "Any ASCII"
	} else if startRune == 0 && endRune == utf8.MaxRune {
		event.DisplayName = "Any unicode"
	} else {
		event.DisplayName = fmt.Sprintf("%c-%c", startRune, endRune)
	}

	return event
}

func engineEventToKey(engineEvent engine.Event) tcell.Key {
	return tcell.Key(engineEvent >> 32)
}

func engineEventToRune(engineEvent engine.Event) rune {
	return rune(engineEvent & 0xFFFFFFFF)
}
