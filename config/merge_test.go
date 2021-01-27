package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMergeRecursive(t *testing.T) {
	testCases := []struct {
		name     string
		base     interface{}
		overlay  interface{}
		expected interface{}
	}{
		{
			name:     "nil / nil",
			base:     nil,
			overlay:  nil,
			expected: nil,
		},
		{
			name:     "nil / int",
			base:     nil,
			overlay:  12,
			expected: 12,
		},
		{
			name:     "int / nil",
			base:     12,
			overlay:  nil,
			expected: 12,
		},
		{
			name:     "int / int",
			base:     12,
			overlay:  34,
			expected: 34,
		},
		{
			name:     "string / string",
			base:     "abcd",
			overlay:  "xyz",
			expected: "xyz",
		},
		{
			name:     "string / int",
			base:     "abcd",
			overlay:  123,
			expected: 123,
		},
		{
			name:     "true / false",
			base:     true,
			overlay:  false,
			expected: false,
		},
		{
			name:     "false / true",
			base:     false,
			overlay:  true,
			expected: true,
		},
		{
			name: "empty map / non-empty map",
			base: map[string]interface{}{},
			overlay: map[string]interface{}{
				"a": 1,
				"b": 2,
			},
			expected: map[string]interface{}{
				"a": 1,
				"b": 2,
			},
		},
		{
			name: "non-empty map / empty map",
			base: map[string]interface{}{
				"a": 1,
				"b": 2,
			},
			overlay: map[string]interface{}{},
			expected: map[string]interface{}{
				"a": 1,
				"b": 2,
			},
		},
		{
			name: "map with overlapping keys, same types",
			base: map[string]interface{}{
				"a": 1,
				"b": 2,
			},
			overlay: map[string]interface{}{
				"b": 3,
				"c": 4,
			},
			expected: map[string]interface{}{
				"a": 1,
				"b": 3,
				"c": 4,
			},
		},
		{
			name: "map with overlapping keys, different types",
			base: map[string]interface{}{
				"a": 1,
				"b": 2,
			},
			overlay: map[string]interface{}{
				"b": "xyz",
				"c": 4,
			},
			expected: map[string]interface{}{
				"a": 1,
				"b": "xyz",
				"c": 4,
			},
		},
		{
			name: "maps with different key types",
			base: map[string]int{
				"a": 1,
				"b": 2,
			},
			overlay: map[int]string{
				3: "c",
				4: "d",
			},
			expected: map[int]string{
				3: "c",
				4: "d",
			},
		},
		{
			name: "int / map",
			base: 12,
			overlay: map[string]int{
				"test": 1,
			},
			expected: map[string]int{
				"test": 1,
			},
		},
		{
			name: "map / int",
			base: map[string]int{
				"test": 1,
			},
			overlay:  12,
			expected: 12,
		},
		{
			name:     "empty slice / slice",
			base:     []string{},
			overlay:  []string{"a", "b"},
			expected: []string{"a", "b"},
		},
		{
			name:     "slice / empty slice",
			base:     []string{"a", "b"},
			overlay:  []string{},
			expected: []string{"a", "b"},
		},
		{
			name:     "slice / slice, same types",
			base:     []string{"a", "b"},
			overlay:  []string{"c", "d"},
			expected: []string{"a", "b", "c", "d"},
		},
		{
			name:     "slice / slice, different types",
			base:     []string{"a", "b"},
			overlay:  []int{3, 4},
			expected: []int{3, 4},
		},
		{
			name:     "slices with heterogeneous types",
			base:     []interface{}{1, "a"},
			overlay:  []interface{}{nil, 2, "b"},
			expected: []interface{}{1, "a", nil, 2, "b"},
		},
		{
			name: "nested slices and structs",
			base: map[string]interface{}{
				"subMap": map[string]interface{}{
					"a": 1,
					"b": 2,
				},
				"subSlice": []interface{}{"x"},
			},
			overlay: map[string]interface{}{
				"subMap": map[string]interface{}{
					"b": 3,
					"c": 4,
				},
				"subSlice": []interface{}{"y"},
			},
			expected: map[string]interface{}{
				"subMap": map[string]interface{}{
					"a": 1,
					"b": 3,
					"c": 4,
				},
				"subSlice": []interface{}{"x", "y"},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			merged := MergeRecursive(tc.base, tc.overlay)
			assert.Equal(t, tc.expected, merged)
		})
	}
}
