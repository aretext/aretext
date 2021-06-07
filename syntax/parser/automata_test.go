package parser

import (
	"io/ioutil"
	"strings"
	"testing"
	"unicode"

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
		{
			name:           "unicode category digit single byte accepts",
			nfa:            NfaForUnicodeCategory(unicode.Digit).SetAcceptAction(99),
			inputString:    "6",
			startPos:       0,
			expectAccepted: true,
			expectEndPos:   1,
			expectActions:  []int{99},
		},
		{
			name:           "unicode category digit single byte rejects",
			nfa:            NfaForUnicodeCategory(unicode.Digit),
			inputString:    "x",
			startPos:       0,
			expectAccepted: false,
		},
		{
			name:           "unicode category digit multi byte accepts",
			nfa:            NfaForUnicodeCategory(unicode.Digit).SetAcceptAction(99),
			inputString:    "\u0660",
			startPos:       0,
			expectAccepted: true,
			expectEndPos:   1,
			expectActions:  []int{99},
		},
		{
			name:           "unicode category digit multi byte rejects",
			nfa:            NfaForUnicodeCategory(unicode.Digit),
			inputString:    "\u0700",
			startPos:       0,
			expectAccepted: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			dfa := tc.nfa.CompileDfa()
			r := &ReadSeekerInput{R: strings.NewReader(tc.inputString)}
			textLen := tc.startPos + uint64(len(tc.inputString))
			matchResult, err := dfa.MatchLongest(r, tc.startPos, textLen)
			require.NoError(t, err)
			assert.Equal(t, tc.expectAccepted, matchResult.Accepted)
			assert.Equal(t, tc.expectEndPos, matchResult.EndPos)
			assert.Equal(t, tc.expectActions, matchResult.Actions)
		})
	}
}

func TestMatchLongestTruncatedText(t *testing.T) {
	nfa := NfaForUnicodeCategory(unicode.Letter).Star().Concat(NfaForEndOfText()).SetAcceptAction(99)
	dfa := nfa.CompileDfa()
	s := "αααβββδδδ"
	r := &ReadSeekerInput{R: strings.NewReader(s)}

	// Set textLen to 3 so the DFA treats the third character as the end of the text,
	// even though the reader outputs more bytes.
	matchResult, err := dfa.MatchLongest(r, 0, 3)
	require.NoError(t, err)
	assert.True(t, matchResult.Accepted)
	assert.Equal(t, uint64(3), matchResult.EndPos)
	assert.Equal(t, []int{99}, matchResult.Actions)

	// Verify that the reader is reset to the end of the match.
	remaining, err := ioutil.ReadAll(r)
	require.NoError(t, err)
	assert.Equal(t, []byte("βββδδδ"), remaining)
}

func TestMinimizeDfa(t *testing.T) {
	builder := NewDfaBuilder()
	s1, s2, s3, s4 := builder.AddState(), builder.AddState(), builder.AddState(), builder.AddState()
	builder.AddTransition(builder.StartState(), 'a', s1)
	builder.AddTransition(builder.StartState(), 'b', s2)
	builder.AddTransition(s1, 'a', s1)
	builder.AddTransition(s1, 'b', s3)
	builder.AddTransition(s2, 'b', s2)
	builder.AddTransition(s2, 'a', s1)
	builder.AddTransition(s3, 'a', s1)
	builder.AddTransition(s3, 'b', s4)
	builder.AddTransition(s4, 'a', s1)
	builder.AddTransition(s4, 'b', s2)
	builder.AddAcceptAction(s4, 99)

	// Expect that the minimized DFA has fewer states than the builder.
	dfa := builder.Build()
	assert.Less(t, dfa.NumStates, builder.NumStates())

	// Expect that the DFA gives correct results.
	testCases := []struct {
		inputString    string
		expectAccepted bool
		expectEndPos   uint64
	}{
		{
			inputString:    "",
			expectAccepted: false,
		},
		{
			inputString:    "a",
			expectAccepted: false,
		},
		{
			inputString:    "b",
			expectAccepted: false,
		},
		{
			inputString:    "abb",
			expectAccepted: true,
			expectEndPos:   3,
		},
		{
			inputString:    "bbabb",
			expectAccepted: true,
			expectEndPos:   5,
		},
		{
			inputString:    "aaaabb",
			expectAccepted: true,
			expectEndPos:   6,
		},
		{
			inputString:    "aba",
			expectAccepted: false,
		},
		{
			inputString:    "abbaaaaaa",
			expectAccepted: true,
			expectEndPos:   3,
		},
	}

	for _, tc := range testCases {
		textLen := uint64(len(tc.inputString))
		r := &ReadSeekerInput{R: strings.NewReader(tc.inputString)}
		matchResult, err := dfa.MatchLongest(r, 0, textLen)
		require.NoError(t, err)
		assert.Equal(t, tc.expectAccepted, matchResult.Accepted)
		assert.Equal(t, tc.expectEndPos, matchResult.EndPos)
		if tc.expectAccepted {
			assert.Equal(t, []int{99}, matchResult.Actions)
		}
	}
}

func BenchmarkNfaForUnicodeCategory(b *testing.B) {
	for n := 0; n < b.N; n++ {
		NfaForUnicodeCategory(unicode.Letter)
	}
}
