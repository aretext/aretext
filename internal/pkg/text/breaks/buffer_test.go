package breaks

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCircularRuneBufferWriteAndConsume(t *testing.T) {
	buf := newCircularRuneBuffer(3)
	buf.write('a')
	buf.write('b')
	buf.write('c')
	assert.Equal(t, 'a', buf.consume())
	assert.Equal(t, 'b', buf.consume())
	assert.Equal(t, 'c', buf.consume())
	buf.write('d')
	assert.Equal(t, 'd', buf.consume())
	buf.write('e')
	buf.write('f')
	assert.Equal(t, 'e', buf.consume())
	assert.Equal(t, 'f', buf.consume())
}

func TestCircularRuneBufferPeek(t *testing.T) {
	buf := newCircularRuneBuffer(3)
	buf.write('a')
	buf.write('b')
	buf.write('c')
	assert.Equal(t, 'a', buf.consume())

	r, ok := buf.peekAtIdx(0)
	assert.True(t, ok)
	assert.Equal(t, 'b', r)

	r, ok = buf.peekAtIdx(1)
	assert.True(t, ok)
	assert.Equal(t, 'c', r)

	_, ok = buf.peekAtIdx(-1)
	assert.False(t, ok)

	_, ok = buf.peekAtIdx(2)
	assert.False(t, ok)
}
