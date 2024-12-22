package editor

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultConfigYamlValid(t *testing.T) {
	rs, err := unmarshalRuleSet(DefaultConfigYaml)
	require.NoError(t, err)
	assert.Greater(t, len(rs), 1)
	require.NoError(t, rs.Validate())

	c := rs.ConfigForPath("test.go")
	assert.Equal(t, "go", c.SyntaxLanguage)
	assert.Equal(t, 4, c.TabSize)
	assert.True(t, c.AutoIndent)
	assert.Equal(t, "olive", c.Styles["lineNum"].Color)
}
