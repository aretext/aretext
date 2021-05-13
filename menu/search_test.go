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
			items:    []Item{},
			expected: []Item{},
		},
		{
			name:              "no items, empty query with emptyQueryShowAll true",
			query:             "",
			emptyQueryShowAll: true,
			items:             []Item{},
			expected:          []Item{},
		},
		{
			name:     "no items, nonempty query",
			query:    "a",
			items:    []Item{},
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
			expected: []Item{},
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
			name:  "some items, prefix match first char",
			query: "a",
			items: []Item{
				{Name: "a"},
				{Name: "ab"},
				{Name: "ac"},
				{Name: "b"},
				{Name: "ba"},
				{Name: "bc"},
			},
			expected: []Item{
				{Name: "a"},
				{Name: "ab"},
				{Name: "ac"},
				{Name: "ba"},
			},
		},
		{
			name:  "some items, prefix match two chars",
			query: "ba",
			items: []Item{
				{Name: "a"},
				{Name: "ab"},
				{Name: "ac"},
				{Name: "b"},
				{Name: "ba"},
				{Name: "bc"},
			},
			expected: []Item{
				{Name: "ba"},
			},
		},
		{
			name:  "some items, prefix match two words",
			query: "foo/se",
			items: []Item{
				{Name: "foo/first.txt"},
				{Name: "foo/second.txt"},
				{Name: "bar/first.txt"},
				{Name: "bar/second.txt"},
			},
			expected: []Item{
				{Name: "foo/second.txt"},
			},
		},
		{
			name:  "some items, prefix match last word",
			query: "fir",
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
			name:  "some items, case insensitive match",
			query: "FoO",
			items: []Item{
				{Name: "fOo/first.txt"},
				{Name: "Foo/second.txt"},
				{Name: "foo/third.txt"},
				{Name: "bar/first.txt"},
				{Name: "bar/second.txt"},
			},
			expected: []Item{
				{Name: "Foo/second.txt"},
				{Name: "fOo/first.txt"},
				{Name: "foo/third.txt"},
			},
		},
		{
			name:  "items with shared prefix, select shorter",
			query: "s",
			items: []Item{
				{Name: "save"},
				{Name: "force save"},
			},
			expected: []Item{
				{Name: "save"},
				{Name: "force save"},
			},
		},
		{
			name:  "find file extension without dot prefix",
			query: "go",
			items: []Item{
				{Name: "foo/bar/test.txt"},
				{Name: "foo/bar/test.go"},
				{Name: "foo/baz/test.go"},
				{Name: "doc.txt"},
				{Name: "main.go"},
			},
			expected: []Item{
				{Name: "foo/bar/test.go"},
				{Name: "foo/baz/test.go"},
				{Name: "main.go"},
			},
		},
		{
			name:  "find file extension with dot prefix",
			query: ".go",
			items: []Item{
				{Name: "foo/bar/test.txt"},
				{Name: "foo/bar/test.go"},
				{Name: "foo/baz/test.go"},
				{Name: "doc.txt"},
				{Name: "main.go"},
			},
			expected: []Item{
				{Name: "foo/bar/test.go"},
				{Name: "foo/baz/test.go"},
				{Name: "main.go"},
			},
		},
		{
			name:  "rank exact match for alias first",
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
			name:  "rank case-insensitive match for alias first",
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
			name:  "rank match at boundary after punctuation",
			query: "foo",
			items: []Item{
				{Name: "baz/bar"},
				{Name: "baz/boofoo"},
				{Name: "baz/foobar"},
				{Name: "baz/barfoo"},
			},
			expected: []Item{
				{Name: "baz/foobar"},
				{Name: "baz/barfoo"},
				{Name: "baz/boofoo"},
			},
		},
		{
			name:  "rank match at boundary after whitespace",
			query: "foo",
			items: []Item{
				{Name: "baz bar"},
				{Name: "baz boofoo"},
				{Name: "baz foobar"},
				{Name: "baz barfoo"},
			},
			expected: []Item{
				{Name: "baz foobar"},
				{Name: "baz barfoo"},
				{Name: "baz boofoo"},
			},
		},
		{
			name:  "rank match at start",
			query: "foo",
			items: []Item{
				{Name: "baz foo"},
				{Name: "foo bar"},
				{Name: "bar foo"},
			},
			expected: []Item{
				{Name: "foo bar"},
				{Name: "bar foo"},
				{Name: "baz foo"},
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

func fakeItems(n int, prefix string) []Item {
	items := make([]Item, 0, n)
	for i := 0; i < n; i++ {
		name := fmt.Sprintf("%s/%d.txt", prefix, i)
		items = append(items, Item{Name: name})
	}
	return items
}
