package file

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func pathWithTrailingSlash(p string) string {
	if len(p) > 0 && p[len(p)-1] == filepath.Separator {
		return p
	}
	return fmt.Sprintf("%s%c", p, filepath.Separator)
}

func pathsWithTrailingSlashes(paths []string) []string {
	result := make([]string, 0, len(paths))
	for _, p := range paths {
		result = append(result, pathWithTrailingSlash(p))
	}
	return result
}

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
			expectedSuffixes: pathsWithTrailingSlashes([]string{"aaa", "aab", "aac", "aba", "abb", "abc", "xyz"}),
		},
		{
			name:             "base directory, no trailing slash",
			prefix:           tmpDir,
			expectedSuffixes: nil,
		},
		{
			name:             "base directory with trailing slash",
			prefix:           pathWithTrailingSlash(tmpDir),
			expectedSuffixes: pathsWithTrailingSlashes([]string{"aaa", "aab", "aac", "aba", "abb", "abc", "xyz"}),
		},
		{
			name:             "first character matches",
			prefix:           filepath.Join(tmpDir, "x"),
			expectedSuffixes: pathsWithTrailingSlashes([]string{"yz"}),
		},
		{
			name:             "first two characters match",
			prefix:           filepath.Join(tmpDir, "ab"),
			expectedSuffixes: pathsWithTrailingSlashes([]string{"a", "b", "c"}),
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
