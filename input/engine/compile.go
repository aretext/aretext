package engine

import (
	"fmt"
	"slices"
	"sort"
	"strconv"
	"strings"
)

// CmdId is a unique identifier for an input command.
type CmdId uint64

// CmdExpr is an expression for a given input command.
type CmdExpr struct {
	CmdId CmdId
	Expr  Expr
}

// Compile transforms a set of expressions (one for each input command) to a state machine.
func Compile(cmdExprs []CmdExpr) (*StateMachine, error) {
	if err := validateCmdExprs(cmdExprs); err != nil {
		return nil, err
	}

	nfa := compileCmdExprsToNFA(cmdExprs)
	sm := compileNFAToStateMachine(nfa)
	sm = minimizeStateMachine(sm)
	return sm, nil
}

func validateCmdExprs(cmdExprs []CmdExpr) error {
	if err := validateCmdIdsUnique(cmdExprs); err != nil {
		return err
	}

	for _, cmdExpr := range cmdExprs {
		err := validateExpr(cmdExpr.Expr, false)
		if err != nil {
			return fmt.Errorf("Invalid expression for cmd %d: %w", cmdExpr.CmdId, err)
		}
	}

	return nil
}

func validateCmdIdsUnique(cmdExprs []CmdExpr) error {
	cmdIds := make(map[CmdId]struct{}, len(cmdExprs))
	for _, cmdExpr := range cmdExprs {
		_, exists := cmdIds[cmdExpr.CmdId]
		if exists {
			return fmt.Errorf("Duplicate command ID detected: %d", cmdExpr.CmdId)
		}
		cmdIds[cmdExpr.CmdId] = struct{}{}
	}
	return nil
}

func validateExpr(expr Expr, inCapture bool) error {
	switch expr := expr.(type) {
	case EventExpr:
		break
	case EventRangeExpr:
		if expr.StartEvent >= expr.EndEvent {
			return fmt.Errorf("Invalid event range [%d, %d]", expr.StartEvent, expr.EndEvent)
		}
	case ConcatExpr:
		for _, child := range expr.Children {
			if err := validateExpr(child, inCapture); err != nil {
				return err
			}
		}
	case AltExpr:
		for _, child := range expr.Children {
			if err := validateExpr(child, inCapture); err != nil {
				return err
			}
		}
	case OptionExpr:
		return validateExpr(expr.Child, inCapture)
	case StarExpr:
		return validateExpr(expr.Child, inCapture)
	case CaptureExpr:
		if inCapture {
			return fmt.Errorf("Nested capture detected: %d", expr.CaptureId)
		}
		return validateExpr(expr.Child, true)
	default:
		return fmt.Errorf("Invalid expression type %T", expr)
	}

	return nil
}

// nfaNode represents one state in a non-deterministic finite automaton (NFA).
type nfaNode struct {
	acceptCmd        *CmdId // non-nil only for accept states.
	emptyTransitions []nfaEmptyTransition
	eventTransitions []nfaEventTransition
}

// nfaEmptyTransition represents an empty (a.k.a "epsilon") transition between NFA states.
type nfaEmptyTransition struct {
	nextNode *nfaNode
}

// nfaEventTransition represents a transition between NFA states triggered by an input event.
type nfaEventTransition struct {
	eventRange eventRange
	nextNode   *nfaNode
	captures   map[CmdId]CaptureId
}

// nfaSubsetEventTransition represents a transition from a subset of NFA states to another subset of NFA states.
type nfaSubsetEventTransition struct {
	eventRange eventRange
	nextNodes  []*nfaNode
	captures   map[CmdId]CaptureId
}

func searchNFA(root *nfaNode, selectFunc func(node *nfaNode) bool) []*nfaNode {
	var matches []*nfaNode
	seen := make(map[*nfaNode]struct{})
	stack := []*nfaNode{root}
	var node *nfaNode
	for len(stack) > 0 {
		node, stack = stack[len(stack)-1], stack[0:len(stack)-1]
		if _, ok := seen[node]; ok {
			continue
		}

		if selectFunc(node) {
			matches = append(matches, node)
		}

		seen[node] = struct{}{}

		for _, t := range node.emptyTransitions {
			stack = append(stack, t.nextNode)
		}

		for _, t := range node.eventTransitions {
			stack = append(stack, t.nextNode)
		}
	}

	return matches
}

func compileCmdExprsToNFA(cmdExprs []CmdExpr) *nfaNode {
	cmdNfaNodes := make([]*nfaNode, 0, len(cmdExprs))
	for _, cmdExpr := range cmdExprs {
		nfaNode := compileCmdExprToNFA(cmdExpr)
		cmdNfaNodes = append(cmdNfaNodes, nfaNode)
	}

	root := &nfaNode{}
	for _, child := range cmdNfaNodes {
		t := nfaEmptyTransition{nextNode: child}
		root.emptyTransitions = append(root.emptyTransitions, t)
	}
	return root
}

func compileCmdExprToNFA(cmdExpr CmdExpr) *nfaNode {
	cmdId := cmdExpr.CmdId

	var compileRec func(Expr) *nfaNode
	compileRec = func(expr Expr) *nfaNode {
		switch expr := expr.(type) {
		case EventExpr:
			return &nfaNode{
				eventTransitions: []nfaEventTransition{
					{
						eventRange: eventRange{
							start: expr.Event,
							end:   expr.Event,
						},
						nextNode: &nfaNode{acceptCmd: &cmdId},
						captures: make(map[CmdId]CaptureId),
					},
				},
			}

		case EventRangeExpr:
			return &nfaNode{
				eventTransitions: []nfaEventTransition{
					{
						eventRange: eventRange{
							start: expr.StartEvent,
							end:   expr.EndEvent,
						},
						nextNode: &nfaNode{acceptCmd: &cmdId},
						captures: make(map[CmdId]CaptureId),
					},
				},
			}

		case ConcatExpr:
			var root, prev *nfaNode
			for _, childExpr := range expr.Children {
				childNode := compileRec(childExpr)
				if root == nil {
					root = childNode
					prev = childNode
				} else {
					prevAcceptNodes := searchNFA(prev, func(node *nfaNode) bool { return node.acceptCmd != nil })
					for _, node := range prevAcceptNodes {
						node.acceptCmd = nil
						t := nfaEmptyTransition{nextNode: childNode}
						node.emptyTransitions = append(node.emptyTransitions, t)
					}
					prev = childNode
				}
			}
			return root

		case AltExpr:
			root := &nfaNode{}
			for _, childExpr := range expr.Children {
				childNode := compileRec(childExpr)
				t := nfaEmptyTransition{nextNode: childNode}
				root.emptyTransitions = append(root.emptyTransitions, t)
			}
			return root

		case OptionExpr:
			child := compileRec(expr.Child)
			acceptNode := &nfaNode{acceptCmd: &cmdId}
			t := nfaEmptyTransition{nextNode: acceptNode}
			childAcceptNodes := searchNFA(child, func(node *nfaNode) bool { return node.acceptCmd != nil })
			for _, node := range childAcceptNodes {
				node.acceptCmd = nil
				node.emptyTransitions = append(node.emptyTransitions, t)
			}
			child.emptyTransitions = append(child.emptyTransitions, t)
			return child

		case StarExpr:
			child := compileRec(expr.Child)
			acceptNode := &nfaNode{
				acceptCmd: &cmdId,
				emptyTransitions: []nfaEmptyTransition{
					{nextNode: child},
				},
			}
			t := nfaEmptyTransition{nextNode: acceptNode}
			childAcceptNodes := searchNFA(child, func(node *nfaNode) bool { return node.acceptCmd != nil })
			for _, node := range childAcceptNodes {
				node.acceptCmd = nil
				node.emptyTransitions = append(node.emptyTransitions, t)
			}
			child.emptyTransitions = append(child.emptyTransitions, t)
			return child

		case CaptureExpr:
			child := compileRec(expr.Child)
			allNodes := searchNFA(child, func(node *nfaNode) bool { return true })
			for _, node := range allNodes {
				for i := 0; i < len(node.eventTransitions); i++ {
					t := &node.eventTransitions[i]
					t.captures[cmdExpr.CmdId] = expr.CaptureId
				}
			}
			return child

		default:
			panic("Invalid expression type")
		}
	}

	return compileRec(cmdExpr.Expr)
}

func compileNFAToStateMachine(root *nfaNode) *StateMachine {
	// Helper to assign a new state for the DFA.
	var numDfaStates uint64
	allocateDfaStateId := func() stateId {
		s := stateId(numDfaStates)
		numDfaStates++
		return s
	}

	// Assign every NFA node a unique ID.
	nodeIdMap := make(map[*nfaNode]int)
	var id int
	allNodes := searchNFA(root, func(node *nfaNode) bool { return true })
	for _, node := range allNodes {
		nodeIdMap[node] = id
		id++
	}

	// Helper to construct a unique composite ID for a subset of NFA nodes.
	idForSubset := func(nodes []*nfaNode) string {
		var sb strings.Builder
		ids := make([]int, 0, len(nodes))
		for _, node := range nodes {
			ids = append(ids, nodeIdMap[node])
		}
		slices.Sort(ids)
		for _, id := range ids {
			sb.WriteString(strconv.Itoa(id))
			sb.WriteRune(':')
		}
		return sb.String()
	}

	// Helper to return the minimum command ID for a subset of NFA nodes.
	minAcceptCmdId := func(nodes []*nfaNode) *CmdId {
		var minAcceptCmd *CmdId
		for _, node := range nodes {
			if minAcceptCmd == nil || (node.acceptCmd != nil && *node.acceptCmd < *minAcceptCmd) {
				minAcceptCmd = node.acceptCmd
			}
		}
		return minAcceptCmd
	}

	// Helper to find all unique set of nodes reachable by empty transitions.
	emptyTransitionClosure := func(roots []*nfaNode) []*nfaNode {
		var closure []*nfaNode
		seen := make(map[*nfaNode]struct{})
		stack := append([]*nfaNode{}, roots...)
		var node *nfaNode
		for len(stack) > 0 {
			node, stack = stack[len(stack)-1], stack[0:len(stack)-1]
			if _, ok := seen[node]; ok {
				continue
			}
			closure = append(closure, node)
			seen[node] = struct{}{}

			for _, t := range node.emptyTransitions {
				stack = append(stack, t.nextNode)
			}
		}
		return closure
	}

	// Helper to calculate the union of multiple node subsets deterministically.
	unionSubsets := func(subsets ...[]*nfaNode) []*nfaNode {
		var union []*nfaNode
		seen := make(map[*nfaNode]struct{})
		for _, subsetNodes := range subsets {
			for _, node := range subsetNodes {
				if _, ok := seen[node]; !ok {
					union = append(union, node)
					seen[node] = struct{}{}
				}
			}
		}
		return union
	}

	// Helper to merge multiple capture maps into a single map.
	unionCaptures := func(captureMaps ...map[CmdId]CaptureId) map[CmdId]CaptureId {
		union := make(map[CmdId]CaptureId)
		for _, cm := range captureMaps {
			for cmdId, captureId := range cm {
				// Validation ensures that capture expressions aren't nested,
				// so we can assume cmdId isn't already set in the map.
				union[cmdId] = captureId
			}
		}
		return union
	}

	// Helper to resolve overlapping event ranges in a set of transitions.
	// The output has at most one transition for a given event.
	consolidateSubsetEventTransitions := func(subset []*nfaNode) []nfaSubsetEventTransition {
		var stack []nfaSubsetEventTransition
		for _, node := range subset {
			for _, t := range node.eventTransitions {
				stack = append(stack, nfaSubsetEventTransition{
					eventRange: t.eventRange,
					nextNodes:  []*nfaNode{t.nextNode},
					captures:   t.captures,
				})
			}
		}

		sort.SliceStable(stack, func(i, j int) bool {
			// Sort descending by start event, so the first transition popped
			// from the end of the stack has the smallest start event.
			return stack[i].eventRange.start > stack[j].eventRange.start
		})

		var transitions []nfaSubsetEventTransition
		var t1, t2 nfaSubsetEventTransition
		for len(stack) > 1 {
			t1, t2, stack = stack[len(stack)-1], stack[len(stack)-2], stack[0:len(stack)-2]
			if t1.eventRange.end < t2.eventRange.start || t1.eventRange.start > t2.eventRange.end {
				// t1 does not overlap t2, so output t1 and push t2 back on the stack.
				transitions = append(transitions, t1)
				stack = append(stack, t2)
			} else if t1.eventRange == t2.eventRange {
				// t1 and t2 have the same event range, so merge them and push back on the stack.
				merged := nfaSubsetEventTransition{
					eventRange: t1.eventRange,
					nextNodes:  unionSubsets(t1.nextNodes, t2.nextNodes),
					captures:   unionCaptures(t1.captures, t2.captures),
				}
				stack = append(stack, merged)
			} else if t1.eventRange.start < t2.eventRange.start {
				// t1 starts before t2, so split t1 up to t2 and push back on the stack.
				t0 := nfaSubsetEventTransition{
					eventRange: eventRange{
						start: t1.eventRange.start,
						end:   t2.eventRange.start - 1,
					},
					nextNodes: t1.nextNodes,
					captures:  t1.captures,
				}
				t1.eventRange.start = t2.eventRange.start
				stack = append(stack, t2, t1, t0)
			} else if t1.eventRange.end != t2.eventRange.end {
				// t1 and t2 start at the same event, but have different end events.
				// Split the longer of the two and push back on the stack.
				if t1.eventRange.end > t2.eventRange.end {
					t1, t2 = t2, t1
				}
				t3 := nfaSubsetEventTransition{
					eventRange: eventRange{
						start: t1.eventRange.end + 1,
						end:   t2.eventRange.end,
					},
					nextNodes: t2.nextNodes,
					captures:  t2.captures,
				}
				t2.eventRange.end = t1.eventRange.end
				stack = append(stack, t3, t2, t1)
			} else {
				// Can happen only if t1 and t2 are in the wrong order.
				panic("unexpected order of transitions in stack")
			}
		}

		if len(stack) > 0 {
			// Only one transition in the stack, so output it.
			transitions = append(transitions, stack[0])
		}

		return transitions
	}

	// Run the subset construction algorithm to convert the NFA to a DFA.
	sm := &StateMachine{
		acceptCmd:   make(map[stateId]CmdId),
		transitions: make(map[stateId][]transition),
	}
	seen := make(map[string]struct{})
	subsetIdToDfaStateId := make(map[string]stateId)
	var subset []*nfaNode
	stack := [][]*nfaNode{{root}}
	for len(stack) > 0 {
		subset, stack = stack[len(stack)-1], stack[0:len(stack)-1]
		subset = emptyTransitionClosure(subset)
		subsetId := idForSubset(subset)
		if _, ok := seen[subsetId]; ok {
			continue
		}

		dfaStateId, ok := subsetIdToDfaStateId[subsetId]
		if !ok {
			dfaStateId = allocateDfaStateId()
			subsetIdToDfaStateId[subsetId] = dfaStateId
		}

		if acceptCmdId := minAcceptCmdId(subset); acceptCmdId != nil {
			sm.acceptCmd[dfaStateId] = *acceptCmdId
		}

		for _, t := range consolidateSubsetEventTransitions(subset) {
			nextSubset := emptyTransitionClosure(t.nextNodes)
			nextSubsetId := idForSubset(nextSubset)
			nextDfaStateId, ok := subsetIdToDfaStateId[nextSubsetId]
			if !ok {
				nextDfaStateId = allocateDfaStateId()
				subsetIdToDfaStateId[nextSubsetId] = nextDfaStateId
			}

			var captures map[CmdId]CaptureId
			if len(t.captures) > 0 {
				captures = t.captures
			}

			sm.transitions[dfaStateId] = append(sm.transitions[dfaStateId], transition{
				eventRange: t.eventRange,
				nextState:  nextDfaStateId,
				captures:   captures,
			})

			stack = append(stack, nextSubset)
		}

		seen[subsetId] = struct{}{}
	}

	sm.numStates = numDfaStates
	return sm
}

// minimizeStateMachine combines redundant states to produce a new, minimal state machine.
// See 5.7 "Minimizing Finite-State Automata" in Grune & Jacobs (2006) Parsing Techniques.
func minimizeStateMachine(sm *StateMachine) *StateMachine {
	var partitions []map[stateId]struct{}
	stateToPartition := make(map[stateId]int)

	// Helper to create a new partition.
	allocatePartition := func() int {
		p := len(partitions)
		partitions = append(partitions, make(map[stateId]struct{}))
		return p
	}

	// Helper to add a state to a partition, removing it from any other partition first.
	addStateToPartition := func(state stateId, partition int) {
		if oldPartition, ok := stateToPartition[state]; ok {
			delete(partitions[oldPartition], state)
			delete(stateToPartition, state)
		}
		partitions[partition][state] = struct{}{}
		stateToPartition[state] = partition
	}

	// Helper to create a string key representing a transition to a partition.
	partitionTransitionKey := func(s stateId) string {
		var sb strings.Builder
		for _, t := range sm.transitions[s] {
			eventStartStr := strconv.FormatUint(uint64(t.eventRange.start), 10)
			sb.WriteString(eventStartStr)

			sb.WriteRune(':')
			eventEndStr := strconv.FormatUint(uint64(t.eventRange.end), 10)
			sb.WriteString(eventEndStr)

			sb.WriteRune(':')
			nextPartitionStr := strconv.Itoa(stateToPartition[t.nextState])
			sb.WriteString(nextPartitionStr)

			sb.WriteRune('|')
		}

		return sb.String()
	}

	// Helper to return the minimum state in a set of states.
	minimumState := func(states map[stateId]struct{}) stateId {
		if len(states) == 0 {
			panic("expected at least one state")
		}

		var i int
		var min stateId
		for s := range states {
			if i == 0 || s < min {
				min = s
			}
			i++
		}
		return min
	}

	// Initially one partition for each accept cmd ID, plus one for all non-accept states.
	acceptCmdToPartition := make(map[CmdId]int)
	nonAcceptPartition := -1
	for state := stateId(0); state < stateId(sm.numStates); state++ {
		if acceptCmdId, ok := sm.acceptCmd[state]; ok {
			partition, ok := acceptCmdToPartition[acceptCmdId]
			if !ok {
				partition = allocatePartition()
				acceptCmdToPartition[acceptCmdId] = partition
			}
			addStateToPartition(state, partition)
		} else {
			if nonAcceptPartition < 0 {
				nonAcceptPartition = allocatePartition()
			}
			addStateToPartition(state, nonAcceptPartition)
		}
	}

	// Iteratively split partitions until the partitions have consistent transitions.
	// "Consistent" means that all transitions for a given event range go to states in
	// the same partition.
retry:
	for _, partitionStates := range partitions {
		sortedPartitionStates := make([]stateId, 0, len(partitionStates))
		for s := range partitionStates {
			sortedPartitionStates = append(sortedPartitionStates, s)
		}
		slices.Sort(sortedPartitionStates)

		var lastKey string
		for _, s := range sortedPartitionStates {
			key := partitionTransitionKey(s)
			if lastKey != "" && key != lastKey {
				// Found a difference, so split the partition.
				var splitStates []stateId
				for _, otherState := range sortedPartitionStates {
					if partitionTransitionKey(otherState) == key {
						splitStates = append(splitStates, otherState)
					}
				}
				newPartition := allocatePartition()
				for _, ss := range splitStates {
					addStateToPartition(ss, newPartition)
				}
				goto retry
			}
			lastKey = key
		}
	}

	// All partitions have consistent transitions, so each partition becomes a state
	// in the minimized state machine.
	minimized := &StateMachine{}
	minimized.numStates = uint64(len(partitions))
	minimized.startState = stateId(stateToPartition[0]) // Original start state was zero, but could map to non-zero partition index.
	minimized.acceptCmd = make(map[stateId]CmdId)
	minimized.transitions = make(map[stateId][]transition, len(partitions))
	for p, states := range partitions {
		s := minimumState(states) // Arbitrarily choose the minimum state to represent this partition.

		var transitions []transition
		for _, t := range sm.transitions[s] {
			var mergedCaptures map[CmdId]CaptureId
			for otherState := range states {
				for _, otherTransition := range sm.transitions[otherState] {
					if otherTransition.eventRange == t.eventRange {
						for cmdId, captureId := range otherTransition.captures {
							if mergedCaptures == nil {
								mergedCaptures = make(map[CmdId]CaptureId)
							}
							mergedCaptures[cmdId] = captureId
						}
					}
				}
			}

			nextPartition := stateToPartition[t.nextState]
			transitions = append(transitions, transition{
				eventRange: t.eventRange,
				nextState:  stateId(nextPartition),
				captures:   mergedCaptures,
			})
		}

		if len(transitions) > 0 {
			minimized.transitions[stateId(p)] = transitions
		}

		if acceptCmdId, ok := sm.acceptCmd[s]; ok {
			minimized.acceptCmd[stateId(p)] = acceptCmdId
		}
	}

	return minimized
}
