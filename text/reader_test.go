package text

import (
	"fmt"
	"io"
	"testing"
	"testing/iotest"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLimitReader(t *testing.T) {
	testCases := []struct {
		name  string
		text  string
		limit uint64
		want  string
	}{
		{
			name:  "empty text",
			text:  "",
			limit: 4,
			want:  "",
		},
		{
			name:  "zero limit",
			text:  "abcdef",
			limit: 0,
			want:  "",
		},
		{
			name:  "limit before end",
			text:  "abcdef",
			limit: 3,
			want:  "abc",
		},
		{
			name:  "limit exactly text length",
			text:  "abcdef",
			limit: 6,
			want:  "abcdef",
		},
		{
			name:  "limit after end",
			text:  "abcdef",
			limit: 10,
			want:  "abcdef",
		},
		{
			name:  "mixed width unicode limit before end",
			text:  "a£፴\U0010AAAAb",
			limit: 4,
			want:  "a£፴\U0010AAAA",
		},
		{
			name:  "mixed width unicode limit exactly text length",
			text:  "a£፴\U0010AAAAb",
			limit: 5,
			want:  "a£፴\U0010AAAAb",
		},
		{
			name:  "combining mark counts as separate rune",
			text:  "e\u0301x",
			limit: 2,
			want:  "e\u0301",
		},
		{
			name:  "emoji sequence counts by rune",
			text:  "👨‍👩‍👧‍👦!",
			limit: 7,
			want:  "👨‍👩‍👧‍👦",
		},
	}

	bufferSizes := []int{1, 2, 3, 4, 5, 8, 64}

	for _, tc := range testCases {
		for _, bufferSize := range bufferSizes {
			t.Run(fmt.Sprintf("%s/buffer size %d", tc.name, bufferSize), func(t *testing.T) {
				tree, err := NewTreeFromString(tc.text)
				require.NoError(t, err)

				reader := NewLimitReader(tree.ReaderAtPosition(0), tc.limit)
				got := readAllWithBufferSize(t, reader, bufferSize)

				assert.Equal(t, tc.want, string(got))
				assert.True(t, utf8.Valid(got), "LimitReader returned invalid UTF-8 bytes: %v", got)
			})
		}
	}
}

func TestLimitReaderWithOneByteReader(t *testing.T) {
	testCases := []struct {
		name  string
		text  string
		limit uint64
		want  string
	}{
		{
			name:  "ascii",
			text:  "abcdef",
			limit: 3,
			want:  "abc",
		},
		{
			name:  "two byte rune",
			text:  "£abc",
			limit: 1,
			want:  "£",
		},
		{
			name:  "three byte rune",
			text:  "፴abc",
			limit: 1,
			want:  "፴",
		},
		{
			name:  "four byte rune",
			text:  "\U0010AAAAabc",
			limit: 1,
			want:  "\U0010AAAA",
		},
		{
			name:  "mixed unicode",
			text:  "a£፴\U0010AAAAb",
			limit: 4,
			want:  "a£፴\U0010AAAA",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tree, err := NewTreeFromString(tc.text)
			require.NoError(t, err)

			reader := NewLimitReader(tree.ReaderAtPosition(0), tc.limit)
			oneByteReader := iotest.OneByteReader(reader)
			got, err := io.ReadAll(oneByteReader)
			require.NoError(t, err)

			assert.Equal(t, tc.want, string(got))
			assert.True(t, utf8.Valid(got), "LimitReader returned invalid UTF-8 bytes: %v", got)
		})
	}
}

func TestLimitReaderDoesNotAdvanceUnderlyingReaderPastLimit(t *testing.T) {
	testCases := []struct {
		name      string
		text      string
		limit     uint64
		want      string
		wantAfter string
	}{
		{
			name:      "ascii",
			text:      "abcdef",
			limit:     3,
			want:      "abc",
			wantAfter: "def",
		},
		{
			name:      "mixed width unicode",
			text:      "a£፴\U0010AAAAb",
			limit:     3,
			want:      "a£፴",
			wantAfter: "\U0010AAAAb",
		},
		{
			name:      "zero limit",
			text:      "abcdef",
			limit:     0,
			want:      "",
			wantAfter: "abcdef",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tree, err := NewTreeFromString(tc.text)
			require.NoError(t, err)

			reader := NewLimitReader(tree.ReaderAtPosition(0), tc.limit)
			got := readAllWithBufferSize(t, reader, len(tc.text)+utf8.UTFMax)

			gotAfter, err := io.ReadAll(&reader.reader)
			require.NoError(t, err)

			assert.Equal(t, tc.want, string(got))
			assert.Equal(t, tc.wantAfter, string(gotAfter))
		})
	}
}

func readAllWithBufferSize(t *testing.T, reader io.Reader, bufferSize int) []byte {
	t.Helper()

	var got []byte
	buf := make([]byte, bufferSize)
	for {
		n, err := reader.Read(buf)
		got = append(got, buf[:n]...)
		if err == io.EOF {
			return got
		}
		require.NoError(t, err)
		require.Greater(t, n, 0, "reader made no progress without returning an error")
	}
}
