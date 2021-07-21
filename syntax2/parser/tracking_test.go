package parser

import (
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/aretext/aretext/text"
)

func TestTrackingRuneIter(t *testing.T) {
	runeIter := text.NewRuneIterForSlice([]rune{'a', 'b', 'c', 'd'})
	trackingIter := NewTrackingRuneIter(runeIter)
	assert.Equal(t, uint64(0), trackingIter.MaxRead())

	r, err := trackingIter.NextRune()
	require.NoError(t, err)
	assert.Equal(t, 'a', r)
	assert.Equal(t, uint64(1), trackingIter.MaxRead())

	cloneIter := trackingIter.Clone()
	r, err = cloneIter.NextRune()
	require.NoError(t, err)
	assert.Equal(t, 'b', r)
	assert.Equal(t, uint64(2), trackingIter.MaxRead())

	r, err = trackingIter.NextRune()
	require.NoError(t, err)
	assert.Equal(t, 'b', r)
	assert.Equal(t, uint64(2), trackingIter.MaxRead())

	r, err = trackingIter.NextRune()
	require.NoError(t, err)
	assert.Equal(t, 'c', r)
	assert.Equal(t, uint64(3), trackingIter.MaxRead())

	r, err = trackingIter.NextRune()
	require.NoError(t, err)
	assert.Equal(t, 'd', r)
	assert.Equal(t, uint64(4), trackingIter.MaxRead())

	_, err = trackingIter.NextRune()
	assert.ErrorIs(t, err, io.EOF)
	assert.Equal(t, uint64(4), trackingIter.MaxRead())

	r, err = cloneIter.NextRune()
	require.NoError(t, err)
	assert.Equal(t, 'c', r)
	assert.Equal(t, uint64(4), trackingIter.MaxRead())
}
