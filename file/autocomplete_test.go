package file

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAutocompleteDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	for _, subdir := range []string{"aaa", "aab", "aac", "aba", "abb", "abc", "xyz"} {
		path := filepath.Join(tmpDir, subdir)
		err := os.Mkdir(path, 0755)
		require.NoError(t, err)
	}

	testCases := []struct {
		name             string
		prefix           string
		chdir            string
		expectedSuffixes []string
	}{
		{
			name:             "empty prefix",
			prefix:           "",
			chdir:            tmpDir,
			expectedSuffixes: []string{"aaa", "aab", "aac", "aba", "abb", "abc", "xyz"},
		},
		{
			name:             "base directory, no trailing slash",
			prefix:           tmpDir,
			expectedSuffixes: nil,
		},
		{
			name:             "base directory with trailing slash",
			prefix:           fmt.Sprintf("%s%c", tmpDir, filepath.Separator),
			expectedSuffixes: []string{"aaa", "aab", "aac", "aba", "abb", "abc", "xyz"},
		},
		{
			name:             "first character matches",
			prefix:           filepath.Join(tmpDir, "x"),
			expectedSuffixes: []string{"yz"},
		},
		{
			name:             "first two characters match",
			prefix:           filepath.Join(tmpDir, "ab"),
			expectedSuffixes: []string{"a", "b", "c"},
		},
		{
			name:             "all characters match",
			prefix:           filepath.Join(tmpDir, "aac"),
			expectedSuffixes: nil,
		},
		{
			name:             "no characters match",
			prefix:           filepath.Join(tmpDir, "m"),
			expectedSuffixes: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.chdir != "" {
				cwd, err := os.Getwd()
				require.NoError(t, err)
				defer func() { os.Chdir(cwd) }()

				err = os.Chdir(tc.chdir)
				require.NoError(t, err)
			}

			suffixes, err := AutocompleteDirectory(tc.prefix)
			require.NoError(t, err)
			assert.Equal(t, tc.expectedSuffixes, suffixes)
		})
	}
}
