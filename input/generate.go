//go:build ignore

package main

import (
	"fmt"
	"os"

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

	// Hack to output graphviz dot file for each state machine.
	dotFilePath := fmt.Sprintf("%s.dot", path)
	fmt.Printf("Generating graphviz dot file for input state machine %s\n", dotFilePath)
	dotData := engine.Render(sm, eventLabelFunc)
	if err := os.WriteFile(dotFilePath, []byte(dotData), 0644); err != nil {
		fmt.Printf("Error writing file %s: %s", dotFilePath, err)
		os.Exit(1)
	}
}

func eventLabelFunc(start, end engine.Event) string {
	return "TODO"
}
