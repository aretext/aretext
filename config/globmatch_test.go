package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGlobMatch(t *testing.T) {
	testCases := []struct {
		name          string
		pattern       string
		path          string
		expectMatched bool
	}{
		{
			name:          "empty pattern, empty path",
			pattern:       "",
			path:          "",
			expectMatched: true,
		},
		{
			name:          "empty pattern, nonempty path",
			pattern:       "",
			path:          "abc",
			expectMatched: false,
		},
		{
			name:          "nonempty pattern, empty path",
			pattern:       "abc",
			path:          "",
			expectMatched: false,
		},
		{
			name:          "single component, exact match",
			pattern:       "foobar",
			path:          "foobar",
			expectMatched: true,
		},
		{
			name:          "single component, mismatch",
			pattern:       "foxbar",
			path:          "foobar",
			expectMatched: false,
		},
		{
			name:          "single component, match with single wildcard",
			pattern:       "*",
			path:          "foobar",
			expectMatched: true,
		},
		{
			name:          "single component, match with single wildcard, empty path",
			pattern:       "*",
			path:          "",
			expectMatched: true,
		},
		{
			name:          "single component, match with characters before single wildcard",
			pattern:       "foo*",
			path:          "foobar",
			expectMatched: true,
		},
		{
			name:          "single component, match with characters after single wildcard",
			pattern:       "*bar",
			path:          "foobar",
			expectMatched: true,
		},
		{
			name:          "single component, mismatch with wildcard",
			pattern:       "x*",
			path:          "foobar",
			expectMatched: false,
		},
		{
			name:          "multiple components, exact match",
			pattern:       "foo/bar/baz",
			path:          "foo/bar/baz",
			expectMatched: true,
		},
		{
			name:          "multiple components with double star wildcard prefix matches",
			pattern:       "**/baz",
			path:          "foo/bar/baz",
			expectMatched: true,
		},
		{
			name:          "multiple components with double star wildcard suffix matches",
			pattern:       "foo/**",
			path:          "foo/bar/baz",
			expectMatched: true,
		},
		{
			name:          "multiple components with double star wildcard prefix+suffix matches",
			pattern:       "foo/**/baz",
			path:          "foo/bar/baz",
			expectMatched: true,
		},
		{
			name:          "multiple components with double star wildcard mismatch",
			pattern:       "**/x",
			path:          "foo/bar/baz",
			expectMatched: false,
		},
		{
			name:          "multiple components with star and double star wildcard match",
			pattern:       "**/test_*.go",
			path:          "foo/bar/test_baz.go",
			expectMatched: true,
		},
		{
			name:          "multiple components with star and double star wildcard mismatch",
			pattern:       "**/test_*.go",
			path:          "foo/bar/baz.go",
			expectMatched: false,
		},
		{
			name:          "double star matches nonempty path",
			pattern:       "**",
			path:          "foo/bar/baz.go",
			expectMatched: true,
		},
		{
			name:          "double star matches empty path",
			pattern:       "**",
			path:          "",
			expectMatched: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			matched := GlobMatch(tc.pattern, tc.path)
			assert.Equal(t, tc.expectMatched, matched)
		})
	}
}
