package clipboard

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClipboardPageNull(t *testing.T) {
	c := New(nil)
	assertClipboardContent(t, c, PageNull, "", false)

	err := c.Set(PageNull, strings.NewReader("abcd"), false)
	require.NoError(t, err)

	assertClipboardContent(t, c, PageNull, "", false)
}

func TestClipboardPageDefault(t *testing.T) {
	c := New(nil)
	assertClipboardContent(t, c, PageDefault, "", false)

	err := c.Set(PageDefault, strings.NewReader("abcd"), false)
	require.NoError(t, err)

	assertClipboardContent(t, c, PageDefault, "abcd", false)
}

func TestClipboardLinewise(t *testing.T) {
	c := New(nil)

	err := c.Set(PageDefault, strings.NewReader("abcd\n"), true)
	require.NoError(t, err)

	assertClipboardContent(t, c, PageDefault, "abcd\n", true)
}

func TestClipboardPageSystemNotConfigured(t *testing.T) {
	testCases := []struct {
		name      string
		pageRune  rune
		operation func(t *testing.T, c *Clipboard, p PageId) error
	}{
		{
			name:     "set plus",
			pageRune: '+',
			operation: func(t *testing.T, c *Clipboard, p PageId) error {
				t.Helper()
				return c.Set(p, strings.NewReader("abcd"), false)
			},
		},
		{
			name:     "set star",
			pageRune: '*',
			operation: func(t *testing.T, c *Clipboard, p PageId) error {
				t.Helper()
				return c.Set(p, strings.NewReader("abcd"), false)
			},
		},
		{
			name:     "get plus",
			pageRune: '+',
			operation: func(t *testing.T, c *Clipboard, p PageId) error {
				t.Helper()
				var buf bytes.Buffer
				_, err := c.Get(p, &buf)
				return err
			},
		},
		{
			name:     "get star",
			pageRune: '*',
			operation: func(t *testing.T, c *Clipboard, p PageId) error {
				t.Helper()
				var buf bytes.Buffer
				_, err := c.Get(p, &buf)
				return err
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			c := New(nil)
			page := PageIdForInputRune(tc.pageRune)

			err := tc.operation(t, c, page)
			require.Error(t, err)
			assert.Contains(t, err.Error(), "system clipboard not configured")
		})
	}
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
		{
			name:         "system plus",
			letter:       '+',
			expectedPage: PageSystem,
		},
		{
			name:         "system star",
			letter:       '*',
			expectedPage: PageSystem,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			page := PageIdForInputRune(tc.letter)
			assert.Equal(t, tc.expectedPage, page)
		})
	}
}

func TestSystemClipboardSetAndGet(t *testing.T) {
	path := filepath.Join(t.TempDir(), "clipboard.txt")
	copyCmd := fmt.Sprintf("cat > %q", path)
	pasteCmd := fmt.Sprintf("cat %q", path)
	c := NewSystemClipboard(copyCmd, pasteCmd, false)

	err := c.Set(strings.NewReader("abcd"), false)
	require.NoError(t, err)
	assertFileContent(t, path, "abcd")

	var buf bytes.Buffer
	linewise, err := c.Get(&buf)
	require.NoError(t, err)
	assert.Equal(t, "abcd", buf.String())
	assert.False(t, linewise)
}

func TestSystemClipboardLinewise(t *testing.T) {
	path := filepath.Join(t.TempDir(), "clipboard.txt")
	copyCmd := fmt.Sprintf("cat > %q", path)
	pasteCmd := fmt.Sprintf("cat %q", path)
	c := NewSystemClipboard(copyCmd, pasteCmd, false)

	err := c.Set(strings.NewReader("abcd"), true)
	require.NoError(t, err)
	assertFileContent(t, path, "abcd\n")

	var buf bytes.Buffer
	linewise, err := c.Get(&buf)
	require.NoError(t, err)
	assert.Equal(t, "abcd", buf.String())
	assert.True(t, linewise)
}

func TestSystemClipboardCommandErrors(t *testing.T) {
	c := NewSystemClipboard("exit 7", "exit 8", false)

	err := c.Set(strings.NewReader("abcd"), false)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "copy command failed")

	var buf bytes.Buffer
	_, err = c.Get(&buf)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "paste command failed")
}

func assertClipboardContent(t *testing.T, c *Clipboard, p PageId, expectedText string, expectedLinewise bool) {
	t.Helper()

	var buf bytes.Buffer
	linewise, err := c.Get(p, &buf)
	require.NoError(t, err)
	assert.Equal(t, expectedText, buf.String())
	assert.Equal(t, expectedLinewise, linewise)
}

func assertFileContent(t *testing.T, path string, expected string) {
	t.Helper()

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, expected, string(data))
}
