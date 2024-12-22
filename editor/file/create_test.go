package file

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidateCreateSuccess(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "test.txt")
	err := ValidateCreate(path)
	require.NoError(t, err)
}

func TestValidateCreateEmptyPath(t *testing.T) {
	err := ValidateCreate("")
	require.EqualError(t, err, "File path is empty")
}

func TestValidateCreatePathEndingWithDirectory(t *testing.T) {
	err := ValidateCreate("foobar/")
	require.EqualError(t, err, "File path must end with a file name")
}

func TestValidateCreateDirectoryDoesNotExist(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "fakeDir/test.txt")
	err := ValidateCreate(path)
	require.ErrorContains(t, err, "Directory does not exist")
}

func TestValidateCreateDirectoryExistsButIsAFile(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "fakeDir")
	f, err := os.Create(path)
	require.NoError(t, err)
	defer f.Close()

	path = filepath.Join(path, "test.txt")
	err = ValidateCreate(path)
	require.ErrorContains(t, err, "Not a directory")
}

func TestValidateCreateFileAlreadyExists(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "test.txt")
	f, err := os.Create(path)
	require.NoError(t, err)
	defer f.Close()

	err = ValidateCreate(path)
	require.ErrorContains(t, err, "File already exists")
}
