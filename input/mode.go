package input

import (
	"log"
	"strings"

	"github.com/gdamore/tcell/v2"

	"github.com/aretext/aretext/input/vm"
	"github.com/aretext/aretext/state"
)

// Mode represents an input mode, which is a way of interpreting key events.
type Mode interface {
	// ProcessKeyEvent interprets the key event according to this mode.
	// It will return any user-initiated action resulting from the keypress
	ProcessKeyEvent(event *tcell.EventKey, config Config) Action

	// InputBufferString returns a string describing buffered input events.
	// It can be displayed to the user to help them understand the input state.
	InputBufferString() string
}

// vmMode is a mode that uses a virtual machine to interpret input.
// This is used to implement normal and visual modes.
type vmMode struct {
	name        string
	runtime     *vm.Runtime
	eventBuffer []vm.Event
	inputBuffer strings.Builder
	commands    []Command
}

func newVmMode(name string, commands []Command) *vmMode {
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

	runtime := vm.NewRuntime(vm.MustCompile(expr))

	return &vmMode{
		name:     name,
		runtime:  runtime,
		commands: commands,
	}
}

func (m *vmMode) ProcessKeyEvent(event *tcell.EventKey, config Config) Action {
	vmEvent := eventKeyToVmEvent(event)
	m.eventBuffer = append(m.eventBuffer, vmEvent)
	if event.Key() == tcell.KeyRune {
		m.inputBuffer.WriteRune(event.Rune())
	}

	action := EmptyAction
	result := m.runtime.ProcessEvent(vmEvent)
	if result.Accepted {
		for _, capture := range result.Captures {
			if int(capture.Id) < len(m.commands) {
				command := m.commands[capture.Id]
				params := capturesToCommandParams(result.Captures, m.eventBuffer)
				action = command.BuildAction(config, params)
				log.Printf(
					"%s mode accepted input for command %q with params %+v and config %+v\n",
					m.name, command.Name,
					params, config,
				)
				break
			}
		}
	}

	if result.Reset {
		m.eventBuffer = m.eventBuffer[:0]
		m.inputBuffer.Reset()
	}

	return action
}

func (m *vmMode) InputBufferString() string {
	return m.inputBuffer.String()
}

// menuMode allows the user to search for and select items in a menu.
type menuMode struct{}

func (m *menuMode) ProcessKeyEvent(event *tcell.EventKey, config Config) Action {
	switch event.Key() {
	case tcell.KeyEscape:
		return HideMenuAndReturnToNormalMode
	case tcell.KeyEnter:
		return ExecuteSelectedMenuItem
	case tcell.KeyUp:
		return MenuSelectionUp
	case tcell.KeyDown:
		return MenuSelectionDown
	case tcell.KeyTab:
		return MenuSelectionDown
	case tcell.KeyRune:
		return AppendRuneToMenuSearch(event.Rune())
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		return DeleteRuneFromMenuSearch
	default:
		return EmptyAction
	}
}

func (m *menuMode) InputBufferString() string {
	return ""
}

// searchMode is used to search the text for a substring.
type searchMode struct{}

func (m *searchMode) ProcessKeyEvent(event *tcell.EventKey, config Config) Action {
	action := m.processKeyEvent(event)
	return func(s *state.EditorState) {
		action(s)
		state.AddToRecordingUserMacro(s, state.MacroAction(action))
	}
}

func (m *searchMode) processKeyEvent(event *tcell.EventKey) Action {
	switch event.Key() {
	case tcell.KeyEscape:
		return AbortSearchAndReturnToNormalMode
	case tcell.KeyEnter:
		return CommitSearchAndReturnToNormalMode
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		// This returns the input mode to normal if the search query is empty.
		return DeleteRuneFromSearchQuery
	case tcell.KeyRune:
		return AppendRuneToSearchQuery(event.Rune())
	default:
		return EmptyAction
	}
}

func (m *searchMode) InputBufferString() string {
	return ""
}

// taskMode is used while a task is running asynchronously.
// This allows the user to cancel the task if it takes too long.
type taskMode struct{}

func (m *taskMode) ProcessKeyEvent(event *tcell.EventKey, config Config) Action {
	switch event.Key() {
	case tcell.KeyEscape:
		return state.CancelTaskIfRunning
	default:
		return EmptyAction
	}
}

func (m *taskMode) InputBufferString() string {
	return ""
}
