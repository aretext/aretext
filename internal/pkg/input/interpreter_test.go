package input

import (
	"testing"

	"github.com/gdamore/tcell"
	"github.com/stretchr/testify/assert"
)

func TestInterpreter(t *testing.T) {
	testCases := []struct {
		name             string
		inputEvents      []tcell.Event
		expectedCommands []string
		config           Config
	}{
		{
			name: "resize",
			inputEvents: []tcell.Event{
				tcell.NewEventResize(123, 456),
			},
			expectedCommands: []string{
				"Composite(Resize(123,456),ScrollToCursor())",
			},
		},
		{
			name: "move cursor left using arrow key",
			inputEvents: []tcell.Event{
				tcell.NewEventKey(tcell.KeyLeft, '\x00', tcell.ModNone),
			},
			expectedCommands: []string{"Composite(MutateCursor(CharInLineLocator(backward, 1, false)),ScrollToCursor())"},
		},
		{
			name: "move cursor right using arrow key",
			inputEvents: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRight, '\x00', tcell.ModNone),
			},
			expectedCommands: []string{"Composite(MutateCursor(CharInLineLocator(forward, 1, false)),ScrollToCursor())"},
		},
		{
			name: "move cursor up using arrow key",
			inputEvents: []tcell.Event{
				tcell.NewEventKey(tcell.KeyUp, '\x00', tcell.ModNone),
			},
			expectedCommands: []string{"Composite(MutateCursor(RelativeLineLocator(backward, 1)),ScrollToCursor())"},
		},
		{
			name: "move cursor down using arrow key",
			inputEvents: []tcell.Event{
				tcell.NewEventKey(tcell.KeyDown, '\x00', tcell.ModNone),
			},
			expectedCommands: []string{"Composite(MutateCursor(RelativeLineLocator(forward, 1)),ScrollToCursor())"},
		},
		{
			name: "move cursor left using 'h' key",
			inputEvents: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, 'h', tcell.ModNone),
			},
			expectedCommands: []string{"Composite(MutateCursor(CharInLineLocator(backward, 1, false)),ScrollToCursor())"},
		},
		{
			name: "move cursor right using 'l' key",
			inputEvents: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
			},
			expectedCommands: []string{"Composite(MutateCursor(CharInLineLocator(forward, 1, false)),ScrollToCursor())"},
		},
		{
			name: "move cursor up using 'k' key",
			inputEvents: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, 'k', tcell.ModNone),
			},
			expectedCommands: []string{"Composite(MutateCursor(RelativeLineLocator(backward, 1)),ScrollToCursor())"},
		},
		{
			name: "move cursor down using 'j' key",
			inputEvents: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, 'j', tcell.ModNone),
			},
			expectedCommands: []string{"Composite(MutateCursor(RelativeLineLocator(forward, 1)),ScrollToCursor())"},
		},
		{
			name: "scroll up using ctrl-u, scroll lines zero",
			inputEvents: []tcell.Event{
				tcell.NewEventKey(tcell.KeyCtrlU, '\x00', tcell.ModNone),
			},
			config:           Config{ScrollLines: 0},
			expectedCommands: []string{"Composite(Composite(MutateCursor(RelativeLineStartLocator(backward, 1)),ScrollLines(backward, 1)),ScrollToCursor())"},
		},
		{
			name: "scroll up using ctrl-u, scroll lines greater than zero",
			inputEvents: []tcell.Event{
				tcell.NewEventKey(tcell.KeyCtrlU, '\x00', tcell.ModNone),
			},
			config:           Config{ScrollLines: 25},
			expectedCommands: []string{"Composite(Composite(MutateCursor(RelativeLineStartLocator(backward, 25)),ScrollLines(backward, 25)),ScrollToCursor())"},
		},
		{
			name: "scroll down using ctrl-d, scroll lines zero",
			inputEvents: []tcell.Event{
				tcell.NewEventKey(tcell.KeyCtrlD, '\x00', tcell.ModNone),
			},
			config:           Config{ScrollLines: 0},
			expectedCommands: []string{"Composite(Composite(MutateCursor(RelativeLineStartLocator(forward, 1)),ScrollLines(forward, 1)),ScrollToCursor())"},
		},
		{
			name: "scroll down using ctrl-d, scroll lines greater than zero",
			inputEvents: []tcell.Event{
				tcell.NewEventKey(tcell.KeyCtrlD, '\x00', tcell.ModNone),
			},
			config:           Config{ScrollLines: 25},
			expectedCommands: []string{"Composite(Composite(MutateCursor(RelativeLineStartLocator(forward, 25)),ScrollLines(forward, 25)),ScrollToCursor())"},
		},
		{
			name: "move cursor to end of line using '$' key",
			inputEvents: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, '$', tcell.ModNone),
			},
			expectedCommands: []string{"Composite(MutateCursor(LineBoundaryLocator(forward, false)),ScrollToCursor())"},
		},
		{
			name: "move cursor to start of line using '0' key",
			inputEvents: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, '0', tcell.ModNone),
			},
			expectedCommands: []string{"Composite(MutateCursor(LineBoundaryLocator(backward, false)),ScrollToCursor())"},
		},
		{
			name: "move cursor to start of line using '^' key",
			inputEvents: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, '^', tcell.ModNone),
			},
			expectedCommands: []string{"Composite(Composite(MutateCursor(LineBoundaryLocator(backward, false)),MutateCursor(NonWhitespaceOrNewlineLocator())),ScrollToCursor())"},
		},
		{
			name: "move cursor to start of first line using 'gg'",
			inputEvents: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, 'g', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'g', tcell.ModNone),
			},
			expectedCommands: []string{"", "Composite(Composite(MutateCursor(LineNumLocator(0)),MutateCursor(NonWhitespaceOrNewlineLocator())),ScrollToCursor())"},
		},
		{
			name: "move cursor to single-digit line number using 'g'",
			inputEvents: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, '5', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'g', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'g', tcell.ModNone),
			},
			expectedCommands: []string{"", "", "Composite(Composite(MutateCursor(LineNumLocator(4)),MutateCursor(NonWhitespaceOrNewlineLocator())),ScrollToCursor())"},
		},
		{
			name: "move cursor to double-digit line number using 'g'",
			inputEvents: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, '1', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, '0', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'g', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'g', tcell.ModNone),
			},
			expectedCommands: []string{"", "", "", "Composite(Composite(MutateCursor(LineNumLocator(9)),MutateCursor(NonWhitespaceOrNewlineLocator())),ScrollToCursor())"},
		},
		{
			name: "move cursor to start of last line using 'G' key",
			inputEvents: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, 'G', tcell.ModNone),
			},
			expectedCommands: []string{"Composite(Composite(MutateCursor(LastLineLocator()),MutateCursor(NonWhitespaceOrNewlineLocator())),ScrollToCursor())"},
		},
		{
			name: "insert and return to normal mode",
			inputEvents: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, 'i', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'a', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'b', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyEscape, '\x00', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
			},
			expectedCommands: []string{
				"",
				"Composite(InsertRune('a'),ScrollToCursor())",
				"Composite(InsertRune('b'),ScrollToCursor())",
				"Composite(MutateCursor(OntoLineLocator()),ScrollToCursor())",
				"Composite(MutateCursor(CharInLineLocator(forward, 1, false)),ScrollToCursor())",
			},
		},
		{
			name: "return to normal mode with escape input as a modifier",
			inputEvents: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, 'i', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'a', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'b', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModCtrl),
			},
			expectedCommands: []string{
				"",
				"Composite(InsertRune('a'),ScrollToCursor())",
				"Composite(InsertRune('b'),ScrollToCursor())",
				"Composite(Composite(MutateCursor(OntoLineLocator()),ScrollToCursor()),Composite(MutateCursor(CharInLineLocator(forward, 1, false)),ScrollToCursor()))",
			},
		},
		{
			name: "insert at beginning of line and return to normal mode",
			inputEvents: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, 'I', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'a', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'b', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyEscape, '\x00', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
			},
			expectedCommands: []string{
				"Composite(Composite(MutateCursor(LineBoundaryLocator(backward, false)),MutateCursor(NonWhitespaceOrNewlineLocator())),ScrollToCursor())",
				"Composite(InsertRune('a'),ScrollToCursor())",
				"Composite(InsertRune('b'),ScrollToCursor())",
				"Composite(MutateCursor(OntoLineLocator()),ScrollToCursor())",
				"Composite(MutateCursor(CharInLineLocator(forward, 1, false)),ScrollToCursor())",
			},
		},
		{
			name: "append and return to normal mode",
			inputEvents: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, 'a', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, '1', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, '2', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyEscape, '\x00', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
			},
			expectedCommands: []string{
				"Composite(MutateCursor(CharInLineLocator(forward, 1, true)),ScrollToCursor())",
				"Composite(InsertRune('1'),ScrollToCursor())",
				"Composite(InsertRune('2'),ScrollToCursor())",
				"Composite(MutateCursor(OntoLineLocator()),ScrollToCursor())",
				"Composite(MutateCursor(CharInLineLocator(forward, 1, false)),ScrollToCursor())",
			},
		},
		{
			name: "append to end of line and return to normal mode",
			inputEvents: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, 'A', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, '1', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, '2', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyEscape, '\x00', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
			},
			expectedCommands: []string{
				"Composite(MutateCursor(LineBoundaryLocator(forward, true)),ScrollToCursor())",
				"Composite(InsertRune('1'),ScrollToCursor())",
				"Composite(InsertRune('2'),ScrollToCursor())",
				"Composite(MutateCursor(OntoLineLocator()),ScrollToCursor())",
				"Composite(MutateCursor(CharInLineLocator(forward, 1, false)),ScrollToCursor())",
			},
		},
		{
			name: "delete character using 'x' key",
			inputEvents: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, 'x', tcell.ModNone),
			},
			expectedCommands: []string{
				"Composite(Composite(Delete(CharInLineLocator(forward, 1, true)),MutateCursor(OntoLineLocator())),ScrollToCursor())",
			},
		},
		{
			name: "delete character using backspace key in insert mode",
			inputEvents: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, 'i', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyBackspace, '\x00', tcell.ModNone),
			},
			expectedCommands: []string{"", "Composite(Delete(CharInLineLocator(backward, 1, true)),ScrollToCursor())"},
		},
		{
			name: "delete character using backspace2 key in insert mode",
			inputEvents: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, 'i', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyBackspace2, '\x00', tcell.ModNone),
			},
			expectedCommands: []string{"", "Composite(Delete(CharInLineLocator(backward, 1, true)),ScrollToCursor())"},
		},
		{
			name: "insert newline",
			inputEvents: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, 'i', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyEnter, '\x00', tcell.ModNone),
			},
			expectedCommands: []string{"", "Composite(InsertRune('\\n'),ScrollToCursor())"},
		},
		{
			name: "insert tab",
			inputEvents: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, 'i', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyTab, '\x00', tcell.ModNone),
			},
			expectedCommands: []string{"", "Composite(InsertRune('\\t'),ScrollToCursor())"},
		},
		{
			name: "open and close menu",
			inputEvents: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, ':', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyEscape, ' ', tcell.ModNone),
			},
			expectedCommands: []string{
				"Composite(ShowMenu(command),ScrollToCursor())",
				"HideMenu()",
			},
		},
		{
			name: "enter menu search text",
			inputEvents: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, ':', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'a', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'b', tcell.ModNone),
			},
			expectedCommands: []string{
				"Composite(ShowMenu(command),ScrollToCursor())",
				"AppendMenuSearch('a')",
				"AppendMenuSearch('b')",
			},
		},
		{
			name: "delete menu search text",
			inputEvents: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, ':', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyBackspace, ' ', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyBackspace2, ' ', tcell.ModNone),
			},
			expectedCommands: []string{
				"Composite(ShowMenu(command),ScrollToCursor())",
				"DeleteMenuSearch()",
				"DeleteMenuSearch()",
			},
		},
		{
			name: "move menu selection up using arrow key",
			inputEvents: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, ':', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyUp, ' ', tcell.ModNone),
			},
			expectedCommands: []string{
				"Composite(ShowMenu(command),ScrollToCursor())",
				"MoveMenuSelection(-1)",
			},
		},
		{
			name: "move menu selection down using arrow key",
			inputEvents: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, ':', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyDown, ' ', tcell.ModNone),
			},
			expectedCommands: []string{
				"Composite(ShowMenu(command),ScrollToCursor())",
				"MoveMenuSelection(1)",
			},
		},
		{
			name: "move menu selection down using tab",
			inputEvents: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, ':', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyTab, ' ', tcell.ModNone),
			},
			expectedCommands: []string{
				"Composite(ShowMenu(command),ScrollToCursor())",
				"MoveMenuSelection(1)",
			},
		},
		{
			name: "execute selected menu item",
			inputEvents: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, ':', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyEnter, ' ', tcell.ModNone),
			},
			expectedCommands: []string{
				"Composite(ShowMenu(command),ScrollToCursor())",
				"ExecuteSelectedMenuItem()",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			interpreter := NewInterpreter()
			commands := make([]string, 0, len(tc.inputEvents))
			for _, event := range tc.inputEvents {
				cmd := interpreter.ProcessEvent(event, tc.config)
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
