package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidate(t *testing.T) {
	testCases := []struct {
		name              string
		buildConfig       func() Config
		expectValid       bool
		expectErrContains string
	}{
		{
			name:        "default",
			buildConfig: DefaultConfig,
			expectValid: true,
		},
		{
			name: "tab size zero",
			buildConfig: func() Config {
				c := DefaultConfig()
				c.TabSize = 0
				return c
			},
			expectErrContains: "TabSize must be greater than zero",
		},
		{
			name: "tab size negative",
			buildConfig: func() Config {
				c := DefaultConfig()
				c.TabSize = -1
				return c
			},
			expectErrContains: "TabSize must be greater than zero",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := tc.buildConfig()
			err := config.Validate()
			if tc.expectValid {
				assert.NoError(t, err)
			} else {
				assert.Contains(t, err.Error(), tc.expectErrContains)
			}
		})
	}
}
