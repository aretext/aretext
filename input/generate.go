//go:build ignore

package main

import (
	"fmt"
	"os"

	"github.com/aretext/aretext/input"
	"github.com/aretext/aretext/input/vm"
)

// This generates virtual machine programs for each of the editor input modes.
// These are compiled, written to disk, then embedded in the aretext binary
// so they can be quickly loaded on program startup.
func main() {
	generateProgram(input.NormalModeProgramPath, input.NormalModeCommands())
	generateProgram(input.InsertModeProgramPath, input.InsertModeCommands())
	generateProgram(input.VisualModeProgramPath, input.VisualModeCommands())
	generateProgram(input.MenuModeProgramPath, input.MenuModeCommands())
	generateProgram(input.SearchModeProgramPath, input.SearchModeCommands())
	generateProgram(input.TaskModeProgramPath, input.TaskModeCommands())
}

func generateProgram(path string, commands []input.Command) {
	fmt.Printf("Generating input program %s\n", path)
	program := compileProgram(commands)
	data := vm.SerializeProgram(program)
	if err := os.WriteFile(path, data, 0644); err != nil {
		fmt.Printf("Error generating %s: %s", path, err)
		os.Exit(1)
	}
}

func compileProgram(commands []input.Command) vm.Program {
	// Build a single expression to recognize any of the commands for this mode.
	// Wrap each command expression in CaptureExpr so we can determine which command
	// was accepted by the virtual machine.
	var expr vm.AltExpr
	for i, c := range commands {
		expr.Children = append(expr.Children, vm.CaptureExpr{
			CaptureId: vm.CaptureId(i),
			Child:     c.BuildExpr(),
		})
	}
	return vm.MustCompile(expr)
}
