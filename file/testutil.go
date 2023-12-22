package file

import (
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func createTestFile(t *testing.T, s string) string {
	tmpDir := t.TempDir()

	filePath := filepath.Join(tmpDir, "test.txt")
	f, err := os.Create(filePath)
	require.NoError(t, err)
	defer f.Close()

	_, err = io.WriteString(f, s)
	require.NoError(t, err)

	return filePath
}

func appendToTestFile(t *testing.T, path string, s string) {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0)
	require.NoError(t, err)
	defer f.Close()

	_, err = io.WriteString(f, s)
	require.NoError(t, err)
}
