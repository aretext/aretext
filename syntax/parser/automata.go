package parser

import (
	"io"

	"github.com/aretext/aretext/text/utf8"
)

// Nfa is a non-deterministic finite automaton.
type Nfa struct {
	// The states of the NFA.
	// There is always at least one state: the start state at index 0.
	states []*nfaState
}

// EmptyLanguageNfa returns an NFA that matches no strings (the empty language).
func EmptyLanguageNfa() *Nfa {
	startState := newNfaState(false)
	return &Nfa{states: []*nfaState{startState}}
}

// EmptyStringNfa returns an NFA that matches only the empty string.
func EmptyStringNfa() *Nfa {
	nfa := EmptyLanguageNfa()
	nfa.states[0].accept = true
	return nfa
}

// NfaForStartOfText returns an NFA that matches the start of the text.
func NfaForStartOfText() *Nfa {
	nfa := EmptyLanguageNfa()
	nfa.states = append(nfa.states, newNfaState(true))
	nfa.states[0].inputTransitions[startOfText] = []int{1}
	return nfa
}

// NfaForEndOfText returns an NFA that matches the end of the text.
func NfaForEndOfText() *Nfa {
	nfa := EmptyLanguageNfa()
	nfa.states = append(nfa.states, newNfaState(true))
	nfa.states[0].inputTransitions[endOfText] = []int{1}
	return nfa
}

// NfaForChars returns an NFA that matches any of the specified chars.
func NfaForChars(chars []byte) *Nfa {
	nfa := EmptyLanguageNfa()
	nfa.states = append(nfa.states, newNfaState(true))
	startState := nfa.states[0]
	for _, c := range chars {
		startState.inputTransitions[int(c)] = []int{1}
	}

	return nfa
}

// NfaForNegatedChars returns an NFA that matches any char EXCEPT the specified chars.
func NfaForNegatedChars(negatedChars []byte) *Nfa {
	nfa := NfaForChars(negatedChars)
	startState := nfa.states[0]

	// Negate all transitions from the start state to the accept state.
	for c := 0; c < 256; c++ {
		states := startState.inputTransitions[c]
		if len(states) > 0 {
			startState.inputTransitions[c] = nil
		} else {
			startState.inputTransitions[c] = []int{1}
		}
	}

	return nfa
}

// SetAcceptAction sets all accept states in the NFA to the specified action.
// This overwrites any actions set previously.
func (nfa *Nfa) SetAcceptAction(action int) *Nfa {
	for _, state := range nfa.states {
		state.acceptActions = []int{action}
	}
	return nfa
}

// Star applies the Kleene star operation to the NFA.
func (nfa *Nfa) Star() *Nfa {
	nfa.states[0].accept = true
	for _, state := range nfa.states {
		if state.accept {
			// Add an empty transition from the accept state back to the start state.
			state.emptyTransitions = insertUniqueSorted(state.emptyTransitions, 0)
		}
	}
	return nfa
}

// Union constructs an NFA from the union of two NFAs.
func (left *Nfa) Union(right *Nfa) *Nfa {
	nfa := EmptyLanguageNfa()

	// Copy the left NFA states into the new NFA, shifting the state numbers
	// by one to account for the new start state.
	for _, state := range left.states {
		newState := state.copyWithShiftedTransitions(1)
		nfa.states = append(nfa.states, newState)
	}

	// Copy the right NFA states into the new NFA, shifting the state numbers
	// to account for both the new start state and the left NFA's states.
	for _, state := range right.states {
		newState := state.copyWithShiftedTransitions(1 + len(left.states))
		nfa.states = append(nfa.states, newState)
	}

	// Add empty transitions from the new start state to start of the left and right NFAs.
	startState := nfa.states[0]
	startState.emptyTransitions = insertUniqueSorted(startState.emptyTransitions, 1)
	startState.emptyTransitions = insertUniqueSorted(startState.emptyTransitions, 1+len(left.states))

	return nfa
}

// Concat constructs an NFA from the concatenation of two NFAs.
// The new NFA has accept actions from both the left and right NFAs.
func (left *Nfa) Concat(right *Nfa) *Nfa {
	nfa := EmptyLanguageNfa()

	// Copy all states from the left NFA, shifting the state numbers to account
	// for new start state.
	leftAcceptActionSet := make(map[int]struct{}, 0)
	for _, state := range left.states {
		newState := state.copyWithShiftedTransitions(1)
		if newState.accept {
			newState.accept = false
			newState.emptyTransitions = insertUniqueSorted(newState.emptyTransitions, len(left.states)+1)

			// Collect the accept actions so we can add them to the new accept states later.
			for _, action := range newState.acceptActions {
				leftAcceptActionSet[action] = struct{}{}
			}
		}
		nfa.states = append(nfa.states, newState)
	}

	// Copy all states from the right NFA, shifting the state numbers to account
	// for both the new start state and the left NFA's states.
	for _, state := range right.states {
		newState := state.copyWithShiftedTransitions(len(left.states) + 1)
		if newState.accept {
			// Include the accept actions from the left NFA in the new accept states.
			for action, _ := range leftAcceptActionSet {
				newState.acceptActions = insertUniqueSorted(newState.acceptActions, action)
			}
		}

		nfa.states = append(nfa.states, newState)
	}

	// Add an empty transition from the new start state to the left NFA's start state.
	startState := nfa.states[0]
	startState.emptyTransitions = insertUniqueSorted(startState.emptyTransitions, 1)

	return nfa
}

// CompileDfa compiles the NFA into an equivalent deterministic finite automaton.
// The DFA has the minimum possible number of states.
// Only accept states with at least one accept action are accepted by the DFA,
// so make sure to call SetAcceptAction at least once before compiling!
func (nfa *Nfa) CompileDfa() *Dfa {
	dfa := NewDfa()

	// Construct the DFA using the subset construction algorithm.
	// See Aho, Alfred V. (2003). Compilers: Principles, Techniques and Tools
	// or Grune, D. (2008). Parsing Techniques: A Practical Guide.
	dfaStateToNfaStates := map[int][]int{
		dfa.StartState: nfa.emptyTransitionsClosure([]int{0}),
	}

	stateSetToDfaState := map[string]int{
		intSliceKey([]int{0}): dfa.StartState,
	}

	var dfaState int
	stack := []int{dfa.StartState}
	for len(stack) > 0 {
		dfaState, stack = stack[len(stack)-1], stack[:len(stack)-1]

		// If any of the corresponding states in the NFA are accept states,
		// mark the new state in the DFA as an accept state.
		for _, nfaStateIdx := range dfaStateToNfaStates[dfaState] {
			nfaState := nfa.states[nfaStateIdx]
			if nfaState.accept {
				for _, action := range nfaState.acceptActions {
					dfa.AddAcceptAction(dfaState, action)
				}
			}
		}

		// Follow transitions from the current state to the next states.
		// If those states don't exist yet in the DFA, create them and
		// push them onto the stack.
		for i := 0; i < maxTransitionsPerState; i++ {
			var nextNfaStates []int
			for _, nfaStateIdx := range dfaStateToNfaStates[dfaState] {
				nfaState := nfa.states[nfaStateIdx]
				for _, s := range nfaState.inputTransitions[i] {
					nextNfaStates = insertUniqueSorted(nextNfaStates, s)
				}
			}

			if len(nextNfaStates) == 0 {
				continue
			}

			// Include all states reachable through empty transitions in the new state.
			nextNfaStates = nfa.emptyTransitionsClosure(nextNfaStates)

			if nextDfaState, ok := stateSetToDfaState[intSliceKey(nextNfaStates)]; ok {
				// The new state already exists, so update it with the new transitions.
				dfa.AddTransition(dfaState, i, nextDfaState)
			} else {
				// The new state does not yet exist, so create it and push it onto the stack.
				nextDfaState = dfa.AddState()
				dfa.AddTransition(dfaState, i, nextDfaState)
				dfaStateToNfaStates[nextDfaState] = nextNfaStates
				stateSetToDfaState[intSliceKey(nextNfaStates)] = nextDfaState
				stack = append(stack, nextDfaState)
			}
		}
	}

	return dfa.Minimize()
}

// emptyTransitionsClosure returns all states reachable from the current states through empty transitions.
func (nfa *Nfa) emptyTransitionsClosure(startStates []int) []int {
	var state int
	reachedStates := make(map[int]struct{}, 0)
	stack := append([]int{}, startStates...)
	for len(stack) > 0 {
		state, stack = stack[len(stack)-1], stack[:len(stack)-1]
		if _, ok := reachedStates[state]; !ok {
			reachedStates[state] = struct{}{}
			for _, nextState := range nfa.states[state].emptyTransitions {
				stack = append(stack, nextState)
			}
		}
	}
	return sortedKeys(reachedStates)
}

// nfaState represents a state in an NFA.
type nfaState struct {
	// Transition from the current state based on an input to other state(s).
	// Inputs can be either input text bytes or start-of-text/end-of-text markers.
	inputTransitions map[int][]int

	// Empty transitions from the current state to other states.
	// The index of the slice is the next state (after the transition).
	// These are sometimes called epsilon transitions.
	emptyTransitions []int

	// Whether this state is an accept state.
	accept bool

	// What actions to take if this state is reached and is an accept state.
	// Note that this may be non-nil even if accept is False; in that case,
	// the value should be ignored.
	acceptActions []int
}

func newNfaState(accept bool) *nfaState {
	return &nfaState{
		accept:           accept,
		inputTransitions: make(map[int][]int, maxTransitionsPerState),
	}
}

// copyWithShiftedTransitions returns a copy with the state indices in transitions incremented by n.
func (state *nfaState) copyWithShiftedTransitions(n int) *nfaState {
	newState := &nfaState{
		inputTransitions: make(map[int][]int, maxTransitionsPerState),
		emptyTransitions: make([]int, 0, len(state.emptyTransitions)),
		accept:           state.accept,
		acceptActions:    append([]int{}, state.acceptActions...),
	}

	for c, transitions := range state.inputTransitions {
		for _, nextState := range transitions {
			newState.inputTransitions[c] = insertUniqueSorted(newState.inputTransitions[c], nextState+n)
		}
	}

	for _, nextState := range state.emptyTransitions {
		newState.emptyTransitions = insertUniqueSorted(newState.emptyTransitions, nextState+n)
	}

	return newState
}

// DfaDeadState represents a state in which the DFA will never accept the string,
// regardless of the remaining input characters.
const DfaDeadState int = 0

// maxTransitionsPerState is the maximum number of transitions from a DFA state.
// 0-255 are transitions on an input char byte.
// 256 is a transition on start-of-text, and 257 is a transition on end-of-text.
const maxTransitionsPerState int = 258
const startOfText int = 256
const endOfText int = 257

// Dfa is a deterministic finite automata.
type Dfa struct {
	// Number of states in the DFA.
	// States are numbered sequentially, starting from zero.
	// State zero is the dead state (input rejected).
	NumStates int

	// The start state of the DFA.
	StartState int

	// Transition based on current state and next input byte.
	// Indices are (currentStateIdx * 256 + inputChar)
	Transitions []int

	// Actions to perform on an accept state.
	// The actions are defined by the user of the DFA.
	AcceptActions [][]int

	// buf is an internal buffer, re-used to amortize the allocation cost.
	buf [16]byte
}

// NewDfa constructs a DFA with only the dead state and the start state.
// This recognizes the empty language (rejects all strings, even the empty string).
func NewDfa() *Dfa {
	numStates := 2 // The dead state and the start state.
	return &Dfa{
		NumStates:     numStates,
		StartState:    1,
		Transitions:   make([]int, maxTransitionsPerState*numStates),
		AcceptActions: make([][]int, numStates),
	}
}

// AddState adds a new state to the DFA, returning the state index.
func (dfa *Dfa) AddState() int {
	state := dfa.NumStates
	dfa.NumStates++
	transitions := make([]int, maxTransitionsPerState)
	dfa.Transitions = append(dfa.Transitions, transitions...)
	dfa.AcceptActions = append(dfa.AcceptActions, nil)
	return state
}

// AddTransition adds a transition from one state to another based on an input character.
func (dfa *Dfa) AddTransition(fromState int, onInput int, toState int) {
	dfa.Transitions[fromState*maxTransitionsPerState+onInput] = toState
}

// AddAcceptAction adds an accept action to take when a state is reached.
// This marks the state as an accept state.
func (dfa *Dfa) AddAcceptAction(state int, action int) {
	dfa.AcceptActions[state] = insertUniqueSorted(dfa.AcceptActions[state], action)
}

// MatchLongest returns the longest match in an input string.
// In some cases, the longest match could be empty (e.g. the regular language for "a*" matches the empty string at the beginning of the string "bbb").
// The reader position is reset to the end of the match, if there is one, or its original position if not.
func (dfa *Dfa) MatchLongest(r InputReader, startPos uint64, textLen uint64) (accepted bool, endPos uint64, lookaheadPos uint64, actions []int, numBytesReadAtLastAccept int, err error) {
	var totalBytesRead int
	pos := startPos
	state := dfa.StartState

	if acceptActions := dfa.AcceptActions[state]; len(acceptActions) > 0 {
		accepted, endPos, actions = true, pos, acceptActions
	}

	if startPos == 0 {
		prevState := state
		state = dfa.Transitions[state*maxTransitionsPerState+startOfText]
		if state == DfaDeadState {
			// The DFA doesn't match start-of-text, so try to recover by restarting at the first character.
			state = prevState
		} else if acceptActions := dfa.AcceptActions[state]; len(acceptActions) > 0 {
			accepted, endPos, actions = true, pos, acceptActions
		}
	}

	for {
		n, err := r.Read(dfa.buf[:])
		if err != nil && err != io.EOF {
			return false, 0, 0, nil, 0, err
		}

		prevTotalBytesRead := totalBytesRead
		totalBytesRead += n

		for i, c := range dfa.buf[:n] {
			pos += uint64(utf8.StartByteIndicator[c])
			state = dfa.Transitions[state*maxTransitionsPerState+int(c)]
			if state == DfaDeadState {
				break
			} else if acceptActions := dfa.AcceptActions[state]; len(acceptActions) > 0 {
				accepted, endPos, numBytesReadAtLastAccept, actions = true, pos, prevTotalBytesRead+i+1, acceptActions
			}
		}

		if err == io.EOF || state == DfaDeadState {
			break
		}
	}

	if pos == textLen {
		state = dfa.Transitions[state*maxTransitionsPerState+endOfText]
		if acceptActions := dfa.AcceptActions[state]; len(acceptActions) > 0 {
			accepted, endPos, numBytesReadAtLastAccept, actions = true, pos, totalBytesRead, acceptActions
		}
	}

	// Reset the reader position to the end of the last match.
	// If there was no match, this resets the reader to its original position.
	if err := r.SeekBackward(uint64(totalBytesRead - numBytesReadAtLastAccept)); err != nil {
		return false, 0, 0, nil, 0, err
	}

	return accepted, endPos, pos, actions, numBytesReadAtLastAccept, nil
}

// Minimize produces an equivalent DFA with the minimum number of states.
func (dfa *Dfa) Minimize() *Dfa {
	// This is the DFA state minimization algorithm from
	// Aho, Alfred V. (2003). Compilers: Principles, Techniques and Tools
	groups := dfa.groupEquivalentStates()
	return dfa.newDfaFromGroups(groups)
}

func (dfa *Dfa) groupEquivalentStates() [][]int {
	groups := dfa.initialGroups()
	for {
		prevNumGroups := len(groups)
		groups = dfa.splitGroupsIfNecessary(groups)
		if prevNumGroups == len(groups) {
			// The groups haven't changed, so they cannot be further split.
			break
		}
	}

	return groups
}

func (dfa *Dfa) initialGroups() [][]int {
	groups := make([][]int, 0, dfa.NumStates)

	// The dead state must come first so that it becomes the first state in the new DFA.
	groups = append(groups, []int{DfaDeadState})

	// Partition the remaining groups by their accept actions.
	partitions := make(map[string][]int, dfa.NumStates)
	for s := 1; s < dfa.NumStates; s++ {
		key := intSliceKey(dfa.AcceptActions[s])
		partitions[key] = append(partitions[key], s)
	}

	forEachPartitionInKeyOrder(partitions, func(states []int) {
		groups = append(groups, states)
	})

	return groups
}

func (dfa *Dfa) indexStatesByGroup(groups [][]int) []int {
	stateToGroup := make([]int, dfa.NumStates)
	for g, states := range groups {
		for _, s := range states {
			stateToGroup[s] = g
		}
	}
	return stateToGroup
}

func (dfa *Dfa) splitGroupsIfNecessary(groups [][]int) [][]int {
	stateToGroup := dfa.indexStatesByGroup(groups)
	newGroups := make([][]int, 0, len(groups))
	for g := 0; g < len(groups); g++ {
		states := groups[g]

		// Cannot split a group with a single state.
		// By construction, all groups must have at least one state.
		if len(states) == 1 {
			newGroups = append(newGroups, states)
			continue
		}

		// Partition states in the group by their transitions to other groups.
		// Two states with the same transitions are considered
		// equivalent and should remain in the same group.
		partitions := make(map[string][]int, len(states))
		for _, s := range states {
			var nextStateGroups [maxTransitionsPerState]int
			for c, nextState := range dfa.Transitions[s*maxTransitionsPerState : (s+1)*maxTransitionsPerState] {
				nextStateGroups[c] = stateToGroup[nextState]
			}
			key := intSliceKey(nextStateGroups[:])
			partitions[key] = append(partitions[key], s)
		}

		forEachPartitionInKeyOrder(partitions, func(states []int) {
			newGroups = append(newGroups, states)
		})
	}

	return newGroups
}

func (dfa *Dfa) newDfaFromGroups(groups [][]int) *Dfa {
	stateToGroup := dfa.indexStatesByGroup(groups)

	newDfa := &Dfa{
		Transitions:   make([]int, 0, len(groups)*maxTransitionsPerState),
		AcceptActions: make([][]int, len(groups)),
	}

	// Add states, one for each group.
	groupToNewState := make([]int, len(groups))
	for g := 0; g < len(groups); g++ {
		newState := newDfa.AddState()
		groupToNewState[g] = newState
	}

	// Add transitions and accept actions.
	for g, states := range groups {
		newState := groupToNewState[g]
		s := states[0] // Arbitrarily choose the first state as representative.

		for i, nextState := range dfa.Transitions[s*maxTransitionsPerState : (s+1)*maxTransitionsPerState] {
			nextGroup := stateToGroup[nextState]
			newNextState := groupToNewState[nextGroup]
			newDfa.AddTransition(newState, i, newNextState)
		}

		for _, a := range dfa.AcceptActions[s] {
			newDfa.AddAcceptAction(newState, a)
		}

		for _, state := range states {
			if state == dfa.StartState {
				startGroup := stateToGroup[state]
				newStartState := groupToNewState[startGroup]
				newDfa.StartState = newStartState
				break
			}
		}
	}

	return newDfa
}
