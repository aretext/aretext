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
		startPos       uint64
		expectAccepted bool
		expectEndPos   uint64
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
			expectEndPos:   0,
			expectActions:  []int{99},
		},
		{
			name:           "empty string nfa accepts empty prefix of non-empty string",
			nfa:            EmptyStringNfa().SetAcceptAction(99),
			inputString:    "a",
			expectAccepted: true,
			expectEndPos:   0,
			expectActions:  []int{99},
		},
		{
			name:           "empty string nfa preserves start position",
			nfa:            EmptyStringNfa().SetAcceptAction(99),
			inputString:    "a",
			startPos:       123,
			expectAccepted: true,
			expectEndPos:   123,
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
			expectEndPos:   1,
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
			expectEndPos:   1,
			expectActions:  []int{99},
		},
		{
			name:           "multi char nfa accepts second char",
			nfa:            NfaForChars([]byte{'a', 'b'}).SetAcceptAction(99),
			inputString:    "b",
			expectAccepted: true,
			expectEndPos:   1,
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
			expectEndPos:   1,
			expectActions:  []int{99},
		},
		{
			name:           "star matches empty string",
			nfa:            NfaForChars([]byte{'a'}).Star().SetAcceptAction(99),
			inputString:    "",
			expectAccepted: true,
			expectEndPos:   0,
			expectActions:  []int{99},
		},
		{
			name:           "star matches single occurrence",
			nfa:            NfaForChars([]byte{'a'}).Star().SetAcceptAction(99),
			inputString:    "a",
			expectAccepted: true,
			expectEndPos:   1,
			expectActions:  []int{99},
		},
		{
			name:           "star matches multiple occurrences",
			nfa:            NfaForChars([]byte{'a'}).Star().SetAcceptAction(99),
			inputString:    "aaaa",
			expectAccepted: true,
			expectEndPos:   4,
			expectActions:  []int{99},
		},
		{
			name:           "star matches multiple occurrences up to non-match",
			nfa:            NfaForChars([]byte{'a'}).Star().SetAcceptAction(99),
			inputString:    "aaabbb",
			expectAccepted: true,
			expectEndPos:   3,
			expectActions:  []int{99},
		},
		{
			name:           "star accepts empty prefix if first character doesn't match",
			nfa:            NfaForChars([]byte{'a'}).Star().SetAcceptAction(99),
			inputString:    "baa",
			expectAccepted: true,
			expectEndPos:   0,
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
			expectEndPos:   1,
			expectActions:  []int{99},
		},
		{
			name:           "union of two chars accepts second char",
			nfa:            NfaForChars([]byte{'a'}).Union(NfaForChars([]byte{'b'})).SetAcceptAction(99),
			inputString:    "b",
			expectAccepted: true,
			expectEndPos:   1,
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
			expectEndPos:   1,
			expectActions:  []int{22},
		},
		{
			name:           "union with different accept actions preserves second accept actions",
			nfa:            NfaForChars([]byte{'a'}).SetAcceptAction(22).Union(NfaForChars([]byte{'b'}).SetAcceptAction(33)),
			inputString:    "b",
			expectAccepted: true,
			expectEndPos:   1,
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
			expectEndPos:   2,
			expectActions:  []int{99},
		},
		{
			name:           "concat of two chars accepts up to end of both chars",
			nfa:            NfaForChars([]byte{'a'}).Concat(NfaForChars([]byte{'b'})).SetAcceptAction(99),
			inputString:    "abcd",
			expectAccepted: true,
			expectEndPos:   2,
			expectActions:  []int{99},
		},
		{
			name:           "concat of two chars preserves actions from both nfas",
			nfa:            NfaForChars([]byte{'a'}).SetAcceptAction(22).Concat(NfaForChars([]byte{'b'}).SetAcceptAction(44)),
			inputString:    "ab",
			expectAccepted: true,
			expectEndPos:   2,
			expectActions:  []int{22, 44},
		},
		{
			name:           "match non-ascii unicode increments end pos by number of runes",
			nfa:            NfaForChars([]byte{0xe4}).Concat(NfaForChars([]byte{0xb8})).Concat(NfaForChars([]byte{0x82})).SetAcceptAction(99),
			inputString:    "\u4E02",
			startPos:       3,
			expectAccepted: true,
			expectEndPos:   4,
			expectActions:  []int{99},
		},
		{
			name:           "start of text accepts",
			nfa:            NfaForStartOfText().Concat(NfaForChars([]byte{'a'})).SetAcceptAction(99),
			inputString:    "a",
			startPos:       0,
			expectAccepted: true,
			expectEndPos:   1,
			expectActions:  []int{99},
		},
		{
			name:           "start of text rejects",
			nfa:            NfaForStartOfText().Concat(NfaForChars([]byte{'a'})).SetAcceptAction(99),
			inputString:    "a",
			startPos:       1,
			expectAccepted: false,
		},
		{
			name:           "end of text accepts",
			nfa:            NfaForChars([]byte{'a'}).Concat(NfaForEndOfText()).SetAcceptAction(99),
			inputString:    "a",
			startPos:       0,
			expectAccepted: true,
			expectEndPos:   1,
			expectActions:  []int{99},
		},
		{
			name:           "end of text rejects",
			nfa:            NfaForChars([]byte{'a'}).Concat(NfaForEndOfText()).SetAcceptAction(99),
			inputString:    "ba",
			startPos:       0,
			expectAccepted: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			dfa := tc.nfa.CompileDfa()
			r := strings.NewReader(tc.inputString)
			textLen := tc.startPos + uint64(len(tc.inputString))
			accepted, endPos, actions, err := dfa.MatchLongest(r, tc.startPos, textLen)
			require.NoError(t, err)
			assert.Equal(t, tc.expectAccepted, accepted)
			assert.Equal(t, tc.expectEndPos, endPos)
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
		textLen := uint64(len(s))

		accepted, endPos, actions, err := dfa.MatchLongest(strings.NewReader(s), 0, textLen)
		require.NoError(t, err)

		newAccepted, newEndPos, newActions, err := newDfa.MatchLongest(strings.NewReader(s), 0, textLen)
		require.NoError(t, err)

		assert.Equal(t, accepted, newAccepted)
		assert.Equal(t, endPos, newEndPos)
		assert.Equal(t, actions, newActions)
	}
}
