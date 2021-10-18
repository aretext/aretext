package state

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLastActionMacro(t *testing.T) {
	var executedNames []string
	buildAction := func(name string) MacroAction {
		return func(s *EditorState) {
			executedNames = append(executedNames, name)
		}
	}

	state := NewEditorState(100, 100, nil, nil)
	ReplayLastActionMacro(state)
	assert.Equal(t, 0, len(executedNames))

	AddToLastActionMacro(state, buildAction("a"))
	AddToLastActionMacro(state, buildAction("b"))
	ReplayLastActionMacro(state)
	assert.Equal(t, []string{"a", "b"}, executedNames)

	executedNames = nil
	ClearLastActionMacro(state)
	ReplayLastActionMacro(state)
	assert.Equal(t, 0, len(executedNames))

	AddToLastActionMacro(state, buildAction("c"))
	ReplayLastActionMacro(state)
	assert.Equal(t, []string{"c"}, executedNames)
}
