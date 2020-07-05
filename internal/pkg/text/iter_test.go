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
			iter := NewForwardRuneIter(reader)
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
	iter := NewForwardRuneIter(reader)
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

func TestForwardRuneIterLookahead(t *testing.T) {
	reader := newSingleByteReader("£ôƊ፴ऴஅ")
	iter := NewForwardRuneIter(reader)
	r, err := iter.NextRune()
	require.NoError(t, err)
	assert.Equal(t, '£', r)
	clonedIter := iter.Clone()
	originalRunes := collectRunes(t, iter)
	clonedRunes := collectRunes(t, clonedIter)
	assert.Equal(t, "ôƊ፴ऴஅ", string(originalRunes))
	assert.Equal(t, "ôƊ፴ऴஅ", string(clonedRunes))
}
