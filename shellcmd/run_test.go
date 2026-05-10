package shellcmd

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunWithStdinAndStdout(t *testing.T) {
	var stdout bytes.Buffer

	err := Run(context.Background(), "cat", nil, strings.NewReader("abcd"), &stdout, nil)
	require.NoError(t, err)
	assert.Equal(t, "abcd", stdout.String())
}
