package app

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultConfigJsonValid(t *testing.T) {
	rs, err := unmarshalRuleSet(DefaultConfigJson)
	require.NoError(t, err)
	assert.Greater(t, len(rs), 1)

	c := rs.ConfigForPath("test.go")
	assert.Equal(t, "go", c.SyntaxLanguage)
	assert.Equal(t, 4, c.TabSize)
	assert.True(t, c.AutoIndent)
}
