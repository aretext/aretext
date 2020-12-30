package exec

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMenuSearch(t *testing.T) {
	testCases := []struct {
		name     string
		query    string
		items    []MenuItem
		expected []MenuItem
	}{
		{
			name:     "no items, empty query",
			query:    "",
			items:    []MenuItem{},
			expected: []MenuItem{},
		},
		{
			name:     "no items, nonempty query",
			query:    "a",
			items:    []MenuItem{},
			expected: []MenuItem{},
		},
		{
			name:  "some items, empty query",
			query: "",
			items: []MenuItem{
				{Name: "a"},
				{Name: "b"},
				{Name: "c"},
			},
			expected: []MenuItem{},
		},
		{
			name:  "some items, prefix match first char",
			query: "a",
			items: []MenuItem{
				{Name: "a"},
				{Name: "ab"},
				{Name: "ac"},
				{Name: "b"},
				{Name: "ba"},
				{Name: "bc"},
			},
			expected: []MenuItem{
				{Name: "a"},
				{Name: "ab"},
				{Name: "ac"},
			},
		},
		{
			name:  "some items, prefix match two chars",
			query: "ba",
			items: []MenuItem{
				{Name: "a"},
				{Name: "ab"},
				{Name: "ac"},
				{Name: "b"},
				{Name: "ba"},
				{Name: "bc"},
			},
			expected: []MenuItem{
				{Name: "ba"},
			},
		},
		{
			name:  "some items, prefix match two words",
			query: "foo/se",
			items: []MenuItem{
				{Name: "foo/first.txt"},
				{Name: "foo/second.txt"},
				{Name: "bar/first.txt"},
				{Name: "bar/second.txt"},
			},
			expected: []MenuItem{
				{Name: "foo/second.txt"},
			},
		},
		{
			name:  "some items, prefix match last word",
			query: "fir",
			items: []MenuItem{
				{Name: "foo/first.txt"},
				{Name: "foo/second.txt"},
				{Name: "bar/first.txt"},
				{Name: "bar/second.txt"},
			},
			expected: []MenuItem{
				{Name: "bar/first.txt"},
				{Name: "foo/first.txt"},
			},
		},
		{
			name:  "some items, case insensitive match",
			query: "FoO",
			items: []MenuItem{
				{Name: "fOo/first.txt"},
				{Name: "Foo/second.txt"},
				{Name: "foo/third.txt"},
				{Name: "bar/first.txt"},
				{Name: "bar/second.txt"},
			},
			expected: []MenuItem{
				{Name: "Foo/second.txt"},
				{Name: "fOo/first.txt"},
				{Name: "foo/third.txt"},
			},
		},
		{
			name:  "some items, some partially match",
			query: "bar baz tes",
			items: []MenuItem{
				{Name: "foo/bar/test.txt"},
				{Name: "foo/bar/test.txt"},
				{Name: "foo/bar/baz/test.txt"},
			},
			expected: []MenuItem{
				{Name: "foo/bar/baz/test.txt"},
			},
		},
		{
			name:  "items with shared prefix words, select shorter",
			query: "s",
			items: []MenuItem{
				{Name: "save"},
				{Name: "force save"},
			},
			expected: []MenuItem{
				{Name: "save"},
				{Name: "force save"},
			},
		},
		{
			name:  "items with shared prefix words, select longer",
			query: "f s",
			items: []MenuItem{
				{Name: "save"},
				{Name: "force save"},
			},
			expected: []MenuItem{
				{Name: "force save"},
			},
		},
		{
			name:  "all separators",
			query: "///",
			items: []MenuItem{
				{Name: " / / - _"},
				{Name: "///"},
				{Name: "   "},
				{Name: " / /"},
			},
			expected: []MenuItem{},
		},
		{
			name:  "empty query, all separators",
			query: "",
			items: []MenuItem{
				{Name: " / / - _"},
				{Name: "///"},
				{Name: "   "},
				{Name: " / /"},
				{Name: "/"},
				{Name: "  "},
			},
			expected: []MenuItem{},
		},
		{
			name:  "find file extension without dot prefix",
			query: "go",
			items: []MenuItem{
				{Name: "foo/bar/test.txt"},
				{Name: "foo/bar/test.go"},
				{Name: "foo/baz/test.go"},
				{Name: "doc.txt"},
				{Name: "main.go"},
			},
			expected: []MenuItem{
				{Name: "foo/bar/test.go"},
				{Name: "foo/baz/test.go"},
				{Name: "main.go"},
			},
		},
		{
			name:  "find file extension with dot prefix",
			query: ".go",
			items: []MenuItem{
				{Name: "foo/bar/test.txt"},
				{Name: "foo/bar/test.go"},
				{Name: "foo/baz/test.go"},
				{Name: "doc.txt"},
				{Name: "main.go"},
			},
			expected: []MenuItem{
				{Name: "foo/bar/test.go"},
				{Name: "foo/baz/test.go"},
				{Name: "main.go"},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			s := &MenuSearch{}
			s.SetQuery(tc.query)
			s.AddItems(tc.items)
			assert.Equal(t, tc.expected, s.Results())
		})
	}
}

func BenchmarkMenuSearchAddItemsEmptyQuery(b *testing.B) {
	items := fakeMenuItems(1000, "foo/bar/baz/bat/test")
	for i := 0; i < b.N; i++ {
		s := &MenuSearch{}
		s.AddItems(items)
	}
}

func BenchmarkMenuSearchAddItemsWithQuery(b *testing.B) {
	items := fakeMenuItems(1000, "foo/bar/baz/bat/test")
	for i := 0; i < b.N; i++ {
		s := &MenuSearch{}
		s.SetQuery("test.txt")
		s.AddItems(items)
	}
}

func BenchmarkMenuSearchSetQuery(b *testing.B) {
	s := &MenuSearch{}
	s.AddItems(fakeMenuItems(1000, "foo/bar/baz/bat/test"))
	for i := 0; i < b.N; i++ {
		if i%2 == 0 {
			s.SetQuery("foo")
		} else {
			s.SetQuery("bar")
		}
	}
}

func fakeMenuItems(n int, prefix string) []MenuItem {
	items := make([]MenuItem, 0, n)
	for i := 0; i < n; i++ {
		name := fmt.Sprintf("%s/%d.txt", prefix, i)
		items = append(items, MenuItem{Name: name})
	}
	return items
}
