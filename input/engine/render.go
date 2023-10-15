package engine

import (
	"fmt"
	"strings"
)

// Hack to render a state machine as a graphviz dot file.
func Render(sm *StateMachine, eventLabelFunc func(start, end Event) string) string {
	var sb strings.Builder

	sb.WriteString(`digraph finite_state_machine {
	fontname="Helvetica,Arial,sans-serif"
	node [fontname="Helvetica,Arial,sans-serif"]
	edge [fontname="Helvetica,Arial,sans-serif"]
	rankdir=LR;
`)

	for state, cmdId := range sm.acceptCmd {
		s := fmt.Sprintf("	node [shape = doublecircle, label=\"%d\"]; %d;\n", cmdId, state)
		sb.WriteString(s)
	}

	sb.WriteString("	node [shape = circle, width=0.2, style=filled, label=\"%s\"];\n")
	for state, transitions := range sm.transitions {
		for _, t := range transitions {
			eventLabel := eventLabelFunc(t.eventRange.start, t.eventRange.end)
			s := fmt.Sprintf("	%d -> %d [label = \"%s\"];\n", state, t.nextState, eventLabel)
			sb.WriteString(s)
		}
	}

	sb.WriteString(`}`)

	return sb.String()
}
