package clipboard

import (
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func readPage(t *testing.T, c *Clipboard, p PageId) (string, bool) {
	t.Helper()
	r, linewise := c.Get(p)
	data, err := io.ReadAll(r)
	require.NoError(t, err)
	return string(data), linewise
}

func TestClipboardSetWriter(t *testing.T) {
	c := New()

	w := c.Set(PageDefault, false)
	n, err := io.WriteString(w, "hello")
	require.NoError(t, err)
	assert.Equal(t, 5, n)

	text, linewise := readPage(t, c, PageDefault)
	assert.Equal(t, "hello", text)
	assert.False(t, linewise)
}

func TestClipboardSetWriterLinewise(t *testing.T) {
	c := New()

	w := c.Set(PageDefault, true)
	io.WriteString(w, "line1\nline2")

	text, linewise := readPage(t, c, PageDefault)
	assert.Equal(t, "line1\nline2", text)
	assert.True(t, linewise)
}

func TestClipboardSetWriterOverwrite(t *testing.T) {
	c := New()

	w := c.Set(PageDefault, false)
	io.WriteString(w, "first")

	w = c.Set(PageDefault, true)
	io.WriteString(w, "second")

	text, linewise := readPage(t, c, PageDefault)
	assert.Equal(t, "second", text)
	assert.True(t, linewise)
}

func TestClipboardGetReader(t *testing.T) {
	c := New()

	// Unset page returns empty reader.
	text, linewise := readPage(t, c, PageDefault)
	assert.Equal(t, "", text)
	assert.False(t, linewise)

	// Set some data and read it back.
	io.WriteString(c.Set(PageDefault, false), "abcd")

	text, linewise = readPage(t, c, PageDefault)
	assert.Equal(t, "abcd", text)
	assert.False(t, linewise)

	// Reading again should return the same data (non-destructive).
	text, linewise = readPage(t, c, PageDefault)
	assert.Equal(t, "abcd", text)
	assert.False(t, linewise)
}

func TestClipboardPageNull(t *testing.T) {
	c := New()

	text, linewise := readPage(t, c, PageNull)
	assert.Equal(t, "", text)
	assert.False(t, linewise)

	w := c.Set(PageNull, false)
	io.WriteString(w, "abcd")

	text, linewise = readPage(t, c, PageNull)
	assert.Equal(t, "", text)
	assert.False(t, linewise)
}

func TestClipboardPageDefault(t *testing.T) {
	c := New()

	text, linewise := readPage(t, c, PageDefault)
	assert.Equal(t, "", text)
	assert.False(t, linewise)

	io.WriteString(c.Set(PageDefault, false), "abcd")

	text, linewise = readPage(t, c, PageDefault)
	assert.Equal(t, "abcd", text)
	assert.False(t, linewise)
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
