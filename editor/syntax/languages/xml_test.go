package languages

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/aretext/aretext/editor/syntax/parser"
)

func TestXmlParseFunc(t *testing.T) {
	testCases := []struct {
		name     string
		text     string
		expected []TokenWithText
	}{
		{
			name: "prologue",
			text: `<?xml version="1.0" encoding="UTF-8"?>`,
			expected: []TokenWithText{
				{Role: xmlTokenRolePrologue, Text: `<?xml version="1.0" encoding="UTF-8"?>`},
			},
		},
		{
			name: "cdata",
			text: `
<script>
   <![CDATA[
      <tag>Hello!</tag>
   ]]>
</script >`,
			expected: []TokenWithText{
				{Role: xmlTokenRoleTag, Text: `<script`}, {Role: xmlTokenRoleTag, Text: `>`},
				{Role: xmlTokenRoleCData, Text: "<![CDATA[\n      <tag>Hello!</tag>\n   ]]>"},
				{Role: xmlTokenRoleTag, Text: `</script`}, {Role: xmlTokenRoleTag, Text: `>`},
			},
		},
		{
			name: "single-line comment",
			text: `<!-- this is a comment -->`,
			expected: []TokenWithText{
				{Role: parser.TokenRoleComment, Text: `<!-- this is a comment -->`},
			},
		},
		{
			name: "multi-line comment",
			text: `<!--
this
is a
comment
-->`,
			expected: []TokenWithText{
				{Role: parser.TokenRoleComment, Text: "<!--\nthis\nis a\ncomment\n-->"},
			},
		},
		{
			name: "tag with attribute double-quoted",
			text: `<meta charset="utf-8">`,
			expected: []TokenWithText{
				{Role: xmlTokenRoleTag, Text: `<meta`},
				{Role: xmlTokenRoleAttrKey, Text: `charset=`},
				{Role: parser.TokenRoleString, Text: `"utf-8"`},
				{Role: xmlTokenRoleTag, Text: `>`},
			},
		},
		{
			name: "tag with attribute single-quoted",
			text: `<meta charset='utf-8'>`,
			expected: []TokenWithText{
				{Role: xmlTokenRoleTag, Text: `<meta`},
				{Role: xmlTokenRoleAttrKey, Text: `charset=`},
				{Role: parser.TokenRoleString, Text: `'utf-8'`},
				{Role: xmlTokenRoleTag, Text: `>`},
			},
		},
		{
			name: "tag with attribute containing equal sign",
			text: `<meta name="viewport" content="width=device-width, initial-scale=1.0">`,
			expected: []TokenWithText{
				{Role: xmlTokenRoleTag, Text: `<meta`},
				{Role: xmlTokenRoleAttrKey, Text: `name=`},
				{Role: parser.TokenRoleString, Text: `"viewport"`},
				{Role: xmlTokenRoleAttrKey, Text: `content=`},
				{Role: parser.TokenRoleString, Text: `"width=device-width, initial-scale=1.0"`},
				{Role: xmlTokenRoleTag, Text: `>`},
			},
		},
		{
			name: "tag with unquoted attribute",
			text: `<link rel=icon type=image/svg+xml href=/favicon/favicon.svg>`,
			expected: []TokenWithText{
				{Role: xmlTokenRoleTag, Text: `<link`},
				{Role: xmlTokenRoleAttrKey, Text: `rel=`},
				{Role: xmlTokenRoleAttrKey, Text: `type=`},
				{Role: xmlTokenRoleAttrKey, Text: `href=`},
				{Role: xmlTokenRoleTag, Text: `>`},
			},
		},
		{
			name: "tag with attribute with dashes",
			text: `<div data-foo='xyz'>`,
			expected: []TokenWithText{
				{Role: xmlTokenRoleTag, Text: `<div`},
				{Role: xmlTokenRoleAttrKey, Text: `data-foo=`},
				{Role: parser.TokenRoleString, Text: `'xyz'`},
				{Role: xmlTokenRoleTag, Text: `>`},
			},
		},
		{
			name: "nested tags",
			text: `
<html>
	<head>
		<title>Title</title>
	</head>
	<body>
		<h1>Title</h1>
		<p>Text</p>
	</body>
</html>`,
			expected: []TokenWithText{
				{Role: xmlTokenRoleTag, Text: `<html`}, {Role: xmlTokenRoleTag, Text: `>`},
				{Role: xmlTokenRoleTag, Text: `<head`}, {Role: xmlTokenRoleTag, Text: `>`},
				{Role: xmlTokenRoleTag, Text: `<title`}, {Role: xmlTokenRoleTag, Text: `>`},
				{Role: xmlTokenRoleTag, Text: `</title`}, {Role: xmlTokenRoleTag, Text: `>`},
				{Role: xmlTokenRoleTag, Text: `</head`}, {Role: xmlTokenRoleTag, Text: `>`},
				{Role: xmlTokenRoleTag, Text: `<body`}, {Role: xmlTokenRoleTag, Text: `>`},
				{Role: xmlTokenRoleTag, Text: `<h1`}, {Role: xmlTokenRoleTag, Text: `>`},
				{Role: xmlTokenRoleTag, Text: `</h1`}, {Role: xmlTokenRoleTag, Text: `>`},
				{Role: xmlTokenRoleTag, Text: `<p`}, {Role: xmlTokenRoleTag, Text: `>`},
				{Role: xmlTokenRoleTag, Text: `</p`}, {Role: xmlTokenRoleTag, Text: `>`},
				{Role: xmlTokenRoleTag, Text: `</body`}, {Role: xmlTokenRoleTag, Text: `>`},
				{Role: xmlTokenRoleTag, Text: `</html`}, {Role: xmlTokenRoleTag, Text: `>`},
			},
		},
		{
			name: "self-closing tag",
			text: `<br/>`,
			expected: []TokenWithText{
				{Role: xmlTokenRoleTag, Text: `<br`}, {Role: xmlTokenRoleTag, Text: `/>`},
			},
		},
		{
			name: "character entity name",
			text: `<p>this is a quote: &quot;</p>`,
			expected: []TokenWithText{
				{Role: xmlTokenRoleTag, Text: `<p`}, {Role: xmlTokenRoleTag, Text: `>`},
				{Role: xmlTokenRoleCharacterEntity, Text: `&quot;`},
				{Role: xmlTokenRoleTag, Text: `</p`}, {Role: xmlTokenRoleTag, Text: `>`},
			},
		},
		{
			name: "character entity unicode codepoint",
			text: `<p>this is unicode: &#1234;</p>`,
			expected: []TokenWithText{
				{Role: xmlTokenRoleTag, Text: `<p`}, {Role: xmlTokenRoleTag, Text: `>`},
				{Role: xmlTokenRoleCharacterEntity, Text: `&#1234;`},
				{Role: xmlTokenRoleTag, Text: `</p`}, {Role: xmlTokenRoleTag, Text: `>`},
			},
		},
		{
			name: "content looks like tag attribute",
			text: `<p>x="123"</p>`,
			expected: []TokenWithText{
				{Role: xmlTokenRoleTag, Text: `<p`}, {Role: xmlTokenRoleTag, Text: `>`},
				{Role: xmlTokenRoleTag, Text: `</p`}, {Role: xmlTokenRoleTag, Text: `>`},
			},
		},
		{
			name:     "invalid tag with space between open delim and name",
			text:     `< p>`,
			expected: []TokenWithText{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tokens := ParseTokensWithText(XmlParseFunc(), tc.text)
			assert.Equal(t, tc.expected, tokens)
		})
	}
}

func BenchmarkXmlParser(b *testing.B) {
	BenchmarkParser(b, XmlParseFunc(), "testdata/xml/example.xml")
}

func BenchmarkXmlParserLongAttrRegression(b *testing.B) {
	// Verify performance issue found by fuzz testing.
	// Attributes that didn't match the pattern x="y" were falling back
	// to error recovery, which was slow for long attribute names.
	BenchmarkParser(b, XmlParseFunc(), "testdata/xml/long-attr-regression.xml")
}

func FuzzXmlParseFunc(f *testing.F) {
	seeds := LoadFuzzTestSeeds(f, "./testdata/xml/*")
	FuzzParser(f, XmlParseFunc(), seeds)
}
