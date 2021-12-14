package menu

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSearch(t *testing.T) {
	testCases := []struct {
		name              string
		query             string
		items             []Item
		emptyQueryShowAll bool
		expected          []Item
	}{
		{
			name:     "no items, empty query",
			query:    "",
			items:    nil,
			expected: nil,
		},
		{
			name:              "no items, empty query with emptyQueryShowAll true",
			query:             "",
			emptyQueryShowAll: true,
			items:             nil,
			expected:          nil,
		},
		{
			name:     "no items, nonempty query",
			query:    "a",
			items:    nil,
			expected: []Item{},
		},
		{
			name:  "some items, empty query",
			query: "",
			items: []Item{
				{Name: "a"},
				{Name: "b"},
				{Name: "c"},
			},
			expected: nil,
		},
		{
			name:              "some items, empty query with emptyQueryShowAll true",
			query:             "",
			emptyQueryShowAll: true,
			items: []Item{
				{Name: "a"},
				{Name: "b"},
				{Name: "c"},
			},
			expected: []Item{
				{Name: "a"},
				{Name: "b"},
				{Name: "c"},
			},
		},
		{
			name:  "exact match for alias",
			query: "w",
			items: []Item{
				{Name: "write one"},
				{Name: "write two", Aliases: []string{"w"}},
				{Name: "write three"},
			},
			expected: []Item{
				{Name: "write two", Aliases: []string{"w"}},
				{Name: "write one"},
				{Name: "write three"},
			},
		},
		{
			name:  "case-insensitive match for alias",
			query: "W",
			items: []Item{
				{Name: "write one"},
				{Name: "write two", Aliases: []string{"w"}},
				{Name: "write three"},
				{Name: "Write capitalized"},
			},
			expected: []Item{
				{Name: "write two", Aliases: []string{"w"}},
				{Name: "Write capitalized"},
				{Name: "write one"},
				{Name: "write three"},
			},
		},
		{
			name:  "commands",
			query: "togle", // deliberate typo, should still fuzzy-match "toggle"
			items: []Item{
				{Name: "quit"},
				{Name: "force quit"},
				{Name: "save document"},
				{Name: "force save document"},
				{Name: "force reload"},
				{Name: "find and open"},
				{Name: "open next document"},
				{Name: "toggle tab expand"},
				{Name: "toggle line numbers"},
			},
			expected: []Item{
				{Name: "toggle tab expand"},
				{Name: "toggle line numbers"},
			},
		},
		{
			name:  "paths",
			query: "firs",
			items: []Item{
				{Name: "foo/first.txt"},
				{Name: "foo/second.txt"},
				{Name: "bar/first.txt"},
				{Name: "bar/second.txt"},
			},
			expected: []Item{
				{Name: "bar/first.txt"},
				{Name: "foo/first.txt"},
			},
		},
		{
			name:  "non-ascii unicode",
			query: "𝓯𝓸",
			items: []Item{
				{Name: "𝓯𝓸𝓸"},
				{Name: "ᵦₐᵣ"},
				{Name: "乃ﾑ乙"},
			},
			expected: []Item{
				{Name: "𝓯𝓸𝓸"},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			s := NewSearch(tc.items, tc.emptyQueryShowAll)
			s.SetQuery(tc.query)
			assert.Equal(t, tc.expected, s.Results())
		})
	}
}

func BenchmarkSearch(b *testing.B) {
	s := NewSearch(fakeItems(1000, "foo/bar/baz/bat/test"), false)
	for i := 0; i < b.N; i++ {
		if i%2 == 0 {
			s.SetQuery("foo")
		} else {
			s.SetQuery("bar")
		}
		s.Results()
	}
}

func BenchmarkIncrementalSearch(b *testing.B) {
	s := NewSearch(fakeItems(1000, "foo/bar/baz/bat/test"), false)
	q := "test/123"
	for i := 0; i < b.N; i++ {
		for i := 1; i < len(q); i++ {
			s.SetQuery(q[0:i])
		}
		s.Results()
	}
}

func fakeItems(n int, prefix string) []Item {
	items := make([]Item, 0, n)
	for i := 0; i < n; i++ {
		name := fmt.Sprintf("%s/%d.txt", prefix, i)
		items = append(items, Item{Name: name})
	}
	return items
}
