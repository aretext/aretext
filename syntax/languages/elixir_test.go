package languages

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/aretext/aretext/syntax/parser"
)

func TestElixirParseFunc(t *testing.T) {
	testCases := []struct {
		name     string
		text     string
		expected []TokenWithText
	}{
		{
			name: "line comment",
			text: "x = 1 # assign one",
			expected: []TokenWithText{
				{Text: "=", Role: parser.TokenRoleOperator},
				{Text: "1", Role: parser.TokenRoleNumber},
				{Text: "# assign one", Role: parser.TokenRoleComment},
			},
		},
		{
			name: "shebang comment",
			text: "#!/usr/bin/env elixir\n",
			expected: []TokenWithText{
				{Text: "#!/usr/bin/env elixir\n", Role: parser.TokenRoleComment},
			},
		},
		{
			name: "keywords in module definition",
			text: "defmodule Foo do\nend",
			expected: []TokenWithText{
				{Text: "defmodule", Role: parser.TokenRoleKeyword},
				{Text: "do", Role: parser.TokenRoleKeyword},
				{Text: "end", Role: parser.TokenRoleKeyword},
			},
		},
		{
			name: "boolean and nil keywords",
			text: "a = true\nb = false\nc = nil",
			expected: []TokenWithText{
				{Text: "=", Role: parser.TokenRoleOperator},
				{Text: "true", Role: parser.TokenRoleKeyword},
				{Text: "=", Role: parser.TokenRoleOperator},
				{Text: "false", Role: parser.TokenRoleKeyword},
				{Text: "=", Role: parser.TokenRoleOperator},
				{Text: "nil", Role: parser.TokenRoleKeyword},
			},
		},
		{
			name:     "predicate identifier is not a keyword",
			text:     "valid?",
			expected: []TokenWithText{},
		},
		{
			name:     "bang identifier is not a keyword",
			text:     "save!",
			expected: []TokenWithText{},
		},
		{
			name: "double quoted string",
			text: `"hello world"`,
			expected: []TokenWithText{
				{Text: `"hello world"`, Role: parser.TokenRoleString},
			},
		},
		{
			name: "double quoted string with escape",
			text: `"line\nbreak"`,
			expected: []TokenWithText{
				{Text: `"line\nbreak"`, Role: parser.TokenRoleString},
			},
		},
		{
			name: "charlist",
			text: `'a charlist'`,
			expected: []TokenWithText{
				{Text: `'a charlist'`, Role: parser.TokenRoleString},
			},
		},
		{
			name: "heredoc",
			text: "\"\"\"\nmultiline\n\"string\"\n\"\"\"",
			expected: []TokenWithText{
				{Text: "\"\"\"\nmultiline\n\"string\"\n\"\"\"", Role: parser.TokenRoleString},
			},
		},
		{
			name: "string interpolation stays within the string token",
			text: `"count: #{n}"`,
			expected: []TokenWithText{
				{Text: `"count: #{n}"`, Role: parser.TokenRoleString},
			},
		},
		{
			name: "regex sigil with modifier",
			text: "~r/foo.*bar/i",
			expected: []TokenWithText{
				{Text: "~r/foo.*bar/i", Role: parser.TokenRoleString},
			},
		},
		{
			name: "word list sigil with paired delimiter",
			text: "~w[one two three]a",
			expected: []TokenWithText{
				{Text: "~w[one two three]a", Role: parser.TokenRoleString},
			},
		},
		{
			name: "string sigil with parens",
			text: "~s(no # comment here)",
			expected: []TokenWithText{
				{Text: "~s(no # comment here)", Role: parser.TokenRoleString},
			},
		},
		{
			name: "heredoc sigil",
			text: "~S\"\"\"\nRaw #{not_interpolated} heredoc.\n\"\"\"",
			expected: []TokenWithText{
				{Text: "~S\"\"\"\nRaw #{not_interpolated} heredoc.\n\"\"\"", Role: parser.TokenRoleString},
			},
		},
		{
			name: "atom",
			text: ":ok",
			expected: []TokenWithText{
				{Text: ":ok", Role: elixirTokenRoleAtom},
			},
		},
		{
			name: "predicate atom",
			text: ":valid?",
			expected: []TokenWithText{
				{Text: ":valid?", Role: elixirTokenRoleAtom},
			},
		},
		{
			name: "quoted atom",
			text: `:"with spaces"`,
			expected: []TokenWithText{
				{Text: `:"with spaces"`, Role: elixirTokenRoleAtom},
			},
		},
		{
			name: "atom in tuple",
			text: "{:ok, result}",
			expected: []TokenWithText{
				{Text: ":ok", Role: elixirTokenRoleAtom},
			},
		},
		{
			name: "double colon is an operator not an atom",
			text: "x :: integer",
			expected: []TokenWithText{
				{Text: "::", Role: parser.TokenRoleOperator},
			},
		},
		{
			name: "module attribute",
			text: "@moduledoc",
			expected: []TokenWithText{
				{Text: "@moduledoc", Role: elixirTokenRoleAttribute},
			},
		},
		{
			name: "attribute with value",
			text: "@timeout 5000",
			expected: []TokenWithText{
				{Text: "@timeout", Role: elixirTokenRoleAttribute},
				{Text: "5000", Role: parser.TokenRoleNumber},
			},
		},
		{
			name: "decimal integer with separators",
			text: "1_000_000",
			expected: []TokenWithText{
				{Text: "1_000_000", Role: parser.TokenRoleNumber},
			},
		},
		{
			name: "hex integer",
			text: "0xCAFE",
			expected: []TokenWithText{
				{Text: "0xCAFE", Role: parser.TokenRoleNumber},
			},
		},
		{
			name: "octal integer",
			text: "0o755",
			expected: []TokenWithText{
				{Text: "0o755", Role: parser.TokenRoleNumber},
			},
		},
		{
			name: "binary integer",
			text: "0b1010",
			expected: []TokenWithText{
				{Text: "0b1010", Role: parser.TokenRoleNumber},
			},
		},
		{
			name: "float",
			text: "3.14",
			expected: []TokenWithText{
				{Text: "3.14", Role: parser.TokenRoleNumber},
			},
		},
		{
			name: "float with exponent",
			text: "1.5e-10",
			expected: []TokenWithText{
				{Text: "1.5e-10", Role: parser.TokenRoleNumber},
			},
		},
		{
			name: "range is two numbers around an operator",
			text: "1..10",
			expected: []TokenWithText{
				{Text: "1", Role: parser.TokenRoleNumber},
				{Text: "..", Role: parser.TokenRoleOperator},
				{Text: "10", Role: parser.TokenRoleNumber},
			},
		},
		{
			name: "char code",
			text: "?a",
			expected: []TokenWithText{
				{Text: "?a", Role: parser.TokenRoleNumber},
			},
		},
		{
			name: "escaped char code",
			text: `?\n`,
			expected: []TokenWithText{
				{Text: `?\n`, Role: parser.TokenRoleNumber},
			},
		},
		{
			name: "pipe operator",
			text: "list |> Enum.map(fn x -> x + 1 end)",
			expected: []TokenWithText{
				{Text: "|>", Role: parser.TokenRoleOperator},
				{Text: ".", Role: parser.TokenRoleOperator},
				{Text: "fn", Role: parser.TokenRoleKeyword},
				{Text: "->", Role: parser.TokenRoleOperator},
				{Text: "+", Role: parser.TokenRoleOperator},
				{Text: "1", Role: parser.TokenRoleNumber},
				{Text: "end", Role: parser.TokenRoleKeyword},
			},
		},
		{
			name: "match and concat operators",
			text: `greeting = "hello" <> name`,
			expected: []TokenWithText{
				{Text: "=", Role: parser.TokenRoleOperator},
				{Text: `"hello"`, Role: parser.TokenRoleString},
				{Text: "<>", Role: parser.TokenRoleOperator},
			},
		},
		{
			name: "full function",
			text: `defmodule Greeter do
  @greeting "Hello"

  def hello(name) do
    IO.puts("#{@greeting}, #{name}!")
    :ok
  end
end`,
			expected: []TokenWithText{
				{Text: "defmodule", Role: parser.TokenRoleKeyword},
				{Text: "do", Role: parser.TokenRoleKeyword},
				{Text: "@greeting", Role: elixirTokenRoleAttribute},
				{Text: `"Hello"`, Role: parser.TokenRoleString},
				{Text: "def", Role: parser.TokenRoleKeyword},
				{Text: "do", Role: parser.TokenRoleKeyword},
				{Text: ".", Role: parser.TokenRoleOperator},
				{Text: `"#{@greeting}, #{name}!"`, Role: parser.TokenRoleString},
				{Text: ":ok", Role: elixirTokenRoleAtom},
				{Text: "end", Role: parser.TokenRoleKeyword},
				{Text: "end", Role: parser.TokenRoleKeyword},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tokens := ParseTokensWithText(ElixirParseFunc(), tc.text)
			assert.Equal(t, tc.expected, tokens)
		})
	}
}

func BenchmarkElixirParser(b *testing.B) {
	ParserBenchmark(b, ElixirParseFunc(), "testdata/elixir/example.ex")
}

func FuzzElixirParseFunc(f *testing.F) {
	seeds := LoadFuzzTestSeeds(f, "./testdata/elixir/*")
	ParserFuzzTest(f, ElixirParseFunc(), seeds)
}
