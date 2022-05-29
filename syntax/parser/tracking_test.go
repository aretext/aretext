package parser

import (
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/aretext/aretext/text"
)

func TestTrackingRuneIter(t *testing.T) {
	tree, err := text.NewTreeFromString("abcd")
	require.NoError(t, err)
	reader := tree.ReaderAtPosition(0)
	trackingIter := NewTrackingRuneIter(reader)
	assert.Equal(t, uint64(0), trackingIter.MaxRead())

	r, err := trackingIter.NextRune()
	require.NoError(t, err)
	assert.Equal(t, 'a', r)
	assert.Equal(t, uint64(1), trackingIter.MaxRead())

	cloneIter := trackingIter
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

	// EOF should include one position past the end of text.
	_, err = trackingIter.NextRune()
	assert.ErrorIs(t, err, io.EOF)
	assert.Equal(t, uint64(5), trackingIter.MaxRead())

	r, err = cloneIter.NextRune()
	require.NoError(t, err)
	assert.Equal(t, 'c', r)
	assert.Equal(t, uint64(5), trackingIter.MaxRead())

	// Read EOF again, shouldn't affect max read.
	_, err = trackingIter.NextRune()
	assert.ErrorIs(t, err, io.EOF)
	assert.Equal(t, uint64(5), trackingIter.MaxRead())
}

func TestTrackingRuneIterLimit(t *testing.T) {
	tree, err := text.NewTreeFromString("abcd")
	require.NoError(t, err)
	reader := tree.ReaderAtPosition(0)
	trackingIter := NewTrackingRuneIter(reader)

	// Set a limit of two runes.
	trackingIter.Limit(2)

	// Read the first rune.
	r, err := trackingIter.NextRune()
	require.NoError(t, err)
	assert.Equal(t, 'a', r)

	// Read the second rune.
	r, err = trackingIter.NextRune()
	require.NoError(t, err)
	assert.Equal(t, 'b', r)

	// Hit the limit, so expect EOF.
	_, err = trackingIter.NextRune()
	require.ErrorIs(t, err, io.EOF)
}
