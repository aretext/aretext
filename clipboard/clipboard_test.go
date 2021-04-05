package clipboard

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClipboardPageNull(t *testing.T) {
	c := New()
	assert.Equal(t, PageContent{}, c.Get(PageNull))
	c.Set(PageNull, PageContent{Text: "abcd"})
	assert.Equal(t, PageContent{}, c.Get(PageNull))
}

func TestClipboardPageDefault(t *testing.T) {
	c := New()
	assert.Equal(t, PageContent{}, c.Get(PageDefault))
	c.Set(PageDefault, PageContent{Text: "abcd"})
	assert.Equal(t, PageContent{Text: "abcd"}, c.Get(PageDefault))
}
