package boundaries

import (
	"testing"
)

func BenchmarkGraphemeClusterBreakFinder(b *testing.B) {
	finder := NewGraphemeClusterBreakFinder()
	for i := 0; i < b.N; i++ {
		finder.ProcessCharacter('a')
	}
}
