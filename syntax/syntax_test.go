package syntax

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/aretext/aretext/syntax/parser"
)

func TestParserForLanguageCachesParseFunc(t *testing.T) {
	testCases := []struct {
		name      string
		language  Language
		expectNil bool
	}{
		{
			name:      "json",
			language:  LanguageJson,
			expectNil: false,
		},
		{
			name:      "plaintext",
			language:  LanguagePlaintext,
			expectNil: true,
		},
		{
			name:      "unknown",
			language:  Language("unknown"),
			expectNil: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// reset the cache
			parseFuncCache = make(map[Language]parser.Func)

			// first retrieval constructs the parser.
			p1 := ParserForLanguage(tc.language)
			if tc.expectNil {
				assert.Nil(t, p1)
			} else {
				assert.NotNil(t, p1)
			}

			// parse func should have been cached.
			pf, ok := parseFuncCache[tc.language]
			require.True(t, ok)
			if tc.expectNil {
				assert.Nil(t, pf)
			} else {
				assert.NotNil(t, pf)
			}

			// retrieve again, should hit the cache
			p2 := ParserForLanguage(tc.language)
			if tc.expectNil {
				assert.Nil(t, p2)
			} else {
				assert.NotNil(t, p2)
			}
		})
	}
}
