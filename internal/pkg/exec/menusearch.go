package exec

import (
	"sort"
	"strings"

	"golang.org/x/text/unicode/norm"
)

// scoredItem is a menu item assigned a similarity score for a given query.
type scoredItem struct {
	// score is a similarity score for a given query
	// High scores indicate high similarity; negative scores indicate a mismatch.
	score int

	// words represent the menu item name normalized and split at spaces/punctuation.
	words []string

	// item is the menu item that has been scored.
	item MenuItem
}

// MenuSearch performs approximate text searches for menu items matching a query string.
type MenuSearch struct {
	scoredItems       []scoredItem
	query             string
	queryWords        []string
	emptyQueryShowAll bool
}

// Query returns the current query.
func (s *MenuSearch) Query() string {
	return s.query
}

// SetQuery updates the query for the search.
func (s *MenuSearch) SetQuery(q string) {
	if s.query == q {
		return
	}

	s.query = q
	s.queryWords = s.splitWords(s.normalize(q))
	for i := 0; i < len(s.scoredItems); i++ {
		candidateWords := s.scoredItems[i].words
		s.scoredItems[i].score = s.calculateScore(candidateWords, s.queryWords)
	}
	s.sortItemsByScore()
}

// AddItems adds more menu items to the search set.
func (s *MenuSearch) AddItems(items []MenuItem) {
	for _, item := range items {
		words := s.splitWords(s.normalize(item.Name))
		s.scoredItems = append(s.scoredItems, scoredItem{
			item:  item,
			words: words,
			score: s.calculateScore(words, s.queryWords),
		})
	}
	s.sortItemsByScore()
}

// Results returns the menu items matching the current query.
// Items are sorted descending by similarity to the query,
// with ties broken by lexicographic ordering.
func (s *MenuSearch) Results() []MenuItem {
	results := make([]MenuItem, 0, len(s.scoredItems))
	for _, si := range s.scoredItems {
		if si.score < 0 {
			break
		}
		results = append(results, si.item)
	}
	return results
}

// sortItemsByScore sorts the result items by their scores.
// Ties are broken by lexicographic ordering.
func (s *MenuSearch) sortItemsByScore() {
	sort.SliceStable(s.scoredItems, func(i, j int) bool {
		s1, s2 := s.scoredItems[i].score, s.scoredItems[j].score
		if s1 == s2 {
			n1, n2 := s.scoredItems[i].item.Name, s.scoredItems[j].item.Name
			return n1 < n2
		}
		return s1 > s2
	})
}

// normalize returns a canonical form of the string for case-insensitive comparison.
func (s *MenuSearch) normalize(x string) string {
	return strings.ToLower(norm.NFC.String(x))
}

// calculateScore returns a similarity score for a candidate and a query.
// This uses a simple heuristic based on the count of query "words"
// that match a (possibly non-contiguous) sequence of candidate "words".
// Every match increases the score, and query words without a match
// always produce a negative score.
// This isn't a perfect similarity measure, but it is fast to evaluate
// and works fairly well for commands and file paths.
func (s *MenuSearch) calculateScore(candidateWords []string, queryWords []string) int {
	// Greedily match words from the query with words in the candidate.
	// It's okay to be greedy because we've defined the similarity score in terms
	// of the number of word matches, ignoring the exact location of those matches in the candidate.
	i, j, score := 0, 0, -1
	if s.emptyQueryShowAll {
		// We want the empty query to show every item, so set the default score to zero.
		// This ensures that every item will be assigned a score of zero,
		// and will be displayed in lexicographic order.
		score = 0
	}

	for i < len(candidateWords) && j < len(queryWords) {
		cw, qw := candidateWords[i], queryWords[j]
		if strings.HasPrefix(cw, qw) {
			// Reward query words that match a word in the candidate.
			points := len(qw)
			if i == 0 && j == 0 {
				points *= 2 // bonus for first word match
			}
			score += points
			j++
		}
		i++
	}

	// If there are query words that didn't match the candidate, classify as a mismatch.
	if j < len(queryWords) {
		return -1
	}

	return score
}

// splitWords separates a string into "words" at space and some punctuation boundaries.
// The main use cases are spaces in commands (e.g. "save and quit" -> ["save", "and", "quit"])
// and file paths (e.g. "foo/bar/baz_test.go" -> ["foo", "bar", "baz", "test", "go"])
func (s *MenuSearch) splitWords(text string) []string {
	wordBuffer := make([]string, 0, 5)
	i := 0
	for j, r := range text {
		if r == ' ' || r == '/' || r == '-' || r == '_' || r == '.' {
			if j > i {
				wordBuffer = append(wordBuffer, text[i:j])
			}
			i = j + 1
		}
	}

	if i < len(text) {
		wordBuffer = append(wordBuffer, text[i:len(text)])
	}

	return wordBuffer
}
