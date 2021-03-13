package input

import (
	"math"
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/stretchr/testify/assert"
)

type acceptResult struct {
	input string
	count *int64
}

var five = int64(5)
var oneHundredAndTwo = int64(102)

func TestParser(t *testing.T) {
	rules := []Rule{
		{
			Name: "a",
			Pattern: []EventMatcher{
				{Key: tcell.KeyRune, Rune: 'a'},
			},
		},
		{
			Name: "ij",
			Pattern: []EventMatcher{
				{Key: tcell.KeyRune, Rune: 'i'},
				{Key: tcell.KeyRune, Rune: 'j'},
			},
		},
		{
			Name: "ik",
			Pattern: []EventMatcher{
				{Key: tcell.KeyRune, Rune: 'i'},
				{Key: tcell.KeyRune, Rune: 'k'},
			},
		},
		{
			Name: "xyz",
			Pattern: []EventMatcher{
				{Key: tcell.KeyRune, Rune: 'x'},
				{Key: tcell.KeyRune, Rune: 'y'},
				{Key: tcell.KeyRune, Rune: 'z'},
			},
		},
		{
			Name: "w with wildcard",
			Pattern: []EventMatcher{
				{Key: tcell.KeyRune, Rune: 'w'},
				{Wildcard: true},
			},
		},
		{
			Name: "special key",
			Pattern: []EventMatcher{
				{Key: tcell.KeyCtrlQ},
			},
		},
		{
			Name: "zero",
			Pattern: []EventMatcher{
				{Key: tcell.KeyRune, Rune: '0'},
			},
		},
	}

	testCases := []struct {
		name        string
		inputEvents []*tcell.EventKey
		expected    []acceptResult
	}{
		{
			name: "no match",
			inputEvents: []*tcell.EventKey{
				tcell.NewEventKey(tcell.KeyRune, 'm', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'n', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'o', tcell.ModNone),
			},
			expected: nil,
		},
		{
			name: "match single char",
			inputEvents: []*tcell.EventKey{
				tcell.NewEventKey(tcell.KeyRune, 'a', tcell.ModNone),
			},
			expected: []acceptResult{
				{input: "a"},
			},
		},
		{
			name: "match special key",
			inputEvents: []*tcell.EventKey{
				tcell.NewEventKey(tcell.KeyCtrlQ, '\x00', tcell.ModNone),
			},
			expected: []acceptResult{
				{input: "\x00"},
			},
		},
		{
			name: "match multiple chars",
			inputEvents: []*tcell.EventKey{
				tcell.NewEventKey(tcell.KeyRune, 'x', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'y', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'z', tcell.ModNone),
			},
			expected: []acceptResult{
				{input: "xyz"},
			},
		},
		{
			name: "match first rule with shared prefix",
			inputEvents: []*tcell.EventKey{
				tcell.NewEventKey(tcell.KeyRune, 'i', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'j', tcell.ModNone),
			},
			expected: []acceptResult{
				{input: "ij"},
			},
		},
		{
			name: "match second rule with shared prefix",
			inputEvents: []*tcell.EventKey{
				tcell.NewEventKey(tcell.KeyRune, 'i', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'k', tcell.ModNone),
			},
			expected: []acceptResult{
				{input: "ik"},
			},
		},
		{
			name: "match rule with wildcard",
			inputEvents: []*tcell.EventKey{
				tcell.NewEventKey(tcell.KeyRune, 'w', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'b', tcell.ModNone),
			},
			expected: []acceptResult{
				{input: "wb"},
			},
		},
		{
			name: "parse count with single digit",
			inputEvents: []*tcell.EventKey{
				tcell.NewEventKey(tcell.KeyRune, '5', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'a', tcell.ModNone),
			},
			expected: []acceptResult{
				{input: "5a", count: &five},
			},
		},
		{
			name: "parse count with multiple digits",
			inputEvents: []*tcell.EventKey{
				tcell.NewEventKey(tcell.KeyRune, '1', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, '0', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, '2', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'a', tcell.ModNone),
			},
			expected: []acceptResult{
				{input: "102a", count: &oneHundredAndTwo},
			},
		},
		{
			name: "reject first input then parse command",
			inputEvents: []*tcell.EventKey{
				tcell.NewEventKey(tcell.KeyRune, 'r', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'i', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'j', tcell.ModNone),
			},
			expected: []acceptResult{
				{input: "ij"},
			},
		},
		{
			name: "accept multiple commands",
			inputEvents: []*tcell.EventKey{
				tcell.NewEventKey(tcell.KeyRune, 'i', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'j', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'a', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'w', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'b', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'x', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'y', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'z', tcell.ModNone),
			},
			expected: []acceptResult{
				{input: "ij"},
				{input: "a"},
				{input: "wb"},
				{input: "xyz"},
			},
		},
		{
			name: "accept multiple commands with counts",
			inputEvents: []*tcell.EventKey{
				tcell.NewEventKey(tcell.KeyRune, 'i', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'j', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, '5', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'a', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'a', tcell.ModNone),
			},
			expected: []acceptResult{
				{input: "ij"},
				{input: "5a", count: &five},
				{input: "a"},
			},
		},
		{
			name: "accept multiple commands resetting on rejected input",
			inputEvents: []*tcell.EventKey{
				tcell.NewEventKey(tcell.KeyRune, 'i', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'r', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'j', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, '5', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'a', tcell.ModNone),
			},
			expected: []acceptResult{
				{input: "5a", count: &five},
			},
		},
		{
			name: "reset on special key with no rule",
			inputEvents: []*tcell.EventKey{
				tcell.NewEventKey(tcell.KeyRune, '2', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'i', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyEscape, '\x00', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, '5', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'a', tcell.ModNone),
			},
			expected: []acceptResult{
				{input: "5a", count: &five},
			},
		},
		{
			name: "zero command",
			inputEvents: []*tcell.EventKey{
				tcell.NewEventKey(tcell.KeyRune, '0', tcell.ModNone),
			},
			expected: []acceptResult{
				{input: "0"},
			},
		},
		{
			name: "zero command followed by command with count",
			inputEvents: []*tcell.EventKey{
				tcell.NewEventKey(tcell.KeyRune, '0', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, '5', tcell.ModNone),
				tcell.NewEventKey(tcell.KeyRune, 'a', tcell.ModNone),
			},
			expected: []acceptResult{
				{input: "0"},
				{input: "5a", count: &five},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var acceptResults []acceptResult
			parser := NewParser(rules)
			for _, e := range tc.inputEvents {
				result := parser.ProcessInput(e)
				if result.Accepted {
					inputRunes := make([]rune, len(result.Input))
					for i, e := range result.Input {
						inputRunes[i] = e.Rune()
					}
					acceptResults = append(acceptResults, acceptResult{
						input: string(inputRunes),
						count: result.Count,
					})
				}
			}
			assert.Equal(t, tc.expected, acceptResults)
		})
	}
}

func TestParseMaxInputLen(t *testing.T) {
	rules := []Rule{
		{
			Name: "a",
			Pattern: []EventMatcher{
				{Key: tcell.KeyRune, Rune: 'a'},
			},
		},
	}
	parser := NewParser(rules)
	for i := 0; i < maxParseInputLen; i++ {
		e := tcell.NewEventKey(tcell.KeyRune, '5', tcell.ModNone)
		result := parser.ProcessInput(e)
		assert.False(t, result.Accepted)
	}

	e := tcell.NewEventKey(tcell.KeyRune, 'a', tcell.ModNone)
	result := parser.ProcessInput(e)
	assert.True(t, result.Accepted)
	assert.Equal(t, 1, len(result.Input))
	assert.Equal(t, 'a', result.Input[0].Rune())
}

func TestParseCountOverflow(t *testing.T) {
	rules := []Rule{
		{
			Name: "a",
			Pattern: []EventMatcher{
				{Key: tcell.KeyRune, Rune: 'a'},
			},
		},
	}
	parser := NewParser(rules)
	for i := 0; i < 100; i++ {
		e := tcell.NewEventKey(tcell.KeyRune, '9', tcell.ModNone)
		result := parser.ProcessInput(e)
		assert.False(t, result.Accepted)
	}

	e := tcell.NewEventKey(tcell.KeyRune, 'a', tcell.ModNone)
	result := parser.ProcessInput(e)
	assert.True(t, result.Accepted)
	assert.Equal(t, 37, len(result.Input))
	assert.True(t, result.Count != nil)
	assert.Equal(t, int64(math.MaxInt64), *result.Count)
}
