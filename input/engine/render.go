package engine

import (
	"fmt"
	"strings"
)

// Hack to render a state machine as a graphviz dot file.
func Render(sm *StateMachine, eventLabelFunc func(start, end Event) string, cmdLabelFunc func(CmdId) string) string {
	var sb strings.Builder

	sb.WriteString(`digraph finite_state_machine {
	fontname="Helvetica,Arial,sans-serif"
	node [fontname="Helvetica,Arial,sans-serif"]
	edge [fontname="Helvetica,Arial,sans-serif"]
	rankdir=LR;
	overlap="orthoxyz";
`)

	// How many states do we have total?
	var maxState stateId
	for s := range sm.transitions {
		if s > maxState {
			maxState = s
		}
	}

	// Node for each state
	for state := stateId(0); state <= maxState; state++ {
		var s string
		if cmdId, ok := sm.acceptCmd[state]; ok {
			cmdLabel := cmdLabelFunc(cmdId)
			s = fmt.Sprintf("	node [shape = box, label=\"%s\", fillcolor=\"#FFFFFF\"]; %d;\n", cmdLabel, state)
		} else {
			s = fmt.Sprintf("	node [shape = circle, width=0.2, style=filled, label=\"\", fillcolor=\"#FFFFFF\"]; %d;\n", state)
		}
		sb.WriteString(s)
	}

	// huge hack, pseudo-states for each edge
	pseudoState := 1000000
	for state, transitions := range sm.transitions {
		for _, t := range transitions {
			eventLabel := eventLabelFunc(t.eventRange.start, t.eventRange.end)

			// hack: split each edge into two so we can inject the label as an intermediate node.
			e1 := fmt.Sprintf("  %d -> %d [dir=none];\n", state, pseudoState)
			s1 := fmt.Sprintf("  node [shape = \"box\", style=\"filled\", fillcolor=\"#E6E6E6\", color=\"#FFFFFF\", label=\"%s\"]; %d;\n", eventLabel, pseudoState)

			e2 := fmt.Sprintf("	%d -> %d [dir=forward, headport=\"w\"];\n", pseudoState, t.nextState)
			if state == t.nextState {
				e2 = fmt.Sprintf("	%d -> %d [dir=forward];\n", pseudoState, t.nextState)
			}

			sb.WriteString(s1)
			sb.WriteString(e1)
			sb.WriteString(e2)

			pseudoState++
		}
	}

	sb.WriteString(`}`)

	return sb.String()
}
