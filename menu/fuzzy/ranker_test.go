package fuzzy

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRanker(t *testing.T) {
	records := []string{
		"barfoo",  // 0
		"xoo",     // 1
		"foo",     // 2
		"foobar",  // 3
		".foobar", // 4
		"foo.bar", // 5
		"xoobar",  // 6
	}
	ranker := newRanker("foobar", len(records))
	for recordId, record := range records {
		ranker.addRecord(recordId, record)
	}
	result := ranker.rankedRecordIds()
	expected := []int{
		3, // "foobar"
		4, // ".foobar"
		5, // "foo.bar"
		6, // "xoobar"
		0, // "barfoo"
		2, // "foo"
		1, // "xoo"
	}
	assert.Equal(t, expected, result)
}
