package text

import (
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func collectRunes(t *testing.T, iter CloneableRuneIter) []rune {
	runes := make([]rune, 0)
	for {
		r, err := iter.NextRune()
		if err == io.EOF {
			break
		} else if err != nil {
			require.NoError(t, err)
		}

		runes = append(runes, r)
	}
	return runes
}

func TestForwardRuneIter(t *testing.T) {
	testCases := []struct {
		name          string
		inputString   string
		expectedRunes []rune
	}{
		{
			name:          "empty string",
			inputString:   "",
			expectedRunes: []rune{},
		},
		{
			name:          "multiple ASCII",
			inputString:   "abcd",
			expectedRunes: []rune{'a', 'b', 'c', 'd'},
		},
		{
			name:        "multi-byte characters",
			inputString: "£ôƊ፴ऴஅ\U0010AAAA\U0010BBBB\U0010CCCC",
			expectedRunes: []rune{
				'£',
				'ô',
				'Ɗ',
				'፴',
				'ऴ',
				'அ',
				'\U0010AAAA',
				'\U0010BBBB',
				'\U0010CCCC',
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			reader := NewCloneableReaderFromString(tc.inputString)
			iter := NewCloneableForwardRuneIter(reader)
			runes := collectRunes(t, iter)
			assert.Equal(t, runes, tc.expectedRunes)
		})
	}
}

func reverse(s string) string {
	bytes := []byte(s)
	reversedBytes := make([]byte, len(bytes))
	for i := 0; i < len(reversedBytes); i++ {
		reversedBytes[i] = bytes[len(bytes)-1-i]
	}
	return string(reversedBytes)
}

func TestBackwardRuneIter(t *testing.T) {
	testCases := []struct {
		name          string
		inputString   string
		expectedRunes []rune
	}{
		{
			name:          "empty string",
			inputString:   "",
			expectedRunes: []rune{},
		},
		{
			name:          "multiple ASCII",
			inputString:   reverse("abcd"),
			expectedRunes: []rune{'d', 'c', 'b', 'a'},
		},
		{
			name:          "two-byte char",
			inputString:   reverse("£"),
			expectedRunes: []rune{'£'},
		},
		{
			name:          "three-byte char",
			inputString:   reverse("ऴ"),
			expectedRunes: []rune{'ऴ'},
		},
		{
			name:          "four-byte char",
			inputString:   reverse("\U0010AAAA"),
			expectedRunes: []rune{'\U0010AAAA'},
		},
		{
			name:        "multi-byte characters",
			inputString: reverse("£ôƊ፴ऴஅ\U0010AAAA\U0010BBBB\U0010CCCC"),
			expectedRunes: []rune{
				'\U0010CCCC',
				'\U0010BBBB',
				'\U0010AAAA',
				'அ',
				'ऴ',
				'፴',
				'Ɗ',
				'ô',
				'£',
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			reader := NewCloneableReaderFromString(tc.inputString)
			iter := NewCloneableBackwardRuneIter(reader)
			runes := collectRunes(t, iter)
			assert.Equal(t, runes, tc.expectedRunes)
		})
	}
}

type singleByteReader struct {
	s string
	i int
}

func newSingleByteReader(s string) CloneableReader {
	return &singleByteReader{s, 0}
}

func (r *singleByteReader) Read(p []byte) (n int, err error) {
	n = copy(p, r.s[r.i:r.i+1])
	r.i++
	if r.i >= len(r.s) {
		err = io.EOF
	}
	return
}

func (r *singleByteReader) Clone() CloneableReader {
	return &singleByteReader{
		s: r.s,
		i: r.i,
	}
}

func TestForwardRuneIterSplitMultibyteRunes(t *testing.T) {
	reader := newSingleByteReader("£ôƊ፴ऴஅ\U0010AAAA\U0010BBBB\U0010CCCC")
	iter := NewCloneableForwardRuneIter(reader)
	runes := collectRunes(t, iter)
	assert.Equal(t, runes, []rune{
		'£',
		'ô',
		'Ɗ',
		'፴',
		'ऴ',
		'அ',
		'\U0010AAAA',
		'\U0010BBBB',
		'\U0010CCCC',
	})
}

func TestBackwardRuneIterSplitMultibyteRunes(t *testing.T) {
	reader := newSingleByteReader(reverse("£ôƊ፴ऴஅ\U0010AAAA\U0010BBBB\U0010CCCC"))
	iter := NewCloneableBackwardRuneIter(reader)
	runes := collectRunes(t, iter)
	assert.Equal(t, runes, []rune{
		'\U0010CCCC',
		'\U0010BBBB',
		'\U0010AAAA',
		'அ',
		'ऴ',
		'፴',
		'Ɗ',
		'ô',
		'£',
	})
}

func TestForwardRuneIterLookahead(t *testing.T) {
	reader := newSingleByteReader("£ôƊ፴ऴஅ")
	iter := NewCloneableForwardRuneIter(reader)
	r, err := iter.NextRune()
	require.NoError(t, err)
	assert.Equal(t, '£', r)
	clonedIter := iter.Clone()
	originalRunes := collectRunes(t, iter)
	clonedRunes := collectRunes(t, clonedIter)
	assert.Equal(t, "ôƊ፴ऴஅ", string(originalRunes))
	assert.Equal(t, "ôƊ፴ऴஅ", string(clonedRunes))
}

func TestBackwardRuneIterLookahead(t *testing.T) {
	reader := newSingleByteReader(reverse("£ôƊ"))
	iter := NewCloneableBackwardRuneIter(reader)
	r, err := iter.NextRune()
	require.NoError(t, err)
	assert.Equal(t, 'Ɗ', r)
	clonedIter := iter.Clone()
	originalRunes := collectRunes(t, iter)
	clonedRunes := collectRunes(t, clonedIter)
	assert.Equal(t, "ô£", string(originalRunes))
	assert.Equal(t, "ô£", string(clonedRunes))
}
