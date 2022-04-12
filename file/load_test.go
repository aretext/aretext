package file

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad(t *testing.T) {
	testCases := []struct {
		name                 string
		fileContents         string
		expectedTreeContents string
	}{
		{
			name:                 "empty",
			fileContents:         "",
			expectedTreeContents: "",
		},
		{
			name:                 "ends with character, no POSIX eof",
			fileContents:         "ab\ncd",
			expectedTreeContents: "ab\ncd",
		},
		{
			name:                 "POSIX eof",
			fileContents:         "abcd\n",
			expectedTreeContents: "abcd",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			filePath := createTestFile(t, tc.fileContents)

			tree, watcher, err := Load(filePath, time.Second)
			require.NoError(t, err)
			defer watcher.Stop()

			assert.Equal(t, tc.expectedTreeContents, tree.String())
		})
	}
}
