package parser

import (
	"sort"
	"strconv"
	"strings"
)

// insertUniqueSorted inserts an integer into a sorted array with unique elements.
// The returned slice is also sorted with unique elements.
// Time complexity is O(n) in the length of the slice, so use this only for short slices.
func insertUniqueSorted(s []int, v int) []int {
	if len(s) == 0 {
		return append(s, v)
	}

	insertIdx := 0
	for insertIdx < len(s) && s[insertIdx] < v {
		insertIdx++
	}

	if insertIdx < len(s) && s[insertIdx] == v {
		return s
	}

	s = append(s, 0)
	for i := len(s) - 1; i > insertIdx; i-- {
		s[i] = s[i-1]
	}
	s[insertIdx] = v
	return s
}

func sortedKeys(m map[int]struct{}) []int {
	result := make([]int, len(m))
	for key := range m {
		result = append(result, key)
	}
	sort.IntSlice(result).Sort()
	return result
}

func intSliceKey(s []int) string {
	var sb strings.Builder
	for _, x := range s {
		sb.WriteString(strconv.Itoa(x))
		sb.WriteString("|")
	}
	return sb.String()
}
