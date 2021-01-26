package config

import (
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSaveAndLoadRuleSet(t *testing.T) {
	rs := RuleSet{
		Rules: []Rule{
			{
				Name:    "default",
				Pattern: "**",
				Config: map[string]interface{}{
					"syntaxLanguage": "undefined",
				},
			},
			{
				Name:    "json",
				Pattern: "**/*.json",
				Config: map[string]interface{}{
					"syntaxLanguage": "json",
				},
			},
		},
	}

	tmpDir, err := ioutil.TempDir("", "")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	path := path.Join(tmpDir, "aretext", "config.json")
	err = SaveRuleSet(path, rs)
	require.NoError(t, err)

	loadedRs, err := LoadRuleSet(path)
	require.NoError(t, err)
	assert.Equal(t, rs, loadedRs)
}
