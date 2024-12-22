package text

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEmptyRuneStack(t *testing.T) {
	var rs RuneStack
	assert.Equal(t, 0, rs.Len())
	assert.Equal(t, "", rs.String())
}

func TestPopEmptyRuneStack(t *testing.T) {
	var rs RuneStack
	ok, r := rs.Pop()
	assert.False(t, ok)
	assert.Equal(t, '\x00', r)
}

func TestRuneStackPushThenPop(t *testing.T) {
	var rs RuneStack

	rs.Push('a')
	assert.Equal(t, 1, rs.Len())
	assert.Equal(t, "a", rs.String())

	rs.Push('b')
	assert.Equal(t, 2, rs.Len())
	assert.Equal(t, "ab", rs.String())

	rs.Push('c')
	assert.Equal(t, 3, rs.Len())
	assert.Equal(t, "abc", rs.String())

	ok, r := rs.Pop()
	assert.True(t, ok)
	assert.Equal(t, 'c', r)
	assert.Equal(t, 2, rs.Len())
	assert.Equal(t, "ab", rs.String())

	ok, r = rs.Pop()
	assert.True(t, ok)
	assert.Equal(t, 'b', r)
	assert.Equal(t, 1, rs.Len())
	assert.Equal(t, "a", rs.String())

	ok, r = rs.Pop()
	assert.True(t, ok)
	assert.Equal(t, 'a', r)
	assert.Equal(t, 0, rs.Len())
	assert.Equal(t, "", rs.String())
}

func TestRuneStackInterleavePushAndPop(t *testing.T) {
	var rs RuneStack

	rs.Push('a')
	assert.Equal(t, 1, rs.Len())
	assert.Equal(t, "a", rs.String())

	rs.Push('b')
	assert.Equal(t, 2, rs.Len())
	assert.Equal(t, "ab", rs.String())

	ok, r := rs.Pop()
	assert.True(t, ok)
	assert.Equal(t, 'b', r)
	assert.Equal(t, 1, rs.Len())
	assert.Equal(t, "a", rs.String())

	rs.Push('x')
	assert.Equal(t, 2, rs.Len())
	assert.Equal(t, "ax", rs.String())

	rs.Push('y')
	assert.Equal(t, 3, rs.Len())
	assert.Equal(t, "axy", rs.String())

	ok, r = rs.Pop()
	assert.True(t, ok)
	assert.Equal(t, 'y', r)
	assert.Equal(t, 2, rs.Len())
	assert.Equal(t, "ax", rs.String())
}

func TestRuneStackRepeatedlyRetrieveString(t *testing.T) {
	var rs RuneStack
	rs.Push('a')
	rs.Push('b')
	rs.Push('c')
	for i := 0; i < 5; i++ {
		assert.Equal(t, "abc", rs.String())
	}
}
