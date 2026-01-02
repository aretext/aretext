package languages

import (
	"errors"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/aretext/aretext/syntax/parser"
	"github.com/aretext/aretext/text"
)

// TokenWithText is a token that includes its text value.
type TokenWithText struct {
	Role parser.TokenRole
	Text string
}

// ParseTokensWithText tokenizes the input string using the specified parse func.
func ParseTokensWithText(f parser.Func, s string) []TokenWithText {
	p := parser.New(f)
	tree, err := text.NewTreeFromString(s)
	if err != nil {
		panic(err)
	}

	stringSlice := func(startPos, endPos uint64) string {
		var sb strings.Builder
		reader := tree.ReaderAtPosition(startPos)
		for i := startPos; i < endPos; i++ {
			r, _, err := reader.ReadRune()
			if err != nil {
				break
			}
			_, err = sb.WriteRune(r)
			if err != nil {
				panic(err)
			}
		}
		return sb.String()
	}

	p.ParseAll(tree)
	tokens := p.TokensIntersectingRange(0, math.MaxUint64)
	tokensWithText := make([]TokenWithText, 0, len(tokens))
	for _, t := range tokens {
		tokensWithText = append(tokensWithText, TokenWithText{
			Role: t.Role,
			Text: stringSlice(t.StartPos, t.EndPos),
		})
	}
	return tokensWithText
}

// ParserBenchmark benchmarks a parser with the input file located at `path`.
func ParserBenchmark(b *testing.B, f parser.Func, path string) {
	data, err := os.ReadFile(path)
	require.NoError(b, err)
	tree, err := text.NewTreeFromString(string(data))
	require.NoError(b, err)

	p := parser.New(f)
	for i := 0; i < b.N; i++ {
		p.ParseAll(tree)
	}
}

// ParserFuzzTest runs a fuzz test on a parser.
func ParserFuzzTest(f *testing.F, parseFunc parser.Func, seeds []string) {
	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, data string) {
		tree, err := text.NewTreeFromString(data)
		if errors.Is(err, text.ErrInvalidUtf8) {
			t.Skip()
		}
		require.NoError(t, err)
		p := parser.New(parseFunc)
		p.ParseAll(tree)
	})
}

// LoadFuzzTestSeeds loads seed data from files matching a glob pattern.
func LoadFuzzTestSeeds(f *testing.F, globPattern string) []string {
	var seeds []string

	matches, err := filepath.Glob(globPattern)
	if err != nil {
		f.Fatalf("filepath.Glob: %s\n", err)
	}

	for _, path := range matches {
		data, err := os.ReadFile(path)
		if err != nil {
			f.Fatalf("os.ReadFile: %s\n", err)
		}

		f.Logf("Loaded seed file %s\n", path)
		seeds = append(seeds, string(data))
	}

	return seeds
}
