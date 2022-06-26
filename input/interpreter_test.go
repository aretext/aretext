package input

import (
	"io"
	"io/ioutil"
	"log"
	"os"
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/aretext/aretext/input/vm"
	"github.com/aretext/aretext/state"
)

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
			name:        "insert then delete",
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
		// NOTE: aretext deletes the empty line, but vim deletes the word after the empty line as well.
		{
			name:        "delete a word on an empty line with next line indented",
			initialText: "a\n\n    bcd",
			events: []tcell.Event{
				tcell.NewEventKey(tcell.KeyRune, 'j', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'd', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'a', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'w', tcell.ModNone),
			},
			expectedCursorPos: 6,
			expectedText:      "a\n    bcd",
		},
		// NOTE: aretext deletes the whitespace up to the word, but vim deletes the word after the whitespace as well.
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
			expectedText:      "abcd   ef",
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
			expectedCursorPos: 10,
			expectedText:      "ab   cd   ",
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
			name:        "change word",
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
				tcell.NewEventKey(tcell.KeyRune, 'i', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'p', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 's', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'u', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyEnter, '\r', tcell.ModNone),
			},
			expectedCursorPos: 6,
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
			expectedCursorPos: 16,
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
			name:        "record and replay user macro",
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
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			interpreter := NewInterpreter()
			editorState := state.NewEditorState(100, 100, nil, nil)

			// Write the initial text to a temp file, which we will load into the editor.
			// Append a final "\n" to the contents as the POSIX end-of-file indicator.
			tmpFile, err := ioutil.TempFile("", "")
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

func TestVerifyGeneratedPrograms(t *testing.T) {
	testCases := []struct {
		name string
		path string
	}{
		{name: "normal mode", path: NormalModeProgramPath},
		{name: "insert mode", path: InsertModeProgramPath},
		{name: "visual mode", path: VisualModeProgramPath},
		{name: "menu mode", path: MenuModeProgramPath},
		{name: "search mode", path: SearchModeProgramPath},
		{name: "task mode", path: TaskModeProgramPath},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			prog := mustLoadProgram(tc.path)
			err := vm.VerifyProgram(prog)
			assert.NoError(t, err)
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
