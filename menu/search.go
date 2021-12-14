package menu

import (
	"sort"
	"strings"

	"github.com/aretext/aretext/menu/fuzzy"
)

const (
	maxSearchItemNameLen = 1024
	maxSearchQueryLen    = 1024
)

// Search performs approximate text searches for menu items matching a query string.
type Search struct {
	query             string
	emptyQueryShowAll bool
	fuzzyIndex        *fuzzy.Index
	aliasIndex        map[string]int
	items             []Item
	results           []Item
}

func NewSearch(items []Item, emptyQueryShowAll bool) *Search {
	itemNames := make([]string, len(items))
	aliasIndex := make(map[string]int, 0)
	for itemId, item := range items {
		// Truncate long names to avoid perf issues when fuzzy searching.
		itemNames[itemId] = truncateString(item.Name, maxSearchItemNameLen)
		for _, alias := range item.Aliases {
			aliasIndex[alias] = itemId
		}
	}

	var results []Item
	if emptyQueryShowAll {
		results = append(results, items...)
		sortItemsInLexicographicOrder(results)
	}

	return &Search{
		emptyQueryShowAll: emptyQueryShowAll,
		fuzzyIndex:        fuzzy.NewIndex(itemNames),
		aliasIndex:        aliasIndex,
		items:             items,
		results:           results,
	}
}

// Query returns the current query.
func (s *Search) Query() string {
	return s.query
}

// SetQuery updates the query for the search.
func (s *Search) SetQuery(q string) {
	if q == s.query {
		return
	} else {
		s.query = q
	}

	if len(q) == 0 {
		if s.emptyQueryShowAll {
			s.results = make([]Item, 0, len(s.items))
			s.results = append(s.results, s.items...)
			sortItemsInLexicographicOrder(s.results)
		} else {
			s.results = nil
		}
		return
	}

	// Truncate long queries to avoid perf issues when fuzzy searching.
	truncatedQuery := truncateString(q, maxSearchQueryLen)
	resultItemIds := s.fuzzyIndex.Search(truncatedQuery)
	results := make([]Item, 0, len(resultItemIds)+1)
	itemIdMatchingAlias := -1
	if itemId, ok := s.aliasIndex[strings.ToLower(truncatedQuery)]; ok {
		itemIdMatchingAlias = itemId
		results = append(results, s.items[itemId])
	}
	for _, itemId := range resultItemIds {
		if itemId != itemIdMatchingAlias {
			results = append(results, s.items[itemId])
		}
	}
	s.results = results
}

// Results returns the menu items matching the current query.
// Items are sorted descending by relevance to the query,
// with ties broken by lexicographic ordering.
func (s *Search) Results() []Item {
	return s.results
}

func truncateString(s string, maxLen int) string {
	if len(s) > maxLen {
		return s[0:maxLen]
	} else {
		return s
	}
}

func sortItemsInLexicographicOrder(items []Item) {
	sort.SliceStable(items, func(i, j int) bool {
		return items[i].Name < items[j].Name
	})
}
