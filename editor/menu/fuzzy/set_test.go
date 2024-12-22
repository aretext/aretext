package fuzzy

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIntSet(t *testing.T) {
	testCases := []struct {
		name              string
		elements          []int
		expectContains    []int
		expectNotContains []int
	}{
		{
			name:              "empty",
			elements:          nil,
			expectContains:    nil,
			expectNotContains: []int{99},
		},
		{
			name:              "single zero",
			elements:          []int{0},
			expectContains:    []int{0},
			expectNotContains: []int{99},
		},
		{
			name:              "single positive int",
			elements:          []int{56},
			expectContains:    []int{56},
			expectNotContains: []int{99},
		},
		{
			name:              "single max int",
			elements:          []int{math.MaxInt},
			expectContains:    []int{math.MaxInt},
			expectNotContains: []int{99},
		},
		{
			name:              "duplicate integers",
			elements:          []int{1, 2, 3, 1, 2, 3, 2, 1},
			expectContains:    []int{1, 2, 3},
			expectNotContains: []int{99},
		},
		{
			name:              "many elements",
			elements:          []int{1895, 5530, 3239, 1406, 2419, 561, 9151, 8228, 9218, 820, 1328, 6488, 3325, 1803, 5499, 9449, 8011, 592, 9483, 905, 2909, 9598, 7349, 2644, 1630, 9702, 4592, 366, 2468, 3300, 7871, 4843},
			expectContains:    []int{1895, 5530, 3239, 1406, 2419, 561, 9151, 8228, 9218, 820, 1328, 6488, 3325, 1803, 5499, 9449, 8011, 592, 9483, 905, 2909, 9598, 7349, 2644, 1630, 9702, 4592, 366, 2468, 3300, 7871, 4843},
			expectNotContains: []int{99},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			set := newIntSet()
			for _, x := range tc.elements {
				set.add(x)
			}
			for _, x := range tc.expectContains {
				assert.Truef(t, set.contains(x), "missing element %d", x)
			}
			for _, x := range tc.expectNotContains {
				assert.Falsef(t, set.contains(x), "extra element %d", x)
			}
			var elements []int
			set.forEach(func(x int) { elements = append(elements, x) })
			assert.ElementsMatch(t, elements, tc.expectContains)
		})
	}
}
