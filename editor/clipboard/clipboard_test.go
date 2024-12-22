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

func TestPageIdForLetter(t *testing.T) {
	testCases := []struct {
		name         string
		letter       rune
		expectedPage PageId
	}{
		{
			name:         "page a",
			letter:       'a',
			expectedPage: PageLetterA,
		},
		{
			name:         "page b",
			letter:       'b',
			expectedPage: PageLetterB,
		},
		{
			name:         "page y",
			letter:       'y',
			expectedPage: PageLetterY,
		},
		{
			name:         "page z",
			letter:       'z',
			expectedPage: PageLetterZ,
		},
		{
			name:         "non-alpha",
			letter:       '!',
			expectedPage: PageNull,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			page := PageIdForLetter(tc.letter)
			assert.Equal(t, tc.expectedPage, page)
		})
	}
}
