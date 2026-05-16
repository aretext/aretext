package languages

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDockerfileParseFunc(t *testing.T) {
	testCases := []struct {
		name     string
		text     string
		expected []TokenWithText
	}{
		// TODO: add test cases
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tokens := ParseTokensWithText(DockerfileParseFunc(), tc.text)
			assert.Equal(t, tc.expected, tokens)
		})
	}
}
