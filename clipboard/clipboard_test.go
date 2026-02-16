package clipboard

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClipboardPageNull(t *testing.T) {
	c := New()
	assertClipboardContent(t, c, PageNull, "", false)

	err := c.Set(PageNull, strings.NewReader("abcd"), false)
	require.NoError(t, err)

	assertClipboardContent(t, c, PageNull, "", false)
}

func TestClipboardPageDefault(t *testing.T) {
	c := New()
	assertClipboardContent(t, c, PageDefault, "", false)

	err := c.Set(PageDefault, strings.NewReader("abcd"), false)
	require.NoError(t, err)

	assertClipboardContent(t, c, PageDefault, "abcd", false)
}

func TestClipboardLinewise(t *testing.T) {
	c := New()

	err := c.Set(PageDefault, strings.NewReader("abcd\n"), true)
	require.NoError(t, err)

	assertClipboardContent(t, c, PageDefault, "abcd\n", true)
}

func TestPageIdForInputRune(t *testing.T) {
	testCases := []struct {
		name         string
		letter       rune
		expectedPage PageId
	}{
		{
			name:         "page a",
			letter:       'a',
			expectedPage: PageLetterA,
		},
		{
			name:         "page b",
			letter:       'b',
			expectedPage: PageLetterB,
		},
		{
			name:         "page y",
			letter:       'y',
			expectedPage: PageLetterY,
		},
		{
			name:         "page z",
			letter:       'z',
			expectedPage: PageLetterZ,
		},
		{
			name:         "non-alpha",
			letter:       '!',
			expectedPage: PageNull,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			page := PageIdForInputRune(tc.letter)
			assert.Equal(t, tc.expectedPage, page)
		})
	}
}

func assertClipboardContent(t *testing.T, c *Clipboard, p PageId, expectedText string, expectedLinewise bool) {
	t.Helper()

	var buf bytes.Buffer
	linewise, err := c.Get(p, &buf)
	require.NoError(t, err)
	assert.Equal(t, expectedText, buf.String())
	assert.Equal(t, expectedLinewise, linewise)
}
