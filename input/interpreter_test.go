package input

import (
	"io"
	"log"
	"os"
	"testing"
	"unicode/utf8"

	"github.com/gdamore/tcell/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/aretext/aretext/input/engine"
	"github.com/aretext/aretext/state"
)

func init() {
	// Suppress noisy log output from these tests.
	log.SetOutput(io.Discard)
}

func TestInterpreterStateIntegration(t *testing.T) {
	testCases := []struct {
		name              string
		initialText       string
		events            []tcell.Event
		expectedCursorPos uint64
		expectedText      string
	}{
		{
			name:        "cursor home row keys",
			initialText: "Lorem ipsum dolor\nsit amet consectetur\nadipiscing elit",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'j', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'j', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'h', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'h', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'k', tcell.ModNone),
			},
			expectedCursorPos: 22,
			expectedText:      "Lorem ipsum dolor\nsit amet consectetur\nadipiscing elit",
		},
		{
			name:        "cursor arrow keys",
			initialText: "Lorem ipsum dolor\nsit amet consectetur\nadipiscing elit",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRight, '\x00', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRight, '\x00', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRight, '\x00', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRight, '\x00', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRight, '\x00', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRight, '\x00', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyDown, '\x00', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyDown, '\x00', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyLeft, '\x00', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyLeft, '\x00', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyUp, '\x00', tcell.ModNone),
			},
			expectedCursorPos: 22,
			expectedText:      "Lorem ipsum dolor\nsit amet consectetur\nadipiscing elit",
		},
		{
			name:        "cursor back",
			initialText: "Lorem ipsum dolor\nsit amet consectetur\nadipiscing elit",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, 'e', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'e', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyBackspace2, '\u007f', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyBackspace2, '\u007f', tcell.ModNone),
			},
			expectedCursorPos: 8,
			expectedText:      "Lorem ipsum dolor\nsit amet consectetur\nadipiscing elit",
		},
		{
			name:        "cursor forward",
			initialText: "Lorem ipsum dolor\nsit amet consectetur\nadipiscing elit",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, ' ', tcell.ModNone),
			},
			expectedCursorPos: 1,
			expectedText:      "Lorem ipsum dolor\nsit amet consectetur\nadipiscing elit",
		},
		{
			name:        "cursor left with count",
			initialText: "Lorem ipsum dolor\nsit amet consectetur\nadipiscing elit",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, '$', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, '5', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'h', tcell.ModNone),
			},
			expectedCursorPos: 11,
			expectedText:      "Lorem ipsum dolor\nsit amet consectetur\nadipiscing elit",
		},
		{
			name:        "cursor right with count",
			initialText: "Lorem ipsum dolor\nsit amet consectetur\nadipiscing elit",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, '5', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
			},
			expectedCursorPos: 5,
			expectedText:      "Lorem ipsum dolor\nsit amet consectetur\nadipiscing elit",
		},
		{
			name:        "cursor down with count",
			initialText: "Lorem ipsum dolor\nsit amet consectetur\nadipiscing elit",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, '2', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'j', tcell.ModNone),
			},
			expectedCursorPos: 39,
			expectedText:      "Lorem ipsum dolor\nsit amet consectetur\nadipiscing elit",
		},
		{
			name:        "cursor up with count",
			initialText: "Lorem ipsum dolor\nsit amet consectetur\nadipiscing elit",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, 'G', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, '2', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'k', tcell.ModNone),
			},
			expectedCursorPos: 0,
			expectedText:      "Lorem ipsum dolor\nsit amet consectetur\nadipiscing elit",
		},
		{
			name:        "cursor start of next line",
			initialText: "Lorem ipsum dolor\n  sit amet consectetur\nadipiscing elit",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyEnter, '\x00', tcell.ModNone),
			},
			expectedCursorPos: 20,
			expectedText:      "Lorem ipsum dolor\n  sit amet consectetur\nadipiscing elit",
		},
		{
			name:        "cursor to next matching in line",
			initialText: "Lorem ipsum dolor\nsit amet consectetur\nadipiscing elit",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, 'f', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'p', tcell.ModNone),
			},
			expectedCursorPos: 7,
			expectedText:      "Lorem ipsum dolor\nsit amet consectetur\nadipiscing elit",
		},
		{
			name:        "cursor to prev matching in line",
			initialText: "Lorem ipsum dolor\nsit amet consectetur\nadipiscing elit",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'j', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'F', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'a', tcell.ModNone),
			},
			expectedCursorPos: 22,
			expectedText:      "Lorem ipsum dolor\nsit amet consectetur\nadipiscing elit",
		},
		{
			name:        "cursor till next matching in line",
			initialText: "Lorem ipsum dolor\nsit amet consectetur\nadipiscing elit",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, 't', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'd', tcell.ModNone),
			},
			expectedCursorPos: 11,
			expectedText:      "Lorem ipsum dolor\nsit amet consectetur\nadipiscing elit",
		},
		{
			name:        "cursor till prev matching in line",
			initialText: "Lorem ipsum dolor\nsit amet consectetur\nadipiscing elit",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, 'j', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'T', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 't', tcell.ModNone),
			},
			expectedCursorPos: 26,
			expectedText:      "Lorem ipsum dolor\nsit amet consectetur\nadipiscing elit",
		},
		{
			name:        "cursor next word start",
			initialText: "Lorem ipsum dolor\nsit amet consectetur\nadipiscing elit",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, 'w', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'w', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'w', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'w', tcell.ModNone),
			},
			expectedCursorPos: 22,
			expectedText:      "Lorem ipsum dolor\nsit amet consectetur\nadipiscing elit",
		},
		{
			name:        "cursor next word start with count",
			initialText: "Lorem ipsum dolor\nsit amet consectetur\nadipiscing elit",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, '5', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'w', tcell.ModNone),
			},
			expectedCursorPos: 27,
			expectedText:      "Lorem ipsum dolor\nsit amet consectetur\nadipiscing elit",
		},
		{
			name:        "cursor next word start including punctuation",
			initialText: "Lorem ipsum,dolor;sit amet consectetur",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, 'W', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'W', tcell.ModNone),
			},
			expectedCursorPos: 22,
			expectedText:      "Lorem ipsum,dolor;sit amet consectetur",
		},
		{
			name:        "cursor next word end of file",
			initialText: "abc",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, 'w', tcell.ModNone),
			},
			expectedCursorPos: 2,
			expectedText:      "abc",
		},
		{
			name:        "cursor next word end of file at newline",
			initialText: "abc\n",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, 'w', tcell.ModNone),
			},
			expectedCursorPos: 4,
			expectedText:      "abc\n",
		},
		{
			name:        "cursor prev word start",
			initialText: "Lorem ipsum dolor\nsit amet consectetur\nadipiscing elit",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, 'j', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'b', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'b', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'b', tcell.ModNone),
			},
			expectedCursorPos: 12,
			expectedText:      "Lorem ipsum dolor\nsit amet consectetur\nadipiscing elit",
		},
		{
			name:        "cursor prev word start including punctuation",
			initialText: "Lorem ipsum,dolor;sit amet consectetur",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, '$', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'B', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'B', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'B', tcell.ModNone),
			},
			expectedCursorPos: 6,
			expectedText:      "Lorem ipsum,dolor;sit amet consectetur",
		},
		{
			name:        "cursor prev word start with count",
			initialText: "Lorem ipsum dolor\nsit amet consectetur\nadipiscing elit",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, 'j', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, '3', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'b', tcell.ModNone),
			},
			expectedCursorPos: 12,
			expectedText:      "Lorem ipsum dolor\nsit amet consectetur\nadipiscing elit",
		},
		{
			name:        "cursor next word end",
			initialText: "Lorem ipsum dolor\nsit amet consectetur\nadipiscing elit",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, 'e', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'e', tcell.ModNone),
			},
			expectedCursorPos: 10,
			expectedText:      "Lorem ipsum dolor\nsit amet consectetur\nadipiscing elit",
		},
		{
			name:        "cursor next word end including punctuation",
			initialText: "Lorem ipsum,dolor;sit amet consectetur",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, 'E', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'E', tcell.ModNone),
			},
			expectedCursorPos: 20,
			expectedText:      "Lorem ipsum,dolor;sit amet consectetur",
		},
		{
			name:        "cursor next word end with count",
			initialText: "Lorem ipsum dolor\nsit amet consectetur\nadipiscing elit",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, '5', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'e', tcell.ModNone),
			},
			expectedCursorPos: 25,
			expectedText:      "Lorem ipsum dolor\nsit amet consectetur\nadipiscing elit",
		},
		{
			name:        "cursor prev paragraph",
			initialText: "Lorem ipsum dolor\n\nsit amet consectetur\nadipiscing\n\nelit\n\n",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, 'j', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'j', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'j', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'j', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'j', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, '{', tcell.ModNone),
			},
			expectedCursorPos: 18,
			expectedText:      "Lorem ipsum dolor\n\nsit amet consectetur\nadipiscing\n\nelit\n\n",
		},
		{
			name:        "cursor next paragraph",
			initialText: "Lorem ipsum dolor\n\nsit amet consectetur\nadipiscing\n\nelit\n\n",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, '}', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, '}', tcell.ModNone),
			},
			expectedCursorPos: 51,
			expectedText:      "Lorem ipsum dolor\n\nsit amet consectetur\nadipiscing\n\nelit\n\n",
		},
		{
			name:        "cursor line start",
			initialText: "Lorem ipsum dolor\n\tsit amet consectetur\n\t\tadipiscing\nelit\n\n",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, 'j', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'w', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'w', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'w', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, '0', tcell.ModNone),
			},
			expectedCursorPos: 18,
			expectedText:      "Lorem ipsum dolor\n\tsit amet consectetur\n\t\tadipiscing\nelit\n\n",
		},
		{
			name:        "cursor line start after indentation",
			initialText: "Lorem ipsum dolor\n\tsit amet consectetur\n\t\tadipiscing\nelit\n\n",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, 'j', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, '^', tcell.ModNone),
			},
			expectedCursorPos: 19,
			expectedText:      "Lorem ipsum dolor\n\tsit amet consectetur\n\t\tadipiscing\nelit\n\n",
		},
		{
			name:        "cursor line end",
			initialText: "Lorem ipsum dolor\nsit amet consectetur\nadipiscing elit",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, '$', tcell.ModNone),
			},
			expectedCursorPos: 16,
			expectedText:      "Lorem ipsum dolor\nsit amet consectetur\nadipiscing elit",
		},
		{
			name:        "cursor start of first line",
			initialText: "Lorem ipsum dolor\nsit amet consectetur\nadipiscing elit",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, 'j', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'j', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'g', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'g', tcell.ModNone),
			},
			expectedCursorPos: 0,
			expectedText:      "Lorem ipsum dolor\nsit amet consectetur\nadipiscing elit",
		},
		{
			name:        "cursor start of line number",
			initialText: "Lorem ipsum dolor\n\nsit amet consectetur\nadipiscing\n\nelit\n\n",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, '4', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'g', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'g', tcell.ModNone),
			},
			expectedCursorPos: 40,
			expectedText:      "Lorem ipsum dolor\n\nsit amet consectetur\nadipiscing\n\nelit\n\n",
		},
		{
			name:        "cursor start of last line",
			initialText: "Lorem ipsum dolor\nsit amet consectetur\nadipiscing elit",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, 'G', tcell.ModNone),
			},
			expectedCursorPos: 39,
			expectedText:      "Lorem ipsum dolor\nsit amet consectetur\nadipiscing elit",
		},
		{
			name:        "cursor matching code block delimiter",
			initialText: `func foo() { fmt.Printf("foo {} bar!") }`,
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, '3', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'w', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, '%', tcell.ModNone),
			},
			expectedCursorPos: 39,
			expectedText:      `func foo() { fmt.Printf("foo {} bar!") }`,
		},
		{
			name:        "cursor prev unmatched open brace",
			initialText: `{ { a { b } c } }`,
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, '6', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'w', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, '[', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, '{', tcell.ModNone),
			},
			expectedCursorPos: 2,
			expectedText:      `{ { a { b } c } }`,
		},
		{
			name:        "cursor next unmatched close brace",
			initialText: `{ { a { b } c } }`,
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, '2', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'w', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, ']', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, '}', tcell.ModNone),
			},
			expectedCursorPos: 14,
			expectedText:      `{ { a { b } c } }`,
		},
		{
			name:        "cursor prev unmatched open paren",
			initialText: `( ( a ( b ) c ) )`,
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, '6', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'w', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, '[', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, '(', tcell.ModNone),
			},
			expectedCursorPos: 2,
			expectedText:      `( ( a ( b ) c ) )`,
		},
		{
			name:        "cursor next unmatched close paren",
			initialText: `( ( a ( b ) c ) )`,
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, '2', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'w', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, ']', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, ')', tcell.ModNone),
			},
			expectedCursorPos: 14,
			expectedText:      `( ( a ( b ) c ) )`,
		},
		{
			name:        "insert",
			initialText: "Lorem ipsum dolor\nsit amet consectetur\nadipiscing elit",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, 'w', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'i', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 't', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'e', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 's', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 't', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, ' ', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyEsc, '\x00', tcell.ModNone),
			},
			expectedCursorPos: 10,
			expectedText:      "Lorem test ipsum dolor\nsit amet consectetur\nadipiscing elit",
		},
		{
			name:        "insert then delete with backspace",
			initialText: "",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, 'i', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'f', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'o', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'o', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'b', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'a', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'r', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyBackspace2, '\u007f', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyBackspace2, '\u007f', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyBackspace2, '\u007f', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyEsc, '\x00', tcell.ModNone),
			},
			expectedCursorPos: 2,
			expectedText:      "foo",
		},
		{
			name:        "insert max rune",
			initialText: "",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, 'i', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, utf8.MaxRune, tcell.ModNone),
			},
			expectedCursorPos: 1,
			expectedText:      "\U0010FFFF",
		},
		{
			name:        "delete with delete key",
			initialText: "foobar baz",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'i', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyDelete, '\x00', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyDelete, '\x00', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyDelete, '\x00', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyEsc, '\x00', tcell.ModNone),
			},
			expectedCursorPos: 0,
			expectedText:      "far baz",
		},
		{
			name:        "insert at start of line",
			initialText: "Lorem ipsum dolor\nsit amet consectetur\nadipiscing elit",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, 'j', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'I', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 't', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'e', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 's', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 't', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, ' ', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyEsc, '\x00', tcell.ModNone),
			},
			expectedCursorPos: 22,
			expectedText:      "Lorem ipsum dolor\ntest sit amet consectetur\nadipiscing elit",
		},
		{
			name:        "insert move cursor right at end of line",
			initialText: "Lorem ipsum dolor\nsit amet consectetur\nadipiscing elit",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, '$', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'i', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRight, '\x00', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyBackspace, '\x00', tcell.ModNone),
			},
			expectedCursorPos: 16,
			expectedText:      "Lorem ipsum dolo\nsit amet consectetur\nadipiscing elit",
		},
		{
			name:        "append",
			initialText: "Lorem ipsum dolor\nsit amet consectetur\nadipiscing elit",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'a', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, ' ', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 't', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'e', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 's', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 't', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyEsc, '\x00', tcell.ModNone),
			},
			expectedCursorPos: 9,
			expectedText:      "Lorem test ipsum dolor\nsit amet consectetur\nadipiscing elit",
		},
		{
			name:        "append at end of line",
			initialText: "Lorem ipsum dolor\nsit amet consectetur\nadipiscing elit",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, 'j', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'A', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, ' ', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 't', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'e', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 's', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 't', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyEsc, '\x00', tcell.ModNone),
			},
			expectedCursorPos: 42,
			expectedText:      "Lorem ipsum dolor\nsit amet consectetur test\nadipiscing elit",
		},
		{
			name:        "new line below",
			initialText: "Lorem ipsum dolor\nsit amet consectetur\nadipiscing elit",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, 'o', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 't', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'e', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 's', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 't', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyEsc, '\x00', tcell.ModNone),
			},
			expectedCursorPos: 21,
			expectedText:      "Lorem ipsum dolor\ntest\nsit amet consectetur\nadipiscing elit",
		},
		{
			name:        "new line above",
			initialText: "Lorem ipsum dolor\nsit amet consectetur\nadipiscing elit",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, 'j', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'O', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 't', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'e', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 's', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 't', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyEsc, '\x00', tcell.ModNone),
			},
			expectedCursorPos: 21,
			expectedText:      "Lorem ipsum dolor\ntest\nsit amet consectetur\nadipiscing elit",
		},
		{
			name:        "join lines",
			initialText: "Lorem ipsum dolor\nsit amet consectetur\nadipiscing elit",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, 'J', tcell.ModNone),
			},
			expectedCursorPos: 17,
			expectedText:      "Lorem ipsum dolor sit amet consectetur\nadipiscing elit",
		},
		{
			name:        "delete next character in line",
			initialText: "Lorem ipsum dolor\nsit amet consectetur\nadipiscing elit",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, 'x', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'x', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'x', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'x', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'x', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'x', tcell.ModNone),
			},
			expectedCursorPos: 0,
			expectedText:      "ipsum dolor\nsit amet consectetur\nadipiscing elit",
		},
		{
			name:        "delete next character in line with count",
			initialText: "Lorem ipsum dolor\nsit amet consectetur\nadipiscing elit",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, '5', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'x', tcell.ModNone),
			},
			expectedCursorPos: 0,
			expectedText:      " ipsum dolor\nsit amet consectetur\nadipiscing elit",
		},
		{
			name:        "delete next character in line using delete key",
			initialText: "Lorem ipsum dolor\nsit amet consectetur\nadipiscing elit",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyDelete, '\x00', tcell.ModNone),
			},
			expectedCursorPos: 0,
			expectedText:      "orem ipsum dolor\nsit amet consectetur\nadipiscing elit",
		},
		{
			name:        "delete next character in line using delete key then paste from default clipboard",
			initialText: "Lorem ipsum dolor\nsit amet consectetur\nadipiscing elit",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyDelete, '\x00', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyDelete, '\x00', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, '$', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'p', tcell.ModNone),
			},
			expectedCursorPos: 15,
			expectedText:      "rem ipsum doloro\nsit amet consectetur\nadipiscing elit",
		},
		{
			name:        "delete line",
			initialText: "Lorem ipsum dolor\nsit amet consectetur\nadipiscing elit",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, 'j', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'd', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'd', tcell.ModNone),
			},
			expectedCursorPos: 18,
			expectedText:      "Lorem ipsum dolor\nadipiscing elit",
		},
		{
			name:        "delete count lines",
			initialText: "Lorem ipsum dolor\nsit amet consectetur\nadipiscing elit",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, '2', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'd', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'd', tcell.ModNone),
			},
			expectedCursorPos: 0,
			expectedText:      "adipiscing elit",
		},
		{
			name:        "delete previous char in line",
			initialText: "Lorem ipsum dolor\nsit amet consectetur\nadipiscing elit",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, 'j', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'd', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'h', tcell.ModNone),
			},
			expectedCursorPos: 22,
			expectedText:      "Lorem ipsum dolor\nsit met consectetur\nadipiscing elit",
		},
		{
			name:        "delete lines below",
			initialText: "Lorem ipsum dolor\nsit amet consectetur\nadipiscing elit",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, 'd', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'j', tcell.ModNone),
			},
			expectedCursorPos: 0,
			expectedText:      "adipiscing elit",
		},
		{
			name:        "delete lines above",
			initialText: "Lorem ipsum dolor\nsit amet consectetur\nadipiscing elit",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, 'j', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'j', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'd', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'k', tcell.ModNone),
			},
			expectedCursorPos: 0,
			expectedText:      "Lorem ipsum dolor",
		},
		{
			name:        "delete next character in line",
			initialText: "Lorem ipsum dolor\nsit amet consectetur\nadipiscing elit",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, 'd', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
			},
			expectedCursorPos: 0,
			expectedText:      "orem ipsum dolor\nsit amet consectetur\nadipiscing elit",
		},
		{
			name:        "delete to end of line",
			initialText: "Lorem ipsum dolor\nsit amet consectetur\nadipiscing elit",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, 'd', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, '$', tcell.ModNone),
			},
			expectedCursorPos: 0,
			expectedText:      "\nsit amet consectetur\nadipiscing elit",
		},
		{
			name:        "delete to end of line shortcut",
			initialText: "Lorem ipsum dolor\nsit amet consectetur\nadipiscing elit",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'D', tcell.ModNone),
			},
			expectedCursorPos: 2,
			expectedText:      "Lor\nsit amet consectetur\nadipiscing elit",
		},
		{
			name:        "delete to start of line",
			initialText: "Lorem ipsum dolor\n\tsit amet consectetur\n\t\tadipiscing\nelit\n\n",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, 'j', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'd', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, '0', tcell.ModNone),
			},
			expectedCursorPos: 18,
			expectedText:      "Lorem ipsum dolor\net consectetur\n\t\tadipiscing\nelit\n\n",
		},
		{
			name:        "delete to start of line after indentation",
			initialText: "Lorem ipsum dolor\n\tsit amet consectetur\n\t\tadipiscing\nelit\n\n",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, 'j', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'd', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, '^', tcell.ModNone),
			},
			expectedCursorPos: 19,
			expectedText:      "Lorem ipsum dolor\n\tt consectetur\n\t\tadipiscing\nelit\n\n",
		},
		{
			name:        "delete to start of next word",
			initialText: "Lorem ipsum dolor\nsit amet consectetur\nadipiscing elit",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, 'd', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'w', tcell.ModNone),
			},
			expectedCursorPos: 0,
			expectedText:      "ipsum dolor\nsit amet consectetur\nadipiscing elit",
		},
		{
			name:        "delete to start of next word on an empty line with next line indented",
			initialText: "a\n\n    bcd",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, 'j', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'd', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'w', tcell.ModNone),
			},
			expectedCursorPos: 6,
			expectedText:      "a\n    bcd",
		},
		{
			name:        "delete to start of next word at end of line with whitespace",
			initialText: "a\n   \nbcd",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, 'j', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, '$', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'd', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'w', tcell.ModNone),
			},
			expectedCursorPos: 3,
			expectedText:      "a\n  \nbcd",
		},
		{
			name:        "delete to start of next word end of document",
			initialText: "abcdef",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'd', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'w', tcell.ModNone),
			},
			expectedCursorPos: 1,
			expectedText:      "ab",
		},
		{
			name:        "delete to start of next word with count",
			initialText: "Lorem ipsum dolor\nsit amet consectetur\nadipiscing elit",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, '4', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'd', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'w', tcell.ModNone),
			},
			expectedCursorPos: 0,
			expectedText:      "amet consectetur\nadipiscing elit",
		},
		{
			name:        "delete to start of next word with verb and object count",
			initialText: "Lorem ipsum dolor\nsit amet consectetur\nadipiscing elit",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, '2', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'd', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, '3', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'w', tcell.ModNone),
			},
			expectedCursorPos: 0,
			expectedText:      "\nadipiscing elit",
		},
		{
			name:        "delete to start of next word including punctuation",
			initialText: "Lorem.ipsum,dolor;sit amet consectetur\nadipiscing elit",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, 'd', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'W', tcell.ModNone),
			},
			expectedCursorPos: 0,
			expectedText:      "amet consectetur\nadipiscing elit",
		},
		{
			name:        "delete a word",
			initialText: "Lorem ipsum dolor\nsit amet consectetur\nadipiscing elit",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'd', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'a', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'w', tcell.ModNone),
			},
			expectedCursorPos: 6,
			expectedText:      "Lorem dolor\nsit amet consectetur\nadipiscing elit",
		},
		// NOTE: vim removes the trailing empty line, but aretext keeps it.
		{
			name:        "delete a word on an empty line with next line indented",
			initialText: "a\n\n    bcd",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, 'j', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'd', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'a', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'w', tcell.ModNone),
			},
			expectedCursorPos: 2,
			expectedText:      "a\n",
		},
		{
			name:        "delete a word in whitespace before word",
			initialText: "ab   cd   ef",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'd', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'a', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'w', tcell.ModNone),
			},
			expectedCursorPos: 2,
			expectedText:      "ab   ef",
		},
		// NOTE: aretext deletes the word, but vim deletes the leading whitespace as well.
		{
			name:        "delete a word on word at end of line with leading whitespace",
			initialText: "ab   cd   ef",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, '$', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'd', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'a', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'w', tcell.ModNone),
			},
			expectedCursorPos: 9,
			expectedText:      "ab   cd   ",
		},
		{
			name:        "delete a word with count",
			initialText: "Lorem ipsum dolor\nsit amet consectetur\nadipiscing elit",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, '5', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'd', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'a', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'w', tcell.ModNone),
			},
			expectedCursorPos: 0,
			expectedText:      "consectetur\nadipiscing elit",
		},
		{
			name:        "delete inner word",
			initialText: "Lorem ipsum dolor\nsit amet consectetur\nadipiscing elit",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, 'j', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'd', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'i', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'w', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyEsc, '\x00', tcell.ModNone),
			},
			expectedCursorPos: 18,
			expectedText:      "Lorem ipsum dolor\n amet consectetur\nadipiscing elit",
		},
		{
			name:        "delete inner word at end of line",
			initialText: "Lorem ipsum dolor\nsit amet consectetur\nadipiscing elit",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, 'w', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'w', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'd', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'i', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'w', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyEsc, '\x00', tcell.ModNone),
			},
			expectedCursorPos: 11,
			expectedText:      "Lorem ipsum \nsit amet consectetur\nadipiscing elit",
		},
		{
			name:        "delete inner word with count",
			initialText: "Lorem ipsum dolor\nsit amet consectetur\nadipiscing elit",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, '7', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'd', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'i', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'w', tcell.ModNone),
			},
			expectedCursorPos: 0,
			expectedText:      "amet consectetur\nadipiscing elit",
		},
		{
			name:        "delete inner string object with double-quote",
			initialText: `"ab c" x`,
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, 'd', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'i', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, '"', tcell.ModNone),
			},
			expectedCursorPos: 1,
			expectedText:      `"" x`,
		},
		{
			name:        "delete a string object with double-quote",
			initialText: `"ab c" x`,
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, 'd', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'a', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, '"', tcell.ModNone),
			},
			expectedCursorPos: 0,
			expectedText:      ` x`,
		},
		{
			name:        "delete inner string object with single-quote",
			initialText: `'ab c' x`,
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, 'd', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'i', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, '\'', tcell.ModNone),
			},
			expectedCursorPos: 1,
			expectedText:      `'' x`,
		},
		{
			name:        "delete a string object with single-quote",
			initialText: `'ab c' x`,
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, 'd', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'a', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, '\'', tcell.ModNone),
			},
			expectedCursorPos: 0,
			expectedText:      ` x`,
		},
		{
			name:        "delete inner string object with backtick-quote",
			initialText: "`ab c` x",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, 'd', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'i', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, '`', tcell.ModNone),
			},
			expectedCursorPos: 1,
			expectedText:      "`` x",
		},
		{
			name:        "delete a string object with backtick-quote",
			initialText: "`ab c` x",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, 'd', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'a', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, '`', tcell.ModNone),
			},
			expectedCursorPos: 0,
			expectedText:      " x",
		},
		{
			name:        "delete to next matching character in line",
			initialText: "Lorem ipsum dolor\nsit amet consectetur\nadipiscing elit",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, 'd', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'f', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'm', tcell.ModNone),
			},
			expectedCursorPos: 0,
			expectedText:      " ipsum dolor\nsit amet consectetur\nadipiscing elit",
		},
		{
			name:        "delete to prev matching char in line",
			initialText: "Lorem ipsum dolor\nsit amet consectetur\nadipiscing elit",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, 'j', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'd', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'F', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 't', tcell.ModNone),
			},
			expectedCursorPos: 25,
			expectedText:      "Lorem ipsum dolor\nsit amensectetur\nadipiscing elit",
		},
		{
			name:        "delete till next matching char in line",
			initialText: "Lorem ipsum dolor\nsit amet consectetur\nadipiscing elit",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, 'd', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 't', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'd', tcell.ModNone),
			},
			expectedCursorPos: 0,
			expectedText:      "dolor\nsit amet consectetur\nadipiscing elit",
		},
		{
			name:        "delete till prev matching char in line",
			initialText: "Lorem ipsum dolor\nsit amet consectetur\nadipiscing elit",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, 'j', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'd', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'T', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 's', tcell.ModNone),
			},
			expectedCursorPos: 19,
			expectedText:      "Lorem ipsum dolor\nsonsectetur\nadipiscing elit",
		},
		{
			name:        "delete a word, then put from default clipboard",
			initialText: "Lorem ipsum dolor\nsit amet consectetur\nadipiscing elit",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, 'd', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'a', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'w', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'o', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyEsc, '\x00', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'p', tcell.ModNone),
			},
			expectedCursorPos: 17,
			expectedText:      "ipsum dolor\nLorem \nsit amet consectetur\nadipiscing elit",
		},
		{
			name:        "change word from start of word",
			initialText: "Lorem ipsum dolor\nsit amet consectetur\nadipiscing elit",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, 'c', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'w', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'f', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'o', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'o', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyEsc, '\x00', tcell.ModNone),
			},
			expectedCursorPos: 2,
			expectedText:      "foo ipsum dolor\nsit amet consectetur\nadipiscing elit",
		},
		{
			name:        "change word from middle of word",
			initialText: "Lorem ipsum dolor\nsit amet consectetur\nadipiscing elit",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'c', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'w', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'f', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'o', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'o', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyEsc, '\x00', tcell.ModNone),
			},
			expectedCursorPos: 5,
			expectedText:      "Lorfoo ipsum dolor\nsit amet consectetur\nadipiscing elit",
		},
		{
			name:        "change word from end of word",
			initialText: "Lorem ipsum dolor\nsit amet consectetur\nadipiscing elit",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, 'e', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'c', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'w', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'f', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'o', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'o', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyEsc, '\x00', tcell.ModNone),
			},
			expectedCursorPos: 6,
			expectedText:      "Lorefoo ipsum dolor\nsit amet consectetur\nadipiscing elit",
		},
		{
			name:        "change word in leading whitespace",
			initialText: "abcd      efghi    jklm",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, 'e', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'c', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'w', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'f', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'o', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'o', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyEsc, '\x00', tcell.ModNone),
			},
			expectedCursorPos: 8,
			expectedText:      "abcd  fooefghi    jklm",
		},
		{
			name:        "change word with count",
			initialText: "Lorem ipsum dolor\nsit amet consectetur\nadipiscing elit",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, '3', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'c', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'w', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'f', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'o', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'o', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyEsc, '\x00', tcell.ModNone),
			},
			expectedCursorPos: 2,
			expectedText:      "foo dolor\nsit amet consectetur\nadipiscing elit",
		},
		{
			name:        "change a word",
			initialText: "Lorem ipsum dolor\nsit amet consectetur\nadipiscing elit",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, 'j', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'c', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'a', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'w', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'f', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'o', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'o', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'b', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'a', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'r', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, ' ', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyEsc, '\x00', tcell.ModNone),
			},
			expectedCursorPos: 28,
			expectedText:      "Lorem ipsum dolor\nsit foobar consectetur\nadipiscing elit",
		},
		{
			name:        "change a word at end of line",
			initialText: "foo=bar\nbaz",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, '$', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'c', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'a', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'w', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'x', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'x', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'x', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyEsc, '\x00', tcell.ModNone),
			},
			expectedCursorPos: 6,
			expectedText:      "foo=xxx\nbaz",
		},
		{
			name:        "change a word with count",
			initialText: "Lorem ipsum dolor\nsit amet consectetur\nadipiscing elit",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, '5', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'c', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'a', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'w', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'f', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'o', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'o', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'b', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'a', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'r', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyEsc, '\x00', tcell.ModNone),
			},
			expectedCursorPos: 5,
			expectedText:      "foobarconsectetur\nadipiscing elit",
		},
		{
			name:        "change inner word",
			initialText: "Lorem ipsum dolor\nsit amet consectetur\nadipiscing elit",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, 'j', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'c', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'i', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'w', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'f', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'o', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'o', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'b', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'a', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'r', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyEsc, '\x00', tcell.ModNone),
			},
			expectedCursorPos: 27,
			expectedText:      "Lorem ipsum dolor\nsit foobar consectetur\nadipiscing elit",
		},
		{
			name:        "change inner word at end of line",
			initialText: "foo=bar\nbaz",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, '$', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'c', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'i', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'w', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'x', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'x', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'x', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyEsc, '\x00', tcell.ModNone),
			},
			expectedCursorPos: 6,
			expectedText:      "foo=xxx\nbaz",
		},
		{
			name:        "change inner word with count",
			initialText: "Lorem ipsum dolor\nsit amet consectetur\nadipiscing elit",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, '8', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'c', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'i', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'w', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'f', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'o', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'o', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'b', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'a', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'r', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyEsc, '\x00', tcell.ModNone),
			},
			expectedCursorPos: 5,
			expectedText:      "foobar consectetur\nadipiscing elit",
		},
		{
			name:        "change inner string object with double-quote",
			initialText: `"abc" x`,
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, 'c', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'i', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, '"', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'f', tcell.ModNone),
			},
			expectedCursorPos: 2,
			expectedText:      `"f" x`,
		},
		{
			name:        "change a string object with double-quote",
			initialText: `"abc" x`,
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, 'c', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'a', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, '"', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'f', tcell.ModNone),
			},
			expectedCursorPos: 1,
			expectedText:      `f x`,
		},
		{
			name:        "change inner string object with single-quote",
			initialText: `'abc' x`,
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, 'c', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'i', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, '\'', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'f', tcell.ModNone),
			},
			expectedCursorPos: 2,
			expectedText:      `'f' x`,
		},
		{
			name:        "change a string object with single-quote",
			initialText: `'abc' x`,
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, 'c', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'a', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, '\'', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'f', tcell.ModNone),
			},
			expectedCursorPos: 1,
			expectedText:      `f x`,
		},
		{
			name:        "change inner string object with backtick-quote",
			initialText: "`abc` x",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, 'c', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'i', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, '`', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'f', tcell.ModNone),
			},
			expectedCursorPos: 2,
			expectedText:      "`f` x",
		},
		{
			name:        "change a string object with backtick-quote",
			initialText: "`abc` x",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, 'c', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'a', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, '`', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'f', tcell.ModNone),
			},
			expectedCursorPos: 1,
			expectedText:      "f x",
		},
		{
			name:        "change to next matching char in line",
			initialText: "Lorem ipsum dolor\nsit amet consectetur\nadipiscing elit",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, 'c', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 't', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 's', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 't', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'e', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 's', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 't', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyEsc, '\x00', tcell.ModNone),
			},
			expectedCursorPos: 3,
			expectedText:      "testsum dolor\nsit amet consectetur\nadipiscing elit",
		},
		{
			name: "change to next matching char in line at end of line",

			initialText: "foobar123\nbaz",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, 'c', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 't', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, '3', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 't', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'e', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 's', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 't', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyEsc, '\x00', tcell.ModNone),
			},
			expectedCursorPos: 3,
			expectedText:      "test3\nbaz",
		},
		{
			name:        "change to prev matching character in line ",
			initialText: "Lorem ipsum dolor\nsit amet consectetur\nadipiscing elit",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, 'j', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'c', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'F', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 's', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'f', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'o', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'o', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'b', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'a', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'r', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyEsc, '\x00', tcell.ModNone),
			},
			expectedCursorPos: 23,
			expectedText:      "Lorem ipsum dolor\nfoobaronsectetur\nadipiscing elit",
		},
		{
			name:        "change till next matching char in line",
			initialText: "Lorem ipsum dolor\nsit amet consectetur\nadipiscing elit",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, 'c', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'f', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'm', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 't', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'e', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 's', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 't', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyEsc, '\x00', tcell.ModNone),
			},
			expectedCursorPos: 3,
			expectedText:      "test ipsum dolor\nsit amet consectetur\nadipiscing elit",
		},
		{
			name:        "change till next matching char in line at end of line",
			initialText: "foobar123\nbaz",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, 'c', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'f', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, '3', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 't', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'e', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 's', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 't', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyEsc, '\x00', tcell.ModNone),
			},
			expectedCursorPos: 3,
			expectedText:      "test\nbaz",
		},
		{
			name:        "change till prev matching char in line",
			initialText: "Lorem ipsum dolor\nsit amet consectetur\nadipiscing elit",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, 'j', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'c', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'T', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 't', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'f', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'o', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'o', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'b', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'a', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'r', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyEsc, '\x00', tcell.ModNone),
			},
			expectedCursorPos: 31,
			expectedText:      "Lorem ipsum dolor\nsit ametfoobarectetur\nadipiscing elit",
		},
		{
			name:        "delete inner paren block (dib)",
			initialText: "abc (def) ghi",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, '5', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'd', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'i', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'b', tcell.ModNone),
			},
			expectedCursorPos: 5,
			expectedText:      "abc () ghi",
		},
		{
			name:        "delete inner paren block (di+openparen)",
			initialText: "abc (def) ghi",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, '5', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'd', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'i', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, '(', tcell.ModNone),
			},
			expectedCursorPos: 5,
			expectedText:      "abc () ghi",
		},
		{
			name:        "delete inner paren block (di+closeparen)",
			initialText: "abc (def) ghi",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, '5', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'd', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'i', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, ')', tcell.ModNone),
			},
			expectedCursorPos: 5,
			expectedText:      "abc () ghi",
		},
		{
			name:        "delete a paren block (dib)",
			initialText: "abc (def) ghi",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, '5', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'd', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'a', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'b', tcell.ModNone),
			},
			expectedCursorPos: 4,
			expectedText:      "abc  ghi",
		},
		{
			name:        "delete inner brace block (diB)",
			initialText: `func() { fmt.Printf("abc") }`,
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, '9', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'd', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'i', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'B', tcell.ModNone),
			},
			expectedCursorPos: 8,
			expectedText:      "func() {}",
		},
		{
			name:        "delete a brace block (daB)",
			initialText: `func() { fmt.Printf("abc") }`,
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, '9', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'd', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'a', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'B', tcell.ModNone),
			},
			expectedCursorPos: 7,
			expectedText:      "func() ",
		},
		{
			name:        "delete inner angle block (di<)",
			initialText: `<html><body /></html>`,
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, '9', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'd', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'i', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, '<', tcell.ModNone),
			},
			expectedCursorPos: 7,
			expectedText:      "<html><></html>",
		},
		{
			name:        "delete an angle block (da<)",
			initialText: `<html><body /></html>`,
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, '9', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'd', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'a', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, '<', tcell.ModNone),
			},
			expectedCursorPos: 6,
			expectedText:      "<html></html>",
		},
		{
			name:        "change inner paren block (cib)",
			initialText: "abc (def) ghi",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, '5', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'c', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'i', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'b', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'x', tcell.ModNone),
			},
			expectedCursorPos: 6,
			expectedText:      "abc (x) ghi",
		},
		{
			name:        "change inner paren block (ci+openparen)",
			initialText: "abc (def) ghi",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, '5', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'c', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'i', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, '(', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'x', tcell.ModNone),
			},
			expectedCursorPos: 6,
			expectedText:      "abc (x) ghi",
		},
		{
			name:        "change inner paren block (ci+closeparen)",
			initialText: "abc (def) ghi",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, '5', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'c', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'i', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, ')', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'x', tcell.ModNone),
			},
			expectedCursorPos: 6,
			expectedText:      "abc (x) ghi",
		},
		{
			name:        "change inner paren block, outside block",
			initialText: "abc (def) ghi",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, 'c', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'i', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'b', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'x', tcell.ModNone), // interpret in normal mode, NOT insert mode.
			},
			expectedCursorPos: 0,
			expectedText:      "bc (def) ghi",
		},
		{
			name:        "change inner brace block (ciB), single line",
			initialText: `func() { fmt.Printf("abc") }`,
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, '9', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'c', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'i', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'B', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'x', tcell.ModNone),
			},
			expectedCursorPos: 9,
			expectedText:      "func() {x}",
		},
		{
			name:        "change inner brace block (ciB), multiple lines",
			initialText: "func() {\n\tfmt.Printf(\"abc\")\n\tfmt.Printf(\"xyz\")\n}",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, '9', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'c', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'i', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'B', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'x', tcell.ModNone),
			},
			expectedCursorPos: 10,
			expectedText:      "func() {\nx\n}",
		},
		{
			name:        "change a brace block (caB)",
			initialText: `func() { fmt.Printf("abc") }`,
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, '9', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'c', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'a', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'B', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'x', tcell.ModNone),
			},
			expectedCursorPos: 8,
			expectedText:      "func() x",
		},
		{
			name:        "change inner angle block (ci<)",
			initialText: `<html><body /></html>`,
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, '9', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'c', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'i', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, '<', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'x', tcell.ModNone),
			},
			expectedCursorPos: 8,
			expectedText:      "<html><x></html>",
		},
		{
			name:        "change an angle block (ca<)",
			initialText: `<html><body /></html>`,
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, '9', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'c', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'a', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, '<', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'x', tcell.ModNone),
			},
			expectedCursorPos: 7,
			expectedText:      "<html>x</html>",
		},
		{
			name:        "replace character",
			initialText: "Lorem ipsum dolor\nsit amet consectetur\nadipiscing elit",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, 'r', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'x', tcell.ModNone),
			},
			expectedCursorPos: 0,
			expectedText:      "xorem ipsum dolor\nsit amet consectetur\nadipiscing elit",
		},
		{
			name:        "replace character with newline",
			initialText: "Lorem ipsum dolor\nsit amet consectetur\nadipiscing elit",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'r', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyEnter, '\r', tcell.ModNone),
			},
			expectedCursorPos: 9,
			expectedText:      "Lorem ip\num dolor\nsit amet consectetur\nadipiscing elit",
		},
		{
			name:        "replace character with tab",
			initialText: "Lorem ipsum dolor\nsit amet consectetur\nadipiscing elit",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'r', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyTab, '\t', tcell.ModNone),
			},
			expectedCursorPos: 8,
			expectedText:      "Lorem ip\tum dolor\nsit amet consectetur\nadipiscing elit",
		},
		{
			name:        "toggle case",
			initialText: "Lorem ipsum dolor\nsit amet consectetur\nadipiscing elit",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, '~', tcell.ModNone),
			},
			expectedCursorPos: 1,
			expectedText:      "lorem ipsum dolor\nsit amet consectetur\nadipiscing elit",
		},
		{
			name:        "indent line",
			initialText: "Lorem ipsum dolor\nsit amet consectetur\nadipiscing elit",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, '>', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, '>', tcell.ModNone),
			},
			expectedCursorPos: 1,
			expectedText:      "\tLorem ipsum dolor\nsit amet consectetur\nadipiscing elit",
		},
		{
			name:        "indent line with count",
			initialText: "Lorem ipsum dolor\nsit amet consectetur\nadipiscing elit",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, '2', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, '>', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, '>', tcell.ModNone),
			},
			expectedCursorPos: 1,
			expectedText:      "\tLorem ipsum dolor\n\tsit amet consectetur\nadipiscing elit",
		},
		{
			name:        "outdent line",
			initialText: "Lorem ipsum dolor\n\tsit amet consectetur\n\t\tadipiscing\nelit\n\n",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, 'j', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, '<', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, '<', tcell.ModNone),
			},
			expectedCursorPos: 18,
			expectedText:      "Lorem ipsum dolor\nsit amet consectetur\n\t\tadipiscing\nelit\n\n",
		},
		{
			name:        "outdent line with count",
			initialText: "Lorem ipsum dolor\n\tsit amet consectetur\n\t\tadipiscing\nelit\n\n",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, 'j', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, '2', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, '<', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, '<', tcell.ModNone),
			},
			expectedCursorPos: 18,
			expectedText:      "Lorem ipsum dolor\nsit amet consectetur\n\tadipiscing\nelit\n\n",
		},
		{
			name:        "yank to start of next word",
			initialText: "Lorem ipsum dolor\nsit amet consectetur\nadipiscing elit",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, 'y', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'w', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'o', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyEsc, '\x00', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'p', tcell.ModNone),
			},
			expectedCursorPos: 23,
			expectedText:      "Lorem ipsum dolor\nLorem \nsit amet consectetur\nadipiscing elit",
		},
		{
			name:        "yank to start of next word with count",
			initialText: "Lorem ipsum dolor\nsit amet consectetur\nadipiscing elit",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, '4', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'y', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'w', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'o', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyEsc, '\x00', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'p', tcell.ModNone),
			},
			expectedCursorPos: 39,
			expectedText:      "Lorem ipsum dolor\nLorem ipsum dolor\nsit \nsit amet consectetur\nadipiscing elit",
		},
		{
			name:        "yank to start of next word including punctuation",
			initialText: "Lorem.ipsum;dolor sit amet consectetur\nadipiscing elit",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, 'y', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'W', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'o', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyEsc, '\x00', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'p', tcell.ModNone),
			},
			expectedCursorPos: 56,
			expectedText:      "Lorem.ipsum;dolor sit amet consectetur\nLorem.ipsum;dolor \nadipiscing elit",
		},
		{
			name:        "yank a word",
			initialText: "Lorem ipsum dolor\nsit amet consectetur\nadipiscing elit",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, 'w', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'y', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'a', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'w', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyEsc, '\x00', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'o', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyEsc, '\x00', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'p', tcell.ModNone),
			},
			expectedCursorPos: 23,
			expectedText:      "Lorem ipsum dolor\nipsum \nsit amet consectetur\nadipiscing elit",
		},
		{
			name:        "yank a word with count",
			initialText: "Lorem ipsum dolor\nsit amet consectetur\nadipiscing elit",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, '4', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'y', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'a', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'w', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyEsc, '\x00', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'o', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyEsc, '\x00', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'p', tcell.ModNone),
			},
			expectedCursorPos: 39,
			expectedText:      "Lorem ipsum dolor\nLorem ipsum dolor\nsit \nsit amet consectetur\nadipiscing elit",
		},
		{
			name:        "yank inner word",
			initialText: "Lorem ipsum dolor\nsit amet consectetur\nadipiscing elit",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, 'y', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'i', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'w', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyEsc, '\x00', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'o', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyEsc, '\x00', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'p', tcell.ModNone),
			},
			expectedCursorPos: 22,
			expectedText:      "Lorem ipsum dolor\nLorem\nsit amet consectetur\nadipiscing elit",
		},
		{
			name:        "yank inner word with count",
			initialText: "Lorem ipsum dolor\nsit amet consectetur\nadipiscing elit",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, '3', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'y', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'i', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'w', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyEsc, '\x00', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'o', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyEsc, '\x00', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'p', tcell.ModNone),
			},
			expectedCursorPos: 28,
			expectedText:      "Lorem ipsum dolor\nLorem ipsum\nsit amet consectetur\nadipiscing elit",
		},
		{
			name:        "yank inner string block with double-quote",
			initialText: `"abc" x`,
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, 'y', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'i', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, '"', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'o', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyEsc, '\x00', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'p', tcell.ModNone),
			},
			expectedCursorPos: 10,
			expectedText:      "\"abc\" x\nabc",
		},
		{
			name:        "yank a string block with double-quote",
			initialText: `"abc" x`,
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, 'y', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'a', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, '"', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'o', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyEsc, '\x00', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'p', tcell.ModNone),
			},
			expectedCursorPos: 12,
			expectedText:      "\"abc\" x\n\"abc\"",
		},
		{
			name:        "yank inner string block with single-quote",
			initialText: `'abc' x`,
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, 'y', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'i', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, '\'', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'o', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyEsc, '\x00', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'p', tcell.ModNone),
			},
			expectedCursorPos: 10,
			expectedText:      "'abc' x\nabc",
		},
		{
			name:        "yank a string block with single-quote",
			initialText: `'abc' x`,
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, 'y', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'a', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, '\'', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'o', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyEsc, '\x00', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'p', tcell.ModNone),
			},
			expectedCursorPos: 12,
			expectedText:      "'abc' x\n'abc'",
		},
		{
			name:        "yank inner string block with backtick-quote",
			initialText: "`abc` x",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, 'y', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'i', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, '`', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'o', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyEsc, '\x00', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'p', tcell.ModNone),
			},
			expectedCursorPos: 10,
			expectedText:      "`abc` x\nabc",
		},
		{
			name:        "yank a string block with backtick-quote",
			initialText: "`abc` x",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, 'y', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'a', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, '`', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'o', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyEsc, '\x00', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'p', tcell.ModNone),
			},
			expectedCursorPos: 12,
			expectedText:      "`abc` x\n`abc`",
		},
		{
			name:        "yank line",
			initialText: "Lorem ipsum dolor\nsit amet consectetur\nadipiscing elit",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, 'y', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'y', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'p', tcell.ModNone),
			},
			expectedCursorPos: 18,
			expectedText:      "Lorem ipsum dolor\nLorem ipsum dolor\nsit amet consectetur\nadipiscing elit",
		},
		{
			name:        "yank to next matching character in line",
			initialText: "Lorem ipsum dolor\nsit amet consectetur\nadipiscing elit",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, 'y', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'f', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'e', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'o', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyEsc, '\x00', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'p', tcell.ModNone),
			},
			expectedCursorPos: 21,
			expectedText:      "Lorem ipsum dolor\nLore\nsit amet consectetur\nadipiscing elit",
		},
		{
			name:        "yank to prev matching character in line",
			initialText: "Lorem ipsum dolor\nsit amet consectetur\nadipiscing elit",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, '$', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'y', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'F', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'o', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyEsc, '\x00', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'p', tcell.ModNone),
			},
			expectedCursorPos: 19,
			expectedText:      "Lorem ipsum dolor\nlo\nsit amet consectetur\nadipiscing elit",
		},
		{
			name:        "yank till next matching character in line",
			initialText: "Lorem ipsum dolor\nsit amet consectetur\nadipiscing elit",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, 'y', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 't', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'e', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'o', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyEsc, '\x00', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'p', tcell.ModNone),
			},
			expectedCursorPos: 20,
			expectedText:      "Lorem ipsum dolor\nLor\nsit amet consectetur\nadipiscing elit",
		},
		{
			name:        "yank till prev matching character in line",
			initialText: "Lorem ipsum dolor\nsit amet consectetur\nadipiscing elit",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, '$', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'y', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'T', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'o', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyEsc, '\x00', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'p', tcell.ModNone),
			},
			expectedCursorPos: 18,
			expectedText:      "Lorem ipsum dolor\no\nsit amet consectetur\nadipiscing elit",
		},
		{
			name:        "put before cursor",
			initialText: "Lorem ipsum dolor\nsit amet consectetur\nadipiscing elit",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, 'y', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'a', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'w', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'P', tcell.ModNone),
			},
			expectedCursorPos: 5,
			expectedText:      "Lorem Lorem ipsum dolor\nsit amet consectetur\nadipiscing elit",
		},
		{
			name:        "search forward",
			initialText: "Lorem ipsum dolor\nsit amet consectetur\nadipiscing elit",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, '/', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'c', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'o', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'n', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyEnter, '\r', tcell.ModNone),
			},
			expectedCursorPos: 27,
			expectedText:      "Lorem ipsum dolor\nsit amet consectetur\nadipiscing elit",
		},
		{
			name:        "search backward",
			initialText: "Lorem ipsum dolor\nsit amet consectetur\nadipiscing elit",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, 'j', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'j', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, '?', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'o', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'r', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyEnter, '\r', tcell.ModNone),
			},
			expectedCursorPos: 15,
			expectedText:      "Lorem ipsum dolor\nsit amet consectetur\nadipiscing elit",
		},
		{
			name:        "find next match",
			initialText: "Lorem ipsum dolor\nsit amet consectetur\nadipiscing elit",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, '/', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'm', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyEnter, '\r', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'n', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'n', tcell.ModNone),
			},
			expectedCursorPos: 23,
			expectedText:      "Lorem ipsum dolor\nsit amet consectetur\nadipiscing elit",
		},
		{
			name:        "find prev match",
			initialText: "Lorem ipsum dolor\nsit amet consectetur\nadipiscing elit",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, 'j', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'j', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, '/', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 't', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyEnter, '\r', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'N', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'N', tcell.ModNone),
			},
			expectedCursorPos: 33,
			expectedText:      "Lorem ipsum dolor\nsit amet consectetur\nadipiscing elit",
		},
		{
			name:        "search forward for word under cursor",
			initialText: "Lorem ipsum dolor\nsit amet consectetur\nlorem ipsum dolor",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, 'w', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, '*', tcell.ModNone),
			},
			expectedCursorPos: 45,
			expectedText:      "Lorem ipsum dolor\nsit amet consectetur\nlorem ipsum dolor",
		},
		{
			name:        "search backward for word under cursor",
			initialText: "Lorem ipsum dolor\nsit amet consectetur\nlorem ipsum dolor",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, 'j', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'j', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'w', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, '#', tcell.ModNone),
			},
			expectedCursorPos: 6,
			expectedText:      "Lorem ipsum dolor\nsit amet consectetur\nlorem ipsum dolor",
		},
		{
			name:        "search forward and delete",
			initialText: "Lorem ipsum dolor\nsit amet consectetur\nlorem ipsum dolor",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, 'd', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, '/', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'd', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'o', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyEnter, '\x00', tcell.ModNone),
			},
			expectedCursorPos: 0,
			expectedText:      "dolor\nsit amet consectetur\nlorem ipsum dolor",
		},
		{
			name:        "search backward and delete",
			initialText: "Lorem ipsum dolor\nsit amet consectetur\nlorem ipsum dolor",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, 'j', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, '$', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'd', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, '?', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'd', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'o', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyEnter, '\x00', tcell.ModNone),
			},
			expectedCursorPos: 15,
			expectedText:      "Lorem ipsum dolr\nlorem ipsum dolor",
		},
		{
			name:        "search and delete to clipboard then paste",
			initialText: "Lorem ipsum dolor\nsit amet consectetur\nlorem ipsum dolor",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, '"', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'a', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'd', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, '/', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'd', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'o', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyEnter, '\x00', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'G', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, '"', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'a', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'P', tcell.ModNone),
			},
			expectedCursorPos: 38,
			expectedText:      "dolor\nsit amet consectetur\nLorem ipsum lorem ipsum dolor",
		},
		{
			name:        "search and delete then undo",
			initialText: "Lorem ipsum dolor\nsit amet consectetur\nlorem ipsum dolor",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, 'D', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'd', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, '/', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'c', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'o', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'n', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyEnter, '\x00', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'u', tcell.ModNone),
			},
			expectedCursorPos: 0,
			expectedText:      "\nsit amet consectetur\nlorem ipsum dolor",
		},
		{
			name:        "search and delete then replay last action",
			initialText: "abc xyz\nabc xyz 123\nabc xyz 456",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, 'd', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, '/', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'x', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'y', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyEnter, '\x00', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'j', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, '.', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, '/', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'a', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyEnter, '\x00', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'j', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, '.', tcell.ModNone),
			},
			expectedCursorPos: 12,
			expectedText:      "xyz\nxyz 123\nxyz 456",
		},
		{
			name:        "search and delete in user macro then replay",
			initialText: "abc xyz\nabc xyz 123\nabc xyz 456",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, ':', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 's', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 't', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'a', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'r', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 't', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyEnter, '\r', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'd', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, '/', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'x', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'y', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyEnter, '\r', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, ':', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 's', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 't', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'o', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'p', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyEnter, '\r', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, ':', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'r', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'e', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'p', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyEnter, '\r', tcell.ModNone),
			},
			expectedCursorPos: 0,
			expectedText:      "xyz 123\nabc xyz 456",
		},
		{
			name:        "search forward and change",
			initialText: "abc xyz 123",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, 'c', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, '/', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'x', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'y', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyEnter, '\r', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'f', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'o', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'o', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyEscape, '\x00', tcell.ModNone),
			},
			expectedCursorPos: 2,
			expectedText:      "fooxyz 123",
		},
		{
			name:        "search backward and change",
			initialText: "abc xyz 123",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, '$', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'c', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, '?', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'x', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'y', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyEnter, '\r', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'f', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'o', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'o', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyEscape, '\x00', tcell.ModNone),
			},
			expectedCursorPos: 8,
			expectedText:      "abc xyfoo3",
		},
		{
			name:        "search and change then replay last action",
			initialText: "abc xyz 123\nabc xyz 123\nabc xyz 123",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, 'c', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, '/', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'x', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'y', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyEnter, '\r', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'f', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'o', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'o', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyEscape, '\x00', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'j', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, '^', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, '.', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'j', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, '^', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, '.', tcell.ModNone),
			},
			expectedCursorPos: 24,
			expectedText:      "fooxyz 123\nfooxyz 123\nfooxyz 123",
		},
		{
			name:        "search forward and yank",
			initialText: "abc xyz 123",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, 'y', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, '/', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'x', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'y', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyEnter, '\r', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'p', tcell.ModNone),
			},
			expectedCursorPos: 4,
			expectedText:      "aabc bc xyz 123",
		},
		{
			name:        "search forward and yank to clipboard",
			initialText: "abc xyz 123",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, '"', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'a', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'y', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, '/', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'x', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'y', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyEnter, '\r', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, '"', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'a', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'p', tcell.ModNone),
			},
			expectedCursorPos: 4,
			expectedText:      "aabc bc xyz 123",
		},
		{
			name:        "search backward and yank",
			initialText: "abc xyz 123",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, '$', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'y', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, '?', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'x', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'y', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyEnter, '\r', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'P', tcell.ModNone),
			},
			expectedCursorPos: 13,
			expectedText:      "abc xyz 12z 123",
		},
		{
			name:        "search backward and yank to clipboard",
			initialText: "abc xyz 123",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, '$', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, '"', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'a', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'y', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, '?', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'x', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'y', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyEnter, '\r', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, '"', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'a', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'P', tcell.ModNone),
			},
			expectedCursorPos: 13,
			expectedText:      "abc xyz 12z 123",
		},
		{
			name:        "previous search query in history",
			initialText: "Lorem ipsum dolor\nsit amet consectetur\nlorem ipsum dolor",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, '/', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'a', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'm', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyEnter, '\x00', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, '/', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'd', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'o', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyEnter, '\x00', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, '/', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyUp, '\x00', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyUp, '\x00', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyEnter, '\x00', tcell.ModNone),
			},
			expectedCursorPos: 22,
			expectedText:      "Lorem ipsum dolor\nsit amet consectetur\nlorem ipsum dolor",
		},
		{
			name:        "next search query in history",
			initialText: "Lorem ipsum dolor\nsit amet consectetur\nlorem ipsum dolor",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, '/', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'a', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'm', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyEscape, '\x00', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, '/', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'd', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'o', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyEscape, '\x00', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, '/', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 's', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'i', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyEscape, '\x00', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, '/', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyUp, '\x00', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyUp, '\x00', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyUp, '\x00', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyDown, '\x00', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyEnter, '\x00', tcell.ModNone),
			},
			expectedCursorPos: 12,
			expectedText:      "Lorem ipsum dolor\nsit amet consectetur\nlorem ipsum dolor",
		},
		{
			name:        "undo",
			initialText: "Lorem ipsum dolor\nsit amet consectetur\nadipiscing elit",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, 'i', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 't', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'e', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 's', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 't', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, ' ', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyEsc, '\x00', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'c', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'i', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'w', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'f', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'o', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'o', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyEsc, '\x00', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'u', tcell.ModNone),
			},
			expectedCursorPos: 5,
			expectedText:      "test Lorem ipsum dolor\nsit amet consectetur\nadipiscing elit",
		},
		{
			name:        "redo",
			initialText: "Lorem ipsum dolor\nsit amet consectetur\nadipiscing elit",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, 'j', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'd', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'd', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'd', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'd', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'u', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyCtrlR, '\x12', tcell.ModCtrl),
			},
			expectedCursorPos: 0,
			expectedText:      "Lorem ipsum dolor",
		},
		{
			name:        "repeat last action delete-a-word",
			initialText: "Lorem ipsum dolor\nsit amet consectetur\nadipiscing elit",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, 'd', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'a', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'w', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, '.', tcell.ModNone),
			},
			expectedCursorPos: 0,
			expectedText:      "dolor\nsit amet consectetur\nadipiscing elit",
		},
		{
			name:        "visual linewise delete",
			initialText: "Lorem ipsum dolor\nsit amet consectetur\nadipiscing elit",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, 'V', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'j', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'd', tcell.ModNone),
			},
			expectedCursorPos: 0,
			expectedText:      "adipiscing elit",
		},
		{
			name:        "visual linewise delete all lines",
			initialText: "Lorem ipsum dolor\nsit amet consectetur\nadipiscing elit",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, 'V', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'G', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'd', tcell.ModNone),
			},
			expectedCursorPos: 0,
			expectedText:      "",
		},
		{
			name:        "visual charwise delete",
			initialText: "Lorem ipsum dolor\nsit amet consectetur\nadipiscing elit",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, 'j', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'v', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'x', tcell.ModNone),
			},
			expectedCursorPos: 24,
			expectedText:      "Lorem ipsum dolor\nsit amectetur\nadipiscing elit",
		},
		{
			name:        "visual change selection",
			initialText: "Lorem ipsum dolor\nsit amet consectetur\nadipiscing elit",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'v', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'c', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'f', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'o', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'o', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'b', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'a', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'r', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyEsc, '\x00', tcell.ModNone),
			},
			expectedCursorPos: 8,
			expectedText:      "Lorfoobarm dolor\nsit amet consectetur\nadipiscing elit",
		},
		{
			name:        "visual mode change case",
			initialText: "Lorem ipsum dolor\nsit amet consectetur\nadipiscing elit",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'v', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, '~', tcell.ModNone),
			},
			expectedCursorPos: 2,
			expectedText:      "LoREM IPSUm dolor\nsit amet consectetur\nadipiscing elit",
		},
		{
			name:        "visual mode indent with count",
			initialText: "Lorem ipsum dolor\nsit amet consectetur\nadipiscing elit",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, 'j', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'V', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'j', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, '4', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, '>', tcell.ModNone),
			},
			expectedCursorPos: 22,
			expectedText:      "Lorem ipsum dolor\n\t\t\t\tsit amet consectetur\n\t\t\t\tadipiscing elit",
		},
		{
			name:        "visual mode indent, then repeat last action",
			initialText: "Lorem ipsum dolor\nsit amet consectetur\nadipiscing elit",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, 'j', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'V', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'j', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, '>', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, '.', tcell.ModNone),
			},
			expectedCursorPos: 20,
			expectedText:      "Lorem ipsum dolor\n\t\tsit amet consectetur\n\t\tadipiscing elit",
		},
		{
			name:        "visual mode outdent",
			initialText: "Lorem ipsum dolor\n\tsit amet consectetur\n\t\tadipiscing\nelit\n\n",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, 'j', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'V', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'j', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, '<', tcell.ModNone),
			},
			expectedCursorPos: 18,
			expectedText:      "Lorem ipsum dolor\nsit amet consectetur\n\tadipiscing\nelit\n\n",
		},
		{
			name:        "visual mode outdent with count",
			initialText: "Lorem ipsum dolor\n\t\tsit amet consectetur\n\t\t\tadipiscing\nelit\n\n",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, 'j', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'V', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'j', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, '2', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, '<', tcell.ModNone),
			},
			expectedCursorPos: 18,
			expectedText:      "Lorem ipsum dolor\nsit amet consectetur\n\tadipiscing\nelit\n\n",
		},
		{
			name:        "visual mode yank",
			initialText: "Lorem ipsum dolor\nsit amet consectetur\nadipiscing elit",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, 'V', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'j', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'y', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'j', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'j', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'p', tcell.ModNone),
			},
			expectedCursorPos: 55,
			expectedText:      "Lorem ipsum dolor\nsit amet consectetur\nadipiscing elit\nLorem ipsum dolor\nsit amet consectetur",
		},
		{
			name:        "visual mode yank to clipboard",
			initialText: "Lorem ipsum dolor\nsit amet consectetur\nadipiscing elit",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, 'V', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'j', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, '"', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'a', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'y', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'j', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'V', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'y', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'j', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'p', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyEsc, '\x00', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, '"', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'a', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'p', tcell.ModNone),
			},
			expectedCursorPos: 76,
			expectedText:      "Lorem ipsum dolor\nsit amet consectetur\nadipiscing elit\nsit amet consectetur\nLorem ipsum dolor\nsit amet consectetur",
		},
		{
			name:        "visual charwise to linewise, then toggle case",
			initialText: "Lorem ipsum dolor\nsit amet consectetur\nadipiscing elit",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, 'v', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'V', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, '~', tcell.ModNone),
			},
			expectedCursorPos: 0,
			expectedText:      "lOREM IPSUM DOLOR\nsit amet consectetur\nadipiscing elit",
		},
		{
			name:        "visual charwise to linewise, then toggle case",
			initialText: "Lorem ipsum dolor\nsit amet consectetur\nadipiscing elit",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, 'V', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'v', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, '~', tcell.ModNone),
			},
			expectedCursorPos: 0,
			expectedText:      "lOREM ipsum dolor\nsit amet consectetur\nadipiscing elit",
		},
		{
			name:        "visual mode, then replay last action",
			initialText: "",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, 'i', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'a', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyEnter, '\r', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyTab, '\t', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'b', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyEnter, '\r', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyTab, '\t', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyTab, '\t', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'c', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyEnter, '\r', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyTab, '\t', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyTab, '\t', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyTab, '\t', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'd', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyEsc, '\x00', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'k', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'k', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'V', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'j', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'j', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, '<', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, '.', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, '.', tcell.ModNone),
			},
			expectedCursorPos: 2,
			expectedText:      "a\nb\nc\nd",
		},
		{
			name:        "edit text, enter and exit visual mode, then replay last action",
			initialText: "abc",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, 'i', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'x', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyEsc, '\x00', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'v', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyEsc, '\x00', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, '.', tcell.ModNone),
			},
			expectedCursorPos: 0,
			expectedText:      "xxabc",
		},
		{
			name:        "visual mode escape to normal mode",
			initialText: "Lorem ipsum dolor\nsit amet consectetur\nadipiscing elit",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, 'V', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyEsc, '\x00', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'o', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 't', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'e', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 's', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 't', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyEsc, '\x00', tcell.ModNone),
			},
			expectedCursorPos: 21,
			expectedText:      "Lorem ipsum dolor\ntest\nsit amet consectetur\nadipiscing elit",
		},
		{
			name:        "visual mode select inner word",
			initialText: "abc def ghi",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'v', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'i', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'w', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, '~', tcell.ModNone),
			},
			expectedCursorPos: 0,
			expectedText:      "ABC def ghi",
		},
		{
			name:        "visual mode select a word",
			initialText: "abc def ghi",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'v', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, '2', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'a', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'w', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'x', tcell.ModNone),
			},
			expectedCursorPos: 0,
			expectedText:      "ghi",
		},
		{
			name:        "visual mode select inner string block with double-quotes",
			initialText: `"abc" x`,
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, 'v', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'i', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, '"', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'x', tcell.ModNone),
			},
			expectedCursorPos: 1,
			expectedText:      `"" x`,
		},
		{
			name:        "visual mode select a string block with double-quotes",
			initialText: `"abc" x`,
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, 'v', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'a', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, '"', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'x', tcell.ModNone),
			},
			expectedCursorPos: 0,
			expectedText:      ` x`,
		},
		{
			name:        "visual mode select inner string block with single-quotes",
			initialText: `'abc' x`,
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, 'v', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'i', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, '\'', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'x', tcell.ModNone),
			},
			expectedCursorPos: 1,
			expectedText:      `'' x`,
		},
		{
			name:        "visual mode select a string block with single-quotes",
			initialText: `'abc' x`,
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, 'v', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'a', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, '\'', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'x', tcell.ModNone),
			},
			expectedCursorPos: 0,
			expectedText:      ` x`,
		},
		{
			name:        "visual mode select inner string block with backtick-quotes",
			initialText: "`abc` x",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, 'v', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'i', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, '`', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'x', tcell.ModNone),
			},
			expectedCursorPos: 1,
			expectedText:      "`` x",
		},
		{
			name:        "visual mode select a string block with backtick-quotes",
			initialText: "`abc` x",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, 'v', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'a', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, '`', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'x', tcell.ModNone),
			},
			expectedCursorPos: 0,
			expectedText:      ` x`,
		},
		{
			name:        "visual mode select inner paren block",
			initialText: "(abc)",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'v', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'i', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'b', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'x', tcell.ModNone),
			},
			expectedCursorPos: 1,
			expectedText:      "()",
		},
		{
			name:        "visual mode select a paren block",
			initialText: "x (abc) yz",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, 'w', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'v', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'a', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'b', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'x', tcell.ModNone),
			},
			expectedCursorPos: 2,
			expectedText:      "x  yz",
		},
		{
			name:        "visual mode select inner brace block",
			initialText: "{abc}",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, 'v', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'i', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'B', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'x', tcell.ModNone),
			},
			expectedCursorPos: 1,
			expectedText:      "{}",
		},
		{
			name:        "visual mode select a brace block",
			initialText: "x {abc} y",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, '4', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'v', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'a', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'B', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'x', tcell.ModNone),
			},
			expectedCursorPos: 2,
			expectedText:      "x  y",
		},
		{
			name:        "visual mode select inner angle block",
			initialText: `<html><body /></html>`,
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, '9', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'v', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'i', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, '>', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'x', tcell.ModNone),
			},
			expectedCursorPos: 7,
			expectedText:      "<html><></html>",
		},
		{
			name:        "visual mode select then delete with delete key",
			initialText: `Lorem ipsum`,
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, 'v', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyDelete, '\x00', tcell.ModNone),
			},
			expectedCursorPos: 0,
			expectedText:      "em ipsum",
		},
		{
			name:        "visual mode select an angle block",
			initialText: `<html><body /></html>`,
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, '9', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'v', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'a', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, '<', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'x', tcell.ModNone),
			},
			expectedCursorPos: 6,
			expectedText:      "<html></html>",
		},
		{
			name:        "record and replay user macro with key binding",
			initialText: "Lorem ipsum dolor\nsit amet consectetur\nadipiscing elit",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, ':', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 's', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 't', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'a', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'r', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 't', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyEnter, '\r', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'i', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, '{', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyEsc, '\x00', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'A', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, '}', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, ',', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyEsc, '\x00', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'j', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, '^', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, ':', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 's', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 't', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'o', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'p', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyEnter, '\r', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, '2', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, '@', tcell.ModNone),
			},
			expectedCursorPos: 45,
			expectedText:      "{Lorem ipsum dolor},\n{sit amet consectetur},\n{adipiscing elit},",
		},
		{
			name:        "record and replay user macro with menu cmd",
			initialText: "Lorem ipsum dolor\nsit amet consectetur\nadipiscing elit",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, ':', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 's', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 't', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'a', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'r', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 't', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyEnter, '\r', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'i', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, '{', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyEsc, '\x00', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'A', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, '}', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, ',', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyEsc, '\x00', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'j', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, '^', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, ':', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 's', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 't', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'o', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'p', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyEnter, '\r', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, ':', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'r', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'e', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'p', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyEnter, '\r', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, '.', tcell.ModNone),
			},
			expectedCursorPos: 45,
			expectedText:      "{Lorem ipsum dolor},\n{sit amet consectetur},\n{adipiscing elit},",
		},
		{
			name:        "record and replay user macro with text search",
			initialText: "foo bar baz bat",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, ':', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 's', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 't', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'a', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'r', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 't', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyEnter, '\r', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, '/', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'b', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'a', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyEnter, '\r', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'x', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, ':', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 's', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 't', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'o', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'p', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyEnter, '\r', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, ':', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'r', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'e', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'p', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyEnter, '\r', tcell.ModNone),
			},
			expectedCursorPos: 7,
			expectedText:      "foo ar az bat",
		},
		{
			name:        "bracketed paste in insert mode",
			initialText: "abc",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, 'A', tcell.ModNone),
				tcell.NewEventPaste(true),
				tcell.NewEventKey(tcell.KeyRune, 'x', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'y', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'z', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyEnter, '\r', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyTab, '\t', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, '1', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, '2', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, '3', tcell.ModNone),
				tcell.NewEventPaste(false),
			},
			expectedCursorPos: 11,
			expectedText:      "abcxyz\n\t123",
		},
		{
			name:        "bracketed paste in search mode",
			initialText: "Lorem ipsum dolor\nsit amet consectetur\nadipiscing elit",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, '/', tcell.ModNone),
				tcell.NewEventPaste(true),
				tcell.NewEventKey(tcell.KeyRune, 'c', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'o', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'n', tcell.ModNone),
				tcell.NewEventPaste(false),
				tcell.NewEventKey(tcell.KeyEnter, '\r', tcell.ModNone),
			},
			expectedCursorPos: 27,
			expectedText:      "Lorem ipsum dolor\nsit amet consectetur\nadipiscing elit",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			interpreter := NewInterpreter()
			editorState := state.NewEditorState(100, 100, nil, nil)

			// Write the initial text to a temp file, which we will load into the editor.
			// Append a final "\n" to the contents as the POSIX end-of-file indicator.
			tmpFile, err := os.CreateTemp(t.TempDir(), "")
			require.NoError(t, err)
			path := tmpFile.Name()
			defer os.Remove(path)

			err = os.WriteFile(path, []byte(tc.initialText+"\n"), 0644)
			require.NoError(t, err)

			// Load the initial text tempfile.
			state.LoadDocument(
				editorState,
				path,
				false,
				func(state.LocatorParams) uint64 { return 0 },
			)

			// Replay the input events.
			for _, event := range tc.events {
				inputCtx := ContextFromEditorState(editorState)
				action := interpreter.ProcessEvent(event, inputCtx)
				action(editorState)
			}

			// Verify that the cursor position and text match what we expect.
			buffer := editorState.DocumentBuffer()
			reader := buffer.TextTree().ReaderAtPosition(0)
			data, err := io.ReadAll(&reader)
			require.NoError(t, err)
			text := string(data)
			assert.Equal(t, tc.expectedCursorPos, buffer.CursorPosition())
			assert.Equal(t, tc.expectedText, text)
		})
	}
}

func TestEnterAndExitVisualModeThenReplayLastAction(t *testing.T) {
	testCases := []struct {
		name               string
		enterVisualModeKey rune
	}{
		{
			name:               "charwise (v)",
			enterVisualModeKey: 'v',
		},
		{
			name:               "linewise (V)",
			enterVisualModeKey: 'V',
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			inputEvents := []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, tc.enterVisualModeKey, tcell.ModNone), // enter visual mode
				tcell.NewEventKey(tcell.KeyEsc, '\x00', tcell.ModNone),                 // exit visual mode
				tcell.NewEventKey(tcell.KeyRune, '.', tcell.ModNone),                   // replay last action
			}

			editorState := state.NewEditorState(100, 100, nil, nil)
			interpreter := NewInterpreter()
			for _, event := range inputEvents {
				inputCtx := ContextFromEditorState(editorState)
				action := interpreter.ProcessEvent(event, inputCtx)
				action(editorState)
			}

			assert.Equal(t, state.InputModeNormal, editorState.InputMode())
		})
	}
}

func inputEventsForBracketedPaste(s string) []tcell.Event {
	inputEvents := make([]tcell.Event, 0, len(s)+2)
	inputEvents = append(inputEvents, tcell.NewEventPaste(true))
	for _, r := range s {
		if r == '\n' {
			inputEvents = append(inputEvents, tcell.NewEventKey(tcell.KeyEnter, '\r', tcell.ModNone))
		} else if r == '\t' {
			inputEvents = append(inputEvents, tcell.NewEventKey(tcell.KeyTab, '\t', tcell.ModNone))
		} else {
			inputEvents = append(inputEvents, tcell.NewEventKey(tcell.KeyRune, r, tcell.ModNone))
		}
	}
	inputEvents = append(inputEvents, tcell.NewEventPaste(false))
	return inputEvents
}

func TestBracketedPasteInSearchMode(t *testing.T) {
	testCases := []struct {
		name          string
		pasteString   string
		expectedQuery string
	}{
		{
			name:          "short input",
			pasteString:   "abc",
			expectedQuery: "abc",
		},
		{
			name:          "long input, truncated",
			pasteString:   "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			expectedQuery: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		},
		{
			name:          "input with newline",
			pasteString:   "abc\nxyz\n123",
			expectedQuery: "abc",
		},
		{
			name:          "input with tab",
			pasteString:   "abc\txyz",
			expectedQuery: "abc\txyz",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			inputEvents := []tcell.Event{tcell.NewEventKey(tcell.KeyRune, '/', tcell.ModNone)}
			inputEvents = append(inputEvents, inputEventsForBracketedPaste(tc.pasteString)...)
			editorState := state.NewEditorState(100, 100, nil, nil)
			interpreter := NewInterpreter()
			for _, event := range inputEvents {
				inputCtx := ContextFromEditorState(editorState)
				action := interpreter.ProcessEvent(event, inputCtx)
				action(editorState)
			}

			assert.Equal(t, state.InputModeSearch, editorState.InputMode())
			searchQuery, _ := editorState.DocumentBuffer().SearchQueryAndDirection()
			assert.Equal(t, tc.expectedQuery, searchQuery)
		})
	}
}

func TestBracketedPasteInMenuMode(t *testing.T) {
	testCases := []struct {
		name          string
		pasteString   string
		expectedQuery string
	}{
		{
			name:          "short input",
			pasteString:   "abc",
			expectedQuery: "abc",
		},
		{
			name:          "long input, truncated",
			pasteString:   "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			expectedQuery: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		},
		{
			name:          "input with newline",
			pasteString:   "abc\nxyz\n123",
			expectedQuery: "abc",
		},
		{
			name:          "input with tab",
			pasteString:   "abc\txyz",
			expectedQuery: "abc\txyz",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			inputEvents := []tcell.Event{tcell.NewEventKey(tcell.KeyRune, ':', tcell.ModNone)}
			inputEvents = append(inputEvents, inputEventsForBracketedPaste(tc.pasteString)...)
			editorState := state.NewEditorState(100, 100, nil, nil)
			interpreter := NewInterpreter()
			for _, event := range inputEvents {
				inputCtx := ContextFromEditorState(editorState)
				action := interpreter.ProcessEvent(event, inputCtx)
				action(editorState)
			}

			assert.Equal(t, state.InputModeMenu, editorState.InputMode())
			assert.Equal(t, tc.expectedQuery, editorState.Menu().SearchQuery())
		})
	}
}

func TestBracketedPasteInNormalAndVisualMode(t *testing.T) {
	testCases := []struct {
		name string
		mode state.InputMode
	}{
		{name: "normal mode", mode: state.InputModeNormal},
		{name: "visual mode", mode: state.InputModeVisual},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var inputEvents []tcell.Event
			if tc.mode == state.InputModeVisual {
				inputEvents = append(inputEvents, tcell.NewEventKey(tcell.KeyRune, 'v', tcell.ModNone))
			}
			inputEvents = append(inputEvents, inputEventsForBracketedPaste("abc")...)

			editorState := state.NewEditorState(100, 100, nil, nil)
			interpreter := NewInterpreter()
			for _, event := range inputEvents {
				inputCtx := ContextFromEditorState(editorState)
				action := interpreter.ProcessEvent(event, inputCtx)
				action(editorState)
			}

			assert.Equal(t, tc.mode, editorState.InputMode())
			assert.Equal(t, state.StatusMsgStyleError, editorState.StatusMsg().Style)
			assert.Equal(t, "Cannot paste in this mode. Press 'i' to enter insert mode", editorState.StatusMsg().Text)
		})
	}
}

func TestBracketedPasteInUserMacro(t *testing.T) {
	var inputEvents []tcell.Event

	// Begin recording a user macro.
	inputEvents = append(inputEvents,
		tcell.NewEventKey(tcell.KeyRune, ':', tcell.ModNone),
		tcell.NewEventKey(tcell.KeyRune, 's', tcell.ModNone),
		tcell.NewEventKey(tcell.KeyRune, 't', tcell.ModNone),
		tcell.NewEventKey(tcell.KeyRune, 'a', tcell.ModNone),
		tcell.NewEventKey(tcell.KeyRune, 'r', tcell.ModNone),
		tcell.NewEventKey(tcell.KeyRune, 't', tcell.ModNone),
		tcell.NewEventKey(tcell.KeyEnter, '\r', tcell.ModNone))

	// Bracketed paste in insert mode.
	inputEvents = append(inputEvents, tcell.NewEventKey(tcell.KeyRune, 'i', tcell.ModNone))
	inputEvents = append(inputEvents, inputEventsForBracketedPaste("abc\n")...)
	inputEvents = append(inputEvents, tcell.NewEventKey(tcell.KeyEscape, '\x00', tcell.ModNone))

	// End macro.
	inputEvents = append(inputEvents,
		tcell.NewEventKey(tcell.KeyRune, ':', tcell.ModNone),
		tcell.NewEventKey(tcell.KeyRune, 's', tcell.ModNone),
		tcell.NewEventKey(tcell.KeyRune, 't', tcell.ModNone),
		tcell.NewEventKey(tcell.KeyRune, 'o', tcell.ModNone),
		tcell.NewEventKey(tcell.KeyRune, 'p', tcell.ModNone),
		tcell.NewEventKey(tcell.KeyEnter, '\r', tcell.ModNone))

	// Replay macro.
	inputEvents = append(inputEvents,
		tcell.NewEventKey(tcell.KeyRune, ':', tcell.ModNone),
		tcell.NewEventKey(tcell.KeyRune, 'r', tcell.ModNone),
		tcell.NewEventKey(tcell.KeyRune, 'e', tcell.ModNone),
		tcell.NewEventKey(tcell.KeyRune, 'p', tcell.ModNone),
		tcell.NewEventKey(tcell.KeyEnter, '\r', tcell.ModNone))

	editorState := state.NewEditorState(100, 100, nil, nil)
	interpreter := NewInterpreter()
	for _, event := range inputEvents {
		inputCtx := ContextFromEditorState(editorState)
		action := interpreter.ProcessEvent(event, inputCtx)
		action(editorState)
	}

	// Expect that the pasted text is inserted twice (once from original paste, once from macro replay).
	buffer := editorState.DocumentBuffer()
	reader := buffer.TextTree().ReaderAtPosition(0)
	data, err := io.ReadAll(&reader)
	require.NoError(t, err)
	text := string(data)
	assert.Equal(t, state.InputModeNormal, editorState.InputMode())
	assert.Equal(t, uint64(8), buffer.CursorPosition())
	assert.Equal(t, "abc\nabc\n", text)
}

func TestLoadGeneratedStateMachines(t *testing.T) {
	testCases := []struct {
		name string
		path string
	}{
		{name: "normal mode", path: NormalModePath},
		{name: "insert mode", path: InsertModePath},
		{name: "visual mode", path: VisualModePath},
		{name: "menu mode", path: MenuModePath},
		{name: "search mode", path: SearchModePath},
		{name: "task mode", path: TaskModePath},
		{name: "textfield mode", path: TextFieldModePath},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			data, err := generatedFiles.ReadFile(tc.path)
			require.NoError(t, err)
			require.NotPanics(t, func() {
				engine.Deserialize(data)
			})
		})
	}
}

func TestCountLimits(t *testing.T) {
	testCases := []string{
		"1025fx",
		"1025Fx",
		"1025tx",
		"1025Tx",
		"1025dd",
		"1025dl",
		"1025x",
		"1025dfx",
		"1025dFx",
		"1025dtx",
		"1025dTx",
		"1025cfx",
		"1025cFx",
		"1025ctx",
		"1025cTx",
		"1025.",
		"1025>>",
		"1025<<",
		"v33>",
		"v33<",
	}

	for _, tc := range testCases {
		t.Run(tc, func(t *testing.T) {
			interpreter := NewInterpreter()
			editorState := state.NewEditorState(100, 100, nil, nil)

			for _, r := range tc {
				event := tcell.NewEventKey(tcell.KeyRune, r, tcell.ModNone)
				inputCtx := ContextFromEditorState(editorState)
				action := interpreter.ProcessEvent(event, inputCtx)
				action(editorState)
			}

			msg := editorState.StatusMsg()
			assert.Equal(t, state.StatusMsgStyleError, msg.Style)
			assert.Contains(t, msg.Text, "count")
		})
	}
}

func TestTextFieldMode(t *testing.T) {
	interpreter := NewInterpreter()
	editorState := state.NewEditorState(100, 100, nil, nil)

	inputEvent := func(event tcell.Event) {
		inputCtx := ContextFromEditorState(editorState)
		action := interpreter.ProcessEvent(event, inputCtx)
		action(editorState)
	}

	// Enter menu mode, select "new document"
	inputEvent(tcell.NewEventKey(tcell.KeyRune, ':', tcell.ModNone))
	inputEvent(tcell.NewEventKey(tcell.KeyRune, 'n', tcell.ModNone))
	inputEvent(tcell.NewEventKey(tcell.KeyRune, 'e', tcell.ModNone))
	inputEvent(tcell.NewEventKey(tcell.KeyRune, 'w', tcell.ModNone))
	inputEvent(tcell.NewEventKey(tcell.KeyEnter, '\x00', tcell.ModNone))

	// Expect that we're in text field mode with a prompt.
	assert.Equal(t, state.InputModeTextField, editorState.InputMode())
	assert.Equal(t, "New document file path:", editorState.TextField().PromptText())
	assert.Equal(t, "", editorState.TextField().InputText())

	// Enter some text.
	inputEvent(tcell.NewEventKey(tcell.KeyRune, 't', tcell.ModNone))
	inputEvent(tcell.NewEventKey(tcell.KeyRune, 'e', tcell.ModNone))
	inputEvent(tcell.NewEventKey(tcell.KeyRune, 's', tcell.ModNone))
	inputEvent(tcell.NewEventKey(tcell.KeyRune, 't', tcell.ModNone))
	inputEvent(tcell.NewEventKey(tcell.KeyRune, '.', tcell.ModNone))
	inputEvent(tcell.NewEventKey(tcell.KeyRune, 't', tcell.ModNone))
	inputEvent(tcell.NewEventKey(tcell.KeyRune, 'x', tcell.ModNone))
	inputEvent(tcell.NewEventKey(tcell.KeyRune, 't', tcell.ModNone))

	// Expect that we're stil in text field mode, and input text is stored.
	assert.Equal(t, state.InputModeTextField, editorState.InputMode())
	assert.Equal(t, "test.txt", editorState.TextField().InputText())

	// Delete some text, then add a new extension.
	inputEvent(tcell.NewEventKey(tcell.KeyBackspace, '\x00', tcell.ModNone))
	inputEvent(tcell.NewEventKey(tcell.KeyBackspace, '\x00', tcell.ModNone))
	inputEvent(tcell.NewEventKey(tcell.KeyBackspace, '\x00', tcell.ModNone))
	inputEvent(tcell.NewEventKey(tcell.KeyRune, 'g', tcell.ModNone))
	inputEvent(tcell.NewEventKey(tcell.KeyRune, 'o', tcell.ModNone))

	// Expect updated input text.
	assert.Equal(t, state.InputModeTextField, editorState.InputMode())
	assert.Equal(t, "test.go", editorState.TextField().InputText())

	// Execute the action (load new file).
	inputEvent(tcell.NewEventKey(tcell.KeyEnter, '\x00', tcell.ModNone))

	// Expect back to normal mode, with the new file path loaded.
	assert.Equal(t, state.InputModeNormal, editorState.InputMode())
	assert.Equal(t, "test.go", editorState.FileWatcher().Path())
}

func BenchmarkNewInterpreter(b *testing.B) {
	for n := 0; n < b.N; n++ {
		NewInterpreter()
	}
}

func BenchmarkProcessEvent(b *testing.B) {
	benchmarks := []struct {
		name   string
		mode   state.InputMode
		events []tcell.Event
	}{
		{
			name: "i",
			mode: state.InputModeNormal,
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, 'i', tcell.ModNone),
			},
		},
		{
			name: "1234gg",
			mode: state.InputModeNormal,
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, '1', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, '2', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, '3', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, '4', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'g', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'g', tcell.ModNone),
			},
		},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			logWriter := log.Writer()
			defer func() {
				log.SetOutput(logWriter)
			}()
			log.SetOutput(io.Discard)

			interpreter := NewInterpreter()
			inputCtx := Context{InputMode: bm.mode}
			b.ResetTimer()
			for _, event := range bm.events {
				interpreter.ProcessEvent(event, inputCtx)
			}
		})
	}
}
