package menu

import (
	"sort"
	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"

	"github.com/aretext/aretext/text"
)

const (
	scoreMismatch = iota
	scoreMatchOffBoundary
	scoreMatchAtBoundaryPastStart
	scoreMatchAtStart
	scoreExactMatchAlias
)

const maxScore = scoreExactMatchAlias

// Search performs approximate text searches for menu items matching a query string.
type Search struct {
	query             string
	emptyQueryShowAll bool
	items             []Item
	normalizedNames   []string
	matchPositions    []uint64
	results           []Item
}

func NewSearch(items []Item, emptyQueryShowAll bool) *Search {
	sortItemsInLexicographicOrder(items)
	normalizedNames := make([]string, len(items))
	for i, item := range items {
		normalizedNames[i] = normalizeString(item.Name)
	}

	results := make([]Item, 0, len(items))
	if emptyQueryShowAll {
		results = append(results, items...)
	}
	return &Search{
		emptyQueryShowAll: emptyQueryShowAll,
		items:             items,
		normalizedNames:   normalizedNames,
		results:           results,
	}
}

// Query returns the current query.
func (s *Search) Query() string {
	return s.query
}

// SetQuery updates the query for the search.
func (s *Search) SetQuery(q string) {
	if s.query == q {
		return
	}
	s.query = q
	s.results = s.results[:0]

	if len(q) == 0 {
		if s.emptyQueryShowAll {
			s.results = append(s.results, s.items...)
		} else {
			s.results = s.results[:0]
		}
		return
	}

	var scoreBuckets [maxScore][]int
	querySearcher := text.NewSearcher(normalizeString(q))
	for i := 0; i < len(s.items); i++ {
		score := s.scoreItemForQuery(
			s.normalizedNames[i],
			s.items[i].Aliases,
			s.query,
			querySearcher,
		)
		if score > 0 {
			bucketIdx := score - 1
			scoreBuckets[bucketIdx] = append(scoreBuckets[bucketIdx], i)
		}
	}

	for bucketIdx := len(scoreBuckets) - 1; bucketIdx >= 0; bucketIdx-- {
		for _, itemIdx := range scoreBuckets[bucketIdx] {
			s.results = append(s.results, s.items[itemIdx])
		}
	}
}

// Results returns the menu items matching the current query.
// Items are sorted descending by similarity to the query,
// with ties broken by lexicographic ordering.
func (s *Search) Results() []Item {
	return s.results
}

func (s *Search) scoreItemForQuery(itemName string, itemAliases []string, query string, querySearcher *text.Searcher) int {
	for _, alias := range itemAliases {
		if normalizeString(alias) == normalizeString(query) {
			return scoreExactMatchAlias
		}
	}

	s.matchPositions = querySearcher.AllInString(itemName, s.matchPositions)
	if len(s.matchPositions) == 0 {
		return scoreMismatch
	}

	if s.matchPositions[0] == 0 {
		return scoreMatchAtStart
	}

	var i int
	var pos uint64
	var prevRune rune
	for _, r := range itemName {
		matchPos := s.matchPositions[i]
		if pos == matchPos {
			if unicode.IsPunct(prevRune) && !unicode.IsPunct(r) {
				return scoreMatchAtBoundaryPastStart
			} else if unicode.IsSpace(prevRune) && !unicode.IsSpace(r) {
				return scoreMatchAtBoundaryPastStart
			}

			i++
			if i == len(s.matchPositions) {
				break
			}
		}
		prevRune = r
		pos++
	}

	return scoreMatchOffBoundary
}

func normalizeString(s string) string {
	return strings.ToLower(norm.NFC.String(s))
}

func sortItemsInLexicographicOrder(items []Item) {
	sort.SliceStable(items, func(i, j int) bool {
		return items[i].Name < items[j].Name
	})
}
