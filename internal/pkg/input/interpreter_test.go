package input

import (
	"testing"

	"github.com/gdamore/tcell"
	"github.com/stretchr/testify/assert"
)

func TestInterpreter(t *testing.T) {
	testCases := []struct {
		name             string
		inputEvents      []*tcell.EventKey
		expectedCommands []string
	}{
		{
			name: "quit using ctrl-c in normal mode",
			inputEvents: []*tcell.EventKey{
				tcell.NewEventKey(tcell.KeyCtrlC, '\x00', tcell.ModNone),
			},
			expectedCommands: []string{"Quit()"},
		},
		{
			name: "quit using ctrl-c in insert mode",
			inputEvents: []*tcell.EventKey{
				tcell.NewEventKey(tcell.KeyRune, 'i', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyCtrlC, '\x00', tcell.ModNone),
			},
			expectedCommands: []string{"", "Quit()"},
		},
		{
			name: "move cursor left using arrow key",
			inputEvents: []*tcell.EventKey{
				tcell.NewEventKey(tcell.KeyLeft, '\x00', tcell.ModNone),
			},
			expectedCommands: []string{"Exec(MutateCursor(CharInLineLocator(backward, 1, false)))"},
		},
		{
			name: "move cursor right using arrow key",
			inputEvents: []*tcell.EventKey{
				tcell.NewEventKey(tcell.KeyRight, '\x00', tcell.ModNone),
			},
			expectedCommands: []string{"Exec(MutateCursor(CharInLineLocator(forward, 1, false)))"},
		},
		{
			name: "move cursor up using arrow key",
			inputEvents: []*tcell.EventKey{
				tcell.NewEventKey(tcell.KeyUp, '\x00', tcell.ModNone),
			},
			expectedCommands: []string{"Exec(MutateCursor(RelativeLineLocator(backward, 1)))"},
		},
		{
			name: "move cursor down using arrow key",
			inputEvents: []*tcell.EventKey{
				tcell.NewEventKey(tcell.KeyDown, '\x00', tcell.ModNone),
			},
			expectedCommands: []string{"Exec(MutateCursor(RelativeLineLocator(forward, 1)))"},
		},
		{
			name: "move cursor left using 'h' key",
			inputEvents: []*tcell.EventKey{
				tcell.NewEventKey(tcell.KeyRune, 'h', tcell.ModNone),
			},
			expectedCommands: []string{"Exec(MutateCursor(CharInLineLocator(backward, 1, false)))"},
		},
		{
			name: "move cursor right using 'l' key",
			inputEvents: []*tcell.EventKey{
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
			},
			expectedCommands: []string{"Exec(MutateCursor(CharInLineLocator(forward, 1, false)))"},
		},
		{
			name: "move cursor up using 'k' key",
			inputEvents: []*tcell.EventKey{
				tcell.NewEventKey(tcell.KeyRune, 'k', tcell.ModNone),
			},
			expectedCommands: []string{"Exec(MutateCursor(RelativeLineLocator(backward, 1)))"},
		},
		{
			name: "move cursor down using 'j' key",
			inputEvents: []*tcell.EventKey{
				tcell.NewEventKey(tcell.KeyRune, 'j', tcell.ModNone),
			},
			expectedCommands: []string{"Exec(MutateCursor(RelativeLineLocator(forward, 1)))"},
		},
		{
			name: "move cursor to end of line using '$' key",
			inputEvents: []*tcell.EventKey{
				tcell.NewEventKey(tcell.KeyRune, '$', tcell.ModNone),
			},
			expectedCommands: []string{"Exec(MutateCursor(LineBoundaryLocator(forward, false)))"},
		},
		{
			name: "move cursor to start of line using '0' key",
			inputEvents: []*tcell.EventKey{
				tcell.NewEventKey(tcell.KeyRune, '0', tcell.ModNone),
			},
			expectedCommands: []string{"Exec(MutateCursor(LineBoundaryLocator(backward, false)))"},
		},
		{
			name: "move cursor to start of line using '^' key",
			inputEvents: []*tcell.EventKey{
				tcell.NewEventKey(tcell.KeyRune, '^', tcell.ModNone),
			},
			expectedCommands: []string{"Exec(Composite(MutateCursor(LineBoundaryLocator(backward, false)),MutateCursor(NonWhitespaceLocator(forward))))"},
		},
		{
			name: "move cursor to start of last line using 'G' key",
			inputEvents: []*tcell.EventKey{
				tcell.NewEventKey(tcell.KeyRune, 'G', tcell.ModNone),
			},
			expectedCommands: []string{"Exec(Composite(MutateCursor(LastLineLocator()),MutateCursor(NonWhitespaceLocator(forward))))"},
		},
		{
			name: "insert and return to normal mode",
			inputEvents: []*tcell.EventKey{
				tcell.NewEventKey(tcell.KeyRune, 'i', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'a', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'b', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyEscape, '\x00', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
			},
			expectedCommands: []string{
				"",
				"Exec(InsertRune('a'))",
				"Exec(InsertRune('b'))",
				"Exec(MutateCursor(OntoLineLocator()))",
				"Exec(MutateCursor(CharInLineLocator(forward, 1, false)))",
			},
		},
		{
			name: "insert at beginning of line and return to normal mode",
			inputEvents: []*tcell.EventKey{
				tcell.NewEventKey(tcell.KeyRune, 'I', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'a', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'b', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyEscape, '\x00', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
			},
			expectedCommands: []string{
				"Exec(Composite(MutateCursor(LineBoundaryLocator(backward, false)),MutateCursor(NonWhitespaceLocator(forward))))",
				"Exec(InsertRune('a'))",
				"Exec(InsertRune('b'))",
				"Exec(MutateCursor(OntoLineLocator()))",
				"Exec(MutateCursor(CharInLineLocator(forward, 1, false)))",
			},
		},
		{
			name: "append and return to normal mode",
			inputEvents: []*tcell.EventKey{
				tcell.NewEventKey(tcell.KeyRune, 'a', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, '1', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, '2', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyEscape, '\x00', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
			},
			expectedCommands: []string{
				"Exec(MutateCursor(CharInLineLocator(forward, 1, true)))",
				"Exec(InsertRune('1'))",
				"Exec(InsertRune('2'))",
				"Exec(MutateCursor(OntoLineLocator()))",
				"Exec(MutateCursor(CharInLineLocator(forward, 1, false)))",
			},
		},
		{
			name: "append to end of line and return to normal mode",
			inputEvents: []*tcell.EventKey{
				tcell.NewEventKey(tcell.KeyRune, 'A', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, '1', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, '2', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyEscape, '\x00', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
			},
			expectedCommands: []string{
				"Exec(MutateCursor(LineBoundaryLocator(forward, true)))",
				"Exec(InsertRune('1'))",
				"Exec(InsertRune('2'))",
				"Exec(MutateCursor(OntoLineLocator()))",
				"Exec(MutateCursor(CharInLineLocator(forward, 1, false)))",
			},
		},
		{
			name: "delete character using 'x' key",
			inputEvents: []*tcell.EventKey{
				tcell.NewEventKey(tcell.KeyRune, 'x', tcell.ModNone),
			},
			expectedCommands: []string{
				"Exec(Composite(Delete(CharInLineLocator(forward, 1, true)),MutateCursor(OntoLineLocator())))",
			},
		},
		{
			name: "delete character using backspace key in insert mode",
			inputEvents: []*tcell.EventKey{
				tcell.NewEventKey(tcell.KeyRune, 'i', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyBackspace, '\x00', tcell.ModNone),
			},
			expectedCommands: []string{"", "Exec(Delete(CharInLineLocator(backward, 1, true)))"},
		},
		{
			name: "delete character using backspace2 key in insert mode",
			inputEvents: []*tcell.EventKey{
				tcell.NewEventKey(tcell.KeyRune, 'i', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyBackspace2, '\x00', tcell.ModNone),
			},
			expectedCommands: []string{"", "Exec(Delete(CharInLineLocator(backward, 1, true)))"},
		},
		{
			name: "insert newline",
			inputEvents: []*tcell.EventKey{
				tcell.NewEventKey(tcell.KeyRune, 'i', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyEnter, '\x00', tcell.ModNone),
			},
			expectedCommands: []string{"", "Exec(InsertRune('\\n'))"},
		},
		{
			name: "insert tab",
			inputEvents: []*tcell.EventKey{
				tcell.NewEventKey(tcell.KeyRune, 'i', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyTab, '\x00', tcell.ModNone),
			},
			expectedCommands: []string{"", "Exec(InsertRune('\\t'))"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			interpreter := NewInterpreter()
			commands := make([]string, 0, len(tc.inputEvents))
			for _, event := range tc.inputEvents {
				cmd := interpreter.ProcessKeyEvent(event)
				if cmd == nil {
					commands = append(commands, "")
				} else {
					commands = append(commands, cmd.String())
				}
			}
			assert.Equal(t, tc.expectedCommands, commands)
		})
	}
}
