package languages

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/aretext/aretext/syntax/parser"
)

func TestProtobufParseFunc(t *testing.T) {
	testCases := []struct {
		name     string
		text     string
		expected []TokenWithText
	}{
		{
			name: "keywords recognized at top-level vs nested",
			text: `
syntax = "proto3";
message Foo {
	int64 syntax = 1;
}
package foo;
`,
			expected: []TokenWithText{
				{Role: parser.TokenRoleKeyword, Text: "syntax"},
				{Role: parser.TokenRoleOperator, Text: "="},
				{Role: parser.TokenRoleString, Text: "\"proto3\""},
				{Role: parser.TokenRoleKeyword, Text: "message"},
				{Role: parser.TokenRoleIdentifier, Text: "Foo"},
				{Role: parser.TokenRoleKeyword, Text: "int64"},
				{Role: parser.TokenRoleIdentifier, Text: "syntax"},
				{Role: parser.TokenRoleOperator, Text: "="},
				{Role: parser.TokenRoleNumber, Text: "1"},
				{Role: parser.TokenRoleKeyword, Text: "package"},
				{Role: parser.TokenRoleIdentifier, Text: "foo"},
			},
		},
		{
			name: "identifier parsing",
			text: `
message Foo {
	int64 foo.message.bar = 1;
}
`,
			expected: []TokenWithText{
				{Role: parser.TokenRoleKeyword, Text: "message"},
				{Role: parser.TokenRoleIdentifier, Text: "Foo"},
				{Role: parser.TokenRoleKeyword, Text: "int64"},
				{Role: parser.TokenRoleIdentifier, Text: "foo.message.bar"},
				{Role: parser.TokenRoleOperator, Text: "="},
				{Role: parser.TokenRoleNumber, Text: "1"},
			},
		},
		{
			name: "grpc service",
			text: `
service SearchService {
  rpc Search (SearchRequest) returns (SearchResponse);
}
`,
			expected: []TokenWithText{
				{Role: parser.TokenRoleKeyword, Text: "service"},
				{Role: parser.TokenRoleIdentifier, Text: "SearchService"},
				{Role: parser.TokenRoleKeyword, Text: "rpc"},
				{Role: parser.TokenRoleIdentifier, Text: "Search"},
				{Role: parser.TokenRoleIdentifier, Text: "SearchRequest"},
				{Role: parser.TokenRoleKeyword, Text: "returns"},
				{Role: parser.TokenRoleIdentifier, Text: "SearchResponse"},
			},
		},
		{
			name: "full example",
			text: `syntax = "proto3";
import public "other.proto";
option java_package = "com.example.foo";
enum EnumAllowingAlias {
  option allow_alias = true;
  UNKNOWN = 0;
  STARTED = 1;
  RUNNING = 2 [(custom_option) = "hello world"];
}
message Outer {
  option (my_option).a = true;
  message Inner {   // Level 2
    int64 ival = 1;
  }
  repeated Inner inner_message = 2;
  EnumAllowingAlias enum_field =3;
  map<int32, string> my_map = 4;
}`,
			expected: []TokenWithText{
				{Role: parser.TokenRoleKeyword, Text: "syntax"},
				{Role: parser.TokenRoleOperator, Text: "="},
				{Role: parser.TokenRoleString, Text: "\"proto3\""},
				{Role: parser.TokenRoleKeyword, Text: "import"},
				{Role: parser.TokenRoleKeyword, Text: "public"},
				{Role: parser.TokenRoleString, Text: "\"other.proto\""},
				{Role: parser.TokenRoleKeyword, Text: "option"},
				{Role: parser.TokenRoleIdentifier, Text: "java_package"},
				{Role: parser.TokenRoleOperator, Text: "="},
				{Role: parser.TokenRoleString, Text: "\"com.example.foo\""},
				{Role: parser.TokenRoleKeyword, Text: "enum"},
				{Role: parser.TokenRoleIdentifier, Text: "EnumAllowingAlias"},
				{Role: parser.TokenRoleKeyword, Text: "option"},
				{Role: parser.TokenRoleIdentifier, Text: "allow_alias"},
				{Role: parser.TokenRoleOperator, Text: "="},
				{Role: parser.TokenRoleKeyword, Text: "true"},
				{Role: parser.TokenRoleIdentifier, Text: "UNKNOWN"},
				{Role: parser.TokenRoleOperator, Text: "="},
				{Role: parser.TokenRoleNumber, Text: "0"},
				{Role: parser.TokenRoleIdentifier, Text: "STARTED"},
				{Role: parser.TokenRoleOperator, Text: "="},
				{Role: parser.TokenRoleNumber, Text: "1"},
				{Role: parser.TokenRoleIdentifier, Text: "RUNNING"},
				{Role: parser.TokenRoleOperator, Text: "="},
				{Role: parser.TokenRoleNumber, Text: "2"},
				{Role: parser.TokenRoleIdentifier, Text: "custom_option"},
				{Role: parser.TokenRoleOperator, Text: "="},
				{Role: parser.TokenRoleString, Text: "\"hello world\""},
				{Role: parser.TokenRoleKeyword, Text: "message"},
				{Role: parser.TokenRoleIdentifier, Text: "Outer"},
				{Role: parser.TokenRoleKeyword, Text: "option"},
				{Role: parser.TokenRoleIdentifier, Text: "my_option"},
				{Role: parser.TokenRoleIdentifier, Text: "a"},
				{Role: parser.TokenRoleOperator, Text: "="},
				{Role: parser.TokenRoleKeyword, Text: "true"},
				{Role: parser.TokenRoleKeyword, Text: "message"},
				{Role: parser.TokenRoleIdentifier, Text: "Inner"},
				{Role: parser.TokenRoleComment, Text: "// Level 2\n"},
				{Role: parser.TokenRoleKeyword, Text: "int64"},
				{Role: parser.TokenRoleIdentifier, Text: "ival"},
				{Role: parser.TokenRoleOperator, Text: "="},
				{Role: parser.TokenRoleNumber, Text: "1"},
				{Role: parser.TokenRoleKeyword, Text: "repeated"},
				{Role: parser.TokenRoleIdentifier, Text: "Inner"},
				{Role: parser.TokenRoleIdentifier, Text: "inner_message"},
				{Role: parser.TokenRoleOperator, Text: "="},
				{Role: parser.TokenRoleNumber, Text: "2"},
				{Role: parser.TokenRoleIdentifier, Text: "EnumAllowingAlias"},
				{Role: parser.TokenRoleIdentifier, Text: "enum_field"},
				{Role: parser.TokenRoleOperator, Text: "="},
				{Role: parser.TokenRoleNumber, Text: "3"},
				{Role: parser.TokenRoleKeyword, Text: "map"},
				{Role: parser.TokenRoleKeyword, Text: "int32"},
				{Role: parser.TokenRoleKeyword, Text: "string"},
				{Role: parser.TokenRoleIdentifier, Text: "my_map"},
				{Role: parser.TokenRoleOperator, Text: "="},
				{Role: parser.TokenRoleNumber, Text: "4"},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tokens := ParseTokensWithText(ProtobufParseFunc(), tc.text)
			assert.Equal(t, tc.expected, tokens)
		})
	}
}
