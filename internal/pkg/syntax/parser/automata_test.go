package parser

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCompileAndMatchLongest(t *testing.T) {
	testCases := []struct {
		name           string
		nfa            *Nfa
		inputString    string
		expectAccepted bool
		expectEnd      int
		expectActions  []int
	}{
		{
			name:           "empty language nfa rejects empty string",
			nfa:            EmptyLanguageNfa(),
			inputString:    "",
			expectAccepted: false,
		},
		{
			name:           "empty language nfa rejects non-empty string",
			nfa:            EmptyLanguageNfa(),
			inputString:    "a",
			expectAccepted: false,
		},
		{
			name:           "empty string nfa accepts empty string",
			nfa:            EmptyStringNfa().SetAcceptAction(99),
			inputString:    "",
			expectAccepted: true,
			expectEnd:      0,
			expectActions:  []int{99},
		},
		{
			name:           "empty string nfa accepts empty prefix of non-empty string",
			nfa:            EmptyStringNfa().SetAcceptAction(99),
			inputString:    "a",
			expectAccepted: true,
			expectEnd:      0,
			expectActions:  []int{99},
		},
		{
			name:           "single char nfa rejects empty string",
			nfa:            NfaForChars([]byte{'a'}).SetAcceptAction(99),
			inputString:    "",
			expectAccepted: false,
		},
		{
			name:           "single char nfa accepts matching char",
			nfa:            NfaForChars([]byte{'a'}).SetAcceptAction(99),
			inputString:    "a",
			expectAccepted: true,
			expectEnd:      1,
			expectActions:  []int{99},
		},
		{
			name:           "single char nfa rejects mismatched char",
			nfa:            NfaForChars([]byte{'a'}).SetAcceptAction(99),
			inputString:    "b",
			expectAccepted: false,
		},
		{
			name:           "single char nfa rejects mismatched char followed by matching char",
			nfa:            NfaForChars([]byte{'a'}).SetAcceptAction(99),
			inputString:    "ba",
			expectAccepted: false,
		},
		{
			name:           "multi char nfa accepts first char",
			nfa:            NfaForChars([]byte{'a', 'b'}).SetAcceptAction(99),
			inputString:    "a",
			expectAccepted: true,
			expectEnd:      1,
			expectActions:  []int{99},
		},
		{
			name:           "multi char nfa accepts second char",
			nfa:            NfaForChars([]byte{'a', 'b'}).SetAcceptAction(99),
			inputString:    "b",
			expectAccepted: true,
			expectEnd:      1,
			expectActions:  []int{99},
		},
		{
			name:           "negated char nfa rejects negated char",
			nfa:            NfaForNegatedChars([]byte{'a'}).SetAcceptAction(99),
			inputString:    "a",
			expectAccepted: false,
		},
		{
			name:           "negated char nfa accepts other chars",
			nfa:            NfaForNegatedChars([]byte{'a'}).SetAcceptAction(99),
			inputString:    "x",
			expectAccepted: true,
			expectEnd:      1,
			expectActions:  []int{99},
		},
		{
			name:           "star matches empty string",
			nfa:            NfaForChars([]byte{'a'}).Star().SetAcceptAction(99),
			inputString:    "",
			expectAccepted: true,
			expectEnd:      0,
			expectActions:  []int{99},
		},
		{
			name:           "star matches single occurrence",
			nfa:            NfaForChars([]byte{'a'}).Star().SetAcceptAction(99),
			inputString:    "a",
			expectAccepted: true,
			expectEnd:      1,
			expectActions:  []int{99},
		},
		{
			name:           "star matches multiple occurrences",
			nfa:            NfaForChars([]byte{'a'}).Star().SetAcceptAction(99),
			inputString:    "aaaa",
			expectAccepted: true,
			expectEnd:      4,
			expectActions:  []int{99},
		},
		{
			name:           "star matches multiple occurrences up to non-match",
			nfa:            NfaForChars([]byte{'a'}).Star().SetAcceptAction(99),
			inputString:    "aaabbb",
			expectAccepted: true,
			expectEnd:      3,
			expectActions:  []int{99},
		},
		{
			name:           "star accepts empty prefix if first character doesn't match",
			nfa:            NfaForChars([]byte{'a'}).Star().SetAcceptAction(99),
			inputString:    "baa",
			expectAccepted: true,
			expectEnd:      0,
			expectActions:  []int{99},
		},
		{
			name:           "union of two chars rejects empty string",
			nfa:            NfaForChars([]byte{'a'}).Union(NfaForChars([]byte{'b'})).SetAcceptAction(99),
			inputString:    "",
			expectAccepted: false,
		},
		{
			name:           "union of two chars accepts first char",
			nfa:            NfaForChars([]byte{'a'}).Union(NfaForChars([]byte{'b'})).SetAcceptAction(99),
			inputString:    "a",
			expectAccepted: true,
			expectEnd:      1,
			expectActions:  []int{99},
		},
		{
			name:           "union of two chars accepts second char",
			nfa:            NfaForChars([]byte{'a'}).Union(NfaForChars([]byte{'b'})).SetAcceptAction(99),
			inputString:    "b",
			expectAccepted: true,
			expectEnd:      1,
			expectActions:  []int{99},
		},
		{
			name:           "union of two chars rejects other char",
			nfa:            NfaForChars([]byte{'a'}).Union(NfaForChars([]byte{'b'})).SetAcceptAction(99),
			inputString:    "x",
			expectAccepted: false,
		},
		{
			name:           "union with different accept actions preserves first accept actions",
			nfa:            NfaForChars([]byte{'a'}).SetAcceptAction(22).Union(NfaForChars([]byte{'b'}).SetAcceptAction(33)),
			inputString:    "a",
			expectAccepted: true,
			expectEnd:      1,
			expectActions:  []int{22},
		},
		{
			name:           "union with different accept actions preserves second accept actions",
			nfa:            NfaForChars([]byte{'a'}).SetAcceptAction(22).Union(NfaForChars([]byte{'b'}).SetAcceptAction(33)),
			inputString:    "b",
			expectAccepted: true,
			expectEnd:      1,
			expectActions:  []int{33},
		},
		{
			name:           "concat of two chars rejects empty string",
			nfa:            NfaForChars([]byte{'a'}).Concat(NfaForChars([]byte{'b'})).SetAcceptAction(99),
			inputString:    "",
			expectAccepted: false,
		},
		{
			name:           "concat of two chars rejects first char only",
			nfa:            NfaForChars([]byte{'a'}).Concat(NfaForChars([]byte{'b'})).SetAcceptAction(99),
			inputString:    "a",
			expectAccepted: false,
		},
		{
			name:           "concat of two chars rejects first followed by wrong char",
			nfa:            NfaForChars([]byte{'a'}).Concat(NfaForChars([]byte{'b'})).SetAcceptAction(99),
			inputString:    "ax",
			expectAccepted: false,
		},
		{
			name:           "concat of two chars accepts first char followed by second char",
			nfa:            NfaForChars([]byte{'a'}).Concat(NfaForChars([]byte{'b'})).SetAcceptAction(99),
			inputString:    "ab",
			expectAccepted: true,
			expectEnd:      2,
			expectActions:  []int{99},
		},
		{
			name:           "concat of two chars accepts up to end of both chars",
			nfa:            NfaForChars([]byte{'a'}).Concat(NfaForChars([]byte{'b'})).SetAcceptAction(99),
			inputString:    "abcd",
			expectAccepted: true,
			expectEnd:      2,
			expectActions:  []int{99},
		},
		{
			name:           "concat of two chars preserves actions from both nfas",
			nfa:            NfaForChars([]byte{'a'}).SetAcceptAction(22).Concat(NfaForChars([]byte{'b'}).SetAcceptAction(44)),
			inputString:    "ab",
			expectAccepted: true,
			expectEnd:      2,
			expectActions:  []int{22, 44},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			dfa := tc.nfa.CompileDfa()
			r := strings.NewReader(tc.inputString)
			accepted, end, actions, err := dfa.MatchLongest(r)
			require.NoError(t, err)
			assert.Equal(t, tc.expectAccepted, accepted)
			assert.Equal(t, tc.expectEnd, end)
			assert.Equal(t, tc.expectActions, actions)
		})
	}
}

func TestMinimizeDfa(t *testing.T) {
	dfa := NewDfa()
	s1, s2, s3, s4 := dfa.AddState(), dfa.AddState(), dfa.AddState(), dfa.AddState()
	dfa.AddTransition(dfa.StartState, 'a', s1)
	dfa.AddTransition(dfa.StartState, 'b', s2)
	dfa.AddTransition(s1, 'a', s1)
	dfa.AddTransition(s1, 'b', s3)
	dfa.AddTransition(s2, 'b', s2)
	dfa.AddTransition(s2, 'a', s1)
	dfa.AddTransition(s3, 'a', s1)
	dfa.AddTransition(s3, 'b', s4)
	dfa.AddTransition(s4, 'a', s1)
	dfa.AddTransition(s4, 'b', s2)
	dfa.AddAcceptAction(s4, 99)

	// Expect that the minimized DFA has fewer states than the original DFA.
	newDfa := dfa.Minimize()
	assert.Less(t, newDfa.NumStates, dfa.NumStates)

	// Expect that the two DFAs give the same results.
	inputStrings := []string{"", "a", "b", "abb", "bbabb", "aaaabb", "aba"}
	for _, s := range inputStrings {
		accepted, end, actions, err := dfa.MatchLongest(strings.NewReader(s))
		require.NoError(t, err)

		newAccepted, newEnd, newActions, err := newDfa.MatchLongest(strings.NewReader(s))
		require.NoError(t, err)

		assert.Equal(t, accepted, newAccepted)
		assert.Equal(t, end, newEnd)
		assert.Equal(t, actions, newActions)
	}
}
