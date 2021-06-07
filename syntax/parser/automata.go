package parser

import (
	"io"
	"unicode"
	"unicode/utf8"

	textUtf8 "github.com/aretext/aretext/text/utf8"
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

// NfaForUnicodeCategory constructs an NFA that matches UTF-8 encoded runes in a unicode category (letter, digit, etc.).
func NfaForUnicodeCategory(rangeTable *unicode.RangeTable) *Nfa {
	var buf [4]byte
	builder := NewDfaBuilder()

	// Construct the initial automata as a trie of the bytes in each rune.
	// Runes that share the same byte prefix will share states/transitions.
	iterRunes(rangeTable, func(r rune) {
		state := builder.StartState()
		n := utf8.EncodeRune(buf[:], r)
		for _, b := range buf[:n] {
			nextState := builder.NextState(state, int(b))
			if nextState == DfaDeadState {
				nextState = builder.AddState()
				builder.AddTransition(state, int(b), nextState)
			}
			state = nextState
		}
		builder.AddAcceptAction(state, 1) // Choose 1 arbitrarily because we'll discard it later.
	})

	// Some categories contain hundreds of thousands of runes, so the automata may have a large number of states.
	// If this automata is embedded in a larger automata, the total number of states can quickly explode.
	// Keep it under control by combining redundant states.
	dfa := builder.Build()
	return convertDfaToNfaWithoutAcceptActions(dfa)
}

func convertDfaToNfaWithoutAcceptActions(dfa *Dfa) *Nfa {
	nfa := EmptyLanguageNfa()

	for state := 1; state < dfa.NumStates; state++ {
		acceptActions := dfa.AcceptActions[state]
		nfaState := newNfaState(len(acceptActions) > 0)
		nfa.states = append(nfa.states, nfaState)
	}

	for i, nextState := range dfa.Transitions {
		if nextState == DfaDeadState {
			continue
		}

		prevState := i / maxTransitionsPerState
		input := i % maxTransitionsPerState
		nfa.states[prevState].inputTransitions[input] = []int{nextState}
	}

	nfa.states[0].emptyTransitions = []int{dfa.StartState}

	return nfa
}

func iterRunes(rangeTable *unicode.RangeTable, f func(rune)) {
	for _, r16 := range rangeTable.R16 {
		for r := rune(r16.Lo); r <= rune(r16.Hi); r += rune(r16.Stride) {
			f(r)
		}
	}
	for _, r32 := range rangeTable.R32 {
		for r := rune(r32.Lo); r <= rune(r32.Hi); r += rune(r32.Stride) {
			f(r)
		}
	}
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
	km := &intSliceKeyMaker{}
	builder := NewDfaBuilder()
	startState := builder.StartState()

	// Construct the DFA using the subset construction algorithm.
	// See Aho, Alfred V. (2003). Compilers: Principles, Techniques and Tools
	// or Grune, D. (2008). Parsing Techniques: A Practical Guide.
	dfaStateToNfaStates := map[int][]int{
		startState: nfa.emptyTransitionsClosure([]int{0}),
	}

	stateSetToDfaState := map[string]int{
		km.makeKey([]int{0}): startState,
	}

	var dfaState int
	stack := []int{startState}
	for len(stack) > 0 {
		dfaState, stack = stack[len(stack)-1], stack[:len(stack)-1]

		// If any of the corresponding states in the NFA are accept states,
		// mark the new state in the DFA as an accept state.
		for _, nfaStateIdx := range dfaStateToNfaStates[dfaState] {
			nfaState := nfa.states[nfaStateIdx]
			if nfaState.accept {
				for _, action := range nfaState.acceptActions {
					builder.AddAcceptAction(dfaState, action)
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

			if nextDfaState, ok := stateSetToDfaState[km.makeKey(nextNfaStates)]; ok {
				// The new state already exists, so update it with the new transitions.
				builder.AddTransition(dfaState, i, nextDfaState)
			} else {
				// The new state does not yet exist, so create it and push it onto the stack.
				nextDfaState = builder.AddState()
				builder.AddTransition(dfaState, i, nextDfaState)
				dfaStateToNfaStates[nextDfaState] = nextNfaStates
				stateSetToDfaState[km.makeKey(nextNfaStates)] = nextDfaState
				stack = append(stack, nextDfaState)
			}
		}
	}

	return builder.Build()
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
		inputTransitions: make(map[int][]int, 0),
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

// dfaState represents a state in the deterministic finite automata.
// It uses a map to represent transitions and avoids materializing transitions to the dead state.
// This is used when building the DFA, since the initial representation may have many redundant states that are eliminated by the minimization algorithm.
// However, the full DFA uses a slice to represent transitions to avoid the runtime cost of map lookups when running the DFA.
type dfaState struct {
	inputTransitions map[int]int
	acceptActions    []int
}

func newDfaState() *dfaState {
	return &dfaState{
		inputTransitions: make(map[int]int, 0),
		acceptActions:    nil,
	}
}

// DfaBuilder constructs a DFA with the minimal number of states.
type DfaBuilder struct {
	states     []*dfaState
	startState int
}

// NewDfaBuilder initializes a new builder.
func NewDfaBuilder() *DfaBuilder {
	deadState, startState := newDfaState(), newDfaState()
	return &DfaBuilder{
		states:     []*dfaState{deadState, startState},
		startState: 1,
	}
}

func (b *DfaBuilder) StartState() int {
	return b.startState
}

func (b *DfaBuilder) NumStates() int {
	return len(b.states)
}

// AddState adds a new state to the DFA, returning the state index.
func (b *DfaBuilder) AddState() int {
	id := len(b.states)
	b.states = append(b.states, newDfaState())
	return id
}

// AddTransition adds a transition from one state to another based on an input.
func (b *DfaBuilder) AddTransition(fromState int, onInput int, toState int) {
	if toState == DfaDeadState {
		// Transitions to the dead state are always represented by an absence of a key in the transition map.
		// It's important to be consistent about this, because later we compare transition maps to determine
		// if two states have the same transitions in the state minimization algorithm.
		delete(b.states[fromState].inputTransitions, onInput)
	}

	b.states[fromState].inputTransitions[onInput] = toState
}

// AddAcceptAction adds an accept action to take when a state is reached.
// This marks the state as an accept state.
func (b *DfaBuilder) AddAcceptAction(state int, action int) {
	s := b.states[state]
	s.acceptActions = insertUniqueSorted(s.acceptActions, action)
}

// NextState returns the next state after a transition based on an input.
func (b *DfaBuilder) NextState(fromState int, onInput int) int {
	return b.states[fromState].inputTransitions[onInput]
}

// Build produces a DFA with the minimal number of states.
func (b *DfaBuilder) Build() *Dfa {
	// This is the DFA state minimization algorithm from
	// Aho, Alfred V. (2003). Compilers: Principles, Techniques and Tools
	groups := b.groupEquivalentStates()
	return b.newDfaFromGroups(groups)
}

func (b *DfaBuilder) groupEquivalentStates() [][]int {
	km := &intSliceKeyMaker{}
	groups := b.initialGroups(km)
	for {
		prevNumGroups := len(groups)
		groups = b.splitGroupsIfNecessary(groups, km)
		if prevNumGroups == len(groups) {
			// The groups haven't changed, so they cannot be further split.
			break
		}
	}

	return groups
}

func (b *DfaBuilder) initialGroups(km *intSliceKeyMaker) [][]int {
	groups := make([][]int, 0, len(b.states))

	// The dead state must come first so that it becomes the first state in the new DFA.
	groups = append(groups, []int{DfaDeadState})

	// Partition the remaining groups by their accept actions.
	partitions := make(map[string][]int, len(b.states))
	for s := 1; s < len(b.states); s++ {
		key := km.makeKey(b.states[s].acceptActions)
		partitions[key] = append(partitions[key], s)
	}

	forEachPartitionInKeyOrder(partitions, func(states []int) {
		groups = append(groups, states)
	})

	return groups
}

func (b *DfaBuilder) indexStatesByGroup(groups [][]int) []int {
	stateToGroup := make([]int, len(b.states))
	for g, states := range groups {
		for _, s := range states {
			stateToGroup[s] = g
		}
	}
	return stateToGroup
}

func (b *DfaBuilder) splitGroupsIfNecessary(groups [][]int, km *intSliceKeyMaker) [][]int {
	stateToGroup := b.indexStatesByGroup(groups)
	newGroups := make([][]int, 0, len(groups))
	for g := 0; g < len(groups); g++ {
		groupStates := groups[g]

		// Fast path for groups that cannot be split.
		if !b.canSplitGroup(groupStates) {
			newGroups = append(newGroups, groupStates)
			continue
		}

		// Partition states in the group by their transitions to other groups.
		// Two states with the same transitions are considered
		// equivalent and should remain in the same group.
		partitions := make(map[string][]int, len(groupStates))
		for _, gs := range groupStates {
			var nextStateGroups [maxTransitionsPerState]int
			for c, nextState := range b.states[gs].inputTransitions {
				nextStateGroups[c] = stateToGroup[nextState]
			}
			key := km.makeKey(nextStateGroups[:])
			partitions[key] = append(partitions[key], gs)
		}

		forEachPartitionInKeyOrder(partitions, func(groupStates []int) {
			newGroups = append(newGroups, groupStates)
		})
	}

	return newGroups
}

func (b *DfaBuilder) canSplitGroup(states []int) bool {
	// Cannot split a group with a single state.
	// By construction, all groups must have at least one state.
	if len(states) == 1 {
		return false
	}

	// Cannot split a group where every state has the same transitions.
	firstStateTransitions := b.states[states[0]].inputTransitions
	for _, s := range states[1:] {
		transitions := b.states[s].inputTransitions
		if len(firstStateTransitions) != len(transitions) {
			return true
		}

		for c, expectedNextState := range firstStateTransitions {
			actualNextState, ok := b.states[s].inputTransitions[c]
			if !ok || actualNextState != expectedNextState {
				return true
			}
		}
	}
	return false
}

func (b *DfaBuilder) newDfaFromGroups(groups [][]int) *Dfa {
	stateToGroup := b.indexStatesByGroup(groups)

	// Initialize a new DFA with one state for each group.
	// Intially, all transitions go to the dead state.
	newDfa := &Dfa{
		NumStates:     len(groups),
		Transitions:   make([]int, len(groups)*maxTransitionsPerState),
		AcceptActions: make([][]int, len(groups)),
	}

	// Add transitions and accept actions.
	for g, groupStates := range groups {
		s := groupStates[0] // Arbitrarily choose the first state as representative.

		for c, nextState := range b.states[s].inputTransitions {
			newDfa.Transitions[g*maxTransitionsPerState+c] = stateToGroup[nextState]
		}

		for _, a := range b.states[s].acceptActions {
			newDfa.AcceptActions[g] = append(newDfa.AcceptActions[g], a)
		}

		for _, s := range groupStates {
			if s == b.startState {
				newDfa.StartState = stateToGroup[s]
				break
			}
		}
	}

	return newDfa
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

// DfaMatchResult represents the result of a running a DFA to find the longest match.
type DfaMatchResult struct {
	Accepted                 bool
	EndPos                   uint64
	LookaheadPos             uint64
	NumBytesReadAtLastAccept int
	Actions                  []int
}

var DfaMatchResultNone = DfaMatchResult{}

// consolidate combines two match results into a single match result.
// It chooses the longest accepting match of the pair.  In the event
// of a tie, it returns an accepting match with the actions from both matches.
func (m DfaMatchResult) consolidate(other DfaMatchResult) DfaMatchResult {
	if m.Accepted && other.Accepted {
		if m.EndPos > other.EndPos {
			return m
		} else if other.EndPos > m.EndPos {
			return other
		} else {
			return DfaMatchResult{
				Accepted:                 true,
				EndPos:                   m.EndPos,
				LookaheadPos:             maxUint64(m.LookaheadPos, other.LookaheadPos),
				NumBytesReadAtLastAccept: m.NumBytesReadAtLastAccept,
				Actions:                  append(m.Actions, other.Actions...),
			}
		}
	} else if m.Accepted {
		return m
	} else if other.Accepted {
		return other
	} else {
		return DfaMatchResultNone
	}
}

// dfaStateSet represents a set of DFA states.
type dfaStateSet []int

// allDead returns whether every state in the set is the dead state.
func (ss dfaStateSet) allDead() bool {
	for _, s := range ss {
		if s != DfaDeadState {
			return false
		}
	}
	return true
}

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

// NextState returns the next state after a transition based on an input.
func (dfa *Dfa) NextState(fromState int, onInput int) int {
	return dfa.Transitions[fromState*maxTransitionsPerState+onInput]
}

// MatchLongest returns the longest match in an input string.
// In some cases, the longest match could be empty (e.g. the regular language for "a*" matches the empty string at the beginning of the string "bbb").
// The reader position is reset to the end of the match, if there is one, or its original position if not.
// startPos and textLen determine the maximum number of runes the DFA will process;
// they also control the behavior of start-of-text (^) and end-of-text ($) patterns.
func (dfa *Dfa) MatchLongest(r InputReader, startPos uint64, textLen uint64) (DfaMatchResult, error) {
	var result DfaMatchResult
	var totalBytesRead, numBytesReadInText int
	pos := startPos
	stateSet := dfaStateSet{dfa.StartState}

	// If at the start of text, we need to run the state machine both with and
	// without the "start-of-text" input.  For example, if the state machine
	// has two patterns "a" and "^a", and the input is "a", we should match
	// both patterns.
	if startPos == 0 {
		stateSet = append(stateSet, dfa.NextState(dfa.StartState, startOfText))
	}

	// Check if the first state is an accept state.
	for _, s := range stateSet {
		if acceptActions := dfa.AcceptActions[s]; len(acceptActions) > 0 {
			result = result.consolidate(DfaMatchResult{
				Accepted: true,
				EndPos:   pos,
				Actions:  acceptActions,
			})
		}
	}

	// Run the state machine on bytes from the input reader.
	for {
		n, err := r.Read(dfa.buf[:])
		if err != nil && err != io.EOF {
			return DfaMatchResultNone, err
		}

		prevTotalBytesRead := totalBytesRead
		totalBytesRead += n

		for i, c := range dfa.buf[:n] {
			pos += uint64(textUtf8.StartByteIndicator[c])
			if pos > textLen {
				// The reader produced more bytes than the length of the text.
				// Exit the loop and treat the last rune read as the end of the text.
				pos = textLen
				break
			}

			numBytesReadInText++

			for stateIdx := range stateSet {
				nextState := dfa.NextState(stateSet[stateIdx], int(c))
				if acceptActions := dfa.AcceptActions[nextState]; len(acceptActions) > 0 {
					result = result.consolidate(DfaMatchResult{
						Accepted:                 true,
						EndPos:                   pos,
						NumBytesReadAtLastAccept: prevTotalBytesRead + i + 1,
						Actions:                  acceptActions,
					})
				}
				stateSet[stateIdx] = nextState
			}

			if stateSet.allDead() {
				break
			}
		}

		if err == io.EOF || stateSet.allDead() || pos >= textLen {
			break
		}
	}

	// Record the maximum position reached by the state machine.
	// This is used during retokenization to decide which tokens to invalidate.
	result.LookaheadPos = pos

	// Check if any pattern matches the end-of-text input.
	if pos == textLen {
		for stateIdx := range stateSet {
			nextState := dfa.NextState(stateSet[stateIdx], endOfText)
			if acceptActions := dfa.AcceptActions[nextState]; len(acceptActions) > 0 {
				result = result.consolidate(DfaMatchResult{
					Accepted:                 true,
					EndPos:                   pos,
					NumBytesReadAtLastAccept: numBytesReadInText,
					Actions:                  acceptActions,
				})
			}
		}
	}

	// Reset the reader position to the end of the last match.
	// If there was no match, this resets the reader to its original position.
	if err := r.SeekBackward(uint64(totalBytesRead - result.NumBytesReadAtLastAccept)); err != nil {
		return DfaMatchResultNone, err
	}

	return result, nil
}
