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
	result := make([]int, 0, len(m))
	for key := range m {
		result = append(result, key)
	}
	sort.IntSlice(result).Sort()
	return result
}

func intSliceKey(s []int) string {
	if len(s) == 0 {
		return ""
	}

	if len(s) == 1 {
		return strconv.Itoa(s[0])
	}

	var sb strings.Builder
	for _, x := range s {
		sb.WriteString(strconv.Itoa(x))
		sb.WriteRune('|')
	}
	return sb.String()
}

func intSliceEqual(x, y []int) bool {
	if len(x) != len(y) {
		return false
	}

	for i := 0; i < len(x); i++ {
		if x[i] != y[i] {
			return false
		}
	}

	return true
}

func forEachPartitionInKeyOrder(partitions map[string][]int, f func(states []int)) {
	partitionKeys := make([]string, 0, len(partitions))
	for key, _ := range partitions {
		partitionKeys = append(partitionKeys, key)
	}
	sort.Strings(partitionKeys)
	for _, key := range partitionKeys {
		f(partitions[key])
	}
}

func minUint64(x uint64, y uint64) uint64 {
	if x < y {
		return x
	} else {
		return y
	}
}

func maxUint64(x uint64, y uint64) uint64 {
	if x > y {
		return x
	} else {
		return y
	}
}

func subtractNoUnderflowUint64(x uint64, y uint64) uint64 {
	if x >= y {
		return x - y
	} else {
		return 0
	}
}
