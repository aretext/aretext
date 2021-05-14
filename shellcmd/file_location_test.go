package shellcmd

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileLocationsFromLines(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected []FileLocation
	}{
		{
			name:     "empty",
			input:    "",
			expected: nil,
		},
		{
			name:     "empty lines",
			input:    "\n\n\n",
			expected: nil,
		},
		{
			name:     "empty lines with whitespace",
			input:    "\n   \t\n\t   \n",
			expected: nil,
		},
		{
			name:  "single line, grep format",
			input: "foo/bar.go:12:    this is a test",
			expected: []FileLocation{
				{
					Path:    "foo/bar.go",
					LineNum: 12,
					Snippet: "this is a test",
				},
			},
		},
		{
			name:  "single line, ripgrep format",
			input: "foo/bar.go:12:34:    this is a test",
			expected: []FileLocation{
				{
					Path:    "foo/bar.go",
					LineNum: 12,
					Snippet: "this is a test",
				},
			},
		},
		{
			name: "multiple lines",
			input: strings.Join([]string{
				"foo/bar.go:12:34:    this is a test",
				"",
				"baz/bat.go:56:78:    and another",
			}, "\n"),
			expected: []FileLocation{
				{
					Path:    "foo/bar.go",
					LineNum: 12,
					Snippet: "this is a test",
				},
				{
					Path:    "baz/bat.go",
					LineNum: 56,
					Snippet: "and another",
				},
			},
		},
		{
			name:  "snippet with colon",
			input: "foobar:12:34:test:with:separator",
			expected: []FileLocation{
				{
					Path:    "foobar",
					LineNum: 12,
					Snippet: "test:with:separator",
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			r := strings.NewReader(tc.input)
			locations, err := FileLocationsFromLines(r)
			require.NoError(t, err)
			assert.Equal(t, tc.expected, locations)
		})
	}
}

func TestFileLocationsFromLinesErrors(t *testing.T) {
	testCases := []struct {
		name        string
		input       string
		expectInErr string
	}{
		{
			name:        "non-numeric line num",
			input:       "foobar.go:abc:test",
			expectInErr: "Invalid line number",
		},
		{
			name:        "one part",
			input:       "foobar",
			expectInErr: "Unsupported format",
		},
		{
			name:        "two parts",
			input:       "foobar:12",
			expectInErr: "Unsupported format",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			r := strings.NewReader(tc.input)
			_, err := FileLocationsFromLines(r)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tc.expectInErr)
		})
	}
}
