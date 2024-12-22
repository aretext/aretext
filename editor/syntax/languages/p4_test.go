package languages

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/aretext/aretext/editor/syntax/parser"
)

func TestP4ParseFunc(t *testing.T) {
	testCases := []struct {
		name     string
		text     string
		expected []TokenWithText
	}{
		{
			name: "line comment",
			text: "// comment",
			expected: []TokenWithText{
				{Text: "// comment", Role: parser.TokenRoleComment},
			},
		},
		{
			name: "block comment",
			text: `
/*
comment
block
*/`,
			expected: []TokenWithText{
				{Text: "/*\ncomment\nblock\n*/", Role: parser.TokenRoleComment},
			},
		},
		{
			name: "numbers",
			text: `
0
7
10
123
123_456
0xDEADBEEF
32w255
32w0d255
32w0xFF
32s0xFF
8w0b10101010
8w0b_1010_1010
8w170
8s0b1010_1010
16w0377
16w0o377
`,
			expected: []TokenWithText{
				{Text: "0", Role: parser.TokenRoleNumber},
				{Text: "7", Role: parser.TokenRoleNumber},
				{Text: "10", Role: parser.TokenRoleNumber},
				{Text: "123", Role: parser.TokenRoleNumber},
				{Text: "123_456", Role: parser.TokenRoleNumber},
				{Text: "0xDEADBEEF", Role: parser.TokenRoleNumber},
				{Text: "32w255", Role: parser.TokenRoleNumber},
				{Text: "32w0d255", Role: parser.TokenRoleNumber},
				{Text: "32w0xFF", Role: parser.TokenRoleNumber},
				{Text: "32s0xFF", Role: parser.TokenRoleNumber},
				{Text: "8w0b10101010", Role: parser.TokenRoleNumber},
				{Text: "8w0b_1010_1010", Role: parser.TokenRoleNumber},
				{Text: "8w170", Role: parser.TokenRoleNumber},
				{Text: "8s0b1010_1010", Role: parser.TokenRoleNumber},
				{Text: "16w0377", Role: parser.TokenRoleNumber},
				{Text: "16w0o377", Role: parser.TokenRoleNumber},
			},
		},
		{
			name:     "decimal cannot start with underscore",
			text:     `_123_456`,
			expected: []TokenWithText{},
		},
		{
			name: "preprocessor directive",
			text: `#include <core.p4>`,
			expected: []TokenWithText{
				{Text: "#include <core.p4>", Role: p4TokenRolePreprocessorDirective},
			},
		},
		{
			name: "annotation with string",
			text: `@name("foo") table t1 { }`,
			expected: []TokenWithText{
				{Text: `@name`, Role: p4TokenRoleAnnotation},
				{Text: `"foo"`, Role: parser.TokenRoleString},
				{Text: `table`, Role: parser.TokenRoleKeyword},
			},
		},
		{
			name: "if statement",
			text: `if (hdr.ipv4.ecn == 1 || hdr.ipv4.ecn == 2)`,
			expected: []TokenWithText{
				{Text: `if`, Role: parser.TokenRoleKeyword},
				{Text: `==`, Role: parser.TokenRoleOperator},
				{Text: `1`, Role: parser.TokenRoleNumber},
				{Text: `||`, Role: parser.TokenRoleOperator},
				{Text: `==`, Role: parser.TokenRoleOperator},
				{Text: `2`, Role: parser.TokenRoleNumber},
			},
		},
		{
			name: "state with transition select",
			text: `
state parse_ethernet {
    packet.extract(hdr.ethernet);
    transition select(hdr.ethernet.etherType) {
        TYPE_IPV4: parse_ipv4;
        default: accept;
    }
}`,
			expected: []TokenWithText{
				{Text: `state`, Role: parser.TokenRoleKeyword},
				{Text: `transition`, Role: parser.TokenRoleKeyword},
				{Text: `select`, Role: parser.TokenRoleKeyword},
				{Text: `:`, Role: parser.TokenRoleOperator},
				{Text: `default`, Role: parser.TokenRoleKeyword},
				{Text: `:`, Role: parser.TokenRoleOperator},
			},
		},
		{
			name: "typedef",
			text: `typedef bit<32> ip4Addr_t;`,
			expected: []TokenWithText{
				{Text: `typedef`, Role: parser.TokenRoleKeyword},
				{Text: `bit`, Role: parser.TokenRoleKeyword},
				{Text: `<`, Role: parser.TokenRoleOperator},
				{Text: `32`, Role: parser.TokenRoleNumber},
				{Text: `>`, Role: parser.TokenRoleOperator},
			},
		},
		{
			name: "const",
			text: `const bit<16> TYPE_IPV4 = 0x800;`,
			expected: []TokenWithText{
				{Text: `const`, Role: parser.TokenRoleKeyword},
				{Text: `bit`, Role: parser.TokenRoleKeyword},
				{Text: `<`, Role: parser.TokenRoleOperator},
				{Text: `16`, Role: parser.TokenRoleNumber},
				{Text: `>`, Role: parser.TokenRoleOperator},
				{Text: `=`, Role: parser.TokenRoleOperator},
				{Text: `0x800`, Role: parser.TokenRoleNumber},
			},
		},
		{
			name: "header",
			text: `
header ethernet_t {
    macAddr_t dstAddr;
    macAddr_t srcAddr;
    bit<16>   etherType;
}
`,
			expected: []TokenWithText{
				{Text: `header`, Role: parser.TokenRoleKeyword},
				{Text: `bit`, Role: parser.TokenRoleKeyword},
				{Text: `<`, Role: parser.TokenRoleOperator},
				{Text: `16`, Role: parser.TokenRoleNumber},
				{Text: `>`, Role: parser.TokenRoleOperator},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tokens := ParseTokensWithText(P4ParseFunc(), tc.text)
			assert.Equal(t, tc.expected, tokens)
		})
	}
}
